package content

import (
	builtin "acctx/content"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

type Skill struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

type Catalog struct {
	SchemaVersion  int     `json:"schema_version"`
	ContentVersion string  `json:"content_version"`
	SupportedYears []int   `json:"supported_years"`
	DefaultPreset  string  `json:"default_preset"`
	Skills         []Skill `json:"skills"`
}

type Bundle struct {
	fs      fs.FS
	Catalog Catalog
}

func coreSkillIDs() []string {
	return []string{
		"acctx-workspace", "acctx-source-review", "acctx-company-profile",
		"acctx-fiscal-year", "acctx-task-workspace", "acctx-software-contract-review",
		"acctx-revenue-recognition", "acctx-bank-reconciliation", "acctx-sales-review",
		"acctx-purchase-review", "acctx-taxpayer-system", "acctx-vat",
		"acctx-corporate-income-tax", "acctx-payroll-tax", "acctx-social-security",
		"acctx-rental-tax", "acctx-electronic-books", "acctx-article-169",
		"acctx-knowledge-based", "acctx-technology-park", "acctx-rd-tax-credit",
		"acctx-financial-statements", "acctx-audit-preparation", "acctx-tax-defense",
		"acctx-social-security-defense", "acctx-compliance-calendar", "acctx-legal-update",
	}
}

func Embedded() (Bundle, error) {
	sub, err := fs.Sub(builtin.FS, "assets")
	if err != nil {
		return Bundle{}, err
	}
	skills := make([]Skill, 0, len(coreSkillIDs()))
	for _, id := range coreSkillIDs() {
		skills = append(skills, Skill{ID: id, Version: "0.2.0", Path: "skills/" + id})
	}
	catalog := Catalog{
		SchemaVersion:  1,
		ContentVersion: "0.3.0",
		SupportedYears: []int{1397, 1398, 1399, 1400, 1401, 1402, 1403, 1404, 1405},
		DefaultPreset:  "ir-software-kb-techpark",
		Skills:         skills,
	}
	return Bundle{fs: sub, Catalog: catalog}, nil
}

func (b Bundle) Read(rel string) ([]byte, error) {
	clean := path.Clean(rel)
	if clean == "." || clean == ".." || clean[0] == '/' || strings.HasPrefix(clean, "../") {
		return nil, fmt.Errorf("invalid asset path")
	}
	return fs.ReadFile(b.fs, clean)
}

func (b Bundle) ReadTree(rel string) (map[string][]byte, error) {
	out := map[string][]byte{}
	err := fs.WalkDir(b.fs, rel, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		value, readErr := fs.ReadFile(b.fs, p)
		if readErr != nil {
			return readErr
		}
		name := strings.TrimPrefix(p, strings.TrimSuffix(rel, "/")+"/")
		out[name] = value
		return nil
	})
	return out, err
}

func (c Catalog) Skill(id string) (Skill, bool) {
	for _, skill := range c.Skills {
		if skill.ID == id {
			return skill, true
		}
	}
	return Skill{}, false
}
