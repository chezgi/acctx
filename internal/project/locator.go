package project

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrGitNotFound = errors.New("git repository not found")
var ErrProjectNotFound = errors.New("acctx project not found")

func find(start string, pred func(string) bool, notFound error) (string, error) {
	a, e := filepath.Abs(start)
	if e != nil {
		return "", e
	}
	if i, e := os.Stat(a); e == nil && !i.IsDir() {
		a = filepath.Dir(a)
	}
	for {
		if pred(a) {
			return a, nil
		}
		p := filepath.Dir(a)
		if p == a {
			return "", notFound
		}
		a = p
	}
}
func GitRoot(start string) (string, error) {
	return find(start, func(d string) bool { _, e := os.Lstat(filepath.Join(d, ".git")); return e == nil }, ErrGitNotFound)
}
func Root(start string) (string, error) {
	return find(start, func(d string) bool {
		i, e := os.Stat(filepath.Join(d, ".acctx", "manifest.yaml"))
		return e == nil && !i.IsDir()
	}, ErrProjectNotFound)
}
