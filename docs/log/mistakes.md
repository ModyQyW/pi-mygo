# Mistakes and corrections

Running log. Every entry has: what happened, how it was caught, what changed so
it cannot happen again. Newest last. Corrections from the owner come here too,
even when they are preferences rather than errors.

## Format

```text
### YYYY-MM-DD — short title
- Symptom:
- Cause:
- Fix:
- Rule: (the durable change, or "none")
```

## 2026-11-18 — Assumed macOS-only distribution

- Symptom: Round 1 recommended shipping macOS first and deferring Windows/Linux.
- Cause: read the absence of a Developer ID certificate as a blocker for all
  distribution, and treated CI cross-builds as later work.
- Fix: the owner asked for cross-platform and auto-update from v0.1. mygo's
  update manifests are Ed25519-signed and independent of Apple code signing, so
  the certificate is not a blocker; it only affects Gatekeeper on first launch.
- Rule: separate "cannot be signed" from "cannot be shipped" before offering a
  staged release plan.

## Environment facts that behave like traps

These are not mistakes made in this repo, but they will bite anyone who touches
the same code, so they are recorded before they cost time.

### Killing pi's stdout reader crashes pi

- Symptom: `node:events:505 ... write EPIPE` from the child after the client
  stops reading.
- Cause: pi writes to stdout without a listener for `error` on the socket.
- Fix: close stdin first, keep draining stdout until the process exits, and only
  then stop reading. Never close the read end while pi is alive.
- Rule: the Go supervisor owns shutdown ordering; there is a test for it.

### Extensions emit UI records before any command

- Symptom: five `extension_ui_request` records (`setStatus`, `setWidget`,
  `notify`) arrive immediately after spawn, with ANSI escapes in `statusText`.
- Cause: the user's real pi extensions assume a TUI and paint on load.
- Fix: treat fire-and-forget records as normal traffic, strip ANSI, never block
  startup on them, and count ignored methods in diagnostics.
- Rule: the RPC client must be usable against a heavily customised pi
  installation, not a clean one.

### 2026-11-18 — M0 gate promised a settings sheet that M3 owns

- Symptom: the M0 gate said "⌘, opens settings", but the roadmap puts the whole
  settings surface in M3, so the gate could never pass on its own phase.
- Cause: wrote the gate from the design mockup instead of from the phase's
  scope list.
- Fix: removed settings from M0 and its gate; the menu has no Settings item
  until it works.
- Rule: a phase gate may only reference things that phase builds. A menu item
  that does nothing is worse than a missing one.
