#! /usr/bin/env bash

# check if git tree is clean
git_tree_state=dirty
if git_status=$(git status --porcelain --untracked=no 2>/dev/null) && [[ -z "${git_status}" ]]; then
    git_tree_state=clean
fi

version_pkg=github.com/dims/maintainers/pkg/version
bin_name=maintainers

MAINTAINERS_INSTALL_PATH="${GOPATH}/bin"

if [ -n "${GOBIN:-}" ]; then
    MAINTAINERS_INSTALL_PATH="${GOBIN}"
fi

if [[ ":$PATH:" != *":$MAINTAINERS_INSTALL_PATH:"* ]]; then
    # setup install path
    export PATH="${PATH}:${MAINTAINERS_INSTALL_PATH}"
fi

# build maintainers
go build -v -trimpath -ldflags "-s -w \
-X $version_pkg.buildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
-X $version_pkg.gitCommit=$(git rev-parse HEAD 2>/dev/null || echo unknown) \
-X $version_pkg.gitTreeState=$git_tree_state \
-X $version_pkg.gitVersion=$(git describe --tags --abbrev=0 || echo unknown)" \
-o "$MAINTAINERS_INSTALL_PATH/$bin_name" . \
|| exit 1

echo "$bin_name installed to $MAINTAINERS_INSTALL_PATH"
