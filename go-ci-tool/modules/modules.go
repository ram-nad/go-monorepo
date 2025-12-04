package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ram-nad/go-monorepo/go-ci-tool/color"
	"github.com/spf13/cobra"
)

const (
	BuildFlag             = "build"
	DownloadFlag          = "download"
	TestFlag              = "test"
	FmtFlag               = "fmt"
	FixFlag               = "fix"
	LintFlag              = "lint"
	TidifyFlag            = "tidify"
	IsTidyFlag            = "is-tidy"
	VersionFlag           = "version"
	CheckLocalReplaceFlag = "check-local-replace"
	ModFlag               = "mod"
)

//nolint:gocognit,cyclop // No better way to deal wit many flags
func GetModulesCommand() *cobra.Command {
	modulesCommand := &cobra.Command{
		Use: "module",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			modProvided := cmd.Flags().Changed(ModFlag)
			modPath, err := cmd.Flags().GetString(ModFlag)
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
					return fmt.Errorf("unable to find current module root: %s", err.Error())
				}
				absModulePath = filepath.Join(cwd, relModulePath)
			}

			moduleDetails, err := GetDetailsForModFile(absModulePath)
			if err != nil {
				return err
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
				return RunTests(
					moduleDetails,
					"test.out.json",
					"coverage.out",
					relModulePath,
				)
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

			checkVersion, err := cmd.Flags().GetBool(VersionFlag)
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
		BoolP(VersionFlag, "v", false, "Check Go module version and compare with minimum supported version")
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
	modulesCommand.Flags().Bool(DownloadFlag, false, "Download module dependencies")
	modulesCommand.Flags().
		Bool(BuildFlag, false, "Build all the packages in the module")

	modulesCommand.Flags().
		StringP(ModFlag, "m", "", "Path to the module root directory for which to run the command. Default is current module root")

	modulesCommand.MarkFlagsMutuallyExclusive(
		VersionFlag,
		CheckLocalReplaceFlag,
		IsTidyFlag,
		TidifyFlag,
		LintFlag,
		FmtFlag,
		FixFlag,
		TestFlag,
		DownloadFlag,
		BuildFlag,
	)

	err := modulesCommand.MarkFlagDirname(ModFlag)
	if err != nil {
		panic(err)
	}

	return modulesCommand
}

//nolint:gocognit,cyclop // The logic to compute modules with diff is complex but it needs to be kept in single place
func GetListModulesCommand() *cobra.Command {
	const listModulesLongHelpDesc = `
List Go modules prsent in current or sub-directories. Current directory is only returned if it is a module root.
Optionally takes a file which should contain file names, then this command only returns a module if any of those files belong to that module.
This is useful to get list of modules whose code has changed in a PR or Push.
`

	const (
		DiffFlag            = "diff"
		SkipNoFileErrorFlag = "skip-no-file-error"
		JSONFlag            = "json"
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

			diffChanged := cmd.Flags().Changed(DiffFlag)
			diffFile, err := cmd.Flags().GetString(DiffFlag)
			if err != nil {
				return err
			}

			skipErrNoFile, err := cmd.Flags().GetBool(SkipNoFileErrorFlag)
			if err != nil {
				return err
			}

			isJSON, err := cmd.Flags().GetBool(JSONFlag)
			if err != nil {
				return err
			}

			var files []string = nil

			if diffChanged {
				if diffFile == "" {
					return fmt.Errorf(
						"invalid value empty string provided for 'diff' flag",
					)
				} else {
					//nolint:gosec // It is fine to read this file
					fileBytes, err := os.ReadFile(diffFile)
					if err != nil {
						if !os.IsNotExist(err) {
							return fmt.Errorf("error while reading diff file: %s, error: %s", diffFile, err.Error())
						} else if !skipErrNoFile {
							return fmt.Errorf("provided diff file does not exist: %s", diffFile)
						}
					} else {
						files = strings.Split(string(fileBytes), "\n")
					}
				}
			}

			if files != nil {
				modulesWithDiff := make(map[string]bool)

				for _, file := range files {
					// Trim spaces from line
					file = strings.Trim(file, " \t\r\f")
					if file == "" {
						continue
					}

					file = filepath.Clean(file)

					// If file not a local path
					if !filepath.IsLocal(file) {
						// Make sure all non-local paths are absolute
						if !filepath.IsAbs(file) {
							file = filepath.Join(cwd, file)
						}

						if !strings.HasPrefix(file, cwd) {
							continue
						}

						file = strings.TrimPrefix(
							strings.TrimPrefix(file, cwd),
							string(filepath.Separator),
						)

						// Only if path was pointing to a current directory itself
						if file == "" {
							file = "."
						}
					}

					for _, module := range allModules {
						if strings.HasPrefix(file, module) || module == "." {
							modulesWithDiff[module] = true
							break
						}
					}
				}

				allModules = make([]string, 0, len(modulesWithDiff))
				for module, hasDiff := range modulesWithDiff {
					if hasDiff {
						allModules = append(allModules, module)
					}
				}
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

	listModulesCommand.Flags().
		StringP(DiffFlag, "d", "", "File containing diff for which we fetch list of modules")
	listModulesCommand.Flags().
		Bool(SkipNoFileErrorFlag, false, "Don't error if diff file does not exist. List all modules in that case")

	listModulesCommand.Flags().Bool(JSONFlag, false, "Output in JSON array format")

	err := listModulesCommand.MarkFlagFilename(DiffFlag)
	if err != nil {
		panic(err)
	}

	return listModulesCommand
}
