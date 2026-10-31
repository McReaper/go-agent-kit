#!/bin/sh
# Runs the tests of the packages a change touches.
set -eu
for pkg in "$@"; do
  go test "./$pkg/..."
done
