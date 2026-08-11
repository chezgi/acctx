package task

import (
	bundle "acctx/internal/content"
	"acctx/internal/plan"
	"acctx/internal/year"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBuildInitPlanCopiesVATTemplate(t *testing.T) {
	root := t.TempDir()
	yearPlan, _, err := year.BuildInitPlan(root, "1405", year.Options{}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plan.Apply(yearPlan); err != nil {
		t.Fatal(err)
	}
	contentBundle, err := bundle.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	taskPlan, model, err := BuildInitPlan(root, "vat", Options{Year: "1405", Period: "Q1"}, contentBundle, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if model.ID != "vat-q1" || model.SkillID != "acctx-vat" {
		t.Fatalf("model=%#v", model)
	}
	if _, err := plan.Apply(taskPlan); err != nil {
		t.Fatal(err)
	}
	templatePath := filepath.Join(root, "accounting", "fiscal-years", "1405", "work", "vat-q1", "templates", "input.csv")
	if _, err := os.Stat(templatePath); err != nil {
		t.Fatal(err)
	}
}

func TestUnknownTaskTypeFails(t *testing.T) {
	contentBundle, _ := bundle.Embedded()
	_, _, err := BuildInitPlan(t.TempDir(), "unknown", Options{Year: "1405"}, contentBundle, time.Now())
	if err == nil {
		t.Fatal("expected unknown task error")
	}
}
