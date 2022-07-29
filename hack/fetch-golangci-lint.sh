#!/usr/bin/env bash
set -euo pipefail

golangci_lint_version="1.47.2"
golangci_lint_sha256="unknown" # set in platform block below

goarch=amd64 # it's 2022
goos="unknown"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
  goos="linux"
  golangci_lint_sha256="2fbd1b5d3c1cda2e5d76d73f40509dd5bf809c60f08a7bce9e6b2bd5612340c0"
elif [[ "$OSTYPE" == "darwin"* ]]; then
  goos="darwin"
  golangci_lint_sha256="337735285ae5e6d93e298c7e7d404534c3aabddb3628436c53b53b5688c2cd5d"
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
