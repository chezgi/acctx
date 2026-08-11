package upgrade

import (
	"acctx/internal/buildinfo"
	bundle "acctx/internal/content"
	"acctx/internal/manifest"
	"acctx/internal/plan"
	"acctx/internal/workspace"
	"time"
)

// Build creates an explicit, drift-aware content upgrade plan. It reuses the
// same materialization rules as init, so vendor files update only when their
// current digest matches the digest recorded by the previous manifest.
func Build(root string, contentBundle bundle.Bundle, build buildinfo.Info, now time.Time) (plan.Plan, error) {
	current, err := manifest.Load(root)
	if err != nil {
		return plan.Plan{}, err
	}
	initPlan, _, err := workspace.BuildProjectInitPlan(root, workspace.InitOptions{
		ProjectID: current.Project.ID,
		Preset:    current.Project.Preset,
	}, contentBundle, build, now)
	if err != nil {
		return plan.Plan{}, err
	}
	return plan.New("project upgrade", root, now, initPlan.Operations), nil
}
