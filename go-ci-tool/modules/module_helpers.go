package modules

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/mod/modfile"
)

type ReplaceInfo struct {
	OldPath    string
	NewPath    string
	OldVersion string
	NewVersion string
}

type ModuleDetails struct {
	Module     string
	ModulePath string
	GoVersion  string
	Replaces   []ReplaceInfo
}

const (
	NotAbsolutePathError = "dir must be an absolute path"
	GoMod                = "go.mod"
)

// FindModuleRoot for the given directory, assuming it is a Go module
// `dir` must be an absolute path
// Returns relative path to the module root
func FindModuleRoot(dir string) (string, error) {
	if !filepath.IsAbs(dir) {
		return "", errors.New(NotAbsolutePathError)
	}

	dir = filepath.Clean(dir)

	d := dir

	for {
		fi, err := os.Stat(filepath.Join(d, GoMod))

		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		if err == nil && !fi.IsDir() {
			rel, err := filepath.Rel(dir, d)
			return rel, err
		}

		parent := filepath.Clean(filepath.Dir(d))

		if d == parent {
			break
		}

		d = parent
	}

	return "", errors.New("not inside a go module")
}

// FindAllModules finds all subdirectories (including current dir)
// that are a separate Go module
// `dir` must be an absolute path
// Returns relative paths to the modules
func FindAllModules(dir string) ([]string, error) {
	if !filepath.IsAbs(dir) {
		return nil, errors.New(NotAbsolutePathError)
	}

	dir = filepath.Clean(dir)

	modules := make([]string, 0)

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Name() == GoMod {
			relPath := strings.TrimPrefix(
				strings.TrimPrefix(filepath.Clean(filepath.Dir(path)), dir),
				string(filepath.Separator),
			)
			if relPath == "" {
				relPath = "."
			}
			modules = append(modules, relPath)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reverse to make sure sub-modules are listed before their parent modules
	slices.Reverse(modules)

	return modules, nil
}

// GetDetailsForModFile returns details of Go module located at `dir`
// dir must be an absolute path
func GetDetailsForModFile(dir string) (ModuleDetails, error) {
	if !filepath.IsAbs(dir) {
		return ModuleDetails{ModulePath: dir}, errors.New(NotAbsolutePathError)
	}

	//nolint:gosec // Safe to read this file
	mod, err := os.ReadFile(filepath.Join(dir, GoMod))
	if err != nil {
		err = fmt.Errorf(
			"unable to read go.mod file for module located at: %s, error: %s",
			dir,
			err.Error(),
		)
		return ModuleDetails{ModulePath: dir}, err
	}

	f, err := modfile.Parse(GoMod, mod, nil)
	if err != nil {
		err = fmt.Errorf(
			"unable to parse go.mod file for module located at: %s, error: %s",
			dir,
			err.Error(),
		)
		return ModuleDetails{ModulePath: dir}, err
	}

	goVersion := f.Go.Version
	moduleName := f.Module.Mod.Path

	replaces := make([]ReplaceInfo, 0)

	for _, r := range f.Replace {
		replaces = append(
			replaces,
			ReplaceInfo{
				OldPath:    r.Old.Path,
				NewPath:    r.New.Path,
				OldVersion: r.Old.Version,
				NewVersion: r.New.Version,
			},
		)
	}

	return ModuleDetails{
		Module:     moduleName,
		ModulePath: dir,
		GoVersion:  goVersion,
		Replaces:   replaces,
	}, nil
}
