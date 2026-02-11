<script>
	import { onMount } from 'svelte';

	let isDark = false;

	onMount(() => {
		// Check localStorage first, then system preference
		const stored = localStorage.getItem('pews-theme');
		if (stored) {
			isDark = stored === 'dark';
		} else {
			isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
		}
		applyTheme();

		// Listen for system preference changes
		const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
		const handler = (e) => {
			if (!localStorage.getItem('pews-theme')) {
				isDark = e.matches;
				applyTheme();
			}
		};
		mediaQuery.addEventListener('change', handler);
		
		return () => mediaQuery.removeEventListener('change', handler);
	});

	function applyTheme() {
		if (isDark) {
			document.documentElement.classList.add('dark');
		} else {
			document.documentElement.classList.remove('dark');
		}
	}

	function toggle() {
		isDark = !isDark;
		localStorage.setItem('pews-theme', isDark ? 'dark' : 'light');
		applyTheme();
	}
</script>

<button
	on:click={toggle}
	class="p-2 rounded-lg hover:bg-[var(--surface-hover)] transition-colors"
	aria-label="Toggle theme"
	title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
>
	{#if isDark}
		<!-- Sun icon -->
		<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-[var(--text-primary)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
		</svg>
	{:else}
		<!-- Moon icon -->
		<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-[var(--text-primary)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
		</svg>
	{/if}
</button>
