package simpleyaml

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFileReadsTopLevelScalarsAndIgnoresNestedValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "company.yaml")
	content := "schema_version: 1\nlegal_name_fa: \"شرکت نمونه\"\nvat_registered: null\naccounts:\n  - iban: IR01\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	values, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if values["legal_name_fa"] != "شرکت نمونه" {
		t.Fatalf("legal_name_fa=%q", values["legal_name_fa"])
	}
	if values["vat_registered"] != "" {
		t.Fatalf("vat_registered=%q", values["vat_registered"])
	}
	if _, ok := values["iban"]; ok {
		t.Fatal("nested field must not be treated as top-level")
	}
}
