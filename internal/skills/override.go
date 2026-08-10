package skills

import (
	"acctx/internal/fsx"
	"acctx/internal/manifest"
	"acctx/internal/plan"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (s Service) OverridePlan(id string, now time.Time) (plan.Plan, error) {
	m, e := manifest.Load(s.Root)
	if e != nil { return plan.Plan{}, e }
	idx := -1
	for i := range m.Managed.Skills { if m.Managed.Skills[i].ID == id { idx = i } }
	if idx < 0 { return plan.Plan{}, fmt.Errorf("skill not found") }
	x := m.Managed.Skills[idx]
	if x.Override != nil { return plan.Plan{}, fmt.Errorf("already overridden") }
	company := "skills/company/" + id
	if _, e := os.Stat(filepath.Join(s.Root, company)); e == nil {
		return plan.New("skill override", s.Root, now, []plan.Operation{{Kind: plan.Conflict, Path: company, Message: "company skill already exists"}}), nil
	}
	ops := []plan.Operation{}
	vendorAbs := filepath.Join(s.Root, x.VendorPath)
	e = filepath.WalkDir(vendorAbs, func(p string, d os.DirEntry, e error) error {
		if e != nil { return e }
		if d.IsDir() { return nil }
		r, _ := filepath.Rel(vendorAbs, p)
		b, e := os.ReadFile(p); if e != nil { return e }
		ops = append(ops, plan.Operation{Kind: plan.Create, Path: filepath.ToSlash(filepath.Join(company, r)), Payload: b, AfterDigest: fsx.BytesDigest(b)})
		return nil
	})
	if e != nil { return plan.Plan{}, e }
	meta := map[string]any{"schema_version": 1, "skill_id": id, "override_type": "full-replacement", "source": map[string]string{"version": x.Version, "digest": x.Digest}}
	mb, _ := json.MarshalIndent(meta, "", "  ")
	ops = append(ops, plan.Operation{Kind: plan.Create, Path: company + "/.acctx-override.yaml", Payload: mb, AfterDigest: fsx.BytesDigest(mb)})
	links := map[string]string{}
	for _, p := range []string{".claude/skills", ".agents/skills"} {
		lp := p + "/" + id
		target := "../../" + company
		ops = append(ops, plan.Operation{Kind: plan.Link, Path: lp, Target: target})
		links[p] = target
	}
	x.ActivePath = company
	x.ProviderLinks = links
	x.Override = &manifest.Override{BasedOnVersion: x.Version, BasedOnDigest: x.Digest}
	m.Managed.Skills[idx] = x
	newm, _ := manifest.Marshal(m)
	oldm, _ := os.ReadFile(filepath.Join(s.Root, ".acctx/manifest.yaml"))
	ops = append(ops, plan.Operation{Kind: plan.Update, Path: ".acctx/manifest.yaml", BeforeDigest: fsx.BytesDigest(oldm), AfterDigest: fsx.BytesDigest(newm), Payload: newm})
	return plan.New("skill override", s.Root, now, ops), nil
}

func (s Service) Diff(id string) (string, error) {
	m, e := manifest.Load(s.Root)
	if e != nil { return "", e }
	var x *manifest.Skill
	for i := range m.Managed.Skills { if m.Managed.Skills[i].ID == id { x = &m.Managed.Skills[i] } }
	if x == nil || x.Override == nil { return "", fmt.Errorf("skill has no override") }
	v, _ := os.ReadFile(filepath.Join(s.Root, x.VendorPath, "SKILL.md"))
	c, _ := os.ReadFile(filepath.Join(s.Root, x.ActivePath, "SKILL.md"))
	if string(v) == string(c) { return "no differences\n", nil }
	var out strings.Builder
	out.WriteString("--- vendor/SKILL.md\n+++ company/SKILL.md\n")
	vl := strings.Split(string(v), "\n"); cl := strings.Split(string(c), "\n"); n := len(vl); if len(cl) > n { n = len(cl) }
	for i := 0; i < n; i++ {
		var a, b string
		if i < len(vl) { a = vl[i] }
		if i < len(cl) { b = cl[i] }
		if a != b { if a != "" { out.WriteString("-" + a + "\n") }; if b != "" { out.WriteString("+" + b + "\n") } }
	}
	return out.String(), nil
}

func (s Service) ResetPlan(id string, now time.Time, force bool) (plan.Plan, error) {
	m, e := manifest.Load(s.Root)
	if e != nil { return plan.Plan{}, e }
	idx := -1
	for i := range m.Managed.Skills { if m.Managed.Skills[i].ID == id { idx = i } }
	if idx < 0 || m.Managed.Skills[idx].Override == nil { return plan.Plan{}, fmt.Errorf("skill has no override") }
	x := m.Managed.Skills[idx]
	if !force {
		vd, _ := treeDigest(filepath.Join(s.Root, x.VendorPath)); cd, _ := treeDigest(filepath.Join(s.Root, x.ActivePath))
		if vd != cd { return plan.New("skill reset", s.Root, now, []plan.Operation{{Kind: plan.Conflict, Path: x.ActivePath, Message: "ACCTX_SKILL_OVERRIDE_MODIFIED"}}), nil }
	}
	ops := []plan.Operation{{Kind: plan.Delete, Path: x.ActivePath}}
	links := map[string]string{}
	for _, p := range []string{".claude/skills", ".agents/skills"} {
		lp := p + "/" + id; t := "../../" + x.VendorPath
		ops = append(ops, plan.Operation{Kind: plan.Link, Path: lp, Target: t}); links[p] = t
	}
	x.ActivePath = x.VendorPath; x.ProviderLinks = links; x.Override = nil; m.Managed.Skills[idx] = x
	nm, _ := manifest.Marshal(m); om, _ := os.ReadFile(filepath.Join(s.Root, ".acctx/manifest.yaml"))
	ops = append(ops, plan.Operation{Kind: plan.Update, Path: ".acctx/manifest.yaml", BeforeDigest: fsx.BytesDigest(om), AfterDigest: fsx.BytesDigest(nm), Payload: nm})
	return plan.New("skill reset", s.Root, now, ops), nil
}

func (s Service) AdoptPlan(id string, now time.Time) (plan.Plan, error) {
	st, e := s.Status(id)
	if e != nil { return plan.Plan{}, e }
	if !st.Valid { return plan.New("skill adopt", s.Root, now, []plan.Operation{{Kind: plan.Conflict, Path: id, Message: "existing skill links are invalid"}}), nil }
	return plan.New("skill adopt", s.Root, now, []plan.Operation{{Kind: plan.Skip, Path: id, Message: "already valid"}}), nil
}
