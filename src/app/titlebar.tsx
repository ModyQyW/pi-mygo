import type { Probe, ProbeState } from "@/mygo";

import { messages } from "@/locales";
import { cn } from "@/lib/cn";

const dotTone: Record<ProbeState, string> = {
  ready: "bg-blue",
  checking: "bg-amber animate-pulse",
  missing: "bg-ink-4",
  unsupported: "bg-danger",
  error: "bg-danger",
};

function probeLabel(probe: Probe): string {
  switch (probe.state) {
    case "ready":
      return probe.version ? `pi ${probe.version}` : messages.envVersion();
    case "checking":
      return messages.envChecking();
    case "missing":
      return messages.envMissing();
    case "unsupported":
      return messages.envUnsupported();
    default:
      return messages.envError();
  }
}

/** Two brackets, echoing the corner motif rather than copying pi's own mark. */
function Mark() {
  return (
    <svg aria-hidden="true" viewBox="0 0 16 16" className="size-4 shrink-0">
      <path d="M2 6V2h4" fill="none" stroke="var(--coral-ink)" strokeWidth="1.6" />
      <path d="M14 10v4h-4" fill="none" stroke="var(--blue-ink)" strokeWidth="1.6" />
    </svg>
  );
}

/**
 * The window's title bar: a drag region under the traffic lights. It carries the
 * app mark, the open session (none yet) and the pi status, nothing else.
 */
export function Titlebar({ probe }: { probe: Probe | null }) {
  return (
    <header className="titlebar drag flex h-[var(--titlebar-height)] items-center gap-3 border-b border-hairline bg-paper pr-4">
      <Mark />
      <span className="font-mono text-[11px] tracking-[0.08em] text-ink-3 uppercase">
        {messages.appName()}
      </span>
      <span aria-hidden="true" className="h-3 w-px bg-hairline-strong" />
      <span className="truncate text-[13px] text-ink-3">{messages.titlebarUntitled()}</span>

      <div className="ml-auto flex items-center gap-2">
        {probe ? (
          <span
            className="no-drag inline-flex items-center gap-1.5 font-mono text-[11px] text-ink-3"
            aria-live="polite"
          >
            <span
              aria-hidden="true"
              className={cn("size-1.5 rounded-full", dotTone[probe.state])}
            />
            {probeLabel(probe)}
          </span>
        ) : null}
      </div>
    </header>
  );
}