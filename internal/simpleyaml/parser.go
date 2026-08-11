package simpleyaml

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseFile reads top-level scalar key/value pairs from a small YAML file.
// Nested maps and sequences are intentionally ignored because company bootstrap
// validation only needs the stable top-level fields owned by acctx.
func ParseFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "-") {
			continue
		}
		if len(raw) > 0 && (raw[0] == ' ' || raw[0] == '\t') {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%s:%d: expected key: value", path, lineNumber)
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, fmt.Errorf("%s:%d: empty key", path, lineNumber)
		}
		value := strings.TrimSpace(parts[1])
		if comment := strings.Index(value, " #"); comment >= 0 {
			value = strings.TrimSpace(value[:comment])
		}
		value = strings.Trim(value, "\"'")
		if value == "null" || value == "~" {
			value = ""
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}
