# Implementation plan

## Definition of done for v0.1

A person with pi installed and signed in can: open the app, see their existing
projects and sessions, start a new session, talk to the agent with streaming
markdown, watch tool calls, steer or queue a follow-up while it runs, stop it,
switch model and thinking level, branch from any message, search the transcript,
rename and export, and change the language, theme and update channel. They can do
all of that with the keyboard, and the app never corrupts pi's own state.

## Phases

Each phase ends with a gate. A phase is not done because code exists; it is done
when the gate's check passes.

### M0 — Foundations

Repo, module, scaffold, window, menu, tokens, i18n, CI skeleton.

- [ ] `mygo init`-shaped project: `main.go`, `mygo.config.ts`, `package.json`,
      `vite.config.ts`, `src/`, `resources/icon.png`, `locales/`.
- [ ] Window: hidden title bar, traffic lights inset, 44 px draught titlebar
      drawn by the page, min 900×600, default 1240×840, restored from state.
- [ ] Menu bar with roles and accelerators; About panel; Quit. Settings moves
      to M3 with the rest of the settings surface, so no menu item is added for
      it here.
- [ ] Design tokens in `src/styles/tokens.css`, light and dark; typography;
      primitives (Button, IconButton, Chip, Field, Sheet, Tooltip, Toast).
- [ ] i18n catalogs `en` + `zh-Hans`, generated client works, language switches
      at runtime.
- [ ] `go vet ./...`, `go test ./...`, `bun run typecheck`, `bun run build` all
      green; `mygo dev` opens the window.

Gate: window shows the paper shell, the environment card reports the real pi
install, dark mode follows the system, `prefers-reduced-motion` removes motion.

### M1 — pi client and the first real conversation

- [ ] `piagent`: resolve executable (`PI_EXECUTABLE`, `PATH`, login shell),
      parse `--version` with a timeout, apply the version gate.
- [ ] RPC client: spawn with `--mode rpc`, LF-only framing, request ids,
      continuous stdout drain, stderr capture ring, orderly shutdown (close
      stdin → drain → wait), kill after 5 s.
- [ ] Session scan: read session headers, group by `cwd`, cache by
      (path, mtime, size), sort by recency.
- [ ] Hub: open/close/new session, one process per session file, snapshot +
      seq patches, reducer for message/turn events.
- [ ] Sidebar with projects and sessions; open and close sessions; new session in
      a project.
- [ ] Transcript: virtualised list, user/assistant text, streaming deltas,
      markdown + code highlighting, scroll anchoring.
- [ ] Composer: send, stop, multiline, drafts restored per session.
- [ ] Error cards for missing/old pi and connection failures.

Gate: a real conversation with the installed pi streams at 60 fps; killing pi
externally shows a crashed banner and "Reopen" works; reopening the app restores
the last session and its draft.

### M2 — The working loop

- [ ] Tool cards with per-tool summaries, smart collapse, diff preview for
      `edit`/`write`, tail for `bash`, auto-expand on error.
- [ ] Thinking blocks, collapsed by default.
- [ ] Images: inline display of message and tool-result images; paste and drop
      become base64 blocks; other files become `@path`.
- [ ] Model and thinking pickers, per session, searchable, with favourites.
- [ ] Slash commands from `get_commands`, `@` file completion, `/skill:`.
- [ ] Steering and follow-up: Enter vs ⌘Enter, queue chips, Esc clears the queue
      into the editor, second Esc aborts.
- [ ] Extension UI: `select`, `confirm`, `input`, `editor` dialogs inline;
      `notify` → toast; `setStatus` → status chips with ANSI stripped.

Gate: a run with five tool calls including a failed one, a pasted screenshot, a
steer and a follow-up completes without losing or reordering text.

### M3 — History and control

- [ ] Session tree drawer: `get_tree`, branch glyphs, labels, keyboard
      navigation, switch/fork/clone, branch summaries.
