package year

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBuildInitPlanUsesJalaliLeapYearAndHistoricalMode(t *testing.T) {
	root := t.TempDir()
	plan, model, err := BuildInitPlan(root, "1399", Options{Mode: ModeHistorical}, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if model.EndsOn != "1399-12-30" || model.RulesetYear != 1399 {
		t.Fatalf("model=%#v", model)
	}
	if plan.Summary.Creates != 4 {
		t.Fatalf("summary=%#v", plan.Summary)
	}
}

func TestYearBefore1397MustBeArchive(t *testing.T) {
	_, _, err := BuildInitPlan(t.TempDir(), "1396", Options{Mode: ModeHistorical}, time.Now())
	if err == nil {
		t.Fatal("expected archive-only error")
	}
}

func TestListAndStatus(t *testing.T) {
	root := t.TempDir()
	plan, _, err := BuildInitPlan(root, "1405", Options{}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range plan.Operations {
		if operation.Kind != "CREATE" {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(operation.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, operation.Payload, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	models, err := List(root)
	if err != nil || len(models) != 1 {
		t.Fatalf("models=%#v err=%v", models, err)
	}
	status, err := ReadStatus(root, "1405")
	if err != nil || status.Model.Mode != ModeOperational {
		t.Fatalf("status=%#v err=%v", status, err)
	}
}
