package exportbundle

import (
	"acctx/internal/evidence"
	"acctx/internal/fsx"
	"acctx/internal/task"
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	TypeTask      = "task"
	TypeAuditPack = "audit-pack"
	TypeTaxPack   = "tax-pack"
)

type Options struct {
	FiscalYear        string
	TaskID            string
	BundleType        string
	IncludeCompany    bool
	IncludeYearInputs bool
	OutputPath        string
	Force             bool
	GeneratedAt       time.Time
	CLIVersion        string
}

type Manifest struct {
	SchemaVersion            int       `json:"schema_version"`
	FormatID                 string    `json:"format_id"`
	FormatVersion            string    `json:"format_version"`
	BundleID                 string    `json:"bundle_id"`
	BundleType               string    `json:"bundle_type"`
	Status                   string    `json:"status"`
	SubmissionPerformed      bool      `json:"submission_performed"`
	ProjectID                string    `json:"project_id"`
	FiscalYear               string    `json:"fiscal_year"`
	TaskID                   string    `json:"task_id"`
	TaskType                 string    `json:"task_type"`
	SkillID                  string    `json:"skill_id"`
	ContentVersion           string    `json:"content_version"`
	CLIVersion               string    `json:"cli_version"`
	GeneratedAt              time.Time `json:"generated_at"`
	SourceDigest             string    `json:"source_digest"`
	EvidenceIndexDigest      string    `json:"evidence_index_digest"`
	FileCount                int       `json:"file_count"`
	TotalBytes               int64     `json:"total_bytes"`
	IncludesCompany          bool      `json:"includes_company"`
	IncludesYearInputs       bool      `json:"includes_year_inputs"`
	FinalHumanReviewRequired bool      `json:"final_human_review_required"`
	RuleVerificationRequired bool      `json:"rule_verification_required"`
	Warnings                 []string  `json:"warnings,omitempty"`
}

type Result struct {
	OutputPath string   `json:"output_path"`
	Size       int64    `json:"size"`
	SHA256     string   `json:"sha256"`
	Manifest   Manifest `json:"manifest"`
}

func WriteTask(root string, options Options) (Result, error) {
	if options.BundleType == "" {
		options.BundleType = TypeTask
	}
	if options.GeneratedAt.IsZero() {
		options.GeneratedAt = time.Now().UTC()
	}
	if options.OutputPath == "" {
		return Result{}, fmt.Errorf("output path is required")
	}
	if err := validateBundleType(root, options); err != nil {
		return Result{}, err
	}
	output, err := fsx.ResolveInside(root, filepath.FromSlash(options.OutputPath))
	if err != nil {
		return Result{}, err
	}
	if info, statErr := os.Lstat(output); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return Result{}, fmt.Errorf("output path is not a regular file")
		}
		if !options.Force {
			return Result{}, fmt.Errorf("output already exists; use --force to replace it")
		}
	} else if !os.IsNotExist(statErr) {
		return Result{}, statErr
	}

	index, err := evidence.BuildTask(root, options.FiscalYear, options.TaskID, evidence.Options{
		IncludeCompany:    options.IncludeCompany,
		IncludeYearInputs: options.IncludeYearInputs,
		ExcludePaths:      []string{options.OutputPath},
	}, options.GeneratedAt)
	if err != nil {
		return Result{}, err
	}
	indexBytes, err := marshalJSON(index)
	if err != nil {
		return Result{}, err
	}
	manifest := Manifest{
		SchemaVersion:            1,
		FormatID:                 "acctx-controlled-bundle",
		FormatVersion:            "1.0.0",
		BundleID:                 stableBundleID(options.BundleType, index.SourceDigest),
		BundleType:               options.BundleType,
		Status:                   "draft",
		SubmissionPerformed:      false,
		ProjectID:                index.ProjectID,
		FiscalYear:               index.FiscalYear,
		TaskID:                   index.TaskID,
		TaskType:                 index.TaskType,
		SkillID:                  index.SkillID,
		ContentVersion:           index.ContentVersion,
		CLIVersion:               options.CLIVersion,
		GeneratedAt:              options.GeneratedAt.UTC(),
		SourceDigest:             index.SourceDigest,
		EvidenceIndexDigest:      fsx.BytesDigest(indexBytes),
		FileCount:                len(index.Files),
		TotalBytes:               index.TotalBytes,
		IncludesCompany:          options.IncludeCompany,
		IncludesYearInputs:       options.IncludeYearInputs,
		FinalHumanReviewRequired: true,
		RuleVerificationRequired: true,
		Warnings:                 append([]string(nil), index.Warnings...),
	}
	manifestBytes, err := marshalJSON(manifest)
	if err != nil {
		return Result{}, err
	}

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return Result{}, err
	}
	temporary, err := os.CreateTemp(filepath.Dir(output), ".acctx-export-*.zip")
	if err != nil {
		return Result{}, err
	}
	temporaryPath := temporary.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(temporaryPath)
		}
	}()

	archive := zip.NewWriter(temporary)
	if err := writeBytes(archive, "bundle-manifest.json", manifestBytes); err != nil {
		_ = archive.Close()
		_ = temporary.Close()
		return Result{}, err
	}
	if err := writeBytes(archive, "evidence-index.json", indexBytes); err != nil {
		_ = archive.Close()
		_ = temporary.Close()
		return Result{}, err
	}
	for _, indexedFile := range index.Files {
		if err := writeSourceFile(archive, root, indexedFile); err != nil {
			_ = archive.Close()
			_ = temporary.Close()
			return Result{}, err
		}
	}
	if err := archive.Close(); err != nil {
		_ = temporary.Close()
		return Result{}, err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return Result{}, err
	}
	if err := temporary.Close(); err != nil {
		return Result{}, err
	}
	if err := os.Rename(temporaryPath, output); err != nil {
		return Result{}, err
	}
	cleanup = false

	info, err := os.Stat(output)
	if err != nil {
		return Result{}, err
	}
	digest, err := fsx.FileDigest(output)
	if err != nil {
		return Result{}, err
	}
	return Result{OutputPath: options.OutputPath, Size: info.Size(), SHA256: digest, Manifest: manifest}, nil
}

