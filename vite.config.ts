import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

// The build writes to web/.dist-tmp; scripts/build-frontend.mjs swaps that
// into web/dist only after the build succeeded, so a failed build never
// leaves a broken page behind. base: "./" keeps the assets relative, which
// both mygo:// and file:// loading need.
export default defineConfig({
  root: "web/src",
  base: "./",
  plugins: [vue()],
  build: {
    outDir: "../../web/.dist-tmp",
    emptyOutDir: true,
    target: "es2022",
    assetsDir: "assets",
    rollupOptions: {
      output: {
        entryFileNames: "assets/[name]-[hash].js",
        chunkFileNames: "assets/[name]-[hash].js",
        assetFileNames: "assets/[name]-[hash][extname]",
      },
    },
  },
});
