# Grilling — Round 3

Frontier after round 2: rendering stack, process lifecycle, keyboard model,
interface language.

---

## Q3-A — How is the transcript rendered and kept fast?

Target: a 20k-message session opens in under a second, streaming stays under
8 ms per frame, scrolling never janks.

**A.** Own pipeline: `@tanstack/react-virtual` for rows, `react-markdown` +
`remark-gfm` + `remark-math` + `rehype-katex`, Shiki highlighting loaded lazily
per language, and a chunker that splits streaming text into blocks so only the
last block re-renders. Diffs rendered by our own two-column view.
**B.** A ready-made streaming markdown component (streamdown or similar).
**C.** Render each block in an isolated webview or iframe.

➡️ **A.** B caps styling and batching control, which is exactly where the app is
supposed to be better than its alternatives. C breaks selection, search and
scroll continuity.

**Answer: A — own rendering pipeline.**

---

## Q3-B — What happens to running work when a tab closes or the app quits?

**A.** One pi process per open session. Closing a tab with a live run asks
(stop / keep running in background); quitting with live runs shows one dialog
listing them (stop all and quit / cancel). Never silently resend a command that
may change files.
**B.** Kill every pi process on tab close and on quit.
**C.** Keep every session's pi alive in the background after the tab closes.

➡️ **A.** B loses work the user believes is running. C accumulates processes and
memory with no visible owner.

**Answer: A — explicit confirmation, never silently resend.**

---

## Q3-C — Keyboard semantics for the composer

**A.** pi semantics: Enter sends, or steers while running; Shift+Enter newline;
⌘Enter queues a follow-up; Esc clears the queue back into the editor, a second
Esc aborts; ⌘K palette, ⌘N new session, ⌘W close tab, ⌘1–9 switch, ⌘F search.
**B.** Enter always sends and queues as follow-up; steering needs an explicit
button.
**C.** Chat-app habits: ⌘Enter sends, Enter newline.

➡️ **A.** Users arrive from the pi TUI; matching its mental model costs nothing
in a GUI and keeps steering a first-class move.

**Answer: A — pi semantics.**

---

## Q3-D — Interface language?

**A.** English + Simplified Chinese from day one through mygo i18n catalogs,
default follows the system, switchable in settings.
**B.** English only.
**C.** Chinese only.

➡️ **A.** The owner writes Chinese; the repo is public and English-speaking
users are the majority of the audience. Retrofitting translation is the expensive
option, so the catalogs exist from the first screen.

**Answer: A — bilingual.**

---

## Settled

- Own transcript pipeline built on virtualisation, incremental chunking, Shiki.
- Per-session pi process, explicit confirmation on tab close and quit, no silent
  retries of file-changing commands.
- pi-compatible keyboard model.
- en + zh-Hans catalogs from the start.
