// utilities.go
// Utility functions used by handlers

package main

import (
	"strings"
)

func startsWith(str string, prefix string) bool {
	return len(str) >= len(prefix) && strings.EqualFold(str[:len(prefix)], prefix)
}

func endsWith(str string, suffix string) bool {
	return len(str) >= len(suffix) && strings.EqualFold(str[len(str)-len(suffix):], suffix)
}
