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

	"github.com/ram-nad/go-monorepo/go-ci-tool/v5/color"
	"github.com/ram-nad/go-monorepo/go-ci-tool/v5/constants"
	customerrors "github.com/ram-nad/go-monorepo/go-ci-tool/v5/custom_errors"
	formattestjson "github.com/ram-nad/go-monorepo/go-ci-tool/v5/format_testjson"
	"golang.org/x/mod/semver"
)

const (
	AllModulesPath = "./..."
	GolangCILint   = "golangci-lint"
	GO             = "go"
	GoWorkOff      = "GOWORK=off"

	DirFileMode = 0o755
)

func runCmd(cmd *exec.Cmd) (stdout, stderr *bytes.Buffer, runErr error) {
	cmd.Stdin = nil
	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return stdout, stderr, cmd.Run()
}

func checkCmdExit(cmd *exec.Cmd, runErr error, label, module string) error {
	if cmd.ProcessState == nil {
		return fmt.Errorf(
			"error while running %s for module %s, error: %s",
			label,
			module,
			runErr.Error(),
		)
	}
	if cmd.ProcessState.ExitCode() != 0 {
		return fmt.Errorf("%s failed for module %s", label, module)
	}
	return nil
}

func printMutedOutput(stdout, stderr *bytes.Buffer, label string) {
	if stdout.Len() > 0 {
		color.Println(color.NoColor)
		color.Print(color.MutedColor, stdout.String())
		color.Println(color.NoColor)
	}
	if stderr.Len() > 0 {
		color.Println(color.NoColor)
		color.Printf(color.MutedColor, "[stderr from %s]\n", label)
		color.Print(color.MutedColor, stderr.String())
		color.Println(color.NoColor)
	}
}

func printRawOutput(stdout, stderr *bytes.Buffer, label string) {
	if stdout.Len() > 0 {
		color.Print(color.NoColor, stdout.String())
	}
	if stderr.Len() > 0 {
		color.Println(color.NoColor)
		color.Printf(color.MutedColor, "[stderr from %s]\n", label)
		color.Print(color.MutedColor, stderr.String())
		color.Println(color.NoColor)
	}
}

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
		color.Printf(
			color.SuccessColorBold,
			"Go module %s is not using any local replaces\n",
			details.Module,
		)
		return nil
	}
}

func CheckModuleTidy(details ModuleDetails) error {
	color.Println(color.InfoColor, "go mod tidy -diff")

	//nolint:gosec // details.ModulePath is not a user input
	cmd := exec.Command(GO, "-C", details.ModulePath, "mod", "tidy", "-diff")
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	const label = "'go mod tidy -diff'"
	stdout, stderr, runErr := runCmd(cmd)
	printMutedOutput(stdout, stderr, label)
	if err := checkCmdExit(cmd, runErr, label, details.Module); err != nil {
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
			color.Printf(
				color.ErrorColorBold,
				"Go module %s is not tidy. Run 'go mod tidy'\n",
				details.Module,
			)
			return customerrors.NewErrNoLog()
		}
		return err
	}
	color.Printf(color.SuccessColorBold, "Go module %s is tidy :)\n", details.Module)
	return nil
}

func RunModuleTidy(details ModuleDetails) error {
	color.Println(color.InfoColor, "go mod tidy")

	//nolint:gosec // details.ModulePath is not a user input
	cmd := exec.Command(GO, "-C", details.ModulePath, "mod", "tidy")
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	const label = "'go mod tidy'"
	stdout, stderr, runErr := runCmd(cmd)
	printMutedOutput(stdout, stderr, label)
	if err := checkCmdExit(cmd, runErr, label, details.Module); err != nil {
		return err
	}
	color.Printf(color.SuccessColorBold, "Go module %s is now tidy.\n", details.Module)
	return nil
}

