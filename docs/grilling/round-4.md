# Grilling — Round 4

Frontier after round 3: session tree presentation, tool-card density, settings
surface, product naming.

---

## Q4-A — How does pi's session tree appear in a GUI?

**A.** Right-side drawer overlay (⌘⇧T): whole tree, current leaf, branch
summaries, labels; ↑↓ navigate, ↵ switch, F fork, L label.
**B.** Permanent right pane.
**C.** No visualisation in v0.1, commands only.

➡️ **A.** B permanently costs 280–360 px and competes with a future changes
panel. C drops one of pi's distinguishing features. A drawer is reachable and
costs nothing when closed.

**Answer: A — right drawer overlay.**

---

## Q4-B — What does a tool call look like by default?

**A.** One collapsed row: icon, tool name, a one-line argument summary (command,
path, query), duration and status. Expanding shows full output. `edit`/`write`
show a collapsed diff preview, `bash` shows the last five lines, errors
auto-expand.
**B.** Fully expanded by default.
**C.** Fully collapsed by default.

➡️ **A.** B lets a 3,000-line build log bury the conversation. C hides what the
agent is doing, which is the thing users watch. A needs the most detail work —
per-tool summary formatting — and that is the point.

**Answer: A — smart collapse.**

---

## Q4-C — Which settings ship?

**A.** Appearance and language; model defaults; input and transcript
preferences; update and diagnostics.

➡️ **All four.** Model defaults are per-session overrides and never write back
to pi's own settings; diagnostics must expose the pi path, detected version and
supported range, because a client of someone else's tool has to explain itself
when it refuses to connect.

**Answer: all four groups.**

---

## Q4-D — What is the app called?

**A.** Repo `pi-mygo`; display name **Pi MyGo**; bundle id `top.modyqyw.pimygo`;
binary `pi-mygo`.
**B.** A separate product name.

➡️ **A.** Descriptive names win for discoverability; the identity work is in the
UI, not in inventing a brand nobody searches for.

**Answer: display name Pi MyGo, bundle id top.modyqyw.pimygo, binary pi-mygo.**

---

## Settled

- Tree drawer, tool cards with smart collapse.
- Four settings groups.
- Naming fixed: Pi MyGo / top.modyqyw.pimygo / pi-mygo.
