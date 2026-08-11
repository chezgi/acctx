//go:build linux

package integration_test

import (
	"acctx/internal/buildinfo"
	"acctx/internal/cli"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCreatesCoreVendorContentAndRelativeSkillLinks(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	code := cli.Execute(
		context.Background(),
		[]string{"--root", root, "init", "--non-interactive", "--yes"},
		cli.Streams{In: bytes.NewBuffer(nil), Out: &out, Err: &errOut, Interactive: false},
		buildinfo.Info{Version: "test"},
	)
	if code != 0 {
		t.Fatalf("code=%d out=%s err=%s", code, out.String(), errOut.String())
	}
	for _, path := range []string{
		"skills/vendor/acctx-vat/SKILL.md",
		"skills/vendor/acctx-tax-defense/SKILL.md",
		"workflows/vendor/quarterly-vat.md",
		"templates/vendor/tasks/vat/input.csv",
		"references/vendor/legal-source-register.yaml",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
	for _, path := range []string{
		".claude/skills/acctx-vat",
		".agents/skills/acctx-tax-defense",
	} {
		target, err := os.Readlink(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		if filepath.IsAbs(target) {
			t.Fatalf("absolute link %s", target)
		}
	}
}
