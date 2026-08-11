//go:build linux

package integration_test

import (
	"acctx/internal/buildinfo"
	"acctx/internal/cli"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestControlledExportAndUpgradeWorkflow(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	run := func(arguments ...string) int {
		stdout.Reset()
		stderr.Reset()
		return cli.Execute(
			context.Background(),
			append([]string{"--root", root}, arguments...),
			cli.Streams{In: bytes.NewBuffer(nil), Out: &stdout, Err: &stderr, Interactive: false},
			buildinfo.Info{Version: "test"},
		)
	}

	if code := run("init", "--non-interactive", "--yes"); code != cli.ExitOK {
		t.Fatalf("init code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Lstat(filepath.Join(root, ".claude", "skills", "acctx-evidence-bundle")); err != nil {
		t.Fatal(err)
	}
	if code := run("year", "init", "1405", "--non-interactive", "--yes"); code != cli.ExitOK {
		t.Fatalf("year code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	if code := run("task", "init", "tax-defense", "--year", "1405", "--non-interactive", "--yes"); code != cli.ExitOK {
		t.Fatalf("task code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	inputRelative := "accounting/fiscal-years/1405/work/tax-defense/inputs/assessment.txt"
	inputPath := filepath.Join(root, filepath.FromSlash(inputRelative))
	if err := os.WriteFile(inputPath, []byte("assessment evidence\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	indexRelative := "accounting/fiscal-years/1405/work/tax-defense/calculations/evidence-index.json"
	if code := run("evidence", "index", "--year", "1405", "--task", "tax-defense", "--include-company", "--output", indexRelative, "--json"); code != cli.ExitOK {
		t.Fatalf("evidence code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	bundleRelative := "accounting/fiscal-years/1405/outputs/tax-defense.zip"
	if code := run("export", "tax-pack", "--year", "1405", "--task", "tax-defense", "--output", bundleRelative, "--json"); code != cli.ExitOK {
		t.Fatalf("export code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	if code := run("export", "verify", "--input", bundleRelative, "--json"); code != cli.ExitOK {
		t.Fatalf("verify code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(bundleRelative))); err != nil {
		t.Fatal(err)
	}

	manifestPath := filepath.Join(root, ".acctx", "manifest.yaml")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	oldManifest := strings.Replace(string(manifestBytes), `"content_version": "0.4.0"`, `"content_version": "0.3.0"`, 1)
	if oldManifest == string(manifestBytes) {
		t.Fatal("content version marker not found")
	}
	if err := os.WriteFile(manifestPath, []byte(oldManifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := run("project", "upgrade", "--non-interactive", "--yes"); code != cli.ExitOK {
		t.Fatalf("upgrade code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	upgraded, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(upgraded), `"content_version": "0.4.0"`) {
		t.Fatal("manifest was not upgraded")
	}
}
