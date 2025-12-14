// Package modules contains code for working with Go modules
package modules

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/ram-nad/go-monorepo/go-ci-tool/v2/color"
	"github.com/ram-nad/go-monorepo/go-ci-tool/v2/constants"
	customerrors "github.com/ram-nad/go-monorepo/go-ci-tool/v2/custom_errors"
	formattestjson "github.com/ram-nad/go-monorepo/go-ci-tool/v2/format_testjson"
	"golang.org/x/mod/semver"
)

const (
	AllModulesPath = "./..."
	GolangCILint   = "golangci-lint"
	GO             = "go"
	GoWorkOff      = "GOWORK=off"
)

func CheckMinVersionSupported(details ModuleDetails) error {
	minSupportedGoVersion := constants.MinSupportedGoVersion()

	color.Printf(color.InfoColor,
		"Go module: %s is using go version %s\n",
		details.Module,
		details.GoVersion,
	)

	c := semver.Compare("v"+details.GoVersion, "v"+minSupportedGoVersion)

	if c > 0 {
		color.Printf(
			color.ErrorColor,
			"Go version of module %s is higher than the minimum supported Go version %s.\n",
			details.Module,
			minSupportedGoVersion,
		)
		return customerrors.NewErrNoLog()
	}

	return nil
}

// CheckReplaceIsNotLocal checks if a module uses a replace
// directive with local path, which is fine for testing
// but not for production code.
func CheckReplaceIsNotLocal(details ModuleDetails) error {
	color.Printf(
		color.InfoColor,
		"Checking Go module: %s for replaces with local path\n",
		details.Module,
	)

	valid := true

	for _, info := range details.Replaces {
		if info.NewPath == "." || info.NewPath == ".." ||
			strings.HasPrefix(info.NewPath, "./") ||
			strings.HasPrefix(info.NewPath, "../") {
			color.Printf(color.ErrorColor,
				"Go module %s is using replace directive with local path '%s'.\n",
				details.Module,
				info.NewPath,
			)
			valid = false
		}
	}

	if !valid {
		return customerrors.NewErrNoLog()
	} else {
		color.Printf(color.SuccessColorBold, "Go module %s is not using any local replaces\n", details.Module)
		return nil
	}
}

func CheckModuleTidy(details ModuleDetails) error {
	color.Println(color.InfoColor, "go mod tidy -diff")

	//nolint:gosec // details.ModulePath is not a user input
	cmd := exec.Command(GO, "-C", details.ModulePath, "mod", "tidy", "-diff")
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	out := bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	// Command Failed to run
	if cmd.ProcessState == nil {
		err = fmt.Errorf(
			"error while running 'go mod tidy -diff' for module %s, error: %s",
			details.Module,
			err.Error(),
		)
		return err
	} else {
		if out.Len() > 0 {
			color.Println(color.NoColor)
			color.Print(color.MutedColor, out.String())
			color.Println(color.NoColor)
		}

		if cmd.ProcessState.ExitCode() != 0 {
			color.Printf(color.ErrorColorBold, "Go module %s is not tidy. Run 'go mod tidy'\n", details.Module)

			return customerrors.NewErrNoLog()
		} else {
			color.Printf(color.SuccessColorBold, "Go module %s is tidy :)\n", details.Module)
			return nil
		}
	}
}

func RunModuleTidy(details ModuleDetails) error {
	color.Println(color.InfoColor, "go mod tidy")

	//nolint:gosec // details.ModulePath is not a user input
	cmd := exec.Command(GO, "-C", details.ModulePath, "mod", "tidy")
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	out := bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	// Command failed to run
	if cmd.ProcessState == nil {
		return fmt.Errorf(
			"error while running 'go mod tidy' for module %s, error: %s",
			details.Module,
			err.Error(),
		)
	} else {
		if out.Len() > 0 {
			color.Println(color.NoColor)
			color.Print(color.MutedColor, out.String())
			color.Println(color.NoColor)
		}

		if cmd.ProcessState.ExitCode() != 0 {
			return fmt.Errorf("'go mod tidy' failed for module %s", details.Module)
		} else {
			color.Printf(color.SuccessColorBold, "Go module %s is now tidy.\n", details.Module)
			return nil
		}
	}
}

