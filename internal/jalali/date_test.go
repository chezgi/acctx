package jalali

import "testing"

func TestLeapAndMonthLengths(t *testing.T) {
	if !IsLeap(1399) || IsLeap(1400) {
		t.Fatal("unexpected Jalali leap-year calculation")
	}
	if DaysInMonth(1399, 12) != 30 || DaysInMonth(1400, 12) != 29 {
		t.Fatal("unexpected Esfand length")
	}
}

func TestDateArithmetic(t *testing.T) {
	date, err := Parse("1404-12-29")
	if err != nil {
		t.Fatal(err)
	}
	plusDays, err := date.AddDays(30)
	if err != nil || plusDays.String() != "1405-01-30" {
		t.Fatalf("plusDays=%s err=%v", plusDays, err)
	}
	plusMonths, err := date.AddMonths(4)
	if err != nil || plusMonths.String() != "1405-04-29" {
		t.Fatalf("plusMonths=%s err=%v", plusMonths, err)
	}
	quarterEnd, _ := Parse("1405-03-31")
	deadline, err := quarterEnd.EndOfMonthAfter(1)
	if err != nil || deadline.String() != "1405-04-31" {
		t.Fatalf("deadline=%s err=%v", deadline, err)
	}
}
