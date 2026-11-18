import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { direction, isMyGo, locale } from "mygo-runtime";

import { App } from "@/app/app";
import { applyTheme } from "@/lib/theme";
import "@/styles/app.css";

// The page can start in the app's language and appearance, before React runs,
// so the first paint is already correct.
if (isMyGo()) {
  document.documentElement.lang = locale();
  document.documentElement.dir = direction();
}
applyTheme("system");

const root = document.getElementById("root");
if (!root) throw new Error("#root is missing from index.html");

createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>,
);