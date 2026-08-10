package content

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

func Digest(b []byte) string { s := sha256.Sum256(b); return "sha256:" + hex.EncodeToString(s[:]) }
func TreeDigest(m map[string][]byte) string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	h := sha256.New()
	for _, k := range ks {
		h.Write([]byte(k))
		h.Write([]byte{0})
		h.Write(m[k])
		h.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}
