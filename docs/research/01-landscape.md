# Landscape: what exists, what we can build on

Verified on 2026-11-18. Facts only; decisions live in `../grilling/`.

## Pi coding agent

- Version installed in this machine's Pi Agent.app: `@earendil-works/pi-coding-agent` 1.5.2. npm `latest` is also 1.5.2.
- Install paths: `curl -fsSL https://pi.dev/install.sh | sh` (self-contained binary), or `npm i -g @earendil-works/pi-coding-agent` (needs Node >= 22.19).
- Three integration surfaces (from `docs/rpc.md`, `docs/sdk.md`, `docs/cli-integration.md` in the 1.5.2 tarball):
  1. **TUI** — the editor you run in a terminal.
  2. **RPC** — `pi --mode rpc`, long-lived child process, JSONL on stdin/stdout. Language-agnostic.
  3. **SDK** — `createAgentSession()` in-process, Node/Bun only.
- Sessions are JSONL files: `~/.pi/agent/sessions/--<escaped-cwd>--/<timestamp>_<id>.jsonl`. Version 3 header, tree entries with `id`/`parentId`. Any client can read and list them from disk.
- Auth lives in `~/.pi/agent/auth.json`. `/login` is an **interactive TUI command** and is not part of RPC's command set (RPC's `get_commands` explicitly excludes built-in TUI commands).
- The RPC protocol has ~35 commands (prompt/steer/follow_up/abort/clear_queue, session ops incl. `get_tree`/`fork`/`clone`/`switch_session`, model + thinking level, compaction, retry, bash, stats, export), a delta-only event stream, and an `extension_ui_request`/`extension_ui_response` sub-protocol for extension dialogs.
- Notable RPC gaps: no session listing, no login, no settings read/write, no file tree, no git. All of those a desktop client must supply itself.

## Pi GPUI (reference, Rust)

- `github.com/bytesbay/pi-gpui`, MIT, 26,820 lines of Rust across 45 files, GPUI 0.2.2.
- Architecture: one `pi --mode rpc` child process **per open session**; JSONL parsed on a thread; owned bridge extension `extensions/pi-gpui-bridge.ts` passed with `--extension` for what RPC lacks.
- Ships its own markdown renderer (pulldown-cmark), syntax highlighting (syntect + two-face), diff view, text selection, code editor, dock/panel system, single-instance, updater, settings JSON.
- Documented principles worth copying (`docs/APP_PRINCIPLES.md`): the agent engine stays in pi; never install/update pi silently; version range with `MIN_PI`/`TESTED_PI`; no shell string assembly; IDs correlate commands; read stdout continuously.
- Reported: 16 MB app, ~0.4 s window, ~75 MB idle. Requires pi installed; no bundled runtime.

## Pi Web (what the user runs today)

- `@agegr/pi-web` 0.15.0, 8.3k stars, MIT. The "Pi Agent.app" on this machine is this package wrapped in **Tauri 2** (`src-tauri/pi-agent-desktop-package.json`).
- Architecture: Next.js 16 app server on `127.0.0.1:30141` + Tauri webview; imports `@earendil-works/pi-*` **as libraries** (SDK in-process), so it bundles pi and its own Node runtime; `node-pty` + `xterm` for a real terminal.
- Frontend: React 19, Tailwind 4, Radix, `react-markdown` + remark-gfm/math/frontmatter/cjk-friendly, KaTeX, Mermaid, `react-syntax-highlighter`, `ansi_up`, `@lobehub/streamdown`.
- Feature bar: session workspace grouped by project, resume/rename/export/delete, in-session branch + independent new session, file browser with git diffs and previews (text, md, image, audio, PDF, DOCX), git worktree switching, provider login and API keys in-app, model tests, plugin packages, skills, i18n (en/zh-Hans/zh-Hant/ja/ru).
- Cost of that architecture: a local HTTP server, a Next.js process, Node, and pi's own Node dependencies all inside the app.

## MyGo 0.10.0

- `github.com/egoist/mygo`, MIT, 6.5k stars, 416 forks, 180 open issues, created 2026-09-26, releases roughly weekly (0.8.3 → 0.10.0 inside two weeks). Changelog states "no breaking changes to the Go API" for 0.10.
- Pure Go, **no cgo**, one binary for macOS/Windows/Linux, plus iOS/Android since 0.10.
- Two window kinds, mixable:
  - **Web page** in the OS webview (WKWebView / WebKitGTK / WebView2). Dev = Vite dev server + HMR; build = frontend embedded via `mygo://localhost`.
  - **Native UI** in Go (`package ui`), GPU-drawn, no webview, no JS.
- Typed IPC: `mygo.Bind(service)` + `mygo generate` emits a TypeScript client with TSDoc from Go comments; `*mygo.Channel[T]` parameters stream values per call; `mygo.NewEvent[T]` for pushed events; `context.Context` cancels when the page goes away; `mygo.CallerWindow(ctx)` for dialogs.
- Desktop APIs: windows, menus, tray, dialogs, notifications, global shortcuts, deep links, file associations, single instance, `App.Path` dirs, preload scripts, custom protocols, dropped files, dark mode, accent color, accessibility.
- Native UI widgets that matter to us: `ui/markdown` (CommonMark + GFM + math, built for **LLM streaming** with per-token parse, even pacing and fade), `ui` code editor, diff viewer, `ui.List` (builds only visible rows; 1M-row list = 14.5 µs/frame), tree, split view, tabs, sidebar, command palette-ish popovers, rich text editor, charts.
- Plugins: fetch, websocket, sqlite, watch, updater, terminal (libghostty-vt, native UI only), glass, game, loading, sfsymbols, automation, permissionflow, render.
- Distribution: `mygo build` → signed `.app` + `.dmg`, Windows installers, deb/rpm/PKGBUILD, notarization, signed auto-updates with delta archives, `darwin/universal`.
- Toolchain on this machine: Go 1.27.4, Bun 1.4.4, Node 22.23.4, macOS 27.2 (arm64). Requirement is Go 1.27+ and Bun for web frontends.

## Why this matters

Three client architectures are already proven for pi. The open question is not "can a desktop pi client exist" but "which shell gives the best app per unit of work". MyGo answers the shell question with a Tauri-class feature set in Go and no cgo; the frontend question is separate and is the first decision to settle.
