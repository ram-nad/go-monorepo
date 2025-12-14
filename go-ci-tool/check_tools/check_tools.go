// Package checktools contains code for checking installed tools and versions
package checktools

import (
	"fmt"
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/ram-nad/go-monorepo/go-ci-tool/v2/color"
	"github.com/ram-nad/go-monorepo/go-ci-tool/v2/constants"
	customerrors "github.com/ram-nad/go-monorepo/go-ci-tool/v2/custom_errors"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

const (
	GO           = "go"
	GoLangCILint = "golangci-lint"
)

func checkIfInstalled(tool string) bool {
	str, err := exec.LookPath(tool)

	if err != nil {
		color.Printf(
			color.ErrorColor,
			"Required Tool %s is not installed in the system\n",
			tool,
		)
		return false
	} else {
		color.Printf(color.InfoColor, "Required Tool %s is installed at %s\n", tool, str)
		return true
	}
}

func checkGoInstalled() (_ bool, err error) {
	minSupportedGoVersion := constants.MinSupportedGoVersion()

	isInstalled := checkIfInstalled(GO)

	if !isInstalled {
		return false, nil
	}

	goVersionString, err := exec.Command(GO, "version").Output()
	if err != nil {
		err = fmt.Errorf("error while checking go version: %s", err.Error())
		return
	}

	if !utf8.Valid(goVersionString) {
		err = fmt.Errorf(
			"got invalid utf8 string while checking go version: %q",
			goVersionString,
		)
		return
	}

	goVersion := strings.Replace(
		strings.Split(string(goVersionString), " ")[2],
		GO,
		"",
		1,
	)

	if !semver.IsValid("v" + goVersion) {
		err = fmt.Errorf(
			"got invalid version string while checking go version: %q",
			goVersionString,
		)
		return
	}

	if semver.Compare(goVersion, minSupportedGoVersion) < 0 {
		color.Printf(
			color.ErrorColorBold,
			"Installed Go Version %s is lower than required version %s. Upgrade your go installation.\n",
			goVersion,
			minSupportedGoVersion,
		)
		return false, nil
	}

	color.Printf(
		color.InfoColor,
		"Minimum Supported Go Version: %s\n",
		minSupportedGoVersion,
	)
	color.Printf(color.InfoColor, "Installed Go Version: %s\n", goVersion)

	color.Printf(color.SuccessColorBold, "Go is installed :)\n")
	return true, nil
}

func checkGoCILintInstalled() (_ bool, err error) {
	golangCILintVersion := constants.GolangCILintVersion()

	isInstalled := checkIfInstalled(GoLangCILint)

	if !isInstalled {
		return false, nil
	}

	versionString, err := exec.Command(GoLangCILint, "version", "--short").Output()
	if err != nil {
		err = fmt.Errorf("error while checking golangci-lint version: %s", err.Error())
		return
	}

	if !utf8.Valid(versionString) {
		err = fmt.Errorf(
			"got invalid utf8 string while checking golangci-lint version: %q",
			versionString,
		)
		return
	}

	version := strings.Trim(string(versionString), " \n")

	if !semver.IsValid("v" + version) {
		err = fmt.Errorf(
			"got invalid version string while checking golangci-lint version: %q",
			version,
		)
		return
	}

	cmpResult := semver.Compare(version, golangCILintVersion)
	var action string

	if cmpResult < 0 {
		action = "Upgrade"
	} else if cmpResult > 0 {
		action = "Downgrade"
	}

	if cmpResult != 0 {
		color.Printf(
			color.ErrorColorBold,
			"Installed golangci-lint version %s. We require it to be %s. %s your golangci-lint installation.\n",
			version,
			golangCILintVersion,
			action,
		)
		return
	}

	color.Printf(
		color.SuccessColorBold,
		"GoLang CI Lint v%s is installed :)\n",
		golangCILintVersion,
	)
	return true, nil
}

func GetCheckInstallationCommand() *cobra.Command {
	checkInstallationCommand := &cobra.Command{
		Use: "check-tools",
		RunE: func(_ *cobra.Command, _ []string) error {
			isGoInstalled, err := checkGoInstalled()
			if err != nil {
				return err
			}

			color.Println(color.NoColor)

			isLintToolInstalled, err := checkGoCILintInstalled()
			if err != nil {
				return err
			}

			if !isGoInstalled || !isLintToolInstalled {
				return customerrors.NewErrNoLog()
			} else {
				return nil
			}
		},
		Args: cobra.NoArgs,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Short:                 "Check if required tools are installed in the system",
		Long:                  "Checks if the required tools are installed in the system with correct versions. Currently Checks for Go (https://go.dev/) and golangci-lint (https://golangci-lint.run/).",
		SilenceErrors:         true,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
	}

	return checkInstallationCommand
}
