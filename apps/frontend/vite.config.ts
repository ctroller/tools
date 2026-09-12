import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [sveltekit()],
	test: {
		environment: 'jsdom',
		include: ['src/**/*.{test,spec}.{js,ts}']
	},
	resolve: {
		conditions: ['browser']
	},
	server: {
		port: 3000,
		strictPort: true,
		proxy: {
			'/api/file-converter': {
				target: process.env.FC_PROXY_TARGET ?? 'http://localhost:8080',
				changeOrigin: true,
				rewrite: (path) => path.replace(/^\/api\/file-converter\/[^/]+/, '')
			}
		}
	}
});
