#!/usr/bin/env bash
set -euo pipefail

golangci_lint_version="1.54.2"
golangci_lint_sha256="unknown" # set in platform block below

goarch=amd64 # it's 2022
goos="unknown"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
  goos="linux"
  golangci_lint_sha256="762ef7c877d9baa4a3ffcc69c88ecf35faf47cd76c1394792d5fecc15f6dc84b"
elif [[ "$OSTYPE" == "darwin"* ]]; then
  goos="darwin"
  golangci_lint_sha256="04d936f68895a9127999fdfa78872a3245d89dd6900c147dc9106c06870b9c5b"
fi

srcdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null 2>&1 && pwd )"

if [ -f "$srcdir/bin/golangci-lint-${golangci_lint_version}" ]; then
    echo "--> Already downloaded"
    exit 0
fi

workdir=$(mktemp -d)

function cleanup {
  rm -rf "$workdir"
}
trap cleanup EXIT

echo "--> Downloading"
curl -sLo "$workdir/download.tgz" "https://github.com/golangci/golangci-lint/releases/download/v${golangci_lint_version}/golangci-lint-${golangci_lint_version}-${goos}-${goarch}.tar.gz"

echo "--> Unpacking"
cd "$workdir"
tar -zxf "$workdir/download.tgz"
mv golangci-lint*/golangci-lint .

echo "--> Verifying"
echo "$golangci_lint_sha256 *golangci-lint" | shasum -a 256 -c -

mkdir -p "$srcdir/bin"
mv golangci-lint "$srcdir/bin/golangci-lint-${golangci_lint_version}"
echo "--> Fetched bin/golangci-lint-${golangci_lint_version}"