- [ ] In-session search: ⌘F, highlight, n/N, auto-expand matching tool cards.
- [ ] Compaction shown inline with token counts; manual compact action.
- [ ] Context gauge, token totals, session cost.
- [ ] Rename, export HTML, delete to trash, reveal session file.
- [ ] Settings: appearance and language, model defaults, input and transcript,
      update and diagnostics.
- [ ] Structured error taxonomy with actions (retry, sign in, compact, switch
      model, copy diagnostics).

Gate: a 5,000-message session opens in under a second, search returns in under
150 ms, forking from message 3 produces a new session that continues correctly.

### M4 — Getting it to people

- [ ] First run: pi missing / old / unauthenticated cards, `pi auth login` in the
      system terminal, polling `pi auth check --json`, automatic continuation.
- [ ] `mygo keygen`, public key committed, private key only in CI secrets,
      `mygo build -upload` creating draft releases.
- [ ] CI matrix: macOS arm64 (verified), Windows x64, Linux x64 (beta).
- [ ] Icon, app bundle metadata, README (en + zh-CN), CONTRIBUTING, AGENTS.md.
- [ ] Update flow: check, download, "Restart to update", never update pi itself.

Gate: a tagged release produces three platform artifacts and a draft GitHub
release; a second machine installs from `install.sh` and updates to the next tag.

### M5 — Polish

- [ ] Motion pass against the design system; reduced-motion audit.
- [ ] Accessibility audit: focus order, names, contrast, screen-reader log.
- [ ] Performance pass against the budgets in `01-architecture.md`.
- [ ] Empty, loading and error states for every surface.
- [ ] Keyboard-only walkthrough of every flow.
- [ ] Cross-platform smoke test notes for Windows and Linux.

Gate: the checklist in `02-design-system.md` passes on macOS with keyboard only.

## Testing

| Layer | Tool | What must be covered |
| --- | --- | --- |
| RPC framing | `go test` | LF split with `U+2028` inside strings, CRLF, partial reads, EPIPE-safe shutdown |
| Reducer | `go test` + recorded fixture | text/thinking/tool/compaction/retry sequences, interrupted tools, out-of-order tool results |
| Version gate | `go test` | table of pi versions around MIN/TESTED |
| Session scan | `go test` with temp dirs | escaped cwd names, broken JSONL, 1k files timing |
| Error classes | `go test` | real pi/provider error strings |
| Settings | `go test` | atomic write, corrupt file recovery, unknown fields preserved |
| Frontend logic | `vitest` | markdown chunker, tool summary formatting, search offsets, queue state |
| Shell flows | Playwright against Vite with a fake `window.mygo` | send/stop, palette, drawer, settings, empty states |
| Manual | macOS screenshots | visual checks the design system lists |

Fixtures: one recorded, sanitised RPC session committed under
`internal/hub/testdata/` so reducer tests replay real traffic.

## Risks

| Risk | Mitigation |
| --- | --- |
| mygo 0.x API churn | Pin v0.10.0; upgrade deliberately with a green gate |
| pi RPC changes | Version gate + feature detection; the reducer owns the vocabulary |
| WKWebView vs WebView2/WebKitGTK differences | Beta label, platform smoke notes, no Chromium-only APIs |
| Large sessions | Virtualisation, Go-side index, memory budget |
| Extensions that assume a TUI | Fire-and-forget UI records are tolerated, ANSI stripped, ignored methods counted |
| Unsigned Windows installer | Documented SmartScreen note; signing deferred until a certificate exists |
| Contract drift between Go and TS | `mygo generate` on every dev build; TypeScript fails the build |

## Working rules for this repo

- Every behaviour change updates the docs in `docs/plan` when it changes a
  decision, and `docs/log/mistakes.md` when it fixes a mistake.
- No feature lands without its empty, loading and error state.
- No new dependency without a line in the PR explaining what breaks without it.
- Performance budgets are checked before a phase is called done.
