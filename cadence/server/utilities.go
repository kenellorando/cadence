// utilities.go
// Utility functions used by handlers

package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/kenellorando/clog"
)

func startsWith(str string, prefix string) bool {
	return len(str) >= len(prefix) && strings.EqualFold(str[:len(prefix)], prefix)
}

func endsWith(str string, suffix string) bool {
	return len(str) >= len(suffix) && strings.EqualFold(str[len(str)-len(suffix):], suffix)
}

func tokenCheck(token string) bool {
	clog.Info("ARIA2Check", fmt.Sprintf("Checking token %s...", token))

	if len(token) != 26 {
		clog.Debug("ARIA2Check", fmt.Sprintf("Token %s does not satisfy length requirements.", token))
		return false
	}

	// Check the whitelist. If this fails, the whitelist is not configured. No panic is thrown, but the bypass is denied.
	b, err := ioutil.ReadFile(c.WhitelistPath)
	if err != nil {
		return false
	}
	s := string(b)

	if strings.Contains(s, token) {
		clog.Info("ARIA2Check", fmt.Sprintf("Token %s is valid.", token))
		return true
	}
	clog.Info("ARIA2Check", fmt.Sprintf("Token %s is invalid.", token))
	return false
}
