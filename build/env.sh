#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
tstdir="$workspace/src/github.com/utereum"
if [ ! -L "$tstdir/go-utchain" ]; then
    mkdir -p "$tstdir"
    cd "$tstdir"
    ln -s ../../../../../. go-utchain
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$tstdir/go-utchain"
PWD="$tstdir/go-utchain"

# Launch the arguments with the configured environment.
exec "$@"
