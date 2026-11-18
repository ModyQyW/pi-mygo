import { defineConfig } from "mygo-cli";

export default defineConfig({
  name: "Pi MyGo",
  identifier: "top.modyqyw.pimygo",
  version: "0.1.0",
  copyright: "© 2026 ModyQyW",
  // `mygo dev` runs devCommand and loads devUrl; `mygo build` runs
  // buildCommand and embeds frontendDist into the app.
  devUrl: "http://localhost:5173",
  devCommand: "bun run dev:web",
  buildCommand: "bun run build:web",
  frontendDist: "dist",
  bindings: "src/mygo.ts",
  out: "build",
  i18n: { source: "en" },
  // Signed updates stay off until `mygo keygen` has produced a key pair and the
  // public key is committed here. See docs/plan/00-plan.md (M4).
  // updates: { publicKey: "…", github: "ModyQyW/pi-mygo" },
});