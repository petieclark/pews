<script>
	// This page just redirects to the [id] page with 'new' as the id
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	onMount(() => {
		goto('/dashboard/communication/drip/new');
	});
</script>
