package calculator

import (
	"acctx/internal/rules"
	"os"
	"path/filepath"
	"testing"
)

func payrollUpper(value int64) *int64 { return &value }

func TestCalculatePayrollTaxAggregatesEmployeeRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payroll.csv")
	content := "employee_id,employee_name_fa,period,gross_taxable_irr,exempt_amount_irr,tax_withheld_irr,evidence_path\n" +
		"E1,علی,01,5000000000,0,10000000,a\n" +
		"E1,علی,02,5000000000,0,10000000,b\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	annual := rules.Annual{Year: 1405, PayrollScope: "ordinary", PayrollBrackets: []rules.Bracket{
		{UpToIRR: payrollUpper(4800000000), RateBasisPoints: 0},
		{UpToIRR: payrollUpper(9600000000), RateBasisPoints: 1000},
		{UpToIRR: payrollUpper(12000000000), RateBasisPoints: 1500},
		{RateBasisPoints: 3000},
	}}
	result, err := CalculatePayrollTax(path, annual)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Valid || len(result.Employees) != 1 || result.CalculatedTaxIRR != 540000000 || result.DifferenceIRR != 520000000 {
		t.Fatalf("result=%#v", result)
	}
}
