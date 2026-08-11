package task

import (
	bundle "acctx/internal/content"
	"acctx/internal/fsx"
	"acctx/internal/plan"
	"acctx/internal/year"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var safeName = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z_-]{0,63}$`)

type Definition struct {
	Type         string `json:"type"`
	SkillID      string `json:"skill_id"`
	TemplateRoot string `json:"template_root,omitempty"`
	TitleFA      string `json:"title_fa"`
}

type Model struct {
	SchemaVersion int       `json:"schema_version"`
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	Year          string    `json:"year"`
	Period        string    `json:"period,omitempty"`
	SkillID       string    `json:"skill_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type Options struct {
	Year   string
	Period string
}

type Status struct {
	Model            Model `json:"task"`
	InputFiles       int   `json:"input_files"`
	TemplateFiles    int   `json:"template_files"`
	CalculationFiles int   `json:"calculation_files"`
	DraftFiles       int   `json:"draft_files"`
}

var definitions = map[string]Definition{
	"vat":                       {Type: "vat", SkillID: "acctx-vat", TemplateRoot: "templates/tasks/vat", TitleFA: "مالیات بر ارزش افزوده"},
	"corporate-tax":             {Type: "corporate-tax", SkillID: "acctx-corporate-income-tax", TemplateRoot: "templates/tasks/corporate-tax", TitleFA: "مالیات عملکرد"},
	"payroll-tax":               {Type: "payroll-tax", SkillID: "acctx-payroll-tax", TemplateRoot: "templates/tasks/payroll-tax", TitleFA: "مالیات حقوق"},
	"tax-defense":               {Type: "tax-defense", SkillID: "acctx-tax-defense", TemplateRoot: "templates/tasks/tax-defense", TitleFA: "دفاع مالیاتی"},
	"audit":                     {Type: "audit", SkillID: "acctx-audit-preparation", TitleFA: "آماده‌سازی حسابرسی"},
	"bank-reconciliation":       {Type: "bank-reconciliation", SkillID: "acctx-bank-reconciliation", TitleFA: "تطبیق بانکی"},
	"financial-statements":      {Type: "financial-statements", SkillID: "acctx-financial-statements", TitleFA: "صورت‌های مالی"},
	"social-security":           {Type: "social-security", SkillID: "acctx-social-security", TitleFA: "تأمین اجتماعی"},
	"social-security-defense":   {Type: "social-security-defense", SkillID: "acctx-social-security-defense", TitleFA: "دفاع تأمین اجتماعی"},
	"taxpayer-system":           {Type: "taxpayer-system", SkillID: "acctx-taxpayer-system", TitleFA: "سامانه مؤدیان"},
	"rd-tax-credit":             {Type: "rd-tax-credit", SkillID: "acctx-rd-tax-credit", TitleFA: "اعتبار مالیاتی تحقیق و توسعه"},
	"revenue-recognition":       {Type: "revenue-recognition", SkillID: "acctx-revenue-recognition", TitleFA: "شناسایی درآمد"},
	"electronic-books":          {Type: "electronic-books", SkillID: "acctx-electronic-books", TitleFA: "دفاتر الکترونیکی"},
	"article-169":               {Type: "article-169", SkillID: "acctx-article-169", TitleFA: "معاملات ماده ۱۶۹"},
	"compliance-calendar":       {Type: "compliance-calendar", SkillID: "acctx-compliance-calendar", TitleFA: "تقویم تکالیف"},
	"monthly-close":             {Type: "monthly-close", SkillID: "acctx-workspace", TitleFA: "بستن ماهانه"},
}

func Definitions() []Definition {
	result := make([]Definition, 0, len(definitions))
	for _, definition := range definitions {
		result = append(result, definition)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Type < result[j].Type })
	return result
}

func DefinitionFor(taskType string) (Definition, bool) {
	definition, ok := definitions[taskType]
	return definition, ok
}

func normalizeSegment(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	if !safeName.MatchString(value) {
		return "", fmt.Errorf("invalid task path segment %q", value)
	}
	return value, nil
}

func instanceID(taskType, period string) (string, error) {
	base, err := normalizeSegment(taskType)
	if err != nil {
		return "", err
	}
	if period == "" {
		return base, nil
	}
	normalizedPeriod, err := normalizeSegment(period)
	if err != nil {
		return "", err
	}
	return base + "-" + normalizedPeriod, nil
}

