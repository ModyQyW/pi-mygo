# Settled decisions

One table, no alternatives. Each row is the outcome of a grilling round; the
reasoning lives beside it in `round-N.md`. Changing a row here is a decision, not
an edit — record why in `../log/mistakes.md` if it was a mistake.

| # | Area | Decision |
| --- | --- | --- |
| 1 | pi integration | Spawn the installed `pi --mode rpc` as a child process, JSONL over stdin/stdout, one process per open session. No Node sidecar, no bundled runtime. |
| 2 | pi support range | `MIN_PI` 1.5.0, `TESTED_PI` 1.5.2. Same major only, newer minors run with a "not tested" note. Feature-detect inside the range; never refuse silently. |
| 3 | Frontend | React 19 + Vite + Tailwind 4 + shadcn/ui primitives inside the mygo WKWebView. |
| 4 | Visual language | Pi paper: warm paper surfaces, ink text, brand colours only for state/focus/charts, corner brackets on panels, mono for metadata. Light and dark. |
| 5 | Scope v0.1 | Agent workspace: projects + sessions sidebar, several live sessions, streaming transcript, tool cards, attachments, steering/follow-up/stop, model + thinking, slash commands, tree drawer, fork/clone, compaction, context/cost, search, rename/export. |
| 6 | Out of v0.1 | File browser, git panel, embedded terminal, in-app login, plugin packages, worktrees, multi-window, MCP management, schedules, remote/SSH. |
| 7 | First run | Guided checks (pi present, version, auth) with actions; login runs `pi auth login <provider>` in the system terminal and polls `pi auth check --json`. The app never writes pi credentials. |
| 8 | Projects | Auto-discovered from `~/.pi/agent/sessions/*/*.jsonl` headers by `cwd`, sorted by recency, plus manual "Add project". |
| 9 | Release | macOS arm64 verified; Windows x64 + Linux built in CI as beta; signed mygo updates on all platforms; no Apple certificate, so ad-hoc signing. |
| 10 | Transcript | `@tanstack/react-virtual` + `react-markdown`/remark/rehype + KaTeX + Shiki (lazy per language) + own incremental block chunker + own diff view. |
| 11 | Process lifecycle | Close tab with a live run → ask stop/background. Quit with live runs → one dialog listing sessions. Never resend a file-changing command automatically. |
| 12 | Keyboard | Enter send/steer, ⇧Enter newline, ⌘Enter follow-up, Esc clears queue then aborts, ⌘K palette, ⌘N new, ⌘W close, ⌘1–9 switch, ⌘F search, ⌘⇧T tree. |
| 13 | Language | English + Simplified Chinese through mygo i18n catalogs; system default, switchable. |
| 14 | Session tree | Right-side drawer overlay. |
| 15 | Tool cards | Smart collapse: one summary row; diff preview for edit/write; last five lines for bash; errors auto-expand. |
| 16 | Settings | Appearance and language; model defaults; input and transcript; update and diagnostics. |
| 17 | Naming | Display `Pi MyGo`, bundle id `top.modyqyw.pimygo`, binary `pi-mygo`, repo `ModyQyW/pi-mygo`. |
| 18 | Errors | Structured error cards with plain-language class, actions, and folded raw text. |
| 19 | Attachments | Images become base64 image blocks; other files insert `@path`; ⌥-drop forces path. |
| 20 | Update keys | `mygo keygen` locally, private key only in GitHub Actions secrets, public key committed, releases created as drafts. |
