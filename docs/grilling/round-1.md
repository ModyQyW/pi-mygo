# Grilling — Round 1

Date: 2026-11-18. Frontier = decisions whose prerequisites are settled by the
facts in `../research/01-landscape.md`. Each question ends with my recommendation;
answering "yes" accepts it.

---

## Q1 — How does the app talk to pi?

Two ways, and they are not close in cost:

**A. `pi --mode rpc` child process (pi-gpui model).** Go spawns the user's
installed pi, speaks JSONL, one process per open session.

**B. Embed `@earendil-works/pi-coding-agent` as a library** inside a Node/bun
sidecar that the Go app supervises (pi-web model, which is what Pi Agent.app
already is).

What B buys: `createAgentSession()`, `SessionManager`, in-process login/OAuth,
custom tools, settings APIs, cumulative event snapshots, no installed-pi
requirement. What B costs: 100–200 MB of Node runtime and pi's dependency tree
inside our app, slower cold start, and we rebuild pi-web with a Go shell around
it — at which point mygo earns almost nothing and Tauri would be the honest shell.

What A costs: pi must be installed and authenticated by the user; no login UI, no
session listing, no settings API, no file/git helpers — Go implements those
(session list = scan `~/.pi/agent/sessions/*/*.jsonl`; settings = read/write
JSON files; git = shell out to `git`). A gets process isolation per session,
the user's real extensions/skills/models, and a ~20 MB app.

➡️ **A.** RPC to the installed pi, with a version gate (`MIN_PI`/`TESTED_PI`) and
feature detection instead of hard failure. Bundling is a possible later
"portable" build, not v1.

---

## Q2 — React + shadcn in the webview, or mygo native UI in Go?

**A. Web frontend** (React 19 + Vite + Tailwind 4 + shadcn/ui) in the OS webview.
The hard rendering parts come off the shelf: streaming markdown, syntax
highlighting, diff, KaTeX/Mermaid, virtualized transcripts, image/PDF preview,
command palette, accessible dialogs. Iterating on visual polish is cheap.
Cost: ~60–75 MB idle, webview rendering is Safari-class on macOS (no Chromium
APIs), a few hundred ms of extra boot.

**B. Native UI in Go** (`package ui`). MyGo already ships the three hardest
widgets for exactly this app: `ui/markdown` with a *streaming* mode built for
LLM replies (per-token parse, even pacing, fade), a code editor, and a diff
viewer, plus `ui.List` that builds only visible rows. Result would be ~15 MB,
instant, no webview, and genuinely distinct from the Electron/Tauri clients.
Cost: everything else is mine to write in Go — composer with IME/paste/
attachments/autocomplete, panels, focus, theming, animation — several thousand
lines against a 0.x API that ships roughly weekly, with no CSS and no ecosystem,
and visual quality degrades to what I can hand-code.

pi-gpui needed 26,820 lines of Rust to reach its current polish with native
rendering. The same surface in React is a fraction of that, which is the whole
argument.

➡️ **A — React + shadcn/ui in the webview**, with the mygo shell supplying
native menus, tray, dialogs, notifications, deep links, single instance and
updates. The user asked whether shadcn is justified: yes, for accessible
primitives (dialog, popover, command, toast) that would otherwise be hand-built
and hand-debugged, and it keeps the design tokens in one place.

---

## Q3 — What must the first version actually do?

**A. Agent workspace (recommended).** Projects and their pi sessions in a
sidebar; open several sessions at once, each with its own pi process; streaming
transcript with markdown, code, diffs and tool-call cards; attachments by paste
and drop; steering and follow-up queue; stop; model + thinking-level picker;
slash commands from pi (`get_commands`); fork/clone/session tree; compaction
shown inline; context/cost meter; rename/delete/export session; keyboard-first
(⌘K palette, ⌘N, ⌘W); settings for the app's own options.

**B. One polished session.** No sidebar, no multi-session, no tree. Ships sooner,
but it is a demo, not a workspace — pi's own value is working across projects.

**C. pi-web parity.** A plus a file browser with git diff, embedded terminal
(xterm + a pty from Go), in-app provider login and API keys, plugin packages,
worktrees, five UI languages.

➡️ **A for v0.1**, with C's items as explicit later releases. One exception I
will put to you in round 2: in-app login, because "install pi and then run
`/login` in a terminal" is a bad first-run experience.

---

## Q4 — Audience and distribution ambition?

**A. Public open-source app, macOS (Apple silicon) first**, Linux next, Windows
only if mygo's WebView2 path proves painless. Code stays cross-platform from day
one (mygo gives that for free); the release pipeline, signing and docs start
with macOS.

**B. Personal tool.** No signing, no notarization, no updater, no docs — fastest
to something usable, worthless to anyone else.

**C. Cross-platform public release from v0.1**, including Windows installers and
signed auto-updates on all three.

Signing/notarization needs an Apple Developer certificate; auto-update needs
release assets in a public repo. Both are decisions about money and process, not
code.

➡️ **A.** Public, macOS-first, MIT, signed + notarized if you have a certificate
(otherwise ad-hoc builds and a "build from source" path), auto-update via mygo's
updater once releases exist.


---

## Answers (2026-11-18)

- **Q1 → A.** RPC to the installed pi. No Node sidecar.
- **Q2 → A.** React 19 + Vite + Tailwind 4 + shadcn/ui in the mygo webview.
- **Q3 → A.** Agent workspace for v0.1.
- **Q4 → A + correction.** Public OSS, but the user overrode "macOS only first":
  cross-platform and auto-update are both in scope, no Apple certificate exists,
  so builds are ad-hoc signed and updates rely on mygo's Ed25519 manifests.
  Repo: `github.com/ModyQyW/pi-mygo`, created and pushed as `ModyQyW`.
