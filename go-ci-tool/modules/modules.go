package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ram-nad/go-monorepo/go-ci-tool/v5/color"
	"github.com/spf13/cobra"
)

const (
	BuildFlag             = "build"
	DownloadFlag          = "download"
	TestFlag              = "test"
	VetFlag               = "vet"
	FmtFlag               = "fmt"
	FixFlag               = "fix"
	LintFlag              = "lint"
	TidifyFlag            = "tidify"
	IsTidyFlag            = "is-tidy"
	CheckVersionFlag      = "check-version"
	CheckLocalReplaceFlag = "check-local-replace"
	CombineCoverageFlag   = "combine-coverage"
	ViewCoverageFlag      = "view-coverage"
	IntegrationFlag       = "integration"
	CombinedFlag          = "combined"
	ModuleFlag            = "module"
	WorkspaceFlag         = "workspace"
)

//nolint:gocognit,cyclop,maintidx // No better way to deal with many flags
func GetModulesCommand() *cobra.Command {
	modulesCommand := &cobra.Command{
		Use: "mod",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			modProvided := cmd.Flags().Changed(ModuleFlag)
			modPath, err := cmd.Flags().GetString(ModuleFlag)
			if err != nil {
				return err
			}

			// Relative path of module from current directory
			var relModulePath string
			// Absolute path of module
			var absModulePath string

			if modProvided {
				if modPath == "" {
					return fmt.Errorf(
						"invalid value empty string provided for 'mod' flag. Omit flag if you want to use current module",
					)
				}

				modPath = filepath.Clean(modPath)
				if filepath.IsAbs(modPath) {
					absModulePath = modPath
					relModulePath = absModulePath
				} else {
					relModulePath = modPath
					absModulePath = filepath.Join(cwd, modPath)
				}
			} else {
				relModulePath, err = FindModuleRoot(cwd)
				if err != nil {
					return fmt.Errorf(
						"unable to find current module root: %s",
						err.Error(),
					)
				}
				absModulePath = filepath.Join(cwd, relModulePath)
			}

			moduleDetails, err := GetDetailsForModFile(absModulePath)
			if err != nil {
				return err
			}

			integration, err := cmd.Flags().GetBool(IntegrationFlag)
			if err != nil {
				return err
			}
			if integration {
				test, err := cmd.Flags().GetBool(TestFlag)
				if err != nil {
					return err
				}
				vet, err := cmd.Flags().GetBool(VetFlag)
				if err != nil {
					return err
				}
				viewCoverage, err := cmd.Flags().GetBool(ViewCoverageFlag)
				if err != nil {
					return err
				}
				if !test && !vet && !viewCoverage {
					return fmt.Errorf(
						"--%s is only valid with --%s, --%s, or --%s",
						IntegrationFlag,
						TestFlag,
						VetFlag,
						ViewCoverageFlag,
					)
				}
			}

			combined, err := cmd.Flags().GetBool(CombinedFlag)
			if err != nil {
				return err
			}
			if combined {
				viewCoverage, err := cmd.Flags().GetBool(ViewCoverageFlag)
				if err != nil {
					return err
				}

				if !viewCoverage {
					return fmt.Errorf(
						"--%s is only valid with --%s",
						CombinedFlag,
						ViewCoverageFlag,
					)
				}
			}

			checkLocalReplace, err := cmd.Flags().GetBool(CheckLocalReplaceFlag)
			if err != nil {
				return err
			}
			if checkLocalReplace {
				return CheckReplaceIsNotLocal(moduleDetails)
			}

			isTidy, err := cmd.Flags().GetBool(IsTidyFlag)
			if err != nil {
				return err
			}
			if isTidy {
				return CheckModuleTidy(moduleDetails)
			}

			tidify, err := cmd.Flags().GetBool(TidifyFlag)
			if err != nil {
				return err
			}
			if tidify {
				return RunModuleTidy(moduleDetails)
			}

			lint, err := cmd.Flags().GetBool(LintFlag)
			if err != nil {
				return err
			}
			if lint {
				return RunGolangCILint(moduleDetails, relModulePath)
			}

			vet, err := cmd.Flags().GetBool(VetFlag)
			if err != nil {
				return err
			}
			if vet {
				return RunVet(moduleDetails, integration)
			}

			fmt, err := cmd.Flags().GetBool(FmtFlag)
			if err != nil {
				return err
			}
			if fmt {
				return RunGolangCILintFmt(moduleDetails)
			}

			fix, err := cmd.Flags().GetBool(FixFlag)
			if err != nil {
				return err
			}
			if fix {
				return RunGolangCILintFix(moduleDetails)
			}

			test, err := cmd.Flags().GetBool(TestFlag)
			if err != nil {
				return err
			}
			if test {
				return RunTests(moduleDetails, relModulePath, integration)
			}

			combineCoverage, err := cmd.Flags().GetBool(CombineCoverageFlag)
			if err != nil {
				return err
			}
			if combineCoverage {
				return CombineCoverage(moduleDetails)
			}

			viewCoverage, err := cmd.Flags().GetBool(ViewCoverageFlag)
			if err != nil {
				return err
			}
			if viewCoverage {
				return ViewCoverage(moduleDetails, integration, combined)
			}

			download, err := cmd.Flags().GetBool(DownloadFlag)
			if err != nil {
				return err
			}
			if download {
				return RunModuleDownload(moduleDetails)
			}

			build, err := cmd.Flags().GetBool(BuildFlag)
			if err != nil {
				return err
			}
			if build {
				return RunModuleBuild(moduleDetails)
			}

			checkVersion, err := cmd.Flags().GetBool(CheckVersionFlag)
			if err != nil {
				return err
			}
			if checkVersion {
				return CheckMinVersionSupported(moduleDetails)
			}

			// Default
			// If no flags are provided, print the module details
			color.Printf(
				color.InfoColor,
				"Module: %s\nPath: %s\nMin Go Version: %s\n",
				moduleDetails.Module,
				moduleDetails.ModulePath,
				moduleDetails.GoVersion,
			)

			return nil
		},
		Args: cobra.NoArgs,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Short:                 "Module specific commands",
		Long:                  "Commands to do specific actions for individual modules. Things like linting, formatting, tidy, testing and more.",
		SilenceErrors:         true,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
	}

	modulesCommand.Flags().
		Bool(CheckVersionFlag, false, "Check module Go version and compare with minimum supported version")
	modulesCommand.Flags().
		Bool(CheckLocalReplaceFlag, false, "Check if module is using any replace directive with a local path")
	modulesCommand.Flags().Bool(IsTidyFlag, false, "Check if the module is tidy")
	modulesCommand.Flags().Bool(TidifyFlag, false, "Run 'go mod tidy' for the module")
	modulesCommand.Flags().Bool(LintFlag, false, "Run 'golangci-lint' for the module")
	modulesCommand.Flags().
		Bool(FmtFlag, false, "Formats the module using 'golangci-lint'")
	modulesCommand.Flags().
		Bool(FixFlag, false, "Fix auto-fixable lint issues in the module")
	modulesCommand.Flags().BoolP(TestFlag, "t", false, "Run Tests for the module")
	modulesCommand.Flags().Bool(VetFlag, false, "Run 'go vet' for the module")
	modulesCommand.Flags().
		Bool(CombineCoverageFlag, false, "Combine normal and integration coverage profiles for the module")
	modulesCommand.Flags().
		Bool(ViewCoverageFlag, false, "Open HTML coverage report for the module in browser")
	modulesCommand.Flags().
		Bool(IntegrationFlag, false, "Run --test or --vet in integration mode (passes -tags=integration); with --view-coverage, opens integration coverage")
	modulesCommand.Flags().
		Bool(CombinedFlag, false, "With --view-coverage, opens the combined (normal + integration) coverage report")
	modulesCommand.Flags().Bool(DownloadFlag, false, "Download module dependencies")
	modulesCommand.Flags().
		BoolP(BuildFlag, "b", false, "Build all the packages in the module")

	modulesCommand.Flags().
		StringP(ModuleFlag, "m", "", "Path to the module root directory for which to run the command. Default is root of current module")

	modulesCommand.MarkFlagsMutuallyExclusive(
		CheckVersionFlag,
		CheckLocalReplaceFlag,
		IsTidyFlag,
		TidifyFlag,
		LintFlag,
		FmtFlag,
		FixFlag,
		TestFlag,
		VetFlag,
		CombineCoverageFlag,
		ViewCoverageFlag,
		DownloadFlag,
		BuildFlag,
	)

	modulesCommand.MarkFlagsMutuallyExclusive(IntegrationFlag, CombinedFlag)

	err := modulesCommand.MarkFlagDirname(ModuleFlag)
	if err != nil {
		panic(err)
	}

	return modulesCommand
}

