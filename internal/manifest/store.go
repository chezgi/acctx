package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func Load(root string) (Model, error) {
	b, e := os.ReadFile(filepath.Join(root, ".acctx", "manifest.yaml"))
	if e != nil {
		return Model{}, e
	}
	var m Model
	if e := json.Unmarshal(b, &m); e != nil {
		return m, e
	}
	return m, Validate(m)
}
func Marshal(m Model) ([]byte, error) {
	sort.Slice(m.Managed.Files, func(i, j int) bool { return m.Managed.Files[i].Path < m.Managed.Files[j].Path })
	sort.Slice(m.Managed.Blocks, func(i, j int) bool { return m.Managed.Blocks[i].Path < m.Managed.Blocks[j].Path })
	sort.Slice(m.Managed.Skills, func(i, j int) bool { return m.Managed.Skills[i].ID < m.Managed.Skills[j].ID })
	return json.MarshalIndent(m, "", "  ")
}
func Validate(m Model) error {
	if m.SchemaVersion != 1 || m.Project.ID == "" || m.Project.Preset == "" {
		return fmt.Errorf("invalid manifest")
	}
	seen := map[string]bool{}
	for _, f := range m.Managed.Files {
		if filepath.IsAbs(f.Path) || seen[f.Path] {
			return fmt.Errorf("invalid managed path %q", f.Path)
		}
		seen[f.Path] = true
	}
	for _, s := range m.Managed.Skills {
		if len(s.ID) < 6 || s.ID[:6] != "acctx-" {
			return fmt.Errorf("invalid skill")
		}
	}
	return nil
}
