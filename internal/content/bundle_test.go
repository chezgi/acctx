package content

import "testing"

func TestEmbeddedCatalog(t *testing.T) {
	b, e := Embedded()
	if e != nil {
		t.Fatal(e)
	}
	if b.Catalog.SupportedYears[0] != 1397 || b.Catalog.SupportedYears[len(b.Catalog.SupportedYears)-1] != 1405 {
		t.Fatalf("years=%v", b.Catalog.SupportedYears)
	}
	if len(b.Catalog.Skills) != 2 {
		t.Fatalf("skills=%v", b.Catalog.Skills)
	}
}