func RunGolangCILintFmt(details ModuleDetails) error {
	color.Println(color.InfoColor, "golanlangci-lint fmt ./...")

	args := []string{"fmt", AllModulesPath}

	cmd := exec.Command(GolangCILint, args...)
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	out := bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	// Command failed to run
	if cmd.ProcessState == nil {
		return fmt.Errorf(
			"error while running 'golangci-lint fmt ./...' for module %s, error: %s",
			details.Module,
			err.Error(),
		)
	} else {
		if out.Len() > 0 {
			color.Println(color.NoColor)
			color.Print(color.MutedColor, out.String())
			color.Println(color.NoColor)
		}

		if cmd.ProcessState.ExitCode() != 0 {
			return fmt.Errorf("'golangci-lint fmt ./...' failed for module %s", details.Module)
		} else {
			color.Printf(color.SuccessColorBold, "Code for module: %s has been formatted.\n", details.Module)
			return nil
		}
	}
}

func RunGolangCILintFix(details ModuleDetails) error {
	color.Println(color.InfoColor, "golanlangci-lint run --fix ./...")

	args := []string{"run", "--fix", AllModulesPath}

	cmd := exec.Command(GolangCILint, args...)
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	out := bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	// Command failed to run
	if cmd.ProcessState == nil {
		return fmt.Errorf(
			"error while running 'golangci-lint run --fix ./...' for module %s, error: %s",
			details.Module,
			err.Error(),
		)
	} else {
		if out.Len() > 0 {
			color.Println(color.NoColor)
			color.Print(color.MutedColor, out.String())
			color.Println(color.NoColor)
		}

		if cmd.ProcessState.ExitCode() != 0 {
			color.Printf(color.ErrorColorBold, "Couldn't fix all lint errors automatically for module %s. Run 'lint' manually to check for other issues.", details.Module)
			return customerrors.NewErrNoLog()
		} else {
			color.Printf(color.SuccessColorBold, "All lint errors for module: %s have been auto-fixed.\n", details.Module)
			return nil
		}
	}
}

func RunGolangCILint(details ModuleDetails, prefix string) error {
	color.Println(color.InfoColor, "golanlangci-lint run ./...")

	args := []string{"run"}

	// Append path prefix if module path is not "."
	if details.ModulePath != "." {
		args = append(args, "--path-prefix", prefix)
	}

	// Force Color Output for CI Env that supports it
	if color.ShouldForceColorOutputForCI() {
		args = append(args, "--color", "always")
	}

	args = append(args, AllModulesPath)

	cmd := exec.Command(GolangCILint, args...)
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout

	err := cmd.Run()

	// Command failed to run
	if cmd.ProcessState == nil {
		return fmt.Errorf(
			"error while running 'golangci-lint run ./...' for module %s, error: %s",
			details.Module,
			err.Error(),
		)
	} else {
		if cmd.ProcessState.ExitCode() != 0 {
			return fmt.Errorf("'golangci-lint run ./...' failed for module %s", details.Module)
		} else {
			color.Printf(color.SuccessColorBold, "Yay! No lint errors for module: %s\n", details.Module)
			return nil
		}
	}
}

