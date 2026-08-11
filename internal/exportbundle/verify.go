package exportbundle

import (
	"acctx/internal/diagnostic"
	"acctx/internal/evidence"
	"acctx/internal/fsx"
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
)

type Verification struct {
	Valid       bool              `json:"valid"`
	Path        string            `json:"path"`
	BundleID    string            `json:"bundle_id,omitempty"`
	BundleType  string            `json:"bundle_type,omitempty"`
	FileCount   int               `json:"file_count"`
	TotalBytes  int64             `json:"total_bytes"`
	Diagnostics []diagnostic.Item `json:"diagnostics,omitempty"`
}

func Verify(path string) (Verification, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return Verification{}, err
	}
	defer reader.Close()

	result := Verification{Path: path}
	entries := map[string]*zip.File{}
	for _, entry := range reader.File {
		name, safeErr := safeArchiveName(entry.Name)
		if safeErr != nil || name != entry.Name {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_UNSAFE_ENTRY", Message: "نام عضو آرشیو امن نیست", Path: entry.Name})
			continue
		}
		if _, duplicate := entries[name]; duplicate {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_DUPLICATE_ENTRY", Message: "عضو تکراری در آرشیو وجود دارد", Path: name})
			continue
		}
		if entry.FileInfo().Mode()&os.ModeSymlink != 0 {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_SYMLINK_ENTRY", Message: "آرشیو نباید Symlink داشته باشد", Path: name})
			continue
		}
		entries[name] = entry
	}

	manifestEntry := entries["bundle-manifest.json"]
	indexEntry := entries["evidence-index.json"]
	if manifestEntry == nil {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_MANIFEST_MISSING", Message: "Bundle manifest موجود نیست", Path: "bundle-manifest.json"})
	}
	if indexEntry == nil {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_INDEX_MISSING", Message: "Evidence index موجود نیست", Path: "evidence-index.json"})
	}
	if manifestEntry == nil || indexEntry == nil {
		result.Valid = false
		return result, nil
	}

	manifestBytes, err := readSmallEntry(manifestEntry, 10<<20)
	if err != nil {
		return Verification{}, err
	}
	indexBytes, err := readSmallEntry(indexEntry, 50<<20)
	if err != nil {
		return Verification{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_MANIFEST_INVALID", Message: err.Error(), Path: "bundle-manifest.json"})
	}
	var index evidence.Index
	if err := json.Unmarshal(indexBytes, &index); err != nil {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_INDEX_INVALID", Message: err.Error(), Path: "evidence-index.json"})
	}
	if hasDiagnosticErrors(result.Diagnostics) {
		return result, nil
	}

	result.BundleID = manifest.BundleID
	result.BundleType = manifest.BundleType
	result.FileCount = len(index.Files)
	result.TotalBytes = index.TotalBytes
	if manifest.SchemaVersion != 1 || manifest.FormatID != "acctx-controlled-bundle" || manifest.FormatVersion != "1.0.0" {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_FORMAT_INVALID", Message: "نسخه یا شناسه قالب Bundle معتبر نیست", Path: "bundle-manifest.json"})
	}
	if manifest.Status != "draft" || manifest.SubmissionPerformed || !manifest.FinalHumanReviewRequired {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_CONTROL_FLAGS_INVALID", Message: "Bundle باید Draft و نیازمند بازبینی انسان باشد", Path: "bundle-manifest.json"})
	}
	if manifest.EvidenceIndexDigest != fsx.BytesDigest(indexBytes) {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_INDEX_DIGEST_MISMATCH", Message: "Digest فهرست شواهد با Manifest برابر نیست", Path: "evidence-index.json"})
	}
	calculatedSourceDigest := evidence.SourceDigest(index.Files)
	if index.SourceDigest != calculatedSourceDigest || manifest.SourceDigest != calculatedSourceDigest {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_SOURCE_DIGEST_MISMATCH", Message: "Digest مجموعه فایل‌ها معتبر نیست", Path: "evidence-index.json"})
	}
	if manifest.FileCount != len(index.Files) || manifest.TotalBytes != index.TotalBytes {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_TOTALS_MISMATCH", Message: "جمع فایل‌ها یا اندازه‌ها با Manifest برابر نیست", Path: "bundle-manifest.json"})
	}

	expected := map[string]evidence.File{}
	for _, indexed := range index.Files {
		archiveName, safeErr := safeArchiveName("files/" + indexed.Path)
		if safeErr != nil {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_INDEX_PATH_INVALID", Message: safeErr.Error(), Path: indexed.Path})
			continue
		}
		if _, duplicate := expected[archiveName]; duplicate {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_INDEX_DUPLICATE_PATH", Message: "مسیر فایل در Index تکراری است", Path: indexed.Path})
			continue
		}
		expected[archiveName] = indexed
	}

	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)
	seen := map[string]bool{}
	for _, name := range names {
		if name == "bundle-manifest.json" || name == "evidence-index.json" {
			continue
		}
		entry := entries[name]
		indexed, ok := expected[name]
		if !ok {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_UNINDEXED_FILE", Message: "فایل آرشیو در Evidence Index ثبت نشده است", Path: name})
			continue
		}
		seen[name] = true
		digest, size, hashErr := hashEntry(entry)
		if hashErr != nil {
			return Verification{}, hashErr
		}
		if digest != indexed.SHA256 {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_FILE_DIGEST_MISMATCH", Message: "Digest فایل آرشیو معتبر نیست", Path: name})
		}
		if size != indexed.Size {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_FILE_SIZE_MISMATCH", Message: "اندازه فایل آرشیو معتبر نیست", Path: name})
		}
	}
	for name := range expected {
		if !seen[name] {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Item{Severity: diagnostic.Error, Code: "ACCTX_BUNDLE_FILE_MISSING", Message: "فایل ثبت‌شده در Index داخل آرشیو نیست", Path: name})
		}
	}
	result.Valid = !hasDiagnosticErrors(result.Diagnostics)
	return result, nil
}

func readSmallEntry(entry *zip.File, limit int64) ([]byte, error) {
	reader, err := entry.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("archive entry %s exceeds limit", entry.Name)
	}
	return data, nil
}

func hashEntry(entry *zip.File) (string, int64, error) {
	reader, err := entry.Open()
	if err != nil {
		return "", 0, err
	}
	defer reader.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, reader)
	if err != nil {
		return "", 0, err
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), size, nil
}

func hasDiagnosticErrors(items []diagnostic.Item) bool {
	for _, item := range items {
		if item.Severity == diagnostic.Error {
			return true
		}
	}
	return false
}
