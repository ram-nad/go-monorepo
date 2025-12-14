// Package color provide utility functions for colored output
package color

import (
	"os"

	"github.com/fatih/color"
	cienv "github.com/ram-nad/go-monorepo/go-ci-tool/v2/ci_env"
)

//nolint:gochecknoglobals // Variable exported to used
var (
	InfoColorBold = color.New(color.FgCyan).Add(color.Bold)
	InfoColor     = color.New(color.FgCyan)

	ErrorColorBold = color.New(color.FgRed).Add(color.Bold)
	ErrorColor     = color.New(color.FgRed)

	WarningColorBold = color.New(color.FgYellow).Add(color.Bold)
	WarningColor     = color.New(color.FgYellow)

	SuccessColorBold = color.New(color.FgGreen).Add(color.Bold)
	SuccessColor     = color.New(color.FgGreen)

	HighLightColorBold = color.New(color.FgMagenta).Add(color.Bold)
	HighLightColor     = color.New(color.FgMagenta)

	MutedColor = color.New(color.Faint)

	NoColor = color.New(color.Reset)
)

func Println(color *color.Color, a ...interface{}) {
	//nolint:errcheck,gosec // Ignoring write errors to stdout
	color.Println(a...)
}

func Printf(c *color.Color, format string, a ...interface{}) {
	//nolint:errcheck,gosec // Ignoring write errors to stdout
	c.Printf(format, a...)
}

func Print(c *color.Color, a ...interface{}) {
	//nolint:errcheck,gosec // Ignoring write errors to stdout
	c.Print(a...)
}

func IsNoColorEnabled() bool {
	return os.Getenv("NO_COLOR") != ""
}

func EnableColorForAll() {
	InfoColorBold.EnableColor()
	InfoColor.EnableColor()

	ErrorColorBold.EnableColor()
	ErrorColor.EnableColor()

	WarningColorBold.EnableColor()
	WarningColor.EnableColor()

	SuccessColorBold.EnableColor()
	SuccessColor.EnableColor()

	HighLightColorBold.EnableColor()
	HighLightColor.EnableColor()

	MutedColor.EnableColor()
	NoColor.EnableColor()
}

func DisableColorForAll() {
	InfoColorBold.DisableColor()
	InfoColor.DisableColor()

	ErrorColorBold.DisableColor()
	ErrorColor.DisableColor()

	WarningColorBold.DisableColor()
	WarningColor.DisableColor()

	SuccessColorBold.DisableColor()
	SuccessColor.DisableColor()

	HighLightColorBold.DisableColor()
	HighLightColor.DisableColor()

	MutedColor.DisableColor()
	NoColor.DisableColor()
}

func ShouldForceColorOutputForCI() bool {
	return cienv.IsCIEnvAndSupportsColor() && !IsNoColorEnabled()
}
