package calculator

import (
	"acctx/internal/diagnostic"
	"acctx/internal/money"
	"acctx/internal/rules"
	"acctx/internal/standardcsv"
	"fmt"
)

type CorporateTaxResult struct {
	Year                    int               `json:"year"`
	RateBasisPoints         int64             `json:"rate_basis_points"`
	BookProfitIRR           int64             `json:"book_profit_irr"`
	AdditionsIRR            int64             `json:"additions_irr"`
	DeductionsIRR           int64             `json:"deductions_irr"`
	TaxableIncomeIRR        int64             `json:"taxable_income_irr"`
	GrossTaxIRR             int64             `json:"gross_tax_irr"`
	TaxCreditsIRR           int64             `json:"tax_credits_irr"`
	NetTaxPayableIRR        int64             `json:"net_tax_payable_irr"`
	IncludedAdjustments     int               `json:"included_adjustments"`
	ExcludedAdjustments     int               `json:"excluded_adjustments"`
	Valid                   bool              `json:"valid"`
	Diagnostics             []diagnostic.Item `json:"diagnostics,omitempty"`
	SourceIDs               []string          `json:"source_ids"`
	FinalVerificationNeeded bool              `json:"final_verification_required"`
}

func CalculateCorporateTax(path string, annual rules.Annual, bookProfitIRR, taxCreditsIRR int64) (CorporateTaxResult, error) {
	if taxCreditsIRR < 0 {
		return CorporateTaxResult{}, fmt.Errorf("tax credits cannot be negative")
	}
	table, report, err := standardcsv.Load("corporate-tax", path)
	if err != nil {
		return CorporateTaxResult{}, err
	}
	result := CorporateTaxResult{
		Year:                    annual.Year,
		RateBasisPoints:         annual.CorporateTaxRateBasisPoints,
		BookProfitIRR:           bookProfitIRR,
		TaxCreditsIRR:           taxCreditsIRR,
		Diagnostics:             append([]diagnostic.Item(nil), report.Diagnostics...),
		SourceIDs:               append([]string(nil), annual.SourceIDs...),
		FinalVerificationNeeded: true,
	}
	if !report.Valid {
		return result, nil
	}

	seen := map[string]bool{}
	for index, row := range table.Rows {
		rowPath := fmt.Sprintf("row:%d", index+2)
		id := row["adjustment_id"]
		if seen[id] {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_CORPORATE_DUPLICATE_ADJUSTMENT", Message: "شناسه تعدیل تکراری است", Path: rowPath + ":adjustment_id"})
			continue
		}
		seen[id] = true
		status := row["review_status"]
		if status != "approved" && status != "accepted" {
			result.ExcludedAdjustments++
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Warning, Code: "ACCTX_CORPORATE_UNAPPROVED_ADJUSTMENT", Message: "تعدیل تأیید نشده و از محاسبه خارج شد", Path: rowPath})
			continue
		}
		amount := parseIRR(row["tax_adjustment_irr"])
		switch row["direction"] {
		case "add":
			result.AdditionsIRR += amount
		case "deduct":
			result.DeductionsIRR += amount
		default:
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_CORPORATE_DIRECTION_INVALID", Message: "جهت تعدیل باید add یا deduct باشد", Path: rowPath + ":direction"})
			continue
		}
		result.IncludedAdjustments++
	}
	result.TaxableIncomeIRR = maxZero(result.BookProfitIRR + result.AdditionsIRR - result.DeductionsIRR)
	result.GrossTaxIRR, err = money.MulBasisPoints(result.TaxableIncomeIRR, result.RateBasisPoints)
	if err != nil {
		return CorporateTaxResult{}, err
	}
	result.NetTaxPayableIRR = maxZero(result.GrossTaxIRR - result.TaxCreditsIRR)
	result.Valid = !hasErrors(result.Diagnostics)
	return result, nil
}
