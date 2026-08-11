package calculator

import (
	"acctx/internal/diagnostic"
	"acctx/internal/money"
	"acctx/internal/rules"
	"acctx/internal/standardcsv"
	"fmt"
	"sort"
)

type EmployeePayrollTax struct {
	EmployeeID        string `json:"employee_id"`
	EmployeeNameFA    string `json:"employee_name_fa"`
	GrossTaxableIRR   int64  `json:"gross_taxable_irr"`
	OtherExemptionsIRR int64 `json:"other_exemptions_irr"`
	ProgressiveBaseIRR int64 `json:"progressive_base_irr"`
	CalculatedTaxIRR  int64  `json:"calculated_tax_irr"`
	TaxWithheldIRR    int64  `json:"tax_withheld_irr"`
	DifferenceIRR     int64  `json:"difference_irr"`
}

type PayrollTaxResult struct {
	Year                    int                  `json:"year"`
	Scope                   string               `json:"scope"`
	Employees               []EmployeePayrollTax `json:"employees"`
	GrossTaxableIRR         int64                `json:"gross_taxable_irr"`
	OtherExemptionsIRR      int64                `json:"other_exemptions_irr"`
	CalculatedTaxIRR        int64                `json:"calculated_tax_irr"`
	TaxWithheldIRR          int64                `json:"tax_withheld_irr"`
	DifferenceIRR           int64                `json:"difference_irr"`
	Valid                   bool                 `json:"valid"`
	Diagnostics             []diagnostic.Item    `json:"diagnostics,omitempty"`
	SourceIDs               []string             `json:"source_ids"`
	FinalVerificationNeeded bool                 `json:"final_verification_required"`
}

type employeeAggregate struct {
	name       string
	gross      int64
	exemptions int64
	withheld   int64
}

func CalculatePayrollTax(path string, annual rules.Annual) (PayrollTaxResult, error) {
	table, report, err := standardcsv.Load("payroll-tax", path)
	if err != nil {
		return PayrollTaxResult{}, err
	}
	result := PayrollTaxResult{
		Year:                    annual.Year,
		Scope:                   annual.PayrollScope,
		Diagnostics:             append([]diagnostic.Item(nil), report.Diagnostics...),
		SourceIDs:               append([]string(nil), annual.SourceIDs...),
		FinalVerificationNeeded: true,
	}
	if !report.Valid {
		return result, nil
	}

	aggregates := map[string]*employeeAggregate{}
	seenPeriod := map[string]bool{}
	for index, row := range table.Rows {
		rowPath := fmt.Sprintf("row:%d", index+2)
		key := row["employee_id"] + "\x00" + row["period"]
		if seenPeriod[key] {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_PAYROLL_DUPLICATE_EMPLOYEE_PERIOD", Message: "کارمند و دوره تکراری است", Path: rowPath})
			continue
		}
		seenPeriod[key] = true
		employee := aggregates[row["employee_id"]]
		if employee == nil {
			employee = &employeeAggregate{name: row["employee_name_fa"]}
			aggregates[row["employee_id"]] = employee
		} else if employee.name != row["employee_name_fa"] {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Warning, Code: "ACCTX_PAYROLL_EMPLOYEE_NAME_MISMATCH", Message: "نام یک شناسه کارمند در ردیف‌ها یکسان نیست", Path: rowPath + ":employee_name_fa"})
		}
		employee.gross += parseIRR(row["gross_taxable_irr"])
		employee.exemptions += parseIRR(row["exempt_amount_irr"])
		employee.withheld += parseIRR(row["tax_withheld_irr"])
	}

	brackets := make([]money.Bracket, 0, len(annual.PayrollBrackets))
	for _, bracket := range annual.PayrollBrackets {
		var upper *int64
		if bracket.UpToIRR != nil {
			value := *bracket.UpToIRR
			upper = &value
		}
		brackets = append(brackets, money.Bracket{UpToIRR: upper, RateBasisPoints: bracket.RateBasisPoints})
	}
	ids := make([]string, 0, len(aggregates))
	for id := range aggregates {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		aggregate := aggregates[id]
		if aggregate.exemptions > aggregate.gross {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Warning, Code: "ACCTX_PAYROLL_EXEMPTION_EXCEEDS_GROSS", Message: "معافیت ثبت‌شده بیش از درآمد عادی است", Path: id})
		}
		base := maxZero(aggregate.gross - aggregate.exemptions)
		calculated, err := money.Progressive(base, brackets)
		if err != nil {
			return PayrollTaxResult{}, err
		}
		item := EmployeePayrollTax{
			EmployeeID:         id,
			EmployeeNameFA:     aggregate.name,
			GrossTaxableIRR:    aggregate.gross,
			OtherExemptionsIRR: aggregate.exemptions,
			ProgressiveBaseIRR: base,
			CalculatedTaxIRR:   calculated,
			TaxWithheldIRR:     aggregate.withheld,
			DifferenceIRR:      calculated - aggregate.withheld,
		}
		result.Employees = append(result.Employees, item)
		result.GrossTaxableIRR += item.GrossTaxableIRR
		result.OtherExemptionsIRR += item.OtherExemptionsIRR
		result.CalculatedTaxIRR += item.CalculatedTaxIRR
		result.TaxWithheldIRR += item.TaxWithheldIRR
	}
	result.DifferenceIRR = result.CalculatedTaxIRR - result.TaxWithheldIRR
	result.Valid = !hasErrors(result.Diagnostics)
	return result, nil
}
