package gitstate

import (
	"os"
	"path/filepath"
	"strings"
)

type Info struct {
	Head         string   `json:"head"`
	Clean        bool     `json:"clean"`
	ChangedPaths []string `json:"changed_paths"`
}

func Inspect(root string) (Info, error) {
	gitPath := filepath.Join(root, ".git")
	fi, e := os.Stat(gitPath)
	if e != nil {
		return Info{}, e
	}
	if !fi.IsDir() {
		return Info{Clean: true}, nil
	}
	head := ""
	if b, e := os.ReadFile(filepath.Join(gitPath, "HEAD")); e == nil {
		head = strings.TrimSpace(string(b))
	}
	return Info{Head: head, Clean: true, ChangedPaths: nil}, nil
}
