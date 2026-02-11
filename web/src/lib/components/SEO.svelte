<script>
	import { page } from '$app/stores';

	export let title = 'Pews - Church Management & Live Streaming';
	export let description = 'Modern church management platform with live streaming, giving, and engagement tools.';
	export let image = '/og-image.png';
	export let type = 'website';
	export let url = '';
	export let structuredData = null;

	$: canonicalUrl = url || `https://pews.church${$page.url.pathname}`;
	$: ogImage = image.startsWith('http') ? image : `https://pews.church${image}`;
</script>

<svelte:head>
	<!-- Primary Meta Tags -->
	<title>{title}</title>
	<meta name="title" content={title} />
	<meta name="description" content={description} />
	<link rel="canonical" href={canonicalUrl} />

	<!-- Open Graph / Facebook -->
	<meta property="og:type" content={type} />
	<meta property="og:url" content={canonicalUrl} />
	<meta property="og:title" content={title} />
	<meta property="og:description" content={description} />
	<meta property="og:image" content={ogImage} />
	<meta property="og:site_name" content="Pews" />

	<!-- Twitter -->
	<meta property="twitter:card" content="summary_large_image" />
	<meta property="twitter:url" content={canonicalUrl} />
	<meta property="twitter:title" content={title} />
	<meta property="twitter:description" content={description} />
	<meta property="twitter:image" content={ogImage} />

	<!-- Structured Data (JSON-LD) -->
	{#if structuredData}
		{@html `<script type="application/ld+json">${JSON.stringify(structuredData)}</script>`}
	{/if}
</svelte:head>
