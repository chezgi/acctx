package fsx

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ResolveInside(root, rel string) (string, error) {
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("absolute path")
	}
	p := filepath.Join(root, filepath.Clean(rel))
	r, e := filepath.Rel(root, p)
	if e != nil {
		return "", e
	}
	if r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root")
	}
	return p, nil
}