func BuildInitPlan(root, taskType string, options Options, contentBundle bundle.Bundle, now time.Time) (plan.Plan, Model, error) {
	definition, ok := DefinitionFor(taskType)
	if !ok {
		return plan.Plan{}, Model{}, fmt.Errorf("unknown task type %q", taskType)
	}
	if options.Year == "" {
		return plan.Plan{}, Model{}, fmt.Errorf("--year is required")
	}
	if _, err := year.Load(root, options.Year); err != nil {
		return plan.Plan{}, Model{}, fmt.Errorf("fiscal year %q is not initialized: %w", options.Year, err)
	}
	id, err := instanceID(taskType, options.Period)
	if err != nil {
		return plan.Plan{}, Model{}, err
	}
	base := filepath.ToSlash(filepath.Join("accounting", "fiscal-years", options.Year, "work", id))

	model := Model{
		SchemaVersion: 1,
		ID:            id,
		Type:          definition.Type,
		Year:          options.Year,
		Period:        options.Period,
		SkillID:       definition.SkillID,
		Status:        "draft",
		CreatedAt:     now.UTC(),
	}
	if existing, loadErr := Load(root, options.Year, id); loadErr == nil {
		if existing.Type != model.Type || existing.Year != model.Year || existing.Period != model.Period || existing.SkillID != model.SkillID {
			return plan.Plan{}, Model{}, fmt.Errorf("existing task configuration conflicts")
		}
		model.CreatedAt = existing.CreatedAt
		model.Status = existing.Status
	}

	descriptor, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return plan.Plan{}, Model{}, err
	}
	descriptor = append(descriptor, '\n')
	readme := []byte(fmt.Sprintf("# %s\n\n- Task ID: `%s`\n- Fiscal year: `%s`\n- Skill: `%s`\n\nPlace or reference original company files under `inputs/`. Complete task templates under `templates/`. Put deterministic command results under `calculations/` and AI-produced drafts under `drafts/`.\n", definition.TitleFA, id, options.Year, definition.SkillID))
	checklist := []byte("# Final checklist\n\n- [ ] Original inputs are preserved.\n- [ ] Every material number is traceable to a source.\n- [ ] Missing facts and assumptions are listed.\n- [ ] Applicable period-specific rules were verified.\n- [ ] Deterministic calculations were run when available.\n- [ ] Draft output was reviewed by an authorized human.\n")

	operations := []plan.Operation{
		fileOperation(root, base+"/task.yaml", descriptor),
		fileOperation(root, base+"/README.md", readme),
		fileOperation(root, base+"/checklist.md", checklist),
		fileOperation(root, base+"/inputs/.gitkeep", nil),
		fileOperation(root, base+"/templates/.gitkeep", nil),
		fileOperation(root, base+"/calculations/.gitkeep", nil),
		fileOperation(root, base+"/drafts/.gitkeep", nil),
	}

	if definition.TemplateRoot != "" {
		templates, readErr := contentBundle.ReadTree(definition.TemplateRoot)
		if readErr != nil && !os.IsNotExist(readErr) && readErr != fs.ErrNotExist {
			return plan.Plan{}, Model{}, readErr
		}
		for relative, data := range templates {
			operations = append(operations, fileOperation(root, filepath.ToSlash(filepath.Join(base, "templates", relative)), data))
		}
	}
	return plan.New("task init", root, now, operations), model, nil
}

func fileOperation(root, relative string, data []byte) plan.Operation {
	path, err := fsx.ResolveInside(root, filepath.FromSlash(relative))
	if err != nil {
		return plan.Operation{Kind: plan.Conflict, Path: relative, Message: err.Error()}
	}
	existing, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return plan.Operation{Kind: plan.Create, Path: relative, AfterDigest: fsx.BytesDigest(data), Payload: data}
	}
	if err != nil {
		return plan.Operation{Kind: plan.Conflict, Path: relative, Message: "existing task path is not a regular readable file"}
	}
	if string(existing) == string(data) {
		return plan.Operation{Kind: plan.Skip, Path: relative, AfterDigest: fsx.BytesDigest(data)}
	}
	return plan.Operation{Kind: plan.Conflict, Path: relative, BeforeDigest: fsx.BytesDigest(existing), Message: "existing task file differs"}
}

func Load(root, fiscalYear, id string) (Model, error) {
	if err := year.ValidateID(fiscalYear); err != nil {
		return Model{}, err
	}
	if !safeName.MatchString(id) {
		return Model{}, fmt.Errorf("invalid task id %q", id)
	}
	path := filepath.Join(root, "accounting", "fiscal-years", fiscalYear, "work", id, "task.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return Model{}, err
	}
	var model Model
	if err := json.Unmarshal(data, &model); err != nil {
		return Model{}, err
	}
	if model.SchemaVersion != 1 || model.ID != id || model.Year != fiscalYear {
		return Model{}, fmt.Errorf("invalid task descriptor")
	}
	if _, ok := DefinitionFor(model.Type); !ok {
		return Model{}, fmt.Errorf("unknown task type in descriptor")
	}
	return model, nil
}

func List(root, fiscalYear string) ([]Model, error) {
	if _, err := year.Load(root, fiscalYear); err != nil {
		return nil, err
	}
	base := filepath.Join(root, "accounting", "fiscal-years", fiscalYear, "work")
	entries, err := os.ReadDir(base)
	if os.IsNotExist(err) {
		return []Model{}, nil
	}
	if err != nil {
		return nil, err
	}
	models := []Model{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		model, err := Load(root, fiscalYear, entry.Name())
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
	return models, nil
}

func ReadStatus(root, fiscalYear, id string) (Status, error) {
	model, err := Load(root, fiscalYear, id)
	if err != nil {
		return Status{}, err
	}
	base := filepath.Join(root, "accounting", "fiscal-years", fiscalYear, "work", id)
	return Status{
		Model:            model,
		InputFiles:       countFiles(filepath.Join(base, "inputs")),
		TemplateFiles:    countFiles(filepath.Join(base, "templates")),
		CalculationFiles: countFiles(filepath.Join(base, "calculations")),
		DraftFiles:       countFiles(filepath.Join(base, "drafts")),
	}, nil
}

func countFiles(root string) int {
	count := 0
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && entry.Name() != ".gitkeep" {
			count++
		}
		return nil
	})
	return count
}
