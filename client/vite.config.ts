import { defineConfig } from 'vitest/config';
import { loadEnv } from 'vite';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), '');
	const clientPort = env.CLIENT_PORT ? Number(env.CLIENT_PORT) : 5173;
	const serverPort = env.SERVER_PORT ?? '3001';

	return {
		plugins: [sveltekit(), tailwindcss()],
		server: {
			port: clientPort,
			strictPort: true,
			proxy: {
				'/api': {
					target: `http://localhost:${serverPort}`,
					changeOrigin: true
				}
			}
		},
		test: {
			expect: { requireAssertions: true },
			projects: [
				{
					extends: './vite.config.ts',
					test: {
						name: 'server',
						environment: 'node',
						include: ['src/**/*.{test,spec}.{js,ts}'],
						exclude: ['src/**/*.svelte.{test,spec}.{js,ts}']
					}
				}
			]
		}
	};
});
