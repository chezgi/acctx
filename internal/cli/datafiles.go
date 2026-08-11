package cli

import (
	"acctx/internal/fsx"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func projectFile(root, value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("file path is required")
	}
	path, err := fsx.ResolveInside(root, filepath.FromSlash(value))
	if err != nil {
		return "", err
	}
	return path, nil
}

func writeJSONOutput(root, relative string, value any) error {
	if relative == "" {
		return nil
	}
	path, err := projectFile(root, relative)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, 0o644); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}
