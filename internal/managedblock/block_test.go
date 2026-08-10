package managedblock

import (
	"errors"
	"strings"
	"testing"
)

func TestMergePreservesAndIdempotent(t *testing.T) {
	a := []byte("# user\nkeep\n")
	r, e := Merge(a, []byte("managed\n"), Markdown)
	if e != nil {
		t.Fatal(e)
	}
	r2, e := Merge(r.Content, []byte("managed\n"), Markdown)
	if e != nil {
		t.Fatal(e)
	}
	if string(r.Content) != string(r2.Content) {
		t.Fatal("not idempotent")
	}
	if !strings.Contains(string(r.Content), "keep") {
		t.Fatal("lost user content")
	}
}
func TestMergeMalformed(t *testing.T) {
	_, e := Merge([]byte("<!-- acctx:begin -->\n"), []byte("x"), Markdown)
	if !errors.Is(e, ErrConflict) {
		t.Fatalf("err=%v", e)
	}
}
