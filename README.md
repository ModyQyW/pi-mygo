# Pi MyGo

A desktop app for [pi](https://pi.dev), the minimal coding agent. Written in Go
with [MyGo](https://mygo.run) and React, it talks to the pi you already have
installed and signed in, so your models, extensions, skills and sessions work
exactly as they do in the terminal.

Status: **pre-alpha**. The plan is [here](docs/plan/00-plan.md); nothing is
released yet.

## Why another pi client

- **Your pi, not a copy.** No bundled runtime, no Node sidecar, no second
  credential store. The app starts `pi --mode rpc` and speaks its protocol.
- **A workspace, not a chat box.** Projects and sessions in a sidebar, several
  live sessions at once, each with its own pi process.
- **A transcript built for reading.** Streaming markdown, code and diffs, tool
  cards that collapse to one line, thinking that stays out of the way until you
  want it, and search that works on a five-thousand-message session.
- **Steering is a first-class move.** Enter sends or steers, ⌘Enter queues a
  follow-up, Esc takes the queue back into the editor.

## Requirements

- pi 1.5.x installed and authenticated (`pi --version`, `pi auth check`)
- macOS 12+ (Apple silicon), or Windows 10+/Linux as beta
- To build: Go 1.27+, Bun

## Develop

```sh
bun install
bun run dev          # mygo dev: Go rebuild + Vite HMR
bun run typecheck
go test ./...
```

`bun run dev` opens a development app whose identifier ends in `.dev`, so it
keeps its data separate from an installed build.

## Build

```sh
bun run build        # mygo build: app bundle, disk image, installers
```

## Documentation

- [`docs/plan/00-plan.md`](docs/plan/00-plan.md) — phases and definition of done
- [`docs/plan/01-architecture.md`](docs/plan/01-architecture.md) — how the pieces
  fit, invariants, performance budgets
- [`docs/plan/02-design-system.md`](docs/plan/02-design-system.md) — tokens,
  type, every surface and its states
- [`docs/research/`](docs/research) — verified facts about pi, MyGo and this
  machine
- [`docs/grilling/`](docs/grilling) — the decisions and how they were reached
- [`docs/log/mistakes.md`](docs/log/mistakes.md) — mistakes, corrections and the
  rules that came out of them

## License

MIT