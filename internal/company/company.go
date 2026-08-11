package company

import (
	"acctx/internal/diagnostic"
	"acctx/internal/simpleyaml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Status struct {
	Root          string            `json:"root"`
	Files         map[string]bool   `json:"files"`
	Values        map[string]string `json:"values"`
	MissingFiles  []string          `json:"missing_files,omitempty"`
	Populated     int               `json:"populated_fields"`
	KnownFields   int               `json:"known_fields"`
}

type Report struct {
	Profile     string            `json:"profile"`
	Valid       bool              `json:"valid"`
	Diagnostics []diagnostic.Item `json:"diagnostics,omitempty"`
}

var profileRequirements = map[string][]string{
	"bootstrap": {
		"identity.legal_name_fa",
		"identity.national_id",
	},
	"complete": {
		"identity.legal_name_fa",
		"identity.national_id",
		"identity.registration_number",
		"identity.legal_type",
		"registrations.economic_number",
		"contacts.registered_address_fa",
		"contacts.postal_code",
	},
	"tax-ready": {
		"identity.legal_name_fa",
		"identity.national_id",
		"identity.registration_number",
		"registrations.economic_number",
		"registrations.taxpayer_unit_code",
		"tax.fiscal_year_default",
	},
	"vat-ready": {
		"identity.legal_name_fa",
		"identity.national_id",
		"registrations.economic_number",
		"tax.fiscal_year_default",
		"tax.vat_registered",
	},
}

var profileFiles = []string{"identity", "registrations", "tax", "contacts", "bank-accounts", "knowledge-based", "technology-park"}

func ReadStatus(root string) (Status, error) {
	status := Status{
		Root:   filepath.Join(root, "accounting", "company"),
		Files:  map[string]bool{},
		Values: map[string]string{},
	}
	for _, name := range profileFiles {
		relative := filepath.ToSlash(filepath.Join("accounting", "company", name+".yaml"))
		path := filepath.Join(root, filepath.FromSlash(relative))
		values, err := simpleyaml.ParseFile(path)
		if os.IsNotExist(err) {
			status.Files[name] = false
			status.MissingFiles = append(status.MissingFiles, relative)
			continue
		}
		if err != nil {
			return Status{}, err
		}
		status.Files[name] = true
		for key, value := range values {
			qualified := name + "." + key
			status.Values[qualified] = value
			status.KnownFields++
			if value != "" {
				status.Populated++
			}
		}
	}
	sort.Strings(status.MissingFiles)
	return status, nil
}

func Validate(root, profile string) (Report, error) {
	required, ok := profileRequirements[profile]
	if !ok {
		return Report{}, fmt.Errorf("unknown company validation profile %q", profile)
	}
	status, err := ReadStatus(root)
	if err != nil {
		return Report{}, err
	}
	diagnostics := []diagnostic.Item{}
	for _, path := range status.MissingFiles {
		diagnostics = append(diagnostics, diagnostic.Item{
			Severity: diagnostic.Error,
			Code:     "ACCTX_COMPANY_FILE_MISSING",
			Message:  "فایل مشخصات شرکت موجود نیست",
			Path:     path,
		})
	}
	for _, field := range required {
		value, exists := status.Values[field]
		if !exists || value == "" {
			diagnostics = append(diagnostics, diagnostic.Item{
				Severity: diagnostic.Error,
				Code:     "ACCTX_COMPANY_REQUIRED_FIELD_MISSING",
				Message:  "فیلد الزامی مشخصات شرکت تکمیل نشده است",
				Path:     field,
			})
		}
	}
	sort.Slice(diagnostics, func(i, j int) bool {
		if diagnostics[i].Code == diagnostics[j].Code {
			return diagnostics[i].Path < diagnostics[j].Path
		}
		return diagnostics[i].Code < diagnostics[j].Code
	})
	return Report{Profile: profile, Valid: len(diagnostics) == 0, Diagnostics: diagnostics}, nil
}