func RunGolangCILintFmt(details ModuleDetails) error {
	color.Println(color.InfoColor, "golanlangci-lint fmt ./...")

	args := []string{"fmt", AllModulesPath}
	args = append(args, "--color", "never")

	cmd := exec.Command(GolangCILint, args...)
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	const label = "'golangci-lint fmt ./...'"
	stdout, stderr, runErr := runCmd(cmd)
	printMutedOutput(stdout, stderr, label)
	if err := checkCmdExit(cmd, runErr, label, details.Module); err != nil {
		return err
	}
	color.Printf(
		color.SuccessColorBold,
		"Code for module: %s has been formatted.\n",
		details.Module,
	)
	return nil
}

func RunGolangCILintFix(details ModuleDetails) error {
	color.Println(color.InfoColor, "golanlangci-lint run --fix ./...")

	args := []string{"run", "--fix", AllModulesPath}
	args = append(args, "--color", "never")

	cmd := exec.Command(GolangCILint, args...)
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	const label = "'golangci-lint run --fix ./...'"
	stdout, stderr, runErr := runCmd(cmd)
	printMutedOutput(stdout, stderr, label)
	if err := checkCmdExit(cmd, runErr, label, details.Module); err != nil {
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
			color.Printf(
				color.ErrorColorBold,
				"Couldn't fix all lint errors automatically for module %s. Run 'lint' manually to check for other issues.",
				details.Module,
			)
			return customerrors.NewErrNoLog()
		}
		return err
	}
	color.Printf(
		color.SuccessColorBold,
		"All lint errors for module: %s have been auto-fixed.\n",
		details.Module,
	)
	return nil
}

func RunGolangCILint(details ModuleDetails, prefix string) error {
	color.Println(color.InfoColor, "golanlangci-lint run ./...")

	args := []string{"run"}

	// Append path prefix if module path is not "."
	if details.ModulePath != "." {
		args = append(args, "--path-prefix", prefix)
	}

	if color.ShouldOutputColor() {
		args = append(args, "--color", "always")
	} else {
		args = append(args, "--color", "never")
	}

	args = append(args, AllModulesPath)

	cmd := exec.Command(GolangCILint, args...)
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	const label = "'golangci-lint run ./...'"
	stdout, stderr, runErr := runCmd(cmd)
	printRawOutput(stdout, stderr, label)
	if err := checkCmdExit(cmd, runErr, label, details.Module); err != nil {
		return err
	}
	color.Printf(
		color.SuccessColorBold,
		"Yay! No lint errors for module: %s\n",
		details.Module,
	)
	return nil
}

func RunVet(details ModuleDetails, integration bool) error {
	vetCmdLabel := "vet"
	if integration {
		vetCmdLabel = "vet (integration)"
		color.Println(color.InfoColor, "go vet -tags=integration ./...")
	} else {
		color.Println(color.InfoColor, "go vet ./...")
	}

	args := []string{"vet"}
	if integration {
		args = append(args, "-tags=integration")
	}
	args = append(args, AllModulesPath)

	cmd := exec.Command(GO, args...)
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	label := "'go " + vetCmdLabel + "'"
	stdout, stderr, runErr := runCmd(cmd)
	printMutedOutput(stdout, stderr, label)
	if err := checkCmdExit(cmd, runErr, label, details.Module); err != nil {
		return err
	}
	color.Printf(
		color.SuccessColorBold,
		"Yay! No %s issues for module: %s\n",
		vetCmdLabel,
		details.Module,
	)
	return nil
}

