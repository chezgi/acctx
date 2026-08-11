package evidence

import "testing"

func TestSourceDigestIgnoresInputOrder(t *testing.T) {
	files := []File{{Path: "b", SHA256: "sha256:2", Size: 2, Category: "input"}, {Path: "a", SHA256: "sha256:1", Size: 1, Category: "input"}}
	first := SourceDigest(files)
	second := SourceDigest([]File{files[1], files[0]})
	if first != second {
		t.Fatalf("first=%s second=%s", first, second)
	}
}
