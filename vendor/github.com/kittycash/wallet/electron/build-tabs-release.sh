#!/usr/bin/env bash
set -e -o pipefail

# installs the node modules for the kittycash tab app
# NOT for the electron build process

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

pushd "$SCRIPTDIR" >/dev/null

cd ../tabs/
yarn
yarn run build
popd >/dev/null
