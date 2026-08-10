package project

import (
	"acctx/internal/gitstate"
	"acctx/internal/manifest"
)

type Status struct {
	Root string `json:"root"`
	ProjectID string `json:"project_id"`
	Preset string `json:"preset"`
	CLIVersion string `json:"cli_version"`
	ContentVersion string `json:"content_version"`
	ManagedFiles int `json:"managed_files"`
	ManagedBlocks int `json:"managed_blocks"`
	ManagedSkills int `json:"managed_skills"`
	SkillOverrides int `json:"skill_overrides"`
	Git gitstate.Info `json:"git"`
}

func ReadStatus(root string) (Status, error) {
	m, e := manifest.Load(root)
	if e != nil { return Status{}, e }
	g, e := gitstate.Inspect(root)
	if e != nil { return Status{}, e }
	o := 0
	for _, s := range m.Managed.Skills { if s.Override != nil { o++ } }
	return Status{root, m.Project.ID, m.Project.Preset, m.Generator.CLIVersion, m.Generator.ContentVersion, len(m.Managed.Files), len(m.Managed.Blocks), len(m.Managed.Skills), o, g}, nil
}
