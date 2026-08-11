package cli

import (
	"acctx/internal/diagnostic"
	"acctx/internal/output"
	"acctx/internal/standardcsv"
	"fmt"
)

func runValidate(a *app, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: acctx validate <kind> --input <path>")
	}
	if args[0] == "kinds" {
		return output.Write(a.streams.Out, a.json, output.Result{Command: "validate kinds", Status: "ok", Data: standardcsv.Kinds()})
	}
	root, err := resolveProjectRoot(a)
	if err != nil {
		return err
	}
	values, _, rest, err := parseFlags(args[1:])
	if err != nil {
		return err
	}
	if len(rest) != 0 || values["--input"] == "" {
		return fmt.Errorf("usage: acctx validate <kind> --input <path>")
	}
	path, err := projectFile(root, values["--input"])
	if err != nil {
		return err
	}
	_, report, err := standardcsv.Load(args[0], path)
	if err != nil {
		return err
	}
	if err := output.Write(a.streams.Out, a.json, output.Result{Command: "validate " + args[0], Status: "ok", Data: report, Diagnostics: report.Diagnostics}); err != nil {
		return err
	}
	if !report.Valid {
		return &ExitError{Code: ExitValidation, Diagnostic: diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_STANDARD_INPUT_INVALID", Message: "فایل استاندارد ورودی معتبر نیست"}}
	}
	return nil
}