func RunTests(
	details ModuleDetails,
	jsonOut string,
	coverageOut string,
	fileOutPath string,
) error {
	if filepath.IsAbs(coverageOut) || filepath.Clean(coverageOut) != coverageOut {
		return fmt.Errorf("coverage output path must be a relative path")
	}

	if filepath.IsAbs(jsonOut) || filepath.Clean(jsonOut) != jsonOut {
		return fmt.Errorf("json output path must be a relative path")
	}

	color.Println(color.InfoColor, "go test ./...")

	//nolint:gosec // coverageOut is validated user input
	cmd := exec.Command(
		GO,
		"test",
		"-cover",
		"-json",
		"-covermode=count",
		"-coverpkg=./...",
		"-coverprofile="+coverageOut,
		AllModulesPath,
	)
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	testOut := formattestjson.NewTestOutState()

	const ReadAllOwnerWritePerm = fs.FileMode(0o644)

	//nolint:gosec // jsonOut is validated user input
	jsonOutFile, errFile := os.OpenFile(
		path.Join(fileOutPath, jsonOut),
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		ReadAllOwnerWritePerm,
	)

	if errFile != nil {
		return fmt.Errorf("error while creating json output file: %s", errFile.Error())
	}

	outWriter := io.MultiWriter(jsonOutFile, testOut)
	errOut := bytes.Buffer{}

	cmd.Stderr = &errOut
	cmd.Stdout = outWriter

	err := cmd.Run()

	// Close the json output file irrespective of the command status
	errClose := jsonOutFile.Close()

	if errClose != nil {
		return fmt.Errorf("error while closing json output file: %s", errClose.Error())
	}

	// Command failed to run
	if cmd.ProcessState == nil {
		return fmt.Errorf(
			"error while running 'go test' for module %s, error: %s",
			details.Module,
			err.Error(),
		)
	} else {
		if errOut.Len() > 0 {
			color.Printf(color.ErrorColor, "Error while running 'go test' for module %s\n", details.Module)
			color.Print(color.MutedColor, errOut.String())
		}

		for pkg, out := range testOut.PackageOut {
			color.Printf(color.InfoColorBold, "Package: %s\n", pkg)

			_, err := os.Stdout.Write(out)
			if err != nil {
				color.Printf(color.ErrorColor, "\nError while writing test output for package %s: %s\n", pkg, err.Error())
			}

			color.Printf(color.SuccessColorBold, "Pass: %d\n", testOut.PackageResult[pkg].PassCount)
			color.Printf(color.ErrorColorBold, "Fail: %d\n", testOut.PackageResult[pkg].FailCount)
			color.Printf(color.WarningColorBold, "Skip: %d\n", testOut.PackageResult[pkg].SkipCount)
		}

		if cmd.ProcessState.ExitCode() != 0 {
			return fmt.Errorf("'go test' failed for module %s", details.Module)
		} else {
			color.Printf(color.SuccessColorBold, "Woohoo! All tests passed for module: %s\n", details.Module)
			return nil
		}
	}
}

func RunModuleDownload(details ModuleDetails) error {
	color.Println(color.InfoColor, "go mod download")

	cmd := exec.Command(GO, "mod", "download")
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	out := bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	// Command failed to run
	if cmd.ProcessState == nil {
		return fmt.Errorf(
			"error while running 'go mod download' for module %s, error: %s",
			details.Module,
			err.Error(),
		)
	} else {
		if out.Len() > 0 {
			color.Println(color.NoColor)
			color.Print(color.MutedColor, out.String())
			color.Println(color.NoColor)
		}

		if cmd.ProcessState.ExitCode() != 0 {
			return fmt.Errorf("'go mod download' failed for module %s", details.Module)
		} else {
			return nil
		}
	}
}

func RunModuleBuild(details ModuleDetails) error {
	color.Println(color.InfoColor, "go build ./...")

	cmd := exec.Command(GO, "build", "-trimpath", "-buildvcs=false", AllModulesPath)
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	out := bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	// Command failed to run
	if cmd.ProcessState == nil {
		return fmt.Errorf(
			"error while running 'go build' for module %s, error: %s",
			details.Module,
			err.Error(),
		)
	} else {
		if out.Len() > 0 {
			color.Println(color.NoColor)
			color.Print(color.MutedColor, out.String())
			color.Println(color.NoColor)
		}

		if cmd.ProcessState.ExitCode() != 0 {
			return fmt.Errorf("'go build' failed for module %s", details.Module)
		} else {
			return nil
		}
	}
}
