# Design system — Pi paper

The look is paper, not glass. Warm surfaces, ink text, hairlines instead of
shadows, small radii, and brand colour spent only where it carries meaning:
state, focus, charts, the logo. Density sits between a chat app and an editor:
comfortable to read for an hour, tight enough to see a whole tool exchange
without scrolling.

## Tokens

Defined once in `src/styles/tokens.css` as CSS custom properties on `:root`, with
a `[data-theme="dark"]` block. Components never use raw hex.

| Token | Light | Dark | Use |
| --- | --- | --- | --- |
| `--paper` | `#FAF7F2` | `#14120F` | app background |
| `--surface` | `#FFFDF9` | `#1B1815` | cards, drawer, palette |
| `--sunken` | `#F3EFE7` | `#100E0C` | user message, code block, inputs |
| `--ink` | `#17161A` | `#F3EDE3` | primary text |
| `--ink-2` | `#4A4640` | `#C9C1B6` | secondary text |
| `--ink-3` | `#7C756C` | `#938A7F` | metadata, placeholders |
| `--ink-4` | `#A9A096` | `#6B645B` | non-text marks, brackets |
| `--hairline` | `#E5DED3` | `#2A2620` | separators, card borders |
| `--hairline-strong` | `#D6CCBE` | `#3A352D` | hover borders |
| `--coral` | `#F09082` | `#F09082` | brand fill, focus ring, send |
| `--coral-ink` | `#C9553E` | `#E8826E` | coral text/links (contrast-safe) |
| `--blue` | `#4D9ABF` | `#5FA8C9` | running state, links |
| `--blue-ink` | `#2F7C9E` | `#7FBBD5` | blue text |
| `--amber` | `#F1BE58` | `#F1BE58` | waiting state, warnings fill |
| `--amber-ink` | `#A06F14` | `#D9A544` | warning text |
| `--success` | `#3C7A57` | `#6FA985` | completed |
| `--danger` | `#B4432F` | `#E0796A` | errors |
| `--danger-bg` | `#FBEDE9` | `#2A1A16` | error card fill |

Accent variables from the desktop are honoured where the OS has one:
`--accent: var(--mygo-accent, var(--coral-ink))`. Text scale multiplies the root:
`font-size: calc(1rem * var(--mygo-text-scale, 1))`.

Contrast floor: body text ≥ 7:1, secondary ≥ 4.5:1, `--ink-4` is never text.

## Type

```css
--font-ui:   -apple-system, "SF Pro Text", "Inter", "PingFang SC", "Noto Sans SC", sans-serif;
--font-read: "New York", ui-serif, Charter, "Songti SC", "Noto Serif CJK SC", serif;
--font-mono: "SF Mono", ui-monospace, "JetBrains Mono", Menlo, monospace;
```

`--font-read` is reserved for app-authored editorial copy: empty states,
onboarding, the About window. Model output is never serif; CJK serif mixing
inside a Latin serif paragraph reads worse than either alone.

| Role | Size / line-height | Weight | Family |
| --- | --- | --- | --- |
| Transcript body | 15.5 / 1.72 | 400 | UI |
| Transcript h2 | 17 / 1.35 | 600 | UI |
| Transcript h3 | 15.5 / 1.4 | 600 | UI |
| UI body, controls | 14 / 1.45 | 400 | UI |
| Metadata, timestamps | 12 / 1.3 | 400 | Mono |
| Section labels | 11 / 1.2, tracking .08em, uppercase | 500 | Mono |
| Page title | 22 / 1.2 | 600 | UI |
| Code | 12.5 / 1.6 | 400 | Mono |

Measure: transcript text is capped at `68ch` and the column at 720 px. Code
blocks scroll horizontally; they never wrap by default, and a wrap toggle is
offered per block.

## Space, radius, elevation

Spacing scale: 2, 4, 8, 12, 16, 24, 32, 48, 64. Controls are 28 px tall (icon
buttons 28×28), rows 44 px, transcript block gap 16 px, turn gap 28 px.

Radii: chips 3 px, cards and inputs 6 px, drawer/palette 10 px, pills 999 px.

Elevation is a hairline plus, for overlays only, `0 12px 32px rgba(23,22,26,.12)`
and `0 1px 0 rgba(255,255,255,.4)` inset in dark. Cards never cast shadows.

Corner brackets: four 6×6 px L-marks, 1 px, `--ink-4`, at the corners of the
transcript container and of expanded tool cards. They are the only decorative
element, and they disappear under `prefers-reduced-motion: reduce` only if they
animate (they do not).

## Chrome

- Window: hidden title bar, traffic lights inset at (18, 18); the page draws
  `--mygo-titlebar-height` of 44 px, using the CSS variables mygo injects.
