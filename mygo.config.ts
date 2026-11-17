import { defineConfig } from "mygo-cli";

// The dev loop runs Vite on 5173 and points the app at it; release builds go
// through npm run build into web/dist, which go:embed packs into the binary.
export default defineConfig({
  name: "guoAppGo",
  identifier: "com.guoappgo.desktop",
  version: "0.0.1",
  out: "build",
  devUrl: "http://localhost:5173",
  devCommand: "npm run dev",
  buildCommand: "npm run build",
  frontendDist: "web/dist",
  // The generated typed client belongs to the frontend tree, not the repo root.
  bindings: "web/src/mygo.ts",
});
