package year

import (
	"acctx/internal/fsx"
	"acctx/internal/plan"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	ModeOperational    = "operational"
	ModeHistorical     = "historical"
	ModeReconstruction = "reconstruction"
	ModeArchive        = "archive"
)

var safeID = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z_-]{0,63}$`)
var jalaliDate = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

type Model struct {
	SchemaVersion int       `json:"schema_version"`
	ID            string    `json:"id"`
	Mode          string    `json:"mode"`
	StartsOn      string    `json:"starts_on"`
	EndsOn        string    `json:"ends_on"`
	RulesetYear   int       `json:"ruleset_year"`
	CreatedAt     time.Time `json:"created_at"`
}

type Options struct {
	Mode        string
	StartsOn    string
	EndsOn      string
	RulesetYear int
}

type Status struct {
	Model       Model `json:"year"`
	InputFiles  int   `json:"input_files"`
	WorkFiles   int   `json:"work_files"`
	OutputFiles int   `json:"output_files"`
}

func ValidateID(id string) error {
	if !safeID.MatchString(id) {
		return fmt.Errorf("invalid fiscal-year id %q", id)
	}
	return nil
}

func validateMode(mode string) error {
	switch mode {
	case ModeOperational, ModeHistorical, ModeReconstruction, ModeArchive:
		return nil
	default:
		return fmt.Errorf("invalid fiscal-year mode %q", mode)
	}
}

func validateDate(value string) error {
	if !jalaliDate.MatchString(value) {
		return fmt.Errorf("invalid Jalali date %q", value)
	}
	parts := strings.Split(value, "-")
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return fmt.Errorf("invalid Jalali date %q", value)
	}
	if month >= 7 && month <= 11 && day > 30 {
		return fmt.Errorf("invalid Jalali date %q", value)
	}
	if month == 12 && day > 30 {
		return fmt.Errorf("invalid Jalali date %q", value)
	}
	return nil
}

func isJalaliLeap(year int) bool {
	breaks := []int{-61, 9, 38, 199, 426, 686, 756, 818, 1111, 1181, 1210, 1635, 2060, 2097, 2192, 2262, 2324, 2394, 2456, 3178}
	previous := breaks[0]
	jump := 0
	for _, current := range breaks[1:] {
		jump = current - previous
		if year < current {
			break
		}
		previous = current
	}
	n := year - previous
	if jump-n < 6 {
		n = n - jump + ((jump+4)/33)*33
	}
	leap := ((n+1)%33 - 1) % 4
	if leap == -1 {
		leap = 4
	}
	return leap == 0
}

func defaultDates(id string) (string, string, int, error) {
	numericYear, err := strconv.Atoi(id)
	if err != nil {
		return "", "", 0, fmt.Errorf("non-numeric fiscal-year ids require --start, --end, and --ruleset-year")
	}
	lastDay := 29
	if isJalaliLeap(numericYear) {
		lastDay = 30
	}
	return fmt.Sprintf("%04d-01-01", numericYear), fmt.Sprintf("%04d-12-%02d", numericYear, lastDay), numericYear, nil
}

func BuildInitPlan(root, id string, options Options, now time.Time) (plan.Plan, Model, error) {
	if err := ValidateID(id); err != nil {
		return plan.Plan{}, Model{}, err
	}
	if options.Mode == "" {
		options.Mode = ModeOperational
	}
	if err := validateMode(options.Mode); err != nil {
		return plan.Plan{}, Model{}, err
	}

	defaultStart, defaultEnd, defaultRuleset, defaultErr := defaultDates(id)
	if options.StartsOn == "" {
		if defaultErr != nil {
			return plan.Plan{}, Model{}, defaultErr
		}
		options.StartsOn = defaultStart
	}
	if options.EndsOn == "" {
		if defaultErr != nil {
			return plan.Plan{}, Model{}, defaultErr
		}
		options.EndsOn = defaultEnd
	}
	if options.RulesetYear == 0 && defaultErr == nil && defaultRuleset >= 1397 {
		options.RulesetYear = defaultRuleset
	}
	if err := validateDate(options.StartsOn); err != nil {
		return plan.Plan{}, Model{}, err
	}
	if err := validateDate(options.EndsOn); err != nil {
		return plan.Plan{}, Model{}, err
	}
	if options.EndsOn < options.StartsOn {
		return plan.Plan{}, Model{}, fmt.Errorf("fiscal-year end precedes start")
	}

	if numericYear, err := strconv.Atoi(id); err == nil && numericYear < 1397 {
		if options.Mode != ModeArchive {
			return plan.Plan{}, Model{}, fmt.Errorf("years before 1397 are archive-only")
		}
		options.RulesetYear = 0
	}
	if options.Mode != ModeArchive && options.RulesetYear < 1397 {
		return plan.Plan{}, Model{}, fmt.Errorf("a supported ruleset year is required")
	}

	model := Model{
		SchemaVersion: 1,
		ID:            id,
		Mode:          options.Mode,
		StartsOn:      options.StartsOn,
		EndsOn:        options.EndsOn,
		RulesetYear:   options.RulesetYear,
		CreatedAt:     now.UTC(),
	}
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return plan.Plan{}, Model{}, err
	}
	data = append(data, '\n')

	base := filepath.ToSlash(filepath.Join("accounting", "fiscal-years", id))
	operations := []plan.Operation{
		fileOperation(root, base+"/year.yaml", data),
		fileOperation(root, base+"/inputs/.gitkeep", nil),
		fileOperation(root, base+"/work/.gitkeep", nil),
		fileOperation(root, base+"/outputs/.gitkeep", nil),
	}
	return plan.New("year init", root, now, operations), model, nil
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
		return plan.Operation{Kind: plan.Conflict, Path: relative, Message: "existing path is not a regular readable file"}
	}
	if string(existing) == string(data) {
		return plan.Operation{Kind: plan.Skip, Path: relative, AfterDigest: fsx.BytesDigest(data)}
	}
	return plan.Operation{Kind: plan.Conflict, Path: relative, BeforeDigest: fsx.BytesDigest(existing), Message: "existing fiscal-year file differs"}
}

func Load(root, id string) (Model, error) {
	if err := ValidateID(id); err != nil {
		return Model{}, err
	}
	path := filepath.Join(root, "accounting", "fiscal-years", id, "year.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return Model{}, err
	}
	var model Model
	if err := json.Unmarshal(data, &model); err != nil {
		return Model{}, err
	}
	if model.SchemaVersion != 1 || model.ID != id {
		return Model{}, fmt.Errorf("invalid fiscal-year descriptor")
	}
	if err := validateMode(model.Mode); err != nil {
		return Model{}, err
	}
	return model, nil
}

func List(root string) ([]Model, error) {
	base := filepath.Join(root, "accounting", "fiscal-years")
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
		model, err := Load(root, entry.Name())
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

func ReadStatus(root, id string) (Status, error) {
	model, err := Load(root, id)
	if err != nil {
		return Status{}, err
	}
	base := filepath.Join(root, "accounting", "fiscal-years", id)
	return Status{
		Model:       model,
		InputFiles:  countFiles(filepath.Join(base, "inputs")),
		WorkFiles:   countFiles(filepath.Join(base, "work")),
		OutputFiles: countFiles(filepath.Join(base, "outputs")),
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
