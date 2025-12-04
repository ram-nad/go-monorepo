// Package listcaches is used to list caches used by the tools
package listcaches

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ram-nad/go-monorepo/go-ci-tool/color"
	customerrors "github.com/ram-nad/go-monorepo/go-ci-tool/custom_errors"
	"github.com/spf13/cobra"
)

type cachePaths struct {
	GoCache           string
	GoLangCILintCache string
	GoModCache        string
}

func computeCache(silent bool) (cachePaths, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		if !silent {
			color.Printf(
				color.ErrorColor,
				"Error while fetching user cache directory: %s\n",
				err.Error(),
			)
		}
		return cachePaths{}, errors.New("error in fetching user cache directory")
	}

	goCILintCache, exists := os.LookupEnv("GOLANGCI_LINT_CACHE")

	if !exists {
		goCILintCache = filepath.Join(userCacheDir, "golangci-lint")
	}

	goCache, err := exec.Command("go", "env", "GOCACHE").Output()
	if err != nil {
		if !silent {
			color.Printf(
				color.ErrorColor,
				"Error while fetching GOCACHE value: %s\n",
				err.Error(),
			)
		}
		return cachePaths{}, errors.New("error in fetching GOCACHE value")
	}

	goModeCache, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		if !silent {
			color.Printf(
				color.ErrorColor,
				"Error while fetching GOMODCACHE value: %s\n",
				err.Error(),
			)
		}
		return cachePaths{}, errors.New("error in fetching GOMODCACHE value")
	}

	const WhiteSpaces = "\n\r\t\f "

	cache := cachePaths{
		GoCache:           strings.Trim(string(goCache), WhiteSpaces),
		GoLangCILintCache: goCILintCache,
		GoModCache:        strings.Trim(string(goModeCache), WhiteSpaces),
	}

	return cache, nil
}

func GetCacheListCommand() *cobra.Command {
	longDesc := `
Lists the different caches used by the tools.
Currently lists the GOCACHE, GOLANGCI_LINT_CACHE and GOMODCACHE.
This is used to set environment variables in CI and as output for debugging.
`
	const EnvFlag = "env"
	const OutFlag = "out"
	const JSONFlag = "json"

	cacheListCommand := &cobra.Command{
		Use: "list-caches",
		RunE: func(cmd *cobra.Command, _ []string) error {
			envFormat, err := cmd.Flags().GetBool(EnvFlag)
			if err != nil {
				return err
			}

			outFormat, err := cmd.Flags().GetBool(OutFlag)
			if err != nil {
				return err
			}

			jsonFormat, err := cmd.Flags().GetBool(JSONFlag)
			if err != nil {
				return err
			}

			cache, err := computeCache(envFormat || outFormat)
			if err != nil {
				if envFormat || outFormat {
					return customerrors.NewErrNoLog()
				}
				return err
			}

			switch {
			case envFormat:
				color.Printf(color.NoColor, "GOCACHE=%s\n", cache.GoCache)
				color.Printf(
					color.NoColor,
					"GOLANGCI_LINT_CACHE=%s\n",
					cache.GoLangCILintCache,
				)
				color.Printf(color.NoColor, "GOMODCACHE=%s\n", cache.GoModCache)
			case outFormat:
				color.Printf(color.NoColor, "gocache-dir=%s\n", cache.GoCache)
				color.Printf(
					color.NoColor,
					"golangci-lint-dir=%s\n",
					cache.GoLangCILintCache,
				)
				color.Printf(color.NoColor, "gomodcache-dir=%s\n", cache.GoModCache)
			case jsonFormat:
				color.Printf(
					color.NoColor,
					"{\"gocache-dir\": \"%s\", \"golangci-lint-dir\": \"%s\", \"gomodcache-dir\": \"%s\"}\n",
					cache.GoCache,
					cache.GoLangCILintCache,
					cache.GoModCache,
				)
			default:
				color.Printf(color.InfoColor, "GOCACHE:%s\n", cache.GoCache)
				color.Printf(
					color.InfoColor,
					"GOLANGCI_LINT_CACHE:%s\n",
					cache.GoLangCILintCache,
				)
				color.Printf(color.InfoColor, "GOMODCACHE:%s\n", cache.GoModCache)
			}

			return nil
		},
		Args: cobra.NoArgs,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Short:                 "List differet caches used by the tools",
		Long:                  longDesc,
		SilenceErrors:         true,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
	}

	cacheListCommand.Flags().
		BoolP(EnvFlag, "e", false, "Print the output in the form of environment variables")
	cacheListCommand.Flags().
		BoolP(JSONFlag, "o", false, "Print the output in the form of key value pairs")
	cacheListCommand.Flags().
		BoolP(OutFlag, "j", false, "Print the output in JSON format")
	cacheListCommand.MarkFlagsMutuallyExclusive(EnvFlag, OutFlag, JSONFlag)

	return cacheListCommand
}
