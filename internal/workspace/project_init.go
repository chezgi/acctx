package workspace

import (
	"acctx/internal/buildinfo"
	bundle "acctx/internal/content"
	"acctx/internal/fsx"
	"acctx/internal/manifest"
	"acctx/internal/plan"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BuildProjectInitPlan applies the base workspace initializer while treating
// accounting/company files as company-owned seed files. They are created when
// missing, but existing values are preserved and are not tracked as immutable
// vendor-managed content.
func BuildProjectInitPlan(root string, options InitOptions, contentBundle bundle.Bundle, build buildinfo.Info, now time.Time) (plan.Plan, manifest.Model, error) {
	operationPlan, model, err := BuildInitPlan(root, options, contentBundle, build, now)
	if err != nil {
		return plan.Plan{}, manifest.Model{}, err
	}

	managedFiles := model.Managed.Files[:0]
	for _, file := range model.Managed.Files {
		if strings.HasPrefix(file.Path, "accounting/company/") {
			continue
		}
		managedFiles = append(managedFiles, file)
	}
	model.Managed.Files = managedFiles

	for index := range operationPlan.Operations {
		operation := &operationPlan.Operations[index]
		if !strings.HasPrefix(operation.Path, "accounting/company/") {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(operation.Path))
		if info, statErr := os.Stat(path); statErr == nil && info.Mode().IsRegular() {
			operation.Kind = plan.Skip
			operation.Payload = nil
			operation.Message = "company-owned file preserved"
			if digest, digestErr := fsx.FileDigest(path); digestErr == nil {
				operation.AfterDigest = digest
			}
		}
	}

	previousFiles := map[string]manifest.File{}
	if previous, loadErr := manifest.Load(root); loadErr == nil {
		for _, file := range previous.Managed.Files {
			previousFiles[file.Path] = file
		}
	}
	ruleFiles, err := contentBundle.ReadTree("rules")
	if err != nil {
		return plan.Plan{}, manifest.Model{}, err
	}
	for relative, data := range ruleFiles {
		path := filepath.ToSlash(filepath.Join("rules", "vendor", relative))
		operation := managedFileOp(root, path, data, previousFiles)
		operationPlan.Operations = append(operationPlan.Operations, operation)
		if operation.Kind != plan.Conflict {
			model.Managed.Files = append(model.Managed.Files, manifest.File{
				Path: path, Digest: fsx.BytesDigest(data), SourceID: "rule:" + relative,
			})
		}
	}
	companyRulePath := "rules/company/.gitkeep"
	companyRuleFile := filepath.Join(root, filepath.FromSlash(companyRulePath))
	if _, statErr := os.Stat(companyRuleFile); os.IsNotExist(statErr) {
		operationPlan.Operations = append(operationPlan.Operations, plan.Operation{Kind: plan.Create, Path: companyRulePath, AfterDigest: fsx.BytesDigest(nil), Payload: nil})
	} else {
		operationPlan.Operations = append(operationPlan.Operations, plan.Operation{Kind: plan.Skip, Path: companyRulePath, Message: "company-owned rule directory preserved"})
	}

	manifestBytes, err := manifest.Marshal(model)
	if err != nil {
		return plan.Plan{}, manifest.Model{}, err
	}
	manifestPath := filepath.Join(root, ".acctx", "manifest.yaml")
	existing, _ := os.ReadFile(manifestPath)
	for index := range operationPlan.Operations {
		operation := &operationPlan.Operations[index]
		if operation.Path != ".acctx/manifest.yaml" {
			continue
		}
		operation.Payload = manifestBytes
		operation.BeforeDigest = fsx.BytesDigest(existing)
		operation.AfterDigest = fsx.BytesDigest(manifestBytes)
		if len(existing) == 0 {
			operation.Kind = plan.Create
		} else if string(existing) == string(manifestBytes) {
			operation.Kind = plan.Skip
		} else {
			operation.Kind = plan.Update
		}
	}

	return plan.New(operationPlan.Command, operationPlan.Root, operationPlan.CreatedAt, operationPlan.Operations), model, nil
}
