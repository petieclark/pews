<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { getToken } from '$lib/api';

	onMount(() => {
		if (getToken()) {
			goto('/dashboard');
		} else {
			goto('/login');
		}
	});
</script>

<div class="min-h-screen flex items-center justify-center">
	<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-teal"></div>
</div>
