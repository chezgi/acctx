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

func TestCompanyYearAndTaskWorkflow(t *testing.T) {
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
	if code := run("company", "validate", "--profile", "bootstrap"); code != cli.ExitValidation {
		t.Fatalf("company validation code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}

	identityPath := filepath.Join(root, "accounting", "company", "identity.yaml")
	identity := "schema_version: 1\nlegal_name_fa: \"شرکت نمونه\"\nnational_id: \"12345678901\"\nregistration_number: \"1234\"\nlegal_type: \"private-joint-stock\"\n"
	if err := os.WriteFile(identityPath, []byte(identity), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := run("init", "--non-interactive", "--yes"); code != cli.ExitOK {
		t.Fatalf("re-init code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	if code := run("company", "validate", "--profile", "bootstrap"); code != cli.ExitOK {
		t.Fatalf("company validation code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}

	if code := run("year", "init", "1399", "--historical", "--non-interactive", "--yes"); code != cli.ExitOK {
		t.Fatalf("year init code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	if code := run("task", "init", "vat", "--year", "1399", "--period", "Q1", "--non-interactive", "--yes"); code != cli.ExitOK {
		t.Fatalf("task init code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}

	for _, relative := range []string{
		"accounting/fiscal-years/1399/year.yaml",
		"accounting/fiscal-years/1399/work/vat-q1/task.yaml",
		"accounting/fiscal-years/1399/work/vat-q1/templates/input.csv",
		"accounting/fiscal-years/1399/work/vat-q1/checklist.md",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(relative))); err != nil {
			t.Fatalf("missing %s: %v", relative, err)
		}
	}

	if code := run("project", "doctor", "--json"); code != cli.ExitOK {
		t.Fatalf("doctor code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
}
