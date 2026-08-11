package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

// SourceDigest returns a stable digest of the sorted project-relative file
// metadata. Generated timestamps are deliberately excluded.
func SourceDigest(files []File) string {
	ordered := append([]File(nil), files...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Path < ordered[j].Path })
	hash := sha256.New()
	for _, file := range ordered {
		fmt.Fprintf(hash, "%s\x00%s\x00%d\x00%s\x00", file.Path, file.SHA256, file.Size, file.Category)
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil))
}
