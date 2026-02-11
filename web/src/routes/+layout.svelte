<script>
	import '../app.css';
	import { onMount } from 'svelte';
	import { initLocale } from '$lib/i18n.js';

	onMount(async () => {
		await initLocale();
	});
</script>

<slot />