- Titlebar contents: app mark (16 px), current project path in mono 12 px,
  session name, then right-aligned run state pill and ⌘K.
- Sidebar: 268 px, resizable 220–380, collapse with ⌘B. Background `--paper`,
  separated by a hairline. Session rows have no card, only a 2 px coral mark on
  the active row.
- Status row above the composer: branch, model, thinking level, context gauge,
  session cost, queue count. Everything is a button that opens the relevant
  control.

## Components

**Session row.** Two lines: name (14 px, truncating middle) and a mono meta line
`main · 2m · 3 tools`. A 6 px dot carries state: idle `--ink-4`, running `--blue`
(pulsing at 1.2 s, disabled with reduced motion), waiting `--amber`, error
`--danger`. Hover raises the row background to `--sunken`; active uses `--sunken`
plus the coral mark.

**Assistant message.** No bubble. Prose, then a hairline-separated meta line on
hover with actions (copy, fork from here, label). User messages are a sunken
block with 6 px radius and a left coral 2 px rule, so the two voices are
distinguishable at a glance without alternating sides.

**Thinking.** Collapsed to a single mono line (`Thinking · 1.2s`) with a
pulsing hairline while streaming. Expanding reveals 13 px `--ink-3` text.

**Tool card.** Header row: icon, tool name, one-line argument summary, duration,
status mark. Body collapsed by default; diffs show a three-context-line preview
with `+`/`-` gutters and soft green/red fills at 12 % opacity; bash shows the last
five lines; errors auto-expand with a danger left rule. Expanded body has a
toolbar: copy output, open file in editor, open full diff.

**Code block.** `--sunken` background, hairline border, language label in mono
11 px at the bottom right, copy button appearing on hover and on focus. Shiki
themes: a light paper theme and a dark ink theme generated from the same token
values, so code never looks pasted in from another app.

**Composer.** A borderless textarea on paper, minimum 2 rows, growing to 40 vh.
Above it, queue chips: `steer · after this tool` in blue tint, `follow-up ·
after it finishes` in ink, each removable. Below it, a row of 28 px controls:
attach, model chip, thinking chip, context gauge, then send or stop. Send button
is coral with ink text; stop is an outline square with a danger icon.

**Command palette.** `--surface`, 10 px radius, hairline, 12 px inset from the
top, 560 px wide. Sections: sessions, commands, models, files. Mono section
labels, 32 px rows, keyboard hint on the right.

**Session tree drawer.** Right side, 360 px, 16 px inset, overlay scrim at 8 %
ink. Each row: branch glyph, message excerpt (truncated, markdown stripped), mono
timestamp, label chip, child count. Current leaf marked with a coral dot.
Keyboard: ↑↓, ↵ switch, F fork, L label, Esc close.

**Error card.** `--danger-bg`, left 2 px `--danger` rule, title in plain language,
one sentence of cause, up to three actions, and a "Details" disclosure with the
raw pi text, monospace and selectable.

**Empty states.** Mark + one sentence in `--font-read` + up to three actions.
Never a spinner alone; every wait shows the operation being waited on.

**Toasts.** Bottom right, 280 px, `--surface`, hairline, auto-dismiss 4 s,
pause on hover, max three stacked.

## Motion

Only `transform` and `opacity`; durations 80 ms (hover), 120 ms (expand), 160 ms
(drawer, palette). Easing `cubic-bezier(.2,.8,.2,1)`. No entrance animation for
streaming content: text appears, it does not slide. A 1 px coral caret marks the
live end of streaming text and stops blinking under reduced motion. Nothing
animates when `prefers-reduced-motion: reduce`.

## Accessibility

- Focus ring: 2 px `--coral` with 2 px offset via `:focus-visible`; never removed.
- Every icon button has an accessible name; icon-only controls get tooltips with
  the shortcut.
- The transcript is a `role="log"` region; only completed messages are announced,
  never deltas.
- The drawer is non-modal; focus moves into it on open and returns to the
  trigger on close; Esc closes.
- Minimum hit target 28×28, 8 px separation for destructive neighbours.
- Text scale and high contrast preferences from the desktop are honoured.

## States every surface must define

| Surface | Loading | Empty | Error |
| --- | --- | --- | --- |
| Sidebar | skeleton rows | "No projects yet" + Add project | scan failed, retry |
| Transcript | "Opening session…" | "Say something" + starters | crash banner, reopen |
| Search | inline count | "No matches" | index failed, retry |
| Tree | skeleton | "This session has one branch" | tree fetch failed |
| Settings | instant | n/a | write failed, retry |
| Model picker | list skeleton | "No models configured" + fix | fetch failed |
