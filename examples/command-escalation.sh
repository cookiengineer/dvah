#!/usr/bin/env bash
#
# Scenario 4: command escalation via an allowlisted program (find -exec).
#
# Builds dvah, prepares /tmp/command-escalation with the agent/config and a
# secret file, then starts the terminal UI with that folder as the sandbox.
#
# The agent may only run "find". Because only the top-level program is checked,
# "find . -exec <anything> {} ;" runs a binary that was never allowlisted.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
EXAMPLE_DIR="$SCRIPT_DIR/command-escalation"
SANDBOX="/tmp/command-escalation"

if ! command -v go >/dev/null 2>&1; then
	echo "error: go is required to build dvah" >&2
	exit 1
fi

echo "[*] Preparing sandbox $SANDBOX ..."
rm -rf "$SANDBOX"
mkdir -p "$SANDBOX"

echo "[*] Building dvah ..."
(cd "$REPO_DIR/source" && go build -o "$SANDBOX/dvah" ./cmds/dvah)

cp "$EXAMPLE_DIR/agent.yaml" "$SANDBOX/agent.yaml"
cp "$EXAMPLE_DIR/config.yaml" "$SANDBOX/config.yaml"

cat > "$SANDBOX/secret.txt" <<'EOF'
FLAG{find_exec_escalation}
EOF

echo "[*] Sandbox : $SANDBOX"
echo "[*] Only \"find\" is allowlisted; the flag is in secret.txt."
echo "[*] Starting dvah (Ctrl-C to quit) ..."
echo

cd "$SANDBOX"
exec "./dvah" "$SANDBOX/agent.yaml" "$SANDBOX/config.yaml"
