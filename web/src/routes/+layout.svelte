<script>
	import '../app.css';
	
	export let data;
</script>

<svelte:head>
	<!-- Organization Structured Data -->
	{#if data?.organizationSchema}
		{@html `<script type="application/ld+json">${JSON.stringify(data.organizationSchema)}</script>`}
	{/if}
</svelte:head>

<slot />
