package calculator

import (
	"acctx/internal/diagnostic"
	"strconv"
)

func parseIRR(value string) int64 {
	parsed, _ := strconv.ParseInt(value, 10, 64)
	return parsed
}

func hasErrors(items []diagnostic.Item) bool {
	for _, item := range items {
		if item.Severity == diagnostic.Error {
			return true
		}
	}
	return false
}

func maxZero(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}
