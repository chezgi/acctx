package money

import "testing"

func ptr(value int64) *int64 { return &value }

func TestMulBasisPointsRoundsHalfUp(t *testing.T) {
	value, err := MulBasisPoints(15, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if value != 2 {
		t.Fatalf("value=%d", value)
	}
}

func TestProgressivePayrollExample1405(t *testing.T) {
	brackets := []Bracket{
		{UpToIRR: ptr(4800000000), RateBasisPoints: 0},
		{UpToIRR: ptr(9600000000), RateBasisPoints: 1000},
		{UpToIRR: ptr(12000000000), RateBasisPoints: 1500},
		{UpToIRR: ptr(14400000000), RateBasisPoints: 2000},
		{UpToIRR: ptr(16800000000), RateBasisPoints: 2500},
		{RateBasisPoints: 3000},
	}
	tax, err := Progressive(10000000000, brackets)
	if err != nil {
		t.Fatal(err)
	}
	// 4.8b at 10% plus 0.4b at 15%.
	if tax != 540000000 {
		t.Fatalf("tax=%d", tax)
	}
}
