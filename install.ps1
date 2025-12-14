$GOLANGCI_LINT_VERSION = "2.7.0"
$GO_MIN_VERSION = "1.25.5"

# Check if Go is installed
if (Get-Command go -ErrorAction SilentlyContinue) {
    Write-Output "[install] Go is already installed"
} else {
    Write-Output "[install] Go is not installed. Please install before running this script"
    return
}

# Check if GOPATH/bin is in the PATH
$GOBIN_PATH = (go env GOPATH) + "\bin"
if ($env:PATH.Contains($GOBIN_PATH)) {
    Write-Output "[install] '$GOBIN_PATH' is already in the PATH"
} else {
    Write-Output "[install] '$GOBIN_PATH' is not in the PATH. Please add it to your PATH"
    return
}

# Check if golangci-lint is installed and matches the required version
$golangciLintVersion = ""
if (Get-Command golangci-lint -ErrorAction SilentlyContinue) { $golangciLintVersion = (golangci-lint version --short 2>$null) }

if ($golangciLintVersion -eq $GOLANGCI_LINT_VERSION) {
    Write-Output "[install] Golang CI Lint is already installed"
} else {
    Write-Output "[install] Installing Golang CI Lint v$GOLANGCI_LINT_VERSION"
    $arch = '386'
    if ((Get-ComputerInfo | Select-Object -ExpandProperty OSArchitecture) -eq "64-bit") { $arch = "amd64" }

    $downloadUrl = "https://github.com/golangci/golangci-lint/releases/download/v$GOLANGCI_LINT_VERSION/golangci-lint-$GOLANGCI_LINT_VERSION-windows-$arch.zip"

    $tmp = New-TemporaryFile
    Invoke-WebRequest -OutFile $tmp $downloadUrl
    Rename-Item -Path $tmp.FullName -NewName ($tmp.Name + ".zip")
    Expand-Archive -Path ($tmp.FullName + ".zip") -Force -DestinationPath ([System.IO.Path]::GetTempPath())
    Remove-Item -Path ($tmp.FullName + ".zip")

    $folderName = "golangci-lint-$GOLANGCI_LINT_VERSION-windows-$arch"
    $golangCILintPath = Join-Path ([System.IO.Path]::GetTempPath()) $folderName
    $GOBIN_PATH_FULL = $GOBIN_PATH + "\"

    if (!(Test-Path -Path $GOBIN_PATH)) {
        New-Item -Path $GOBIN_PATH -Type Directory
    }

    Move-Item -Path $golangCILintPath\golangci-lint.exe -Destination $GOBIN_PATH -Force
}

# Build the go-tool tool
Write-Output "[install] Building the go-ci-tool tool :)"

$VERSION_SUBSTITUION = "main.version=2.0.0"
$GO_MIN_VERSION_SUBSTITUTION = "github.com/ram-nad/go-monorepo/go-ci-tool/v2/constants.minGoVersion=$GO_MIN_VERSION"
$GOLANG_CI_LINT_VERSION_SUBSTITUTION = "github.com/ram-nad/go-monorepo/go-ci-tool/v2/constants.minGolangCILintVersion=$GOLANGCI_LINT_VERSION"

powershell -Command { $env:GOWORK="off"; go install -C go-ci-tool -trimpath -buildvcs=false -ldflags="-w -X $VERSION_SUBSTITUION -X $GO_MIN_VERSION_SUBSTITUTION -X $GOLANG_CI_LINT_VERSION_SUBSTITUTION . }
