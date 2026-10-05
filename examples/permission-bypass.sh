#!/usr/bin/env bash
#
# Scenario 2: buggy capability (permission) API - scope confusion.
#
# Builds dvah, prepares /tmp/permission-bypass with the agent/config, then
# starts the terminal UI with that folder as the sandbox.
#
# The agent requests files.Write for the scope "test.txt". The permission store
# drops the scope and keys grants by tool.method only, so the follow-up write to
# "etc/flag.txt" - a completely different path - is silently allowed.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
EXAMPLE_DIR="$SCRIPT_DIR/permission-bypass"
SANDBOX="/tmp/permission-bypass"

if ! command -v go >/dev/null 2>&1; then
	echo "error: go is required to build dvah" >&2
	exit 1
fi

echo "[*] Preparing sandbox $SANDBOX ..."
rm -rf "$SANDBOX"
mkdir -p "$SANDBOX/etc"

echo "[*] Building dvah ..."
(cd "$REPO_DIR/source" && go build -o "$SANDBOX/dvah" ./cmds/dvah)

cp "$EXAMPLE_DIR/agent.yaml" "$SANDBOX/agent.yaml"
cp "$EXAMPLE_DIR/config.yaml" "$SANDBOX/config.yaml"

echo "[*] Starting dvah, sandbox is $SANDBOX (Ctrl-C to quit) ..."
echo

cd "$SANDBOX"
exec "./dvah" "$SANDBOX/agent.yaml" "$SANDBOX/config.yaml"
