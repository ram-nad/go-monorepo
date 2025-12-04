package main

import (
	"errors"
	"os"

	checktools "github.com/ram-nad/go-monorepo/go-ci-tool/check_tools"
	"github.com/ram-nad/go-monorepo/go-ci-tool/color"
	customerrors "github.com/ram-nad/go-monorepo/go-ci-tool/custom_errors"
	listcaches "github.com/ram-nad/go-monorepo/go-ci-tool/list_caches"
	"github.com/ram-nad/go-monorepo/go-ci-tool/modules"
	"github.com/spf13/cobra"
)

// Populated by the build system
var version = ""

func main() {
	// Explicitly enable color output for CI that supports it
	if color.ShouldForceColorOutputForCI() {
		color.EnableColorForAll()
	}

	command := os.Args[0]

	const longHelp = `
go-ci-tool is a utility to make your life easy (How? Check README).
It provides a set of commands to build, lint, format and test the Go code.
`

	rootCmd := &cobra.Command{
		Use:     command,
		Short:   "go-ci-tool makes your life easier",
		Long:    longHelp,
		Version: version,
		Args:    cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			return errors.New(
				"unknown command. run \"go-ci-tool help\" to see available commands",
			)
		},
		DisableAutoGenTag: true,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	rootCmd.InitDefaultCompletionCmd()

	rootCmd.AddCommand(checktools.GetCheckInstallationCommand())
	rootCmd.AddCommand(listcaches.GetCacheListCommand())
	rootCmd.AddCommand(modules.GetModulesCommand())
	rootCmd.AddCommand(modules.GetListModulesCommand())

	err := rootCmd.Execute()
	if err != nil {
		if !errors.Is(err, customerrors.NewErrNoLog()) {
			color.Printf(color.ErrorColorBold, "\n[go-ci-tool]: %s\n", err.Error())
		}

		os.Exit(1)
	}
}
