/** The appearance preference; the settings UI will own it in a later phase. */
export type ThemePreference = "system" | "light" | "dark";

function darkQuery(): MediaQueryList {
  return window.matchMedia("(prefers-color-scheme: dark)");
}

/**
 * Sets the appearance on the document. Tokens switch by attribute, so the page
 * never needs to remount to change theme.
 */
export function applyTheme(preference: ThemePreference): void {
  const dark = preference === "dark" || (preference === "system" && darkQuery().matches);
  document.documentElement.dataset.theme = dark ? "dark" : "light";
}

/**
 * Applies a preference now and whenever the system's appearance changes.
 * Returns a function that stops watching.
 */
export function watchSystemTheme(preference: ThemePreference): () => void {
  const query = darkQuery();
  const onChange = () => applyTheme(preference);
  query.addEventListener("change", onChange);
  applyTheme(preference);
  return () => query.removeEventListener("change", onChange);
}