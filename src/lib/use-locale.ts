import { useEffect, useState } from "react";
import { isMyGo, onLocaleChange } from "mygo-runtime";

/**
 * Re-renders when the app's language changes.
 *
 * Generated messages read the current language each time they are called, so a
 * re-render is all a language switch needs.
 */
export function useLocaleVersion(): number {
  const [version, setVersion] = useState(0);

  useEffect(() => {
    if (!isMyGo()) return;
    return onLocaleChange(() => setVersion((current) => current + 1));
  }, []);

  return version;
}