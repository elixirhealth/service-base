#!/usr/bin/env bash

set -eou pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GIT_HOOKS_DIR="${DIR}/git-hooks"

for filepath in ${GIT_HOOKS_DIR}/* ; do
    filename=$(basename ${filepath})
    cp ${filepath} .git/hooks/${filename}
    echo "installed ${filename}"
done
