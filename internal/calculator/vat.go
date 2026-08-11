package calculator

import (
	"acctx/internal/diagnostic"
	"acctx/internal/money"
	"acctx/internal/rules"
	"acctx/internal/standardcsv"
	"fmt"
)

type VATResult struct {
	Year                    int               `json:"year"`
	RateBasisPoints         int64             `json:"rate_basis_points"`
	Rows                    int               `json:"rows"`
	SalesNetIRR             int64             `json:"sales_net_irr"`
	OutputVATIRR            int64             `json:"output_vat_irr"`
	PurchaseNetIRR          int64             `json:"purchase_net_irr"`
	EligibleInputVATIRR     int64             `json:"eligible_input_vat_irr"`
	NetVATPayableIRR        int64             `json:"net_vat_payable_irr"`
	Valid                   bool              `json:"valid"`
	Diagnostics             []diagnostic.Item `json:"diagnostics,omitempty"`
	SourceIDs               []string          `json:"source_ids"`
	FinalVerificationNeeded bool              `json:"final_verification_required"`
}

func CalculateVAT(path string, annual rules.Annual) (VATResult, error) {
	table, report, err := standardcsv.Load("vat", path)
	if err != nil {
		return VATResult{}, err
	}
	result := VATResult{
		Year:                    annual.Year,
		RateBasisPoints:         annual.VATRateBasisPoints,
		Rows:                    len(table.Rows),
		Diagnostics:             append([]diagnostic.Item(nil), report.Diagnostics...),
		SourceIDs:               append([]string(nil), annual.SourceIDs...),
		FinalVerificationNeeded: true,
	}
	if !report.Valid {
		result.Valid = false
		return result, nil
	}

	seen := map[string]bool{}
	for index, row := range table.Rows {
		rowPath := fmt.Sprintf("row:%d", index+2)
		transactionID := row["transaction_id"]
		if seen[transactionID] {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_VAT_DUPLICATE_TRANSACTION", Message: "شناسه تراکنش تکراری است", Path: rowPath + ":transaction_id"})
			continue
		}
		seen[transactionID] = true
		netAmount := parseIRR(row["net_amount_irr"])
		vatAmount := parseIRR(row["vat_amount_irr"])
		direction := row["direction"]
		status := row["tax_status"]

		switch direction {
		case "sale":
			result.SalesNetIRR += netAmount
			switch status {
			case "taxable":
				result.OutputVATIRR += vatAmount
				appendVATRateDiagnostic(&result, rowPath, netAmount, vatAmount)
			case "exempt", "out-of-scope":
				if vatAmount != 0 {
					result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_VAT_NON_TAXABLE_WITH_VAT", Message: "برای فروش معاف یا خارج از شمول مالیات درج شده است", Path: rowPath})
				}
			case "unknown":
				result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_VAT_STATUS_UNKNOWN", Message: "وضعیت مالیاتی فروش تعیین نشده است", Path: rowPath})
			default:
				result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_VAT_STATUS_INVALID", Message: "وضعیت مالیاتی فروش معتبر نیست", Path: rowPath + ":tax_status"})
			}
		case "purchase":
			result.PurchaseNetIRR += netAmount
			switch status {
			case "eligible-credit":
				result.EligibleInputVATIRR += vatAmount
				appendVATRateDiagnostic(&result, rowPath, netAmount, vatAmount)
			case "noncreditable":
				// The VAT remains a cost and is intentionally excluded from credit.
			case "exempt", "out-of-scope":
				if vatAmount != 0 {
					result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_VAT_NON_TAXABLE_WITH_VAT", Message: "برای خرید معاف یا خارج از شمول مالیات درج شده است", Path: rowPath})
				}
			case "unknown":
				result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_VAT_STATUS_UNKNOWN", Message: "وضعیت اعتبار خرید تعیین نشده است", Path: rowPath})
			default:
				result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_VAT_STATUS_INVALID", Message: "وضعیت مالیاتی خرید معتبر نیست", Path: rowPath + ":tax_status"})
			}
		default:
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_VAT_DIRECTION_INVALID", Message: "جهت تراکنش باید sale یا purchase باشد", Path: rowPath + ":direction"})
		}
	}
	result.NetVATPayableIRR = result.OutputVATIRR - result.EligibleInputVATIRR
	result.Valid = !hasErrors(result.Diagnostics)
	return result, nil
}

func appendVATRateDiagnostic(result *VATResult, rowPath string, netAmount, declaredVAT int64) {
	expected, err := money.MulBasisPoints(netAmount, result.RateBasisPoints)
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_VAT_RATE_CALCULATION_FAILED", Message: err.Error(), Path: rowPath})
		return
	}
	if expected != declaredVAT {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Warning, Code: "ACCTX_VAT_RATE_MISMATCH", Message: fmt.Sprintf("مالیات درج‌شده %d ریال و مالیات حاصل از نرخ عمومی %d ریال است", declaredVAT, expected), Path: rowPath})
	}
}
