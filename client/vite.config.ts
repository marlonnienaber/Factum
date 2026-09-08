import { defineConfig } from 'vitest/config';
import { loadEnv } from 'vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { svelteTesting } from '@testing-library/svelte/vite';

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), '');
	const clientPort = env.CLIENT_PORT ? Number(env.CLIENT_PORT) : 5173;
	const serverPort = env.SERVER_PORT ?? '3001';

	return {
		plugins: [sveltekit(), svelteTesting()],
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
			passWithNoTests: true,
			coverage: {
				provider: 'v8',
				reporter: ['text-summary', 'html'],
				reportsDirectory: './coverage',
				include: ['src/**/*.{ts,svelte}'],
				exclude: [
					'src/app.d.ts',
					'src/lib/utils/testing.ts',
					'src/lib/utils/testing-setup.ts',
					'src/**/*.d.ts',
					'src/**/*.test.{js,ts}',
					'src/**/*.spec.{js,ts}',
					'src/**/*.svelte.test.{js,ts}',
					'src/**/*.svelte.spec.{js,ts}'
				]
			},
			projects: [
				{
					extends: './vite.config.ts',
					test: {
						name: 'client',
						environment: 'jsdom',
						setupFiles: ['./src/lib/utils/testing-setup.ts'],
						include: ['src/**/*.svelte.{test,spec}.{js,ts}']
					}
				},
				{
					extends: './vite.config.ts',
					test: {
						name: 'unit',
						environment: 'node',
						include: ['src/**/*.{test,spec}.{js,ts}'],
						exclude: ['src/**/*.svelte.{test,spec}.{js,ts}']
					}
				}
			]
		}
	};
});
