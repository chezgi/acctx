package standardcsv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReportsMissingHeaderAndInvalidMoney(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vat.csv")
	content := "transaction_id,date_jalali,invoice_number,counterparty_national_id,net_amount_irr,vat_amount_irr,direction,tax_status\n1,1405-01-01,A,1,not-number,100,sale,taxable\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, report, err := Load("vat", path)
	if err != nil {
		t.Fatal(err)
	}
	if report.Valid || len(report.Diagnostics) < 2 {
		t.Fatalf("report=%#v", report)
	}
}

func TestKindsAreStable(t *testing.T) {
	kinds := Kinds()
	if len(kinds) != 4 || kinds[0] != "corporate-tax" || kinds[3] != "vat" {
		t.Fatalf("kinds=%v", kinds)
	}
}
