import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		// `npm run dev` talks to a locally running Cadence server so the UI can be
		// developed against real data without rebuilding the Go image.
		proxy: {
			'/api': {
				target: process.env.CADENCE_API ?? 'http://localhost:8080',
				changeOrigin: true
			}
		}
	}
});
