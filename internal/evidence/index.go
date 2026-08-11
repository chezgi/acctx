package evidence

import (
	"acctx/internal/fsx"
	"acctx/internal/manifest"
	"acctx/internal/task"
	"acctx/internal/year"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type File struct {
	Path      string `json:"path"`
	Category  string `json:"category"`
	Size      int64  `json:"size"`
	SHA256    string `json:"sha256"`
	MediaType string `json:"media_type"`
}

type Index struct {
	SchemaVersion             int       `json:"schema_version"`
	Scope                     string    `json:"scope"`
	ProjectID                 string    `json:"project_id"`
	FiscalYear                string    `json:"fiscal_year"`
	TaskID                    string    `json:"task_id"`
	TaskType                  string    `json:"task_type"`
	SkillID                   string    `json:"skill_id"`
	ContentVersion            string    `json:"content_version"`
	GeneratedAt               time.Time `json:"generated_at"`
	SourceDigest              string    `json:"source_digest"`
	Files                     []File    `json:"files"`
	TotalBytes                int64     `json:"total_bytes"`
	Warnings                  []string  `json:"warnings,omitempty"`
	FinalHumanReviewRequired  bool      `json:"final_human_review_required"`
}

type Options struct {
	IncludeCompany   bool
	IncludeYearInputs bool
	ExcludePaths     []string
}

type sourceRoot struct {
	Relative string
	Category string
	TaskRoot bool
}

func BuildTask(root, fiscalYear, taskID string, options Options, now time.Time) (Index, error) {
	projectManifest, err := manifest.Load(root)
	if err != nil {
		return Index{}, err
	}
	if _, err := year.Load(root, fiscalYear); err != nil {
		return Index{}, err
	}
	taskModel, err := task.Load(root, fiscalYear, taskID)
	if err != nil {
		return Index{}, err
	}

	roots := []sourceRoot{
		{Relative: filepath.ToSlash(filepath.Join("accounting", "fiscal-years", fiscalYear, "year.yaml")), Category: "year-context"},
		{Relative: filepath.ToSlash(filepath.Join("accounting", "fiscal-years", fiscalYear, "work", taskID)), Category: "task", TaskRoot: true},
	}
	if options.IncludeCompany {
		roots = append(roots, sourceRoot{Relative: "accounting/company", Category: "company"})
	}
	if options.IncludeYearInputs {
		roots = append(roots, sourceRoot{Relative: filepath.ToSlash(filepath.Join("accounting", "fiscal-years", fiscalYear, "inputs")), Category: "year-input"})
	}

	excluded := map[string]bool{}
	for _, relative := range options.ExcludePaths {
		clean, err := cleanRelative(relative)
		if err != nil {
			return Index{}, err
		}
		excluded[clean] = true
	}

	files := []File{}
	seen := map[string]bool{}
	for _, selected := range roots {
		if err := collect(root, selected, excluded, seen, &files); err != nil {
			return Index{}, err
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })

	hash := sha256.New()
	var totalBytes int64
	taskInputCount := 0
	for _, file := range files {
		fmt.Fprintf(hash, "%s\x00%s\x00%d\x00%s\x00", file.Path, file.SHA256, file.Size, file.Category)
		totalBytes += file.Size
		if file.Category == "task-input" {
			taskInputCount++
		}
	}
	warnings := []string{}
	if taskInputCount == 0 {
		warnings = append(warnings, "Task workspace contains no files under inputs/; referenced evidence may not be portable.")
	}
	return Index{
		SchemaVersion:            1,
		Scope:                    "task",
		ProjectID:                projectManifest.Project.ID,
		FiscalYear:               fiscalYear,
		TaskID:                   taskID,
		TaskType:                 taskModel.Type,
		SkillID:                  taskModel.SkillID,
		ContentVersion:           projectManifest.Generator.ContentVersion,
		GeneratedAt:              now.UTC(),
		SourceDigest:             "sha256:" + hex.EncodeToString(hash.Sum(nil)),
		Files:                    files,
		TotalBytes:               totalBytes,
		Warnings:                 warnings,
		FinalHumanReviewRequired: true,
	}, nil
}

func collect(root string, selected sourceRoot, excluded, seen map[string]bool, files *[]File) error {
	relative, err := cleanRelative(selected.Relative)
	if err != nil {
		return err
	}
	absolute, err := fsx.ResolveInside(root, filepath.FromSlash(relative))
	if err != nil {
		return err
	}
	info, err := os.Lstat(absolute)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("bundle scope contains symlink %s", relative)
	}
	if info.Mode().IsRegular() {
		return appendFile(root, absolute, selected.Category, excluded, seen, files)
	}
	if !info.IsDir() {
		return fmt.Errorf("bundle scope contains unsupported file type %s", relative)
	}
	return filepath.WalkDir(absolute, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == absolute {
			return nil
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		projectRelative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		projectRelative = filepath.ToSlash(projectRelative)
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("bundle scope contains symlink %s", projectRelative)
		}
		if entry.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("bundle scope contains unsupported file type %s", projectRelative)
		}
		if entry.Name() == ".gitkeep" {
			return nil
		}
		category := selected.Category
		if selected.TaskRoot {
			taskRelative, err := filepath.Rel(absolute, path)
			if err != nil {
				return err
			}
			category = taskCategory(filepath.ToSlash(taskRelative))
		}
		return appendFile(root, path, category, excluded, seen, files)
	})
}

func appendFile(root, absolute, category string, excluded, seen map[string]bool, files *[]File) error {
	relative, err := filepath.Rel(root, absolute)
	if err != nil {
		return err
	}
	relative, err = cleanRelative(filepath.ToSlash(relative))
	if err != nil {
		return err
	}
	if excluded[relative] || seen[relative] || filepath.Base(relative) == ".gitkeep" {
		return nil
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return err
	}
	digest, err := fsx.FileDigest(absolute)
	if err != nil {
		return err
	}
	mediaType := mime.TypeByExtension(strings.ToLower(filepath.Ext(relative)))
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	seen[relative] = true
	*files = append(*files, File{Path: relative, Category: category, Size: info.Size(), SHA256: digest, MediaType: mediaType})
	return nil
}

func taskCategory(relative string) string {
	first := relative
	if index := strings.IndexByte(relative, '/'); index >= 0 {
		first = relative[:index]
	}
	switch first {
	case "inputs":
		return "task-input"
	case "templates":
		return "task-template"
	case "calculations":
		return "calculation"
	case "drafts":
		return "draft"
	default:
		return "task-metadata"
	}
}

func cleanRelative(value string) (string, error) {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(value)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || filepath.IsAbs(filepath.FromSlash(value)) {
		return "", fmt.Errorf("invalid project-relative path %q", value)
	}
	return clean, nil
}
