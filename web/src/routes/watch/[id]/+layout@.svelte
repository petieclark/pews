<script>
	import '../../../app.css';
	import { page } from '$app/stores';
	
	$: meta = $page.data?.meta || {};
</script>

<svelte:head>
	<title>{meta.title || 'Watch Live'}</title>
	
	<!-- Open Graph Meta Tags -->
	<meta property="og:title" content={meta.title || 'Watch Live'} />
	<meta property="og:description" content={meta.description || 'Join us for our live service'} />
	<meta property="og:type" content="video.other" />
	<meta property="og:url" content={meta.url || ''} />
	{#if meta.image}
		<meta property="og:image" content={meta.image} />
	{/if}
	
	<!-- Twitter Card Meta Tags -->
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content={meta.title || 'Watch Live'} />
	<meta name="twitter:description" content={meta.description || 'Join us for our live service'} />
	{#if meta.image}
		<meta name="twitter:image" content={meta.image} />
	{/if}
</svelte:head>

<!-- Minimal public layout - no dashboard navigation -->
<div class="min-h-screen" style="background-color: #0a0a0a; color: #ffffff">
	<slot />
</div>
