package skills

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Metadata struct { ID string `json:"id"`; Version string `json:"version"`; Title string `json:"title"` }

func ReadMetadata(dir string) (Metadata, error) {
	f, err := os.Open(filepath.Join(dir, "SKILL.md"))
	if err != nil { return Metadata{}, err }
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() || strings.TrimSpace(s.Text()) != "---" { return Metadata{}, fmt.Errorf("missing frontmatter") }
	m := Metadata{}
	closed := false
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "---" { closed = true; break }
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 { continue }
		key := strings.TrimSpace(parts[0]); value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		switch key { case "id": m.ID=value; case "version":m.Version=value; case "title", "title_fa":m.Title=value }
	}
	if err := s.Err(); err != nil { return Metadata{}, err }
	if !closed { return Metadata{}, fmt.Errorf("unclosed frontmatter") }
	if !strings.HasPrefix(m.ID, "acctx-") || m.Version == "" || filepath.Base(dir) != m.ID { return Metadata{}, fmt.Errorf("invalid skill metadata") }
	return m, nil
}
