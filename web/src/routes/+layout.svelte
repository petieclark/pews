<script>
	import '../app.css';
	import Toast from '$lib/Toast.svelte';
	import InstallPrompt from '$lib/InstallPrompt.svelte';
</script>

<Toast />
<slot />

<InstallPrompt />
