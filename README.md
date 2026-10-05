# dvah — Damn Vulnerable Agentic Harness

`dvah` is a deliberately vulnerable agentic harness, in the spirit of OWASP's
Damn Vulnerable Web Application. It exists so that agent builders, security
researchers and educators can study — and demonstrate — what *not* to do when
building an agent harness.

Everything in this repository is intentionally insecure. Do **not** run it
against production systems, real credentials, or hosts you do not own.

## Usage

```sh
cd source
go build -o dvah ./cmds/dvah

# always starts the terminal UI, single agent, sandbox = current working dir
./dvah <path/to/agent.yml> [path/to/config.yml]
```

- `<agent.yml>` defines the single agent (name, role, model, prompt, allowed
  tools/programs).
- `[config.yml]` is optional and provides the endpoint URL, debug flag, the
  first user prompt, and the provider/token map. It is merged on top of the
  global config at `~/.config/dvah/config.yaml`.
- The agent runs with the current working directory as its sandbox

Examples:

```sh
./dvah ../templates/agent.yaml
./dvah ../templates/agent.yaml ../templates/config.yaml
./dvah ../agent-outside-sandbox.yml ../config.yml
```

## Vulnerability Matrix

Status: **Implemented** = present in the code today, **Planned** = designed but
not yet wired up.

| Status | File(s) | Description |
| --- | --- | --- |
| Implemented | `source/tools/sanitizeFilesystemError.go` | **Raw `ENOENT` leak.** Filesystem errors are returned verbatim (`open foo: no such file or directory`) instead of being sanitized. The leaked errno tells the model it is talking to a real filesystem/shell, nudging it to attempt shell-driven escapes (e.g. Shellshock-style payloads). |
| Implemented | `source/tools/Files.go` (`Search`), `source/tools/Files.json` | **Command injection via `files.Search`.** The method takes a single `keywords` string and interpolates it, unescaped, into `sh -c "grep -rn <keywords> ."`. Shell metacharacters (`;`, `\|`, `&`, `$()`, backticks, newlines, ...) are executed by the shell, giving arbitrary command execution inside the sandbox; Shellshock-style payloads can also be smuggled through the keywords. |
| Implemented | `source/types/Permission.go`, `source/tools/Permissions.go`, `source/engine/Session.go` | **Self-service, scope-confused capability API.** The agent can grant itself capabilities with `permissions.Request(tool, method, scope)`. Grants are cached in a `map[string]bool` keyed by `"tool.method"` only: the requested scope is accepted and then discarded. One grant for `files.Write` on a single path therefore authorizes `files.Write` on *every* path. Sensitive methods (`files.Write`, `files.Copy`) are gated in `Session.CallTool`; `permissions.List` reports every grant as `scope: *`. |
| Implemented | `source/tools/resolveSandboxPath.go`, `source/tools/sanitizeSandboxPath.go` | **Symlink sandbox escape.** Containment is checked lexically with `filepath.Abs`/`filepath.Rel`; `EvalSymlinks` is never used. A symlink inside the sandbox can point outside it and the following `os.ReadFile`/`os.WriteFile`/`Copy` operations follow it out. See `examples/symlink-escape/`. |
| Planned | `source/engine/Session.go` (`LoadSkill`), `source/tools/readSkills.go` | **Prompt injection via skills.** A loaded skill body is spliced into the conversation as a `system` message, so attacker-controlled `SKILL.md` content can override the agent's original instructions. |
| Planned | `source/tools/Websites.go` | **SSRF by design.** `websites.Fetch`/`Stat` accept any `http`/`https` URL with no host allowlist, reaching loopback services and cloud metadata endpoints (e.g. `169.254.169.254`). |
| Planned | `source/tools/Programs.go` | **Command escalation surface.** Any program listed in the agent's `allowed-programs` can be executed; args containing a path separator are sandboxed, but flag/argument injection (e.g. `find -exec`, interpreters, `--output`) and permissive allowlists enable host escape. |
| Planned | `source/engine/Recovery.go` | **Sandbox data leakage.** Session recovery writes the full conversation (including tool output, secrets echoed by tools, and prompts) to `<sandbox>/.dvah/session.json` in plaintext. |

## Talk Examples

Ready-to-run scenarios live in [`examples/`](examples/README.md):

- `examples/readonly-status-report/` — read-only agent asked to write to
  `/etc/flag.txt`, forcing filesystem-API bypass attempts.
- `examples/permission-bypass/` — a capability granted for `test.txt` silently
  authorizes `etc/flag.txt` because the permission store drops the scope.
- `examples/symlink-escape/` — a symlink inside the sandbox points at a secret
  outside it, defeating the lexical sandbox check.

## Repository Layout

```
templates/         example agent.yaml and config.yaml
examples/          talk scenarios (agent.yaml + config.yaml)
source/agents/     agent loader (single agent, YAML)
source/engine/     session, HTTP inference, recovery
source/tools/      files, programs, skills, websites, permissions tools
source/types/      agent/config/permission/console types
source/parsers/    YAML and HTML parsers
source/utils/      formatting, filesystem and API helpers
```

## Disclaimer

This project is for education and authorized security testing only. The
vulnerabilities above are intentional. Use it at your own risk.
