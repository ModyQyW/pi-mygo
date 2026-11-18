# Verified facts about this machine and its toolchain

Checked 2026-11-18 on the machine the app is being built on. Every line was
produced by a command; none is inferred from documentation alone.

## Toolchain

```console
$ go version            go version go1.27.4 darwin/arm64
$ bun --version         1.4.4
$ node --version        v26.13.1
$ pi --version          1.5.2
$ npm view @earendil-works/pi-coding-agent version   1.5.2
$ xcode-select -p       /Applications/Xcode.app/Contents/Developer
$ security find-identity -v -p codesigning   → 0 valid identities found
```

MyGo v0.10.0 is the version in the module cache
(`~/go/pkg/mod/github.com/egoist/mygo@v0.10.0`), which is what the project pins.

## pi RPC behaves like this against the real installation

A `get_state` + `get_commands` round trip over `pi --mode rpc --no-session`
returned, before any command response:

- five `extension_ui_request` records from the user's own extensions
  (`setWidget` ×2, `setStatus` ×3, `notify` ×1), emitted unprompted at startup;
- `setStatus.statusText` containing raw ANSI escapes
  (`\u001b[38;2;135;144;149m…`), because those extensions were written for a TUI.

Implications:

1. The client must treat fire-and-forget UI records as normal traffic and strip
   ANSI before rendering.
2. Status text is not a stable API; it must never block session startup.

Closing the reader while pi is alive produced an unhandled `EPIPE` crash in the
child (`node:events:505 ... write EPIPE`) instead of a clean exit.

Implication: the Go supervisor owns shutdown. Close stdin, keep draining stdout
until the process exits, then stop reading. Killing the pipe first crashes pi and
may leave a session file mid-write.

`pi --mode rpc` accepts `--session <path|id>`, `--session-id <id>`,
`--session-dir <dir>`, `--name`, `--provider`, `--model`, `--thinking`,
`--extension`, `--no-extensions`, `--no-session`. RPC has no command to list
sessions, log in, read or write settings, or browse files.

Authentication:

```console
$ pi auth login <provider> [--api-key] [--no-browser]   # docs: needs an interactive terminal
$ pi auth check --provider <p> --json                   # exit 0 ready / 1 not ready / 2 invalid
$ pi auth logout <provider>
$ pi auth print-api-key --provider <p>
```

## Update signing

`mygo keygen [-o dir]` writes `mygo-update.key` (secret) and `mygo-update.pub`.
`mygo build -upload` signs per-platform update manifests, and with
`updates.github` creates a **draft** GitHub release tagged with the version.
Signing identity and update signing are unrelated: an ad-hoc signed macOS app can
still update itself, because `install.sh` verifies the Ed25519 manifest signature
and `curl` downloads carry no quarantine flag.

Windows: an unsigned installer triggers SmartScreen. Linux: the install script
installs per user without root, and needs WebKitGTK present.

## Session files

Sessions live in `~/.pi/agent/sessions/--<escaped-cwd>--/<timestamp>_<id>.jsonl`,
version 3, JSONL, a header with `cwd` and a tree of entries with `id`/`parentId`.
The app reads them to build the project list and never writes them.

## What this environment cannot do

- `screencapture` fails with "could not create image from display": the agent's
  process has no Screen Recording permission. Visual verification of the window
  therefore has to come from the owner, or from the mygo automation plugin once
  it is added. Process-level verification (`Pi MyGo Dev` appears as a foreground
  app, Vite serves the bundle) is available and is what the M0 report used.
- There is no Apple code-signing identity, so `mygo build` signs ad hoc. That is
  enough to run locally and to install from a `curl`-downloaded update, but
  Gatekeeper will warn on a downloaded zip on another Mac.
