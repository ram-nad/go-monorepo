// Package constants contains helpers for global constants
package constants

import (
	"os"

	"golang.org/x/mod/semver"
)

var minGoVersion = ""
var minGolangCILintVersion = ""

func MinSupportedGoVersion() string {
	envValue := os.Getenv("MIN_SUPPORTED_GO_VERSION")

	if envValue != "" && semver.IsValid("v"+envValue) {
		return envValue
	} else {
		return minGoVersion
	}
}

func GolangCILintVersion() string {
	envValue := os.Getenv("GOLANGCI_LINT_VERSION")

	if envValue != "" && semver.IsValid("v"+envValue) {
		return envValue
	} else {
		return minGolangCILintVersion
	}
}
