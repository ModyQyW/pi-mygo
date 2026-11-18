import { useEffect, useState } from "react";

import type { Info, Probe } from "@/mygo";
import { Button } from "@/components/ui/button";
import { messages } from "@/locales";
import { cn } from "@/lib/cn";

const INSTALL_COMMAND = "curl -fsSL https://pi.dev/install.sh | sh";

type Tone = "ok" | "warn" | "danger" | "muted";

const toneClass: Record<Tone, string> = {
  ok: "text-success border-success/40",
  warn: "text-amber-ink border-amber/50",
  danger: "text-danger border-danger/40",
  muted: "text-ink-3 border-hairline-strong",
};

function Chip({ tone, children }: { tone: Tone; children: string }) {
  return (
    <span
      className={cn(
        "inline-flex h-5 items-center rounded-[var(--radius-chip)] border px-1.5",
        "font-mono text-[11px] leading-none",
        toneClass[tone],
      )}
    >
      {children}
    </span>
  );
}

function compatChip(probe: Probe) {
  switch (probe.compat) {
    case "ok":
      return <Chip tone="ok">{messages.compatOk()}</Chip>;
    case "untested":
      return <Chip tone="warn">{messages.compatUntested()}</Chip>;
    case "too_old":
      return <Chip tone="danger">{messages.compatTooOld()}</Chip>;
    case "too_new":
      return <Chip tone="danger">{messages.compatTooNew()}</Chip>;
    default:
      return null;
  }
}

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="pt-0.5 font-mono text-[11px] tracking-[0.08em] text-ink-3 uppercase">
        {label}
      </dt>
      <dd className="min-w-0 text-[13px] text-ink-2">{children}</dd>
    </>
  );
}

function CopyCommand() {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!copied) return;
    const timer = window.setTimeout(() => setCopied(false), 1600);
    return () => window.clearTimeout(timer);
  }, [copied]);

  return (
    <div className="mt-3 flex items-center gap-2">
      <code className="min-w-0 flex-1 overflow-x-auto rounded-[var(--radius-chip)] bg-sunken px-2 py-1.5 font-mono text-[12px] whitespace-nowrap text-ink-2">
        {INSTALL_COMMAND}
      </code>
      <Button
        variant="ghost"
        onClick={() => {
          void navigator.clipboard.writeText(INSTALL_COMMAND).then(() => setCopied(true));
        }}
      >
        {copied ? messages.envCopied() : messages.envCopy()}
      </Button>
    </div>
  );
}

/** "1.5.2" -> "1.x": the major series pi releases in. */
function series(version: string): string {
  const major = version.split(".")[0];
  return major ? `${major}.x` : version;
}

function ProbeDetail({ probe, info }: { probe: Probe; info: Info | null }) {
  switch (probe.state) {
    case "checking":
      return <p className="text-[13px] text-ink-3">{messages.envChecking()}</p>;
    case "missing":
      return (
        <div>
          <p className="text-[13px] text-ink-2">{messages.envDetailMissing()}</p>
          <p className="mt-3 text-[13px] text-ink-3">{messages.envInstallHint()}</p>
          <CopyCommand />
        </div>
      );
    case "unsupported":
      return (
        <p className="text-[13px] text-danger">
          {probe.compat === "too_new"
            ? messages.envDetailSeries(info ? series(info.minPi) : "—")
            : messages.envDetailMin(info?.minPi ?? "—")}
        </p>
      );
    case "error":
      return <p className="text-[13px] text-danger">{probe.detail}</p>;
    default:
      return probe.detail ? <p className="text-[13px] text-ink-3">{messages.envDetailTested()}</p> : null;
  }
}

/**
 * The environment panel: what pi was found, whether this app supports it, and
 * what to do about it. It is the first-run experience until session discovery
 * lands.
 */
export function EnvironmentCard({
  info,
  onRecheck,
  rechecking,
}: {
  info: Info | null;
  onRecheck: () => void;
  rechecking: boolean;
}) {
  const probe = info?.pi;

  return (
    <section
      aria-labelledby="environment-title"
      className="bracketed mt-8 bg-surface px-6 py-5"
    >
      <div className="bracket-b" aria-hidden="true" />
      <div className="flex items-center justify-between gap-4">
        <h2
          id="environment-title"
          className="font-mono text-[11px] tracking-[0.08em] text-ink-3 uppercase"
        >
          {messages.envTitle()}
        </h2>
        <Button variant="ghost" onClick={onRecheck} disabled={rechecking}>
          {messages.envRecheck()}
        </Button>
      </div>

      {probe ? (
        <dl className="mt-4 grid grid-cols-[minmax(0,9rem)_minmax(0,1fr)] items-baseline gap-x-6 gap-y-2.5">
          <Row label={messages.envPi()}>
            {probe.path ? (
              <span className="font-mono text-[12px] break-all" title={probe.path}>
                {probe.path}
              </span>
            ) : (
              <span className="text-ink-3">—</span>
            )}
          </Row>
          <Row label={messages.envVersion()}>
            <span className="inline-flex items-center gap-2">
              <span className="font-mono text-[12px]">{probe.version || "—"}</span>
              {compatChip(probe)}
            </span>
          </Row>
          <Row label={messages.envApp()}>
            <span className="font-mono text-[12px]">{info?.appVersion ?? "—"}</span>
          </Row>
          <Row label={messages.envPlatform()}>
            <span className="font-mono text-[12px]">
              {info ? `${info.platform}/${info.arch}` : "—"}
            </span>
          </Row>
        </dl>
      ) : null}

      <div className="mt-4 border-t border-hairline pt-4">
        {probe ? <ProbeDetail probe={probe} info={info} /> : <p className="text-[13px] text-ink-3">{messages.envChecking()}</p>}
      </div>
    </section>
  );
}