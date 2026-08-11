package rules

import (
	bundle "acctx/internal/content"
	"encoding/json"
	"fmt"
	"sort"
)

type Bracket struct {
	UpToIRR        *int64 `json:"up_to_irr,omitempty"`
	RateBasisPoints int64 `json:"rate_basis_points"`
}

type Annual struct {
	Year                         int       `json:"year"`
	VATRateBasisPoints           int64     `json:"vat_rate_basis_points"`
	CorporateTaxRateBasisPoints  int64     `json:"corporate_tax_rate_basis_points"`
	PayrollScope                 string    `json:"payroll_scope"`
	PayrollBrackets              []Bracket `json:"payroll_brackets"`
	SourceIDs                    []string  `json:"source_ids"`
}

type document struct {
	SchemaVersion int      `json:"schema_version"`
	Jurisdiction  string   `json:"jurisdiction"`
	Years         []Annual `json:"years"`
}

type Set struct {
	years map[int]Annual
}

func Load(contentBundle bundle.Bundle) (Set, error) {
	data, err := contentBundle.Read("rules/ir/annual.json")
	if err != nil {
		return Set{}, err
	}
	var value document
	if err := json.Unmarshal(data, &value); err != nil {
		return Set{}, err
	}
	if value.SchemaVersion != 1 || value.Jurisdiction != "IR" {
		return Set{}, fmt.Errorf("unsupported annual rule document")
	}
	set := Set{years: map[int]Annual{}}
	for _, annual := range value.Years {
		if annual.Year < 1397 || annual.VATRateBasisPoints <= 0 || annual.CorporateTaxRateBasisPoints <= 0 || len(annual.PayrollBrackets) < 2 {
			return Set{}, fmt.Errorf("invalid rules for year %d", annual.Year)
		}
		if _, exists := set.years[annual.Year]; exists {
			return Set{}, fmt.Errorf("duplicate rules for year %d", annual.Year)
		}
		var previous int64
		for index, bracket := range annual.PayrollBrackets {
			if bracket.RateBasisPoints < 0 || bracket.RateBasisPoints > 10000 {
				return Set{}, fmt.Errorf("invalid payroll rate for year %d", annual.Year)
			}
			if bracket.UpToIRR == nil {
				if index != len(annual.PayrollBrackets)-1 {
					return Set{}, fmt.Errorf("open payroll bracket must be last for year %d", annual.Year)
				}
				continue
			}
			if *bracket.UpToIRR <= previous {
				return Set{}, fmt.Errorf("payroll brackets are not increasing for year %d", annual.Year)
			}
			previous = *bracket.UpToIRR
		}
		set.years[annual.Year] = annual
	}
	return set, nil
}

func (set Set) Year(year int) (Annual, error) {
	annual, ok := set.years[year]
	if !ok {
		return Annual{}, fmt.Errorf("no embedded Iranian rule set for year %d", year)
	}
	return annual, nil
}

func (set Set) Years() []int {
	years := make([]int, 0, len(set.years))
	for year := range set.years {
		years = append(years, year)
	}
	sort.Ints(years)
	return years
}
