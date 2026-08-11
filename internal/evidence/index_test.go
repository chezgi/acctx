package evidence

import (
	bundle "acctx/internal/content"
	"acctx/internal/manifest"
	"acctx/internal/plan"
	"acctx/internal/task"
	"acctx/internal/year"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func createProjectTask(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	manifestModel := manifest.Model{SchemaVersion: 1, Project: manifest.Project{ID: "example", Preset: "ir-software-kb-techpark", InitializedAt: time.Unix(1, 0).UTC()}, Generator: manifest.Generator{CLIVersion: "test", ContentVersion: "0.4.0", ContentDigest: "sha256:test"}}
	data, err := manifest.Marshal(manifestModel)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".acctx"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".acctx", "manifest.yaml"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	yearPlan, _, err := year.BuildInitPlan(root, "1405", year.Options{}, time.Unix(2, 0).UTC())
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
	taskPlan, _, err := task.BuildInitPlan(root, "vat", task.Options{Year: "1405", Period: "Q1"}, contentBundle, time.Unix(3, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plan.Apply(taskPlan); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(root, "accounting", "fiscal-years", "1405", "work", "vat-q1", "inputs", "sales.csv")
	if err := os.WriteFile(input, []byte("id,amount\n1,100\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestBuildTaskProducesSortedStableIndex(t *testing.T) {
	root := createProjectTask(t)
	generatedAt := time.Unix(10, 0).UTC()
	first, err := BuildTask(root, "1405", "vat-q1", Options{}, generatedAt)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildTask(root, "1405", "vat-q1", Options{}, generatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if first.SourceDigest != second.SourceDigest || len(first.Files) == 0 {
		t.Fatalf("first=%#v second=%#v", first, second)
	}
	for index := 1; index < len(first.Files); index++ {
		if first.Files[index-1].Path > first.Files[index].Path {
			t.Fatal("files are not sorted")
		}
	}
	foundInput := false
	for _, file := range first.Files {
		if file.Category == "task-input" && filepath.Base(file.Path) == "sales.csv" {
			foundInput = true
		}
	}
	if !foundInput {
		t.Fatal("task input missing from index")
	}
}

func TestBuildTaskRejectsSymlink(t *testing.T) {
	root := createProjectTask(t)
	taskInputs := filepath.Join(root, "accounting", "fiscal-years", "1405", "work", "vat-q1", "inputs")
	if err := os.Symlink("sales.csv", filepath.Join(taskInputs, "link.csv")); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildTask(root, "1405", "vat-q1", Options{}, time.Now()); err == nil {
		t.Fatal("expected symlink rejection")
	}
}