func GetListModulesCommand() *cobra.Command {
	const listModulesLongHelpDesc = `
List Go modules prsent in current or sub-directories. Current directory is only returned if it is a module root.
`

	const (
		JSONFlag = "json"
	)

	listModulesCommand := &cobra.Command{
		Use: "list-modules",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			allModules, err := FindAllModules(cwd)
			if err != nil {
				return err
			}

			isJSON, err := cmd.Flags().GetBool(JSONFlag)
			if err != nil {
				return err
			}

			if isJSON {
				out, err := json.Marshal(allModules)
				if err != nil {
					return fmt.Errorf(
						"error while formatting modules to JSON: %s",
						err.Error(),
					)
				}

				_, err = os.Stdout.Write(append(out, '\n'))
				if err != nil {
					return fmt.Errorf(
						"error while writing JSON output: %s",
						err.Error(),
					)
				}
			} else {
				for _, module := range allModules {
					color.Println(color.InfoColor, module)
				}
			}

			return nil
		},
		Args: cobra.NoArgs,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Short:                 "List all module roots in current or subdirectories",
		Long:                  listModulesLongHelpDesc,
		SilenceErrors:         true,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
	}

	listModulesCommand.Flags().Bool(JSONFlag, false, "Output in JSON array format")

	return listModulesCommand
}
