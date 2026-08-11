package content

import "testing"

func TestEmbeddedCatalog(t *testing.T) {
	bundle, err := Embedded()
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Catalog.ContentVersion != "0.4.0" {
		t.Fatalf("content version=%s", bundle.Catalog.ContentVersion)
	}
	if bundle.Catalog.SupportedYears[0] != 1397 || bundle.Catalog.SupportedYears[len(bundle.Catalog.SupportedYears)-1] != 1405 {
		t.Fatalf("years=%v", bundle.Catalog.SupportedYears)
	}
	if got, want := len(bundle.Catalog.Skills), 28; got != want {
		t.Fatalf("skill count=%d want=%d", got, want)
	}
	if _, ok := bundle.Catalog.Skill("acctx-evidence-bundle"); !ok {
		t.Fatal("acctx-evidence-bundle missing from catalog")
	}
	if files, err := bundle.ReadTree("workflows"); err != nil || len(files) < 6 {
		t.Fatalf("workflows=%d err=%v", len(files), err)
	}
	if rules, err := bundle.Read("rules/ir/annual.json"); err != nil || len(rules) == 0 {
		t.Fatalf("rules bytes=%d err=%v", len(rules), err)
	}
}
