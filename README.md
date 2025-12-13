## Go Monorepo Setup

Go modules are code that is supposed to be versioned and released separately. This monorepo will contain various modules that may be inter-dependent on each other, in which case they should import the dependencies like any other external modules only. Each module will also be tested separately.

### go.work

Use Go workspaces when making changes to multiple modules together. `go.work` files are not supposed to be committed to VCS. Use it locally to make easily test changes across multiple modules. All the tests/builds are run with `GOWORK=off` to ensure they run independent of workspaces.

### CI

We validate following things in CI for all the modules:

1. `go.mod` is valid and tidy
2. `go.mod` doesn't contain any replaces with local modules. This is sometimes required during testing, but should be reverted before submitting the changes.
3. Module uses Go Version lower than minimum supported Go version.
4. Code is formatted and correctly and follows the best practices, we use `GolangCI Lint` for this.
5. We are able to download dependencies and build the modules.

The CI step is optimised to execute fast and use minimal compute. To that effect, it caches all the downloaded Go modules, build directories (for faster builds) and even GolangCI-Lint cache for faster linting. Morever, it only runs the above validations for modules that are modified in a PR/commit.

### Using the setup in your own GitHub repository

1. Use the `.github/workflows/go-ci.yml` workflow in your repository (Check [ci.yaml](.github/workflows/ci.yml)for example)

### Using the setup locally

Note: This will override the pre-existing binaries of `GolangCI-Lint` and `go-ci-tool`.

1. Clone [this repository](https://github.com/ram-nad/go-monorepo)
2. Install Go [https://go.dev/doc/install](https://go.dev/doc/install)
3. Add `$(go env GOPATH)/bin` (Path to binaries installed by Go) to your `PATH`
4. Run `install.sh` in Linux/Mac or `install.ps1` in Windows
5. You can optionally change the minimum Go version and Golang CI Lint version in install scripts before running them

### Resources

1. Writing Tests: https://research.swtch.com/testing
