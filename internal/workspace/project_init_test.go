package workspace

import (
	"acctx/internal/buildinfo"
	bundle "acctx/internal/content"
	"acctx/internal/plan"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProjectInitPreservesEditedCompanyFilesAndDropsManagedOwnership(t *testing.T) {
	root := t.TempDir()
	contentBundle, err := bundle.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	first, _, err := BuildProjectInitPlan(root, InitOptions{ProjectID: "example"}, contentBundle, buildinfo.Info{Version: "test"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plan.Apply(first); err != nil {
		t.Fatal(err)
	}
	identityPath := filepath.Join(root, "accounting", "company", "identity.yaml")
	if err := os.WriteFile(identityPath, []byte("schema_version: 1\nlegal_name_fa: شرکت نمونه\nnational_id: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, model, err := BuildProjectInitPlan(root, InitOptions{ProjectID: "example"}, contentBundle, buildinfo.Info{Version: "test"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range second.Operations {
		if operation.Path == "accounting/company/identity.yaml" && operation.Kind != plan.Skip {
			t.Fatalf("identity operation=%#v", operation)
		}
	}
	for _, file := range model.Managed.Files {
		if strings.HasPrefix(file.Path, "accounting/company/") {
			t.Fatalf("company file remained managed: %s", file.Path)
		}
	}
}
