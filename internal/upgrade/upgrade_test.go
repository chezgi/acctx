package upgrade

import (
	"acctx/internal/buildinfo"
	bundle "acctx/internal/content"
	"acctx/internal/fsx"
	"acctx/internal/manifest"
	"acctx/internal/plan"
	"acctx/internal/workspace"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBuildUpdatesUnmodifiedVendorContentAndPreservesProjectData(t *testing.T) {
	root := t.TempDir()
	contentBundle, err := bundle.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	initial, _, err := workspace.BuildProjectInitPlan(root, workspace.InitOptions{ProjectID: "example"}, contentBundle, buildinfo.Info{Version: "old-cli"}, time.Unix(1, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plan.Apply(initial); err != nil {
		t.Fatal(err)
	}

	vendorPath := filepath.Join(root, "workflows", "vendor", "monthly-close.md")
	oldContent := []byte("old vendor workflow\n")
	if err := os.WriteFile(vendorPath, oldContent, 0o644); err != nil {
		t.Fatal(err)
	}
	model, err := manifest.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	model.Generator.ContentVersion = "0.3.0"
	for index := range model.Managed.Files {
		if model.Managed.Files[index].Path == "workflows/vendor/monthly-close.md" {
			model.Managed.Files[index].Digest = fsx.BytesDigest(oldContent)
		}
	}
	manifestBytes, err := manifest.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".acctx", "manifest.yaml"), manifestBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	identityPath := filepath.Join(root, "accounting", "company", "identity.yaml")
	identity := []byte("schema_version: 1\nlegal_name_fa: شرکت نمونه\nnational_id: 1\n")
	if err := os.WriteFile(identityPath, identity, 0o644); err != nil {
		t.Fatal(err)
	}

	upgradePlan, err := Build(root, contentBundle, buildinfo.Info{Version: "new-cli"}, time.Unix(2, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if upgradePlan.HasConflicts() {
		t.Fatalf("upgrade plan has conflicts: %#v", upgradePlan)
	}
	foundUpdate := false
	for _, operation := range upgradePlan.Operations {
		if operation.Path == "workflows/vendor/monthly-close.md" && operation.Kind == plan.Update {
			foundUpdate = true
		}
	}
	if !foundUpdate {
		t.Fatal("expected vendor workflow update")
	}
	if _, err := plan.Apply(upgradePlan); err != nil {
		t.Fatal(err)
	}
	actualIdentity, err := os.ReadFile(identityPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(actualIdentity) != string(identity) {
		t.Fatal("company-owned identity file changed during upgrade")
	}
}
