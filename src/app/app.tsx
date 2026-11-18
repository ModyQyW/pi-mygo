import { useCallback, useEffect, useState } from "react";
import { isMyGo } from "mygo-runtime";

import { App as PiApp, events, type Info } from "@/mygo";
import { Sidebar } from "@/app/sidebar";
import { Titlebar } from "@/app/titlebar";
import { EnvironmentCard } from "@/features/environment/environment-card";
import { messages } from "@/locales";
import { useLocaleVersion } from "@/lib/use-locale";
import { watchSystemTheme } from "@/lib/theme";

/** The welcome copy: the only place the reading serif appears. */
function Welcome() {
  return (
    <div className="max-w-[52ch]">
      <h1 className="font-read text-[30px] leading-tight text-ink">{messages.appName()}</h1>
      <p className="mt-1 font-read text-[16px] text-ink-2">{messages.tagline()}</p>
      <p className="mt-4 text-[13.5px] leading-relaxed text-ink-3">{messages.welcomeBody()}</p>
    </div>
  );
}

function FailureNotice({ message }: { message: string }) {
  return (
    <div className="bracketed mt-8 border-l-2 border-danger bg-danger-bg px-5 py-4">
      <div className="bracket-b" aria-hidden="true" />
      <p className="font-mono text-[11px] tracking-[0.08em] text-danger uppercase">
        {messages.envError()}
      </p>
      <p className="mt-1.5 text-[13px] break-words text-ink-2">{message}</p>
    </div>
  );
}

export function App() {
  // Re-render when the app's language changes; the generated messages read the
  // current language on every call.
  useLocaleVersion();

  const [info, setInfo] = useState<Info | null>(null);
  const [failure, setFailure] = useState<string | null>(null);
  const [rechecking, setRechecking] = useState(false);

  useEffect(() => watchSystemTheme("system"), []);

  useEffect(() => {
    if (!isMyGo()) {
      setFailure(
        "This page is open in a browser, outside the app. Bound Go methods and events exist only inside a Pi MyGo window.",
      );
      return;
    }

    let alive = true;
    const off = events.environmentChanged.on((probe) => {
      if (alive) setInfo((previous) => (previous ? { ...previous, pi: probe } : previous));
    });

    PiApp.info()
      .then((value) => {
        if (alive) setInfo(value);
      })
      .catch((error: unknown) => {
        if (alive) setFailure(error instanceof Error ? error.message : String(error));
      });

    return () => {
      alive = false;
      off();
    };
  }, []);

  const onRecheck = useCallback(() => {
    setRechecking(true);
    PiApp.recheck()
      .then((probe) => {
        setFailure(null);
        setInfo((previous) => (previous ? { ...previous, pi: probe } : previous));
      })
      .catch((error: unknown) => {
        setFailure(error instanceof Error ? error.message : String(error));
      })
      .finally(() => setRechecking(false));
  }, []);

  return (
    <div className="grid h-dvh grid-rows-[auto_minmax(0,1fr)] bg-paper">
      <Titlebar probe={info?.pi ?? null} />
      <div className="grid min-h-0 grid-cols-[268px_minmax(0,1fr)]">
        <Sidebar />
        <main className="min-h-0 overflow-y-auto">
          <div className="mx-auto flex min-h-full max-w-[760px] flex-col justify-center px-8 py-12">
            <Welcome />
            {failure ? (
              <FailureNotice message={failure} />
            ) : (
              <EnvironmentCard info={info} onRecheck={onRecheck} rechecking={rechecking} />
            )}
          </div>
        </main>
      </div>
    </div>
  );
}