# Grilling — Round 5

Frontier after round 4: error presentation, in-session search, attachment
handling, update signing keys.

---

## Q5-A — How are failures shown?

pi reports provider errors, rate limits, context overflow, retry and compaction
failures, tool failures and process crashes through the event stream.

**A.** Structured error cards: classify into a small set (not signed in, invalid
key, rate limited, overloaded, context overflow, unknown model, network, tool
failure, process crash), one sentence of plain language plus actions (retry,
switch model, compact, copy diagnostics), raw text folded inside.
**B.** Show pi's raw error text.
**C.** Toast only.

➡️ **A.** The classification is the product: `429` and
`context_length_exceeded` are obvious to the person who wrote the client and
opaque to everyone else. Raw text stays available, never hidden.

**Answer: A — structured error cards.**

---

## Q5-B — Is in-session search part of v0.1?

**A.** Yes: ⌘F, match highlighting, n/N navigation, "3 of 17", auto-expanding
tool cards that contain a match.
**B.** v0.2.

➡️ **A.** Long sessions are the normal case; without search the transcript
becomes unreadable and users fall back to the terminal.

**Answer: A — search in v0.1.**

---

## Q5-C — How do dropped files and pasted images reach pi?

**A.** Images become base64 `images` blocks on `prompt`/`steer`/`follow_up`; other
files insert `@relative/path` text, which pi reads itself. ⌥-drop forces a path
for images too. Attachments are removable before sending.
**B.** Every file becomes a path.
**C.** Every image becomes a temp file path.

➡️ **A.** pi's RPC accepts images directly, so the model sees the picture without
a tool round trip; paths keep large files out of the JSON stream.

**Answer: A — inline images, paths for files.**

---

## Q5-D — How are update signing keys handled?

mygo signs update manifests with an Ed25519 key pair from `mygo keygen`; the
private key must stay out of the repository.

**A.** Generate locally, keep the private key in a file outside the repo, put it
into GitHub Actions secrets with `gh secret set` without printing it, commit only
the public key into `mygo.config.ts`. CI runs `mygo build -upload`, which creates
a draft release for review.
**B.** Leave updates unconfigured until the owner supplies keys.
**C.** Ship without auto-update.

➡️ **A.** The user chose auto-update in round 1; a draft release keeps publishing
a human decision.

**Answer: A — generate and store in secrets, publish drafts.**

---

## Settled

- Structured error cards with actions and raw detail.
- ⌘F search in v0.1.
- Images inline, files as `@path`.
- Ed25519 keys generated locally, stored as CI secrets, public key committed,
  releases as drafts.
