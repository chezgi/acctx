package company

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateBootstrapRequiresNameAndNationalID(t *testing.T) {
	root := t.TempDir()
	companyDir := filepath.Join(root, "accounting", "company")
	if err := os.MkdirAll(companyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"identity.yaml":        "schema_version: 1\nlegal_name_fa: \"\"\nnational_id: \"\"\n",
		"registrations.yaml":   "schema_version: 1\n",
		"tax.yaml":             "schema_version: 1\n",
		"contacts.yaml":        "schema_version: 1\n",
		"bank-accounts.yaml":   "schema_version: 1\naccounts: []\n",
		"knowledge-based.yaml": "schema_version: 1\n",
		"technology-park.yaml": "schema_version: 1\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(companyDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	report, err := Validate(root, "bootstrap")
	if err != nil {
		t.Fatal(err)
	}
	if report.Valid || len(report.Diagnostics) != 2 {
		t.Fatalf("report=%#v", report)
	}
}