func validateBundleType(root string, options Options) error {
	taskModel, err := task.Load(root, options.FiscalYear, options.TaskID)
	if err != nil {
		return err
	}
	switch options.BundleType {
	case TypeTask:
		return nil
	case TypeAuditPack:
		if taskModel.Type != "audit" {
			return fmt.Errorf("audit-pack requires an audit task")
		}
		return nil
	case TypeTaxPack:
		allowed := map[string]bool{
			"vat": true, "corporate-tax": true, "payroll-tax": true,
			"tax-defense": true, "taxpayer-system": true, "article-169": true,
			"electronic-books": true, "rd-tax-credit": true,
		}
		if !allowed[taskModel.Type] {
			return fmt.Errorf("task type %q is not a tax-pack task", taskModel.Type)
		}
		return nil
	default:
		return fmt.Errorf("unknown bundle type %q", options.BundleType)
	}
}

func marshalJSON(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func stableBundleID(bundleType, sourceDigest string) string {
	sum := sha256.Sum256([]byte(bundleType + "\x00" + sourceDigest))
	return "BND-" + strings.ToUpper(hex.EncodeToString(sum[:8]))
}

func writeBytes(archive *zip.Writer, name string, data []byte) error {
	writer, err := archive.CreateHeader(stableHeader(name))
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

func writeSourceFile(archive *zip.Writer, root string, indexed evidence.File) error {
	archiveName, err := safeArchiveName("files/" + indexed.Path)
	if err != nil {
		return err
	}
	absolute, err := fsx.ResolveInside(root, filepath.FromSlash(indexed.Path))
	if err != nil {
		return err
	}
	info, err := os.Lstat(absolute)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("source changed to unsupported file type: %s", indexed.Path)
	}
	file, err := os.Open(absolute)
	if err != nil {
		return err
	}
	defer file.Close()
	writer, err := archive.CreateHeader(stableHeader(archiveName))
	if err != nil {
		return err
	}
	hash := sha256.New()
	written, err := io.Copy(writer, io.TeeReader(file, hash))
	if err != nil {
		return err
	}
	digest := "sha256:" + hex.EncodeToString(hash.Sum(nil))
	if written != indexed.Size || digest != indexed.SHA256 {
		return fmt.Errorf("source changed during export: %s", indexed.Path)
	}
	return nil
}

func stableHeader(name string) *zip.FileHeader {
	clean, _ := safeArchiveName(name)
	header := &zip.FileHeader{Name: clean, Method: zip.Deflate}
	header.SetMode(0o644)
	header.Modified = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
	return header
}

func safeArchiveName(name string) (string, error) {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(name)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
		return "", fmt.Errorf("unsafe archive entry %q", name)
	}
	return clean, nil
}
