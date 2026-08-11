package cli

import (
	"acctx/internal/diagnostic"
	"acctx/internal/exportbundle"
	"acctx/internal/output"
	"fmt"
	"time"
)

type exportFormat struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

func runExport(a *app, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("export subcommand required")
	}
	if args[0] == "formats" {
		return output.Write(a.streams.Out, a.json, output.Result{Command: "export formats", Status: "ok", Data: []exportFormat{
			{ID: "task", Description: "generic task bundle"},
			{ID: "audit-pack", Description: "audit task bundle with company profile and year inputs"},
			{ID: "tax-pack", Description: "tax task bundle with company profile"},
		}})
	}
	root, err := resolveProjectRoot(a)
	if err != nil {
		return err
	}
	values, flags, rest, err := parseFlags(args[1:])
	if err != nil {
		return err
	}
	if len(rest) != 0 {
		return fmt.Errorf("unexpected export arguments")
	}

	if args[0] == "verify" {
		if values["--input"] == "" {
			return fmt.Errorf("usage: acctx export verify --input <bundle.zip>")
		}
		path, err := projectFile(root, values["--input"])
		if err != nil {
			return err
		}
		verification, err := exportbundle.Verify(path)
		if err != nil {
			return err
		}
		if err := output.Write(a.streams.Out, a.json, output.Result{Command: "export verify", Status: "ok", Data: verification, Diagnostics: verification.Diagnostics}); err != nil {
			return err
		}
		if !verification.Valid {
			return &ExitError{Code: ExitValidation, Diagnostic: diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_VERIFICATION_FAILED", Message: "اعتبار Bundle تأیید نشد"}}
		}
		return nil
	}

	if err := a.linuxMutating(); err != nil {
		return err
	}
	if values["--year"] == "" || values["--task"] == "" || values["--output"] == "" {
		return fmt.Errorf("--year, --task, and --output are required")
	}
	options := exportbundle.Options{
		FiscalYear:        values["--year"],
		TaskID:            values["--task"],
		OutputPath:        values["--output"],
		Force:             flags["--force"],
		IncludeCompany:    flags["--include-company"],
		IncludeYearInputs: flags["--include-year-inputs"],
		GeneratedAt:       time.Now().UTC(),
		CLIVersion:        a.build.Version,
	}
	switch args[0] {
	case "task":
		options.BundleType = exportbundle.TypeTask
	case "audit-pack":
		options.BundleType = exportbundle.TypeAuditPack
		options.IncludeCompany = true
		options.IncludeYearInputs = true
	case "tax-pack":
		options.BundleType = exportbundle.TypeTaxPack
		options.IncludeCompany = true
	default:
		return fmt.Errorf("unknown export subcommand")
	}
	result, err := exportbundle.WriteTask(root, options)
	if err != nil {
		return err
	}
	return output.Write(a.streams.Out, a.json, output.Result{Command: "export " + args[0], Status: "ok", Data: result})
}
