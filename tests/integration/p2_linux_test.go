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

func TestRulesValidationAndVATCalculation(t *testing.T) {
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
		return cli.Execute(context.Background(), append([]string{"--root", root}, arguments...), cli.Streams{In: bytes.NewBuffer(nil), Out: &stdout, Err: &stderr, Interactive: false}, buildinfo.Info{Version: "test"})
	}
	if code := run("init", "--non-interactive", "--yes"); code != cli.ExitOK {
		t.Fatalf("init code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "rules", "vendor", "ir", "annual.json")); err != nil {
		t.Fatal(err)
	}
	if code := run("year", "init", "1405", "--non-interactive", "--yes"); code != cli.ExitOK {
		t.Fatalf("year code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	if code := run("task", "init", "vat", "--year", "1405", "--period", "Q1", "--non-interactive", "--yes"); code != cli.ExitOK {
		t.Fatalf("task code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	inputRelative := "accounting/fiscal-years/1405/work/vat-q1/templates/input.csv"
	inputPath := filepath.Join(root, filepath.FromSlash(inputRelative))
	content := "transaction_id,date_jalali,invoice_number,counterparty_national_id,net_amount_irr,vat_amount_irr,direction,tax_status,evidence_path\nS1,1405-01-01,A,1,1000000,100000,sale,taxable,a.pdf\nP1,1405-01-02,B,2,400000,40000,purchase,eligible-credit,b.pdf\n"
	if err := os.WriteFile(inputPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := run("validate", "vat", "--input", inputRelative, "--json"); code != cli.ExitOK {
		t.Fatalf("validate code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	outputRelative := "accounting/fiscal-years/1405/work/vat-q1/calculations/result.json"
	if code := run("calc", "vat", "--input", inputRelative, "--year", "1405", "--output", outputRelative, "--json"); code != cli.ExitOK {
		t.Fatalf("calc code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(outputRelative))); err != nil {
		t.Fatal(err)
	}
	if code := run("calc", "deadline", "--event", "tax-assessment-objection", "--date", "1404-12-29", "--json"); code != cli.ExitOK {
		t.Fatalf("deadline code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
}
