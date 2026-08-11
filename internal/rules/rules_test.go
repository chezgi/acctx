package rules

import (
	bundle "acctx/internal/content"
	"testing"
)

func TestLoadContainsEverySupportedYear(t *testing.T) {
	contentBundle, err := bundle.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	set, err := Load(contentBundle)
	if err != nil {
		t.Fatal(err)
	}
	years := set.Years()
	if len(years) != 9 || years[0] != 1397 || years[len(years)-1] != 1405 {
		t.Fatalf("years=%v", years)
	}
	annual, err := set.Year(1403)
	if err != nil {
		t.Fatal(err)
	}
	if annual.VATRateBasisPoints != 1000 {
		t.Fatalf("VAT rate=%d", annual.VATRateBasisPoints)
	}
}
