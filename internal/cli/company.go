package cli

import (
	"acctx/internal/company"
	"acctx/internal/diagnostic"
	"acctx/internal/output"
	"fmt"
)

func runCompany(a *app, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("company subcommand required")
	}
	root, err := resolveProjectRoot(a)
	if err != nil {
		return err
	}
	switch args[0] {
	case "status":
		status, err := company.ReadStatus(root)
		if err != nil {
			return err
		}
		return output.Write(a.streams.Out, a.json, output.Result{Command: "company status", Status: "ok", Data: status})
	case "validate":
		values, _, _, err := parseFlags(args[1:])
		if err != nil {
			return err
		}
		profile := values["--profile"]
		if profile == "" {
			profile = "bootstrap"
		}
		report, err := company.Validate(root, profile)
		if err != nil {
			return err
		}
		if err := output.Write(a.streams.Out, a.json, output.Result{Command: "company validate", Status: "ok", Data: report, Diagnostics: report.Diagnostics}); err != nil {
			return err
		}
		if !report.Valid {
			return &ExitError{
				Code: ExitValidation,
				Diagnostic: diagnostic.Item{
					Severity: diagnostic.Error,
					Code:     "ACCTX_COMPANY_VALIDATION_FAILED",
					Message:  "مشخصات شرکت برای پروفایل انتخاب‌شده کامل نیست",
				},
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown company subcommand")
	}
}
