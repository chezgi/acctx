package calculator

import (
	"acctx/internal/rules"
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateVATIncludesOnlyEligiblePurchaseCredit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vat.csv")
	content := "transaction_id,date_jalali,invoice_number,counterparty_national_id,net_amount_irr,vat_amount_irr,direction,tax_status,evidence_path\n" +
		"S1,1405-01-01,A,1,1000000,100000,sale,taxable,a.pdf\n" +
		"P1,1405-01-02,B,2,400000,40000,purchase,eligible-credit,b.pdf\n" +
		"P2,1405-01-03,C,3,200000,20000,purchase,noncreditable,c.pdf\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := CalculateVAT(path, rules.Annual{Year: 1405, VATRateBasisPoints: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Valid || result.OutputVATIRR != 100000 || result.EligibleInputVATIRR != 40000 || result.NetVATPayableIRR != 60000 {
		t.Fatalf("result=%#v", result)
	}
}

func TestCalculateVATFlagsUnknownStatus(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vat.csv")
	content := "transaction_id,date_jalali,invoice_number,counterparty_national_id,net_amount_irr,vat_amount_irr,direction,tax_status,evidence_path\nS1,1405-01-01,A,1,100,0,sale,unknown,a\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := CalculateVAT(path, rules.Annual{Year: 1405, VATRateBasisPoints: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid {
		t.Fatal("unknown status must invalidate the result")
	}
}
