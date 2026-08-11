package standardcsv

import (
	"acctx/internal/diagnostic"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Schema struct {
	Kind            string
	RequiredHeaders []string
	IntegerHeaders  []string
}

type Table struct {
	Headers []string            `json:"headers"`
	Rows    []map[string]string `json:"rows"`
}

type Report struct {
	Kind        string            `json:"kind"`
	Path        string            `json:"path"`
	Rows        int               `json:"rows"`
	Valid       bool              `json:"valid"`
	Diagnostics []diagnostic.Item `json:"diagnostics,omitempty"`
}

var schemas = map[string]Schema{
	"vat": {
		Kind: "vat",
		RequiredHeaders: []string{"transaction_id", "date_jalali", "invoice_number", "counterparty_national_id", "net_amount_irr", "vat_amount_irr", "direction", "tax_status", "evidence_path"},
		IntegerHeaders:  []string{"net_amount_irr", "vat_amount_irr"},
	},
	"corporate-tax": {
		Kind: "corporate-tax",
		RequiredHeaders: []string{"adjustment_id", "account_or_source", "description", "book_amount_irr", "tax_adjustment_irr", "direction", "legal_source_id", "evidence_path", "review_status"},
		IntegerHeaders:  []string{"book_amount_irr", "tax_adjustment_irr"},
	},
	"payroll-tax": {
		Kind: "payroll-tax",
		RequiredHeaders: []string{"employee_id", "employee_name_fa", "period", "gross_taxable_irr", "exempt_amount_irr", "tax_withheld_irr", "evidence_path"},
		IntegerHeaders:  []string{"gross_taxable_irr", "exempt_amount_irr", "tax_withheld_irr"},
	},
	"tax-defense": {
		Kind: "tax-defense",
		RequiredHeaders: []string{"issue_id", "issue_title", "disputed_amount_irr", "authority_position", "company_response", "legal_source_id", "evidence_path", "status"},
		IntegerHeaders:  []string{"disputed_amount_irr"},
	},
}

func Kinds() []string {
	kinds := make([]string, 0, len(schemas))
	for kind := range schemas {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds
}

func SchemaFor(kind string) (Schema, bool) {
	schema, ok := schemas[kind]
	return schema, ok
}

func Load(kind, path string) (Table, Report, error) {
	schema, ok := SchemaFor(kind)
	if !ok {
		return Table{}, Report{}, fmt.Errorf("unknown standard CSV kind %q", kind)
	}
	file, err := os.Open(path)
	if err != nil {
		return Table{}, Report{}, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	headers, err := reader.Read()
	if err == io.EOF {
		report := Report{Kind: kind, Path: path, Valid: false, Diagnostics: []diagnostic.Item{{Severity: diagnostic.Error, Code: "ACCTX_CSV_EMPTY", Message: "فایل CSV خالی است", Path: path}}}
		return Table{}, report, nil
	}
	if err != nil {
		return Table{}, Report{}, err
	}
	for index := range headers {
		headers[index] = strings.TrimSpace(headers[index])
	}

	diagnostics := []diagnostic.Item{}
	positions := map[string]int{}
	for index, header := range headers {
		if header == "" {
			diagnostics = append(diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_CSV_EMPTY_HEADER", Message: "نام ستون خالی است", Path: fmt.Sprintf("column:%d", index+1)})
			continue
		}
		if _, exists := positions[header]; exists {
			diagnostics = append(diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_CSV_DUPLICATE_HEADER", Message: "نام ستون تکراری است", Path: header})
			continue
		}
		positions[header] = index
	}
	for _, header := range schema.RequiredHeaders {
		if _, exists := positions[header]; !exists {
			diagnostics = append(diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_CSV_REQUIRED_HEADER_MISSING", Message: "ستون الزامی موجود نیست", Path: header})
		}
	}

	rows := []map[string]string{}
	rowNumber := 1
	for {
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		rowNumber++
		if readErr != nil {
			return Table{}, Report{}, fmt.Errorf("CSV row %d: %w", rowNumber, readErr)
		}
		if len(record) != len(headers) {
			diagnostics = append(diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_CSV_FIELD_COUNT", Message: "تعداد مقادیر ردیف با ستون‌ها برابر نیست", Path: fmt.Sprintf("row:%d", rowNumber)})
			continue
		}
		row := map[string]string{}
		empty := true
		for index, header := range headers {
			value := strings.TrimSpace(record[index])
			if value != "" {
				empty = false
			}
			row[header] = value
		}
		if empty {
			continue
		}
		for _, header := range schema.RequiredHeaders {
			if row[header] == "" {
				diagnostics = append(diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_CSV_REQUIRED_VALUE_MISSING", Message: "مقدار الزامی خالی است", Path: fmt.Sprintf("row:%d:%s", rowNumber, header)})
			}
		}
		for _, header := range schema.IntegerHeaders {
			value := row[header]
			if value == "" {
				continue
			}
			parsed, parseErr := strconv.ParseInt(value, 10, 64)
			if parseErr != nil || parsed < 0 {
				diagnostics = append(diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_CSV_INVALID_IRR", Message: "مبلغ باید عدد صحیح ریالی و غیرمنفی باشد", Path: fmt.Sprintf("row:%d:%s", rowNumber, header)})
			}
		}
		rows = append(rows, row)
	}
	return Table{Headers: headers, Rows: rows}, Report{Kind: kind, Path: path, Rows: len(rows), Valid: len(diagnostics) == 0, Diagnostics: diagnostics}, nil
}
