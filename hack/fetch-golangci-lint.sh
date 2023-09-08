#!/usr/bin/env bash
set -euo pipefail

golangci_lint_version="1.54.2"
golangci_lint_sha256="unknown" # set in platform block below

goarch=amd64 # it's 2022
goos="unknown"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
  goos="linux"
  golangci_lint_sha256="17c9ca05253efe833d47f38caf670aad2202b5e6515879a99873fabd4c7452b3"
elif [[ "$OSTYPE" == "darwin"* ]]; then
  goos="darwin"
  golangci_lint_sha256="925c4097eae9e035b0b052a66d0a149f861e2ab611a4e677c7ffd2d4e05b9b89"
fi

srcdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null 2>&1 && pwd )"

if [ -f "$srcdir/bin/golangci-lint-${golangci_lint_version}" ]; then
    echo "--> Already downloaded"
    exit 0
fi

workdir=$(mktemp -d)

function cleanup {
  echo $workdir
echo foo #  rm -rf "$workdir"
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
