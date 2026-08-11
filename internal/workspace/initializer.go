package workspace

import (
	"acctx/internal/buildinfo"
	bundle "acctx/internal/content"
	"acctx/internal/fsx"
	"acctx/internal/managedblock"
	"acctx/internal/manifest"
	"acctx/internal/plan"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type InitOptions struct {
	ProjectID string
	Preset    string
}

func readMaybe(path string) ([]byte, error) {
	value, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return value, err
}

func managedFileOp(root, path string, data []byte, previous map[string]manifest.File) plan.Operation {
	destination := filepath.Join(root, filepath.FromSlash(path))
	info, err := os.Lstat(destination)
	after := fsx.BytesDigest(data)
	if os.IsNotExist(err) {
		return plan.Operation{Kind: plan.Create, Path: path, AfterDigest: after, Payload: data}
	}
	if err != nil || !info.Mode().IsRegular() {
		return plan.Operation{Kind: plan.Conflict, Path: path, Message: "مسیر موجود قابل مدیریت نیست"}
	}
	actual, _ := fsx.FileDigest(destination)
	if actual == after {
		return plan.Operation{Kind: plan.Skip, Path: path, AfterDigest: after}
	}
	if old, ok := previous[path]; ok && actual == old.Digest {
		return plan.Operation{Kind: plan.Update, Path: path, BeforeDigest: actual, AfterDigest: after, Payload: data}
	}
	return plan.Operation{Kind: plan.Conflict, Path: path, Message: "فایل موجود تحت مدیریت acctx نیست یا تغییر کرده است"}
}

func materializeVendorTree(
	root string,
	contentBundle bundle.Bundle,
	sourceRoot string,
	destinationRoot string,
	sourceID string,
	previous map[string]manifest.File,
) ([]plan.Operation, []manifest.File, error) {
	files, err := contentBundle.ReadTree(sourceRoot)
	if err != nil {
		return nil, nil, err
	}
	operations := make([]plan.Operation, 0, len(files))
	managed := make([]manifest.File, 0, len(files))
	for relativePath, data := range files {
		path := filepath.ToSlash(filepath.Join(destinationRoot, relativePath))
		operation := managedFileOp(root, path, data, previous)
		operations = append(operations, operation)
		if operation.Kind != plan.Conflict {
			managed = append(managed, manifest.File{
				Path:     path,
				Digest:   fsx.BytesDigest(data),
				SourceID: sourceID + ":" + relativePath,
			})
		}
	}
	return operations, managed, nil
}

func BuildInitPlan(root string, options InitOptions, contentBundle bundle.Bundle, build buildinfo.Info, now time.Time) (plan.Plan, manifest.Model, error) {
	if options.ProjectID == "" {
		options.ProjectID = filepath.Base(root)
	}
	if options.Preset == "" {
		options.Preset = contentBundle.Catalog.DefaultPreset
	}
	if options.Preset != contentBundle.Catalog.DefaultPreset {
		return plan.Plan{}, manifest.Model{}, fmt.Errorf("ACCTX_PRESET_NOT_FOUND: %s", options.Preset)
	}

	var old *manifest.Model
	if existing, err := manifest.Load(root); err == nil {
		old = &existing
		if existing.Project.ID != options.ProjectID || existing.Project.Preset != options.Preset {
			return plan.Plan{}, manifest.Model{}, fmt.Errorf("project configuration conflict")
		}
	}

	previousFiles := map[string]manifest.File{}
	previousSkills := map[string]manifest.Skill{}
	if old != nil {
		for _, file := range old.Managed.Files {
			previousFiles[file.Path] = file
		}
		for _, skill := range old.Managed.Skills {
			previousSkills[skill.ID] = skill
		}
	}

	operations := []plan.Operation{}
	model := manifest.Model{
		SchemaVersion: 1,
		Project: manifest.Project{
			ID:            options.ProjectID,
			Preset:        options.Preset,
			InitializedAt: now.UTC(),
		},
		Generator: manifest.Generator{
			CLIVersion:     build.Version,
			ContentVersion: contentBundle.Catalog.ContentVersion,
			ContentDigest:  fsx.BytesDigest([]byte(contentBundle.Catalog.ContentVersion)),
		},
	}
	if old != nil {
		model.Project.InitializedAt = old.Project.InitializedAt
	}

	blockDefinitions := []struct {
		path  string
		body  []byte
		style managedblock.Style
	}{
		{
			"AGENTS.md",
			[]byte("## acctx project\n\n- Manifest: `.acctx/manifest.yaml`\n- Company: `accounting/company/`\n- Fiscal years: `accounting/fiscal-years/`\n- Skills: `skills/`\n- Workflows: `workflows/vendor/`\n- Templates: `templates/vendor/`\n- References: `references/vendor/`\n- Read the relevant skill before work.\n- Use acctx for deterministic operations only.\n"),
			managedblock.Markdown,
		},
		{
			"CLAUDE.md",
			[]byte("@AGENTS.md\n\nClaude Code skills are available under `.claude/skills/`.\n"),
			managedblock.Markdown,
		},
		{
			".gitignore",
			[]byte(".acctx/cache/\n.acctx/staging/\n.acctx/tmp/\n"),
			managedblock.Gitignore,
		},
	}
	for _, definition := range blockDefinitions {
		existing, err := readMaybe(filepath.Join(root, definition.path))
		if err != nil {
			return plan.Plan{}, model, err
		}
		merged, err := managedblock.Merge(existing, definition.body, definition.style)
		if err != nil {
			operations = append(operations, plan.Operation{Kind: plan.Conflict, Path: definition.path, Message: err.Error()})
			continue
		}
		kind := plan.Skip
		if len(existing) == 0 {
			kind = plan.Create
		} else if merged.Changed {
			kind = plan.Update
		}
		operations = append(operations, plan.Operation{
			Kind:         kind,
			Path:         definition.path,
			BeforeDigest: fsx.BytesDigest(existing),
			AfterDigest:  fsx.BytesDigest(merged.Content),
			Payload:      merged.Content,
		})
		styleName := "markdown"
		if definition.path == ".gitignore" {
			styleName = "gitignore"
		}
		model.Managed.Blocks = append(model.Managed.Blocks, manifest.Block{
			Path:       definition.path,
			BodyDigest: fsx.BytesDigest(bytes.Trim(definition.body, "\r\n")),
			Style:      styleName,
		})
	}

	companyFiles := map[string]string{
		"accounting/company/identity.yaml":        "schema_version: 1\nlegal_name_fa: \"\"\nlegal_name_en: \"\"\nnational_id: \"\"\nregistration_number: \"\"\nlegal_type: \"\"\n",
		"accounting/company/registrations.yaml":   "schema_version: 1\neconomic_number: \"\"\ntaxpayer_unit_code: \"\"\nregistration_date_jalali: \"\"\n",
		"accounting/company/tax.yaml":             "schema_version: 1\nfiscal_year_default: \"\"\nvat_registered: null\ntax_office_name_fa: \"\"\n",
		"accounting/company/contacts.yaml":        "schema_version: 1\nregistered_address_fa: \"\"\npostal_code: \"\"\nphone: \"\"\nemail: \"\"\n",
		"accounting/company/bank-accounts.yaml":   "schema_version: 1\naccounts: []\n",
		"accounting/company/knowledge-based.yaml": "schema_version: 1\nstatus: unknown\napprovals: []\n",
		"accounting/company/technology-park.yaml": "schema_version: 1\nmember: true\npark_name_fa: \"\"\napprovals: []\n",
	}
	for path, body := range companyFiles {
		operation := managedFileOp(root, path, []byte(body), previousFiles)
		operations = append(operations, operation)
		if operation.Kind != plan.Conflict {
			model.Managed.Files = append(model.Managed.Files, manifest.File{
				Path:     path,
				Digest:   fsx.BytesDigest([]byte(body)),
				SourceID: "bootstrap:" + path,
			})
		}
	}

	for _, path := range []string{
		"accounting/fiscal-years/.gitkeep",
		"skills/company/.gitkeep",
		"templates/company/.gitkeep",
		"workflows/company/.gitkeep",
		"references/company/.gitkeep",
		"examples/.gitkeep",
	} {
		operation := managedFileOp(root, path, []byte{}, previousFiles)
		operations = append(operations, operation)
		if operation.Kind != plan.Conflict {
			model.Managed.Files = append(model.Managed.Files, manifest.File{
				Path:     path,
				Digest:   fsx.BytesDigest(nil),
				SourceID: "workspace-placeholder",
			})
		}
	}

	vendorTrees := []struct {
		source      string
		destination string
		sourceID    string
	}{
		{"workflows", "workflows/vendor", "workflow"},
		{"templates", "templates/vendor", "template"},
		{"references", "references/vendor", "reference"},
	}
	for _, tree := range vendorTrees {
		treeOperations, managedFiles, err := materializeVendorTree(
			root,
			contentBundle,
			tree.source,
			tree.destination,
			tree.sourceID,
			previousFiles,
		)
		if err != nil {
			return plan.Plan{}, model, err
		}
		operations = append(operations, treeOperations...)
		model.Managed.Files = append(model.Managed.Files, managedFiles...)
	}

	for _, skill := range contentBundle.Catalog.Skills {
		files, err := contentBundle.ReadTree(skill.Path)
		if err != nil {
			return plan.Plan{}, model, err
		}
		vendorPath := "skills/vendor/" + skill.ID
		for relativePath, data := range files {
			path := filepath.ToSlash(filepath.Join(vendorPath, relativePath))
			operation := managedFileOp(root, path, data, previousFiles)
			operations = append(operations, operation)
			if operation.Kind != plan.Conflict {
				model.Managed.Files = append(model.Managed.Files, manifest.File{
					Path:     path,
					Digest:   fsx.BytesDigest(data),
					SourceID: "skill:" + skill.ID,
				})
			}
		}

		activePath := vendorPath
		providerLinks := map[string]string{}
		if previousSkill, ok := previousSkills[skill.ID]; ok && previousSkill.Override != nil {
			activePath = previousSkill.ActivePath
		}
		for _, provider := range []string{".claude/skills", ".agents/skills"} {
			linkPath := filepath.ToSlash(filepath.Join(provider, skill.ID))
			target := "../../" + activePath
			destination := filepath.Join(root, filepath.FromSlash(linkPath))
			info, err := os.Lstat(destination)
			kind := plan.Link
			if err == nil {
				if info.Mode()&os.ModeSymlink != 0 {
					actual, _ := os.Readlink(destination)
					if actual == target {
						kind = plan.Skip
					} else {
						kind = plan.Conflict
					}
				} else {
					kind = plan.Conflict
				}
			} else if !os.IsNotExist(err) {
				return plan.Plan{}, model, err
			}
			operations = append(operations, plan.Operation{
				Kind:    kind,
				Path:    linkPath,
				Target:  target,
				Message: "provider path conflict",
			})
			if kind != plan.Conflict {
				providerLinks[provider] = target
			}
		}

		tree, _ := contentBundle.ReadTree(skill.Path)
		model.Managed.Skills = append(model.Managed.Skills, manifest.Skill{
			ID:            skill.ID,
			Version:       skill.Version,
			VendorPath:    vendorPath,
			ActivePath:    activePath,
			Digest:        bundle.TreeDigest(tree),
			ProviderLinks: providerLinks,
		})
	}

	manifestBytes, err := manifest.Marshal(model)
	if err != nil {
		return plan.Plan{}, model, err
	}
	manifestPath := ".acctx/manifest.yaml"
	existingManifest, _ := readMaybe(filepath.Join(root, manifestPath))
	kind := plan.Create
	if len(existingManifest) > 0 {
		if string(existingManifest) == string(manifestBytes) {
			kind = plan.Skip
		} else {
			kind = plan.Update
		}
	}
	operations = append(operations, plan.Operation{
		Kind:         kind,
		Path:         manifestPath,
		BeforeDigest: fsx.BytesDigest(existingManifest),
		AfterDigest:  fsx.BytesDigest(manifestBytes),
		Payload:      manifestBytes,
	})
	return plan.New("init", root, now, operations), model, nil
}
