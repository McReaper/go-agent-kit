#!/bin/sh
# Runs every check the CI runs. It must pass before a change is done.
set -u
KIT_BIN="${KIT_BIN:-/c/Users/victo/AppData/Local/Temp/claude/C--Users-victo-work-perso-t7-companion/67a948a2-b45c-40cd-9b71-4a0307f6987e/scratchpad/crash/bin}"
fail=0
for goos in linux windows; do
  GOOS=$goos golangci-lint run --config .kit/golangci.yml ./... || fail=1
  GOOS=$goos "$KIT_BIN/kitvet.exe" ./... || fail=1
done
go test ./... || fail=1
"$KIT_BIN/kit.exe" guard --base base || fail=1
if [ $fail -eq 0 ]; then echo "check.sh: all checks pass"; else echo "check.sh: FAILED"; fi
exit $fail
