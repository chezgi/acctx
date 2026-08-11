package calculator

import "testing"

func TestCalculateDeadlineRules(t *testing.T) {
	corporate, err := CalculateDeadline("corporate-tax-return", "1404-12-29")
	if err != nil || corporate.BaselineDeadlineJalali != "1405-04-29" {
		t.Fatalf("corporate=%#v err=%v", corporate, err)
	}
	vat, err := CalculateDeadline("vat-return", "1405-03-31")
	if err != nil || vat.BaselineDeadlineJalali != "1405-04-31" {
		t.Fatalf("vat=%#v err=%v", vat, err)
	}
	objection, err := CalculateDeadline("tax-assessment-objection", "1404-12-29")
	if err != nil || objection.BaselineDeadlineJalali != "1405-01-30" {
		t.Fatalf("objection=%#v err=%v", objection, err)
	}
}

func TestDeadlineEventsAreStable(t *testing.T) {
	events := DeadlineEvents()
	if len(events) != 6 || events[0].Event != "corporate-tax-return" {
		t.Fatalf("events=%#v", events)
	}
}
