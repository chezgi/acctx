package money

import "fmt"

type Bracket struct {
	UpToIRR        *int64
	RateBasisPoints int64
}

func MulBasisPoints(amountIRR, basisPoints int64) (int64, error) {
	if amountIRR < 0 {
		return 0, fmt.Errorf("negative amount is not supported")
	}
	if basisPoints < 0 || basisPoints > 10000 {
		return 0, fmt.Errorf("basis points must be between 0 and 10000")
	}
	whole := (amountIRR / 10000) * basisPoints
	remainderProduct := (amountIRR % 10000) * basisPoints
	return whole + (remainderProduct+5000)/10000, nil
}

func Progressive(amountIRR int64, brackets []Bracket) (int64, error) {
	if amountIRR < 0 {
		return 0, fmt.Errorf("negative amount is not supported")
	}
	if len(brackets) == 0 {
		return 0, fmt.Errorf("at least one bracket is required")
	}
	var lower int64
	var tax int64
	for index, bracket := range brackets {
		if bracket.RateBasisPoints < 0 || bracket.RateBasisPoints > 10000 {
			return 0, fmt.Errorf("invalid bracket rate")
		}
		upper := amountIRR
		if bracket.UpToIRR != nil {
			if *bracket.UpToIRR <= lower {
				return 0, fmt.Errorf("brackets must be strictly increasing")
			}
			if upper > *bracket.UpToIRR {
				upper = *bracket.UpToIRR
			}
		} else if index != len(brackets)-1 {
			return 0, fmt.Errorf("open bracket must be last")
		}
		if upper > lower {
			piece, err := MulBasisPoints(upper-lower, bracket.RateBasisPoints)
			if err != nil {
				return 0, err
			}
			tax += piece
		}
		if amountIRR <= upper {
			return tax, nil
		}
		if bracket.UpToIRR != nil {
			lower = *bracket.UpToIRR
		}
	}
	return tax, nil
}
