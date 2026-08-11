package workspace

import (
	"acctx/internal/buildinfo"
	bundle "acctx/internal/content"
	"acctx/internal/manifest"
	"acctx/internal/plan"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProjectInitPreservesSkillOverrideMetadata(t *testing.T) {
	root := t.TempDir()
	contentBundle, err := bundle.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	initial, _, err := BuildProjectInitPlan(root, InitOptions{ProjectID: "example"}, contentBundle, buildinfo.Info{Version: "test"}, time.Unix(1, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plan.Apply(initial); err != nil {
		t.Fatal(err)
	}
	model, err := manifest.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	for index := range model.Managed.Skills {
		if model.Managed.Skills[index].ID == "acctx-vat" {
			model.Managed.Skills[index].ActivePath = "skills/company/acctx-vat"
			model.Managed.Skills[index].Override = &manifest.Override{BasedOnVersion: model.Managed.Skills[index].Version, BasedOnDigest: model.Managed.Skills[index].Digest}
		}
	}
	manifestBytes, err := manifest.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".acctx", "manifest.yaml"), manifestBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	companySkill := filepath.Join(root, "skills", "company", "acctx-vat")
	if err := os.MkdirAll(companySkill, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, provider := range []string{".claude", ".agents"} {
		link := filepath.Join(root, provider, "skills", "acctx-vat")
		if err := os.Remove(link); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("../../skills/company/acctx-vat", link); err != nil {
			t.Fatal(err)
		}
	}

	_, upgraded, err := BuildProjectInitPlan(root, InitOptions{ProjectID: "example"}, contentBundle, buildinfo.Info{Version: "test2"}, time.Unix(2, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	for _, skill := range upgraded.Managed.Skills {
		if skill.ID == "acctx-vat" {
			if skill.Override == nil || skill.ActivePath != "skills/company/acctx-vat" {
				t.Fatalf("override lost: %#v", skill)
			}
			return
		}
	}
	t.Fatal("acctx-vat missing")
}