func RunTests(
	details ModuleDetails,
	fileOutPath string,
	integration bool,
) error {
	jsonOut := "test.out.json"
	coverageOut := "coverage.out"
	coverdataDir := ".coverdata"
	if integration {
		jsonOut = "test.integration.out.json"
		coverageOut = "coverage.integration.out"
		coverdataDir = ".coverdata.integration"
	}

	coverdataDirAbs := filepath.Join(details.ModulePath, coverdataDir)
	if err := os.RemoveAll(coverdataDirAbs); err != nil {
		return fmt.Errorf(
			"error while cleaning coverage data directory: %s",
			err.Error(),
		)
	}
	if err := os.MkdirAll(coverdataDirAbs, fs.FileMode(DirFileMode)); err != nil {
		return fmt.Errorf(
			"error while creating coverage data directory: %s",
			err.Error(),
		)
	}

	testCmdLabel := "test"
	if integration {
		testCmdLabel = "test (integration)"
		color.Println(color.InfoColor, "go test -tags=integration ./...")
	} else {
		color.Println(color.InfoColor, "go test ./...")
	}

	testArgs := []string{
		"test",
		"-cover",
		"-json",
		"-covermode=count",
		"-coverpkg=./...",
	}
	if integration {
		testArgs = append(testArgs, "-tags=integration")
	}
	testArgs = append(testArgs, AllModulesPath)
	testArgs = append(testArgs, "-args", "-test.gocoverdir="+coverdataDirAbs)

	cmd := exec.Command(GO, testArgs...)
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)
	cmd.Stdin = nil

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

	runErr := cmd.Run()

	// Close the json output file irrespective of the command status
	errClose := jsonOutFile.Close()

	if errClose != nil {
		return fmt.Errorf("error while closing json output file: %s", errClose.Error())
	}

	label := "'go " + testCmdLabel + "'"

	if cmd.ProcessState != nil {
		if errOut.Len() > 0 {
			color.Printf(
				color.ErrorColor,
				"Error while running %s for module %s\n",
				label,
				details.Module,
			)
			color.Print(color.MutedColor, errOut.String())
		}

		for pkg, out := range testOut.PackageOut {
			color.Printf(color.InfoColorBold, "Package: %s\n", pkg)
			color.Print(color.NoColor, string(out))

			color.Printf(
				color.SuccessColorBold,
				"Pass: %d\n",
				testOut.PackageResult[pkg].PassCount,
			)
			color.Printf(
				color.ErrorColorBold,
				"Fail: %d\n",
				testOut.PackageResult[pkg].FailCount,
			)
			color.Printf(
				color.WarningColorBold,
				"Skip: %d\n",
				testOut.PackageResult[pkg].SkipCount,
			)
		}

		coverageProfilePath := filepath.Join(details.ModulePath, coverageOut)
		if errCov := writeCoverageProfile(
			coverdataDirAbs,
			coverageProfilePath,
			details.Module,
		); errCov != nil {
			color.Printf(
				color.ErrorColor,
				"Error generating coverage profile after %s for module %s: %s\n",
				testCmdLabel,
				details.Module,
				errCov.Error(),
			)
		}
	}

	if err := checkCmdExit(cmd, runErr, label, details.Module); err != nil {
		return err
	}
	color.Printf(
		color.SuccessColorBold,
		"Woohoo! All %s cases passed for module: %s\n",
		testCmdLabel,
		details.Module,
	)
	return nil
}

func writeCoverageProfile(srcDir, dstPath, module string) error {
	if _, err := os.Stat(srcDir); err != nil {
		return fmt.Errorf(
			"coverage data directory not accessible: %s: %s",
			srcDir,
			err.Error(),
		)
	}

	const label = "'go tool covdata textfmt'"
	//nolint:gosec // srcDir and dstPath are internally constructed
	cmd := exec.Command(GO, "tool", "covdata", "textfmt", "-i="+srcDir, "-o="+dstPath)
	cmd.Env = append(os.Environ(), GoWorkOff)

	stdout, stderr, runErr := runCmd(cmd)
	printMutedOutput(stdout, stderr, label)
	return checkCmdExit(cmd, runErr, label, module)
}

