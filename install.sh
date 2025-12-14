GOLANGCI_LINT_VERSION="2.7.0"
GO_MIN_VERSION="1.25.5"

# Check if Go is installed
if (go version >/dev/null 2>&1) then
    echo "[install] Go is already installed"
    true
else
    echo "[install] Go is not installed. Please install before running this script"
    false
    return
fi

# Check if GOPATH/bin is in the PATH
GOBIN_PATH=$(go env GOPATH)/bin
if [[ ":$PATH:" == *":$GOBIN_PATH:"* ]]
then
    echo "[install] '$GOBIN_PATH' is already in the PATH"
    true
else
    echo "[install] '$GOBIN_PATH' is not in the PATH. Please add it to your PATH"
    false
    return
fi

# Check if golangci-lint is installed and matches the required version
if (golangci-lint version >/dev/null 2>&1) && [ $(golangci-lint version --short) == $GOLANGCI_LINT_VERSION ]
then
    echo "[install] Golang CI Lint is already installed"
    true
else
    echo "[install] Installing Golang CI Lint v$GOLANGCI_LINT_VERSION"
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $GOBIN_PATH v$GOLANGCI_LINT_VERSION
fi

# Build the go-tool tool
echo "[install] Building the go-ci-tool :)"

VERSION_SUBSTITUION="main.version=2.0.0"
GO_MIN_VERSION_SUBSTITUTION="github.com/ram-nad/go-monorepo/go-ci-tool/v2/constants.minGoVersion=$GO_MIN_VERSION"
GOLANG_CI_LINT_VERSION_SUBSTITUTION="github.com/ram-nad/go-monorepo/go-ci-tool/v2/constants.minGolangCILintVersion=$GOLANGCI_LINT_VERSION"

GOWORK=off go install -C go-ci-tool -trimpath -buildvcs=false -ldflags="-w -X $VERSION_SUBSTITUION -X $GO_MIN_VERSION_SUBSTITUTION -X $GOLANG_CI_LINT_VERSION_SUBSTITUTION" .
