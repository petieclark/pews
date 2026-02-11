import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 5173,
		host: true
	},
	build: {
		// Enable minification
		minify: 'terser',
		// Tree shaking optimizations
		rollupOptions: {
			output: {
				manualChunks: {
					// Separate vendor chunks for better caching
					'svelte-core': ['svelte', '@sveltejs/kit']
				}
			}
		},
		// Target modern browsers for smaller bundles
		target: 'es2020'
	}
});
