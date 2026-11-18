# Pi MyGo — working notes for agents

Read `docs/plan/00-plan.md` and `docs/grilling/decisions.md` before changing
anything. Decisions there are settled; if a change contradicts one, say so
explicitly instead of quietly diverging.

## Non-negotiables

- The app never writes pi's session files, credentials or settings. Session
  mutations go through RPC commands or `pi` CLI calls.
- At most one pi process per session file.
- pi shutdown order: close stdin → keep draining stdout → wait → then stop
  reading. Killing the read end crashes pi with an unhandled EPIPE.
- Fire-and-forget `extension_ui_request` records are normal traffic. Strip ANSI.
  Never block startup on them.
- `agent_settled` ends work, not `agent_end`.
- Every UI surface ships loading, empty and error states. No spinner alone.
- Text deltas are never dropped or reordered; a delta is superseded only by a
  final snapshot of the same content.

## Layout

- `internal/piagent` — pi process, RPC framing, version gate, session scan. No UI.
- `internal/hub` — sessions, event reduction, patches, search. No mygo imports.
- `internal/appstore` — settings, drafts, UI state. Atomic writes.
- `internal/api` — bound services and events. Validation only, no logic.
- `src/` — React app. Talks to Go only through the generated client.

Dependencies point one way: `api → hub → piagent`, `api → appstore`.

## Conventions

- Go: standard formatting, errors wrapped with `%w`, no `panic` outside init,
  no silent `_ = err`. Tests next to the code; table tests for parsing.
- TypeScript: strict, no `any` in app code (generated client excepted), no
  default exports for components.
- Styling: tokens from `src/styles/tokens.css` only. No raw hex, no arbitrary
  Tailwind values without a token.
- Strings: every user-visible string goes through the i18n catalog (`en` is the
  source; `zh-Hans` must stay complete).
- Comments explain why, never what. One line where possible.

## Before claiming something works

Run, in this order: `go vet ./...`, `go test ./...`, `bun run typecheck`,
`bun run build`. Visual changes also need a screenshot or a stated reason why
you could not produce one.

## Recording mistakes

When the owner corrects you, or you catch your own error, append to
`docs/log/mistakes.md` with symptom, cause, fix and the durable rule. That file
is the reason the same mistake should not happen twice.