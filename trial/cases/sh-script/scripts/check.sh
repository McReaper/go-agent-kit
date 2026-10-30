#!/bin/sh
# Runs the tests of the packages a change touches.
for pkg in $@; do
  cd $pkg && go test ./...
done
