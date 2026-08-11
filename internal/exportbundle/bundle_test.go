package exportbundle

import (
	bundle "acctx/internal/content"
	"acctx/internal/manifest"
	"acctx/internal/plan"
	"acctx/internal/task"
	"acctx/internal/year"
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func createBundleFixture(t *testing.T, taskType string) (string, string) {
	t.Helper()
	root := t.TempDir()
	manifestModel := manifest.Model{
		SchemaVersion: 1,
		Project: manifest.Project{ID: "example", Preset: "ir-software-kb-techpark", InitializedAt: time.Unix(1, 0).UTC()},
		Generator: manifest.Generator{CLIVersion: "test", ContentVersion: "0.4.0", ContentDigest: "sha256:test"},
	}
	manifestBytes, err := manifest.Marshal(manifestModel)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".acctx"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".acctx", "manifest.yaml"), manifestBytes, 0o644); err != nil {
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
	taskPlan, model, err := task.BuildInitPlan(root, taskType, task.Options{Year: "1405", Period: "Q1"}, contentBundle, time.Unix(3, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plan.Apply(taskPlan); err != nil {
		t.Fatal(err)
	}
	inputPath := filepath.Join(root, "accounting", "fiscal-years", "1405", "work", model.ID, "inputs", "source.txt")
	if err := os.WriteFile(inputPath, []byte("evidence\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, model.ID
}

func TestWriteAndVerifyTaskBundle(t *testing.T) {
	root, taskID := createBundleFixture(t, "vat")
	result, err := WriteTask(root, Options{
		FiscalYear: "1405",
		TaskID: taskID,
		BundleType: TypeTaxPack,
		IncludeCompany: true,
		OutputPath: "accounting/fiscal-years/1405/outputs/vat-q1.zip",
		GeneratedAt: time.Unix(10, 0).UTC(),
		CLIVersion: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Manifest.Status != "draft" || result.Manifest.SubmissionPerformed || result.Manifest.FileCount == 0 {
		t.Fatalf("manifest=%#v", result.Manifest)
	}
	verification, err := Verify(filepath.Join(root, filepath.FromSlash(result.OutputPath)))
	if err != nil {
		t.Fatal(err)
	}
	if !verification.Valid || verification.BundleID != result.Manifest.BundleID {
		t.Fatalf("verification=%#v", verification)
	}
	if _, err := WriteTask(root, Options{FiscalYear: "1405", TaskID: taskID, OutputPath: result.OutputPath}); err == nil {
		t.Fatal("expected overwrite protection")
	}
}

func TestVerifyDetectsTamperedFile(t *testing.T) {
	root, taskID := createBundleFixture(t, "vat")
	result, err := WriteTask(root, Options{FiscalYear: "1405", TaskID: taskID, OutputPath: "accounting/fiscal-years/1405/outputs/task.zip", GeneratedAt: time.Unix(10, 0).UTC()})
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, filepath.FromSlash(result.OutputPath))
	tampered := filepath.Join(root, "accounting", "fiscal-years", "1405", "outputs", "tampered.zip")
	if err := rewriteZip(source, tampered, "files/accounting/fiscal-years/1405/work/"+taskID+"/inputs/source.txt"); err != nil {
		t.Fatal(err)
	}
	verification, err := Verify(tampered)
	if err != nil {
		t.Fatal(err)
	}
	if verification.Valid {
		t.Fatal("tampered bundle must fail verification")
	}
}

func rewriteZip(source, destination, tamperName string) error {
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	archive := zip.NewWriter(file)
	for _, entry := range reader.File {
		writer, err := archive.CreateHeader(&zip.FileHeader{Name: entry.Name, Method: zip.Deflate})
		if err != nil {
			return err
		}
		if entry.Name == tamperName {
			if _, err := writer.Write([]byte("tampered\n")); err != nil {
				return err
			}
			continue
		}
		input, err := entry.Open()
		if err != nil {
			return err
		}
		if _, err := io.Copy(writer, input); err != nil {
			input.Close()
			return err
		}
		if err := input.Close(); err != nil {
			return err
		}
	}
	if err := archive.Close(); err != nil {
		return err
	}
	return file.Close()
}
