# Grilling — Round 2

Frontier after round 1: visual identity, first-run path, workspace discovery,
release staging. All four depend only on settled round-1 decisions.

---

## Q2-A — Which visual direction?

Three candidates: Pi paper (warm paper, ink text, brand accents, corner brackets,
mono metadata), IDE neutral (gray, dense, one accent), macOS native (vibrancy,
system font, HIG chrome).

➡️ **Pi paper.** The app should be recognisable in a screenshot. The risk —
gimmick — is contained by using brand colour only for state, focus and charts,
never for body text, and by keeping the transcript itself quiet.

**Answer: Pi paper.**

---

## Q2-B — How does a user get from install to a first conversation?

pi authenticates through `/login` in the TUI. `pi auth login <provider>` exists
but the docs say it needs an interactive terminal; `pi auth check --json` is
non-interactive and tells readiness.

**A.** Guided: detect missing/old pi and missing auth, then run
`pi auth login <provider>` in the system terminal and poll `pi auth check`.
**B.** Embed a terminal (mygo terminal plugin) and log in inside the app.
**C.** Collect an API key and write `~/.pi/agent/auth.json` ourselves.

➡️ **A.** B adds a native library dependency and pty ownership for one screen.
C bypasses pi's credential handling and can corrupt a working install.

**Answer: A — guided, system terminal, poll auth check.**

---

## Q2-C — Where do "projects" come from?

**A.** Scan `~/.pi/agent/sessions/*/*.jsonl` headers for `cwd`, group, sort by
recency, plus "Add project" for empty directories.
**B.** Only user-added directories.
**C.** Auto-discover with a hide list.

➡️ **A.** Sessions created in the terminal must appear; an empty sidebar on first
launch would be a bad first impression. Hide lists are a v0.2 refinement.

**Answer: A — auto-discover + manual add.**

---

## Q2-D — Platform and update staging?

Context: no code-signing identity on this machine; mygo updates are Ed25519
manifest signatures, independent of Apple certificates.

**A.** macOS arm64 is the verified target; CI also emits Windows x64 and Linux
packages marked beta; updater enabled on all three.
**B.** macOS only in v0.1.
**C.** All three at equal quality before v0.1.

➡️ **A.** Cross-platform is a mygo strength, but only macOS can be verified
locally. Shipping untested builds as beta is honest; claiming parity is not.

**Answer: A — macOS first, cross-platform beta, updater on.**

---

## Settled

- Pi paper visual language.
- Guided first run; login happens in the system terminal; app never writes
  pi credentials.
- Projects auto-discovered from session headers, plus manual add.
- macOS-first release, Windows/Linux beta, signed updates on all platforms.
