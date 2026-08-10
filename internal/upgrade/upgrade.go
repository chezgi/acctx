package upgrade

import (
	bundle "acctx/internal/content"
	"acctx/internal/manifest"
	"acctx/internal/plan"
	"time"
)

func Build(root string, b bundle.Bundle, now time.Time) (plan.Plan, error) {
	m, e := manifest.Load(root)
	if e != nil { return plan.Plan{}, e }
	if m.Generator.ContentVersion == b.Catalog.ContentVersion {
		return plan.New("project upgrade", root, now, []plan.Operation{{Kind: plan.Skip, Path: ".acctx/manifest.yaml", Message: "content already current"}}), nil
	}
	return plan.New("project upgrade", root, now, []plan.Operation{{Kind: plan.Conflict, Path: ".acctx/manifest.yaml", Message: "upgrade between content versions requires migration support"}}), nil
}
