package cli

import (
	"acctx/internal/calculator"
	"acctx/internal/diagnostic"
	"acctx/internal/output"
	"acctx/internal/rules"
	"fmt"
	"strconv"
)

func runCalc(a *app, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("calculator subcommand required")
	}
	if args[0] == "deadline-events" {
		return output.Write(a.streams.Out, a.json, output.Result{Command: "calc deadline-events", Status: "ok", Data: calculator.DeadlineEvents()})
	}
	root, err := resolveProjectRoot(a)
	if err != nil {
		return err
	}
	values, _, rest, err := parseFlags(args[1:])
	if err != nil {
		return err
	}
	if len(rest) != 0 {
		return fmt.Errorf("unexpected calculator arguments")
	}

	switch args[0] {
	case "deadline":
		if values["--event"] == "" || values["--date"] == "" {
			return fmt.Errorf("usage: acctx calc deadline --event <id> --date <jalali>")
		}
		result, err := calculator.CalculateDeadline(values["--event"], values["--date"])
		if err != nil {
			return err
		}
		if err := writeJSONOutput(root, values["--output"], result); err != nil {
			return err
		}
		return output.Write(a.streams.Out, a.json, output.Result{Command: "calc deadline", Status: "ok", Data: result})
	case "vat", "corporate-tax", "payroll-tax":
		if values["--input"] == "" || values["--year"] == "" {
			return fmt.Errorf("--input and --year are required")
		}
		inputPath, err := projectFile(root, values["--input"])
		if err != nil {
			return err
		}
		year, err := strconv.Atoi(values["--year"])
		if err != nil {
			return fmt.Errorf("invalid --year: %w", err)
		}
		ruleSet, err := rules.Load(a.bundle)
		if err != nil {
			return err
		}
		annual, err := ruleSet.Year(year)
		if err != nil {
			return err
		}
		switch args[0] {
		case "vat":
			result, err := calculator.CalculateVAT(inputPath, annual)
			if err != nil {
				return err
			}
			return emitCalculation(a, root, "calc vat", values["--output"], result, result.Valid, result.Diagnostics)
		case "payroll-tax":
			result, err := calculator.CalculatePayrollTax(inputPath, annual)
			if err != nil {
				return err
			}
			return emitCalculation(a, root, "calc payroll-tax", values["--output"], result, result.Valid, result.Diagnostics)
		case "corporate-tax":
			bookProfit, err := requiredInt64(values, "--book-profit")
			if err != nil {
				return err
			}
			credits, err := optionalInt64(values, "--tax-credits")
			if err != nil {
				return err
			}
			result, err := calculator.CalculateCorporateTax(inputPath, annual, bookProfit, credits)
			if err != nil {
				return err
			}
			return emitCalculation(a, root, "calc corporate-tax", values["--output"], result, result.Valid, result.Diagnostics)
		}
	}
	return fmt.Errorf("unknown calculator %q", args[0])
}

func emitCalculation(a *app, root, command, outputPath string, value any, valid bool, diagnostics []diagnostic.Item) error {
	if err := writeJSONOutput(root, outputPath, value); err != nil {
		return err
	}
	if err := output.Write(a.streams.Out, a.json, output.Result{Command: command, Status: "ok", Data: value, Diagnostics: diagnostics}); err != nil {
		return err
	}
	if !valid {
		return &ExitError{Code: ExitValidation, Diagnostic: diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_CALCULATION_INPUT_INVALID", Message: "محاسبه دارای خطای ورودی یا طبقه‌بندی حل‌نشده است"}}
	}
	return nil
}

func requiredInt64(values map[string]string, name string) (int64, error) {
	if values[name] == "" {
		return 0, fmt.Errorf("%s is required", name)
	}
	return parseCLIInt64(values[name], name)
}

func optionalInt64(values map[string]string, name string) (int64, error) {
	if values[name] == "" {
		return 0, nil
	}
	return parseCLIInt64(values[name], name)
}

func parseCLIInt64(value, name string) (int64, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	return parsed, nil
}
