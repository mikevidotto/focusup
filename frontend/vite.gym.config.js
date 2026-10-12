import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { fileURLToPath } from "node:url";

// The phone page served by cmd/gymserver, which embeds this output.
// Run via `npm run build:gym` (or scripts/build-gym.sh).
export default defineConfig({
    root: fileURLToPath(new URL("./gym", import.meta.url)),
    base: "./",
    plugins: [svelte()],
    build: {
        outDir: fileURLToPath(new URL("../cmd/gymserver/dist/app", import.meta.url)),
        emptyOutDir: true,
    },
    server: {
        // `npx vite --config vite.gym.config.js` while gymserver runs locally
        proxy: { "/api": "http://localhost:8787" },
    },
});
