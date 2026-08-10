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

func Embedded() (Bundle, error) {
	sub, err := fs.Sub(builtin.FS, "assets")
	if err != nil {
		return Bundle{}, err
	}
	c := Catalog{SchemaVersion: 1, ContentVersion: "0.1.0", SupportedYears: []int{1397, 1398, 1399, 1400, 1401, 1402, 1403, 1404, 1405}, DefaultPreset: "ir-software-kb-techpark", Skills: []Skill{{ID: "acctx-workspace", Version: "0.1.0", Path: "skills/acctx-workspace"}, {ID: "acctx-source-review", Version: "0.1.0", Path: "skills/acctx-source-review"}}}
	return Bundle{fs: sub, Catalog: c}, nil
}
func (b Bundle) Read(rel string) ([]byte, error) {
	clean := path.Clean(rel)
	if clean == "." || clean == ".." || clean[0] == '/' || (len(clean) >= 3 && clean[:3] == "../") {
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
		v, e := fs.ReadFile(b.fs, p)
		if e != nil {
			return e
		}
		r := strings.TrimPrefix(p, strings.TrimSuffix(rel, "/")+"/")
		out[r] = v
		return nil
	})
	return out, err
}
func (c Catalog) Skill(id string) (Skill, bool) {
	for _, s := range c.Skills {
		if s.ID == id {
			return s, true
		}
	}
	return Skill{}, false
}