func CombineCoverage(details ModuleDetails) error {
	coverdataDir := filepath.Join(details.ModulePath, ".coverdata")
	coverdataIntegrationDir := filepath.Join(
		details.ModulePath,
		".coverdata.integration",
	)
	coverdataCombinedDir := filepath.Join(details.ModulePath, ".coverdata.combined")
	coverageCombinedFile := filepath.Join(details.ModulePath, "coverage.combined.out")

	color.Printf(
		color.InfoColor,
		"Combining coverage profiles for module %s\n",
		details.Module,
	)

	for _, dir := range []string{coverdataDir, coverdataIntegrationDir} {
		if _, err := os.Stat(dir); err != nil {
			color.Printf(
				color.WarningColor,
				"No coverage data found at %s for module %s\n",
				dir,
				details.Module,
			)
		}
		if err := os.MkdirAll(dir, fs.FileMode(DirFileMode)); err != nil {
			return fmt.Errorf(
				"error while creating coverage data directory %s: %s",
				dir,
				err.Error(),
			)
		}
	}

	if err := os.RemoveAll(coverdataCombinedDir); err != nil {
		return fmt.Errorf(
			"error while cleaning combined coverage data directory: %s",
			err.Error(),
		)
	}
	if err := os.MkdirAll(coverdataCombinedDir, fs.FileMode(DirFileMode)); err != nil {
		return fmt.Errorf(
			"error while creating combined coverage data directory: %s",
			err.Error(),
		)
	}

	color.Println(color.InfoColor, "go tool covdata merge")

	//nolint:gosec // arguments are internally constructed
	mergeCmd := exec.Command(GO, "tool", "covdata", "merge",
		"-i="+coverdataDir+","+coverdataIntegrationDir,
		"-o="+coverdataCombinedDir,
	)
	mergeCmd.Dir = details.ModulePath
	mergeCmd.Env = append(os.Environ(), GoWorkOff)

	const mergeLabel = "'go tool covdata merge'"
	mergeStdout, mergeStderr, mergeErr := runCmd(mergeCmd)
	printMutedOutput(mergeStdout, mergeStderr, mergeLabel)
	if err := checkCmdExit(mergeCmd, mergeErr, mergeLabel, details.Module); err != nil {
		return err
	}

	if err := writeCoverageProfile(
		coverdataCombinedDir,
		coverageCombinedFile,
		details.Module,
	); err != nil {
		return err
	}

	color.Printf(
		color.SuccessColorBold,
		"Combined coverage profile written to %s for module %s\n",
		coverageCombinedFile,
		details.Module,
	)
	return nil
}

func ViewCoverage(details ModuleDetails, integration, combined bool) error {
	coverageFile := "coverage.out"
	label := "'go tool cover -html'"
	hint := "--test"
	if integration {
		coverageFile = "coverage.integration.out"
		label = "'go tool cover -html (integration)'"
		hint = "--test --integration"
	} else if combined {
		coverageFile = "coverage.combined.out"
		label = "'go tool cover -html (combined)'"
		hint = "--combine-coverage"
	}

	coveragePath := filepath.Join(details.ModulePath, coverageFile)

	if _, err := os.Stat(coveragePath); err != nil {
		return fmt.Errorf(
			"coverage profile not found at %s: %s. Run %s first",
			coveragePath,
			err.Error(),
			hint,
		)
	}

	color.Printf(color.InfoColor, "go tool cover -html=%s\n", coverageFile)

	//nolint:gosec // coveragePath is internally constructed
	cmd := exec.Command(GO, "tool", "cover", "-html="+coveragePath)
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	stdout, stderr, runErr := runCmd(cmd)
	printMutedOutput(stdout, stderr, label)
	return checkCmdExit(cmd, runErr, label, details.Module)
}

func RunModuleDownload(details ModuleDetails) error {
	color.Println(color.InfoColor, "go mod download")

	cmd := exec.Command(GO, "mod", "download")
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	const label = "'go mod download'"
	stdout, stderr, runErr := runCmd(cmd)
	printMutedOutput(stdout, stderr, label)
	return checkCmdExit(cmd, runErr, label, details.Module)
}

func RunModuleBuild(details ModuleDetails) error {
	color.Println(color.InfoColor, "go build ./...")

	cmd := exec.Command(GO, "build", "-trimpath", "-buildvcs=false", AllModulesPath)
	cmd.Dir = details.ModulePath
	cmd.Env = append(os.Environ(), GoWorkOff)

	const label = "'go build ./...'"
	stdout, stderr, runErr := runCmd(cmd)
	printMutedOutput(stdout, stderr, label)
	return checkCmdExit(cmd, runErr, label, details.Module)
}
