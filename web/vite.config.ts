import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
	plugins: [react(), tailwindcss()],
	server: {
		port: 5173,
		// In dev, forward API calls to the Go server so cookies stay same-origin.
		proxy: {
			"/api": "http://localhost:8080",
		},
	},
	build: {
		outDir: "dist",
		emptyOutDir: true,
	},
});
