package calculator

import (
	"acctx/internal/rules"
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateCorporateTaxUsesApprovedAdjustments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "adjustments.csv")
	content := "adjustment_id,account_or_source,description,book_amount_irr,tax_adjustment_irr,direction,legal_source_id,evidence_path,review_status\n" +
		"A1,expense,non-deductible,100,200000000,add,S1,a.pdf,approved\n" +
		"A2,income,benefit,100,100000000,deduct,S2,b.pdf,accepted\n" +
		"A3,expense,draft,100,900000000,add,S3,c.pdf,draft\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := CalculateCorporateTax(path, rules.Annual{Year: 1405, CorporateTaxRateBasisPoints: 2500}, 1000000000, 50000000)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Valid || result.TaxableIncomeIRR != 1100000000 || result.GrossTaxIRR != 275000000 || result.NetTaxPayableIRR != 225000000 || result.ExcludedAdjustments != 1 {
		t.Fatalf("result=%#v", result)
	}
}
