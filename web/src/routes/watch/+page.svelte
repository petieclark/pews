<script lang="ts">
	import { onMount } from 'svelte';

	let liveStream: any = null;
	let loading = true;
	let error = '';

	onMount(async () => {
		await loadLiveStream();
		loading = false;
	});

	async function loadLiveStream() {
		try {
			const response = await fetch('/api/streaming/live', {
				headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
			});
			
			if (response.ok) {
				const data = await response.json();
				if (data && data.is_live) {
					liveStream = data;
				}
			}
		} catch (err) {
			console.error('Failed to load live stream:', err);
			error = 'Unable to connect to streaming service';
		}
	}

	function getEmbedUrl(url: string) {
		// Convert YouTube/Vimeo URLs to embed format
		if (url.includes('youtube.com/watch')) {
			const videoId = new URL(url).searchParams.get('v');
			return `https://www.youtube.com/embed/${videoId}`;
		}
		if (url.includes('youtu.be/')) {
			const videoId = url.split('youtu.be/')[1].split('?')[0];
			return `https://www.youtube.com/embed/${videoId}`;
		}
		if (url.includes('vimeo.com/')) {
			const videoId = url.split('vimeo.com/')[1].split('?')[0];
			return `https://player.vimeo.com/video/${videoId}`;
		}
		return url;
	}
</script>

<svelte:head>
	<title>Watch Live</title>
</svelte:head>

<div class="min-h-screen bg-surface">
	<!-- Hero Section -->
	<div class="bg-gradient-to-r from-[var(--teal)] to-[var(--navy)] text-white py-20">
		<div class="container mx-auto px-4 max-w-6xl text-center">
			<h1 class="text-5xl font-bold mb-4">Watch Live</h1>
			<p class="text-xl opacity-90">Join us online for our live services</p>
		</div>
	</div>

	<div class="container mx-auto px-4 max-w-6xl py-12">
		{#if loading}
			<div class="flex justify-center items-center py-24">
				<div class="animate-spin rounded-full h-16 w-16 border-b-2 border-[var(--teal)]"></div>
			</div>
		{:else if liveStream}
			<!-- Live Stream Active -->
			<div class="bg-surface border border-custom rounded-lg shadow-lg overflow-hidden">
				<div class="bg-red-600 text-white px-6 py-3 flex items-center justify-between">
					<div class="flex items-center gap-3">
						<span class="inline-block w-3 h-3 bg-white rounded-full animate-pulse"></span>
						<span class="font-bold uppercase text-sm">Live Now</span>
					</div>
					<span class="text-sm">{liveStream.title || 'Live Service'}</span>
				</div>

				<div class="aspect-video w-full bg-black">
					{#if liveStream.stream_url}
						<iframe
							src={getEmbedUrl(liveStream.stream_url)}
							class="w-full h-full"
							frameborder="0"
							allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
							allowfullscreen
						></iframe>
					{:else}
						<div class="flex items-center justify-center h-full text-white">
							<p>Stream URL not available</p>
						</div>
					{/if}
				</div>

				{#if liveStream.description}
					<div class="p-6">
						<p class="text-[var(--text-primary)]">{liveStream.description}</p>
					</div>
				{/if}
			</div>
		{:else}
			<!-- No Active Stream -->
			<div class="bg-surface border border-custom rounded-lg shadow-lg p-12 text-center">
				<div class="mb-6">
					<svg class="w-24 h-24 mx-auto text-[var(--text-secondary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
					</svg>
				</div>
				<h2 class="text-3xl font-bold text-[var(--text-primary)] mb-4">No Active Stream</h2>
				<p class="text-lg text-[var(--text-secondary)] mb-8">
					There's no live service at the moment. Check back during service times!
				</p>

				{#if error}
					<p class="text-sm text-red-600 dark:text-red-400 mb-4">{error}</p>
				{/if}

				<!-- Service Times -->
				<div class="mt-8 bg-[var(--surface-hover)] dark:bg-gray-800 rounded-lg p-6 inline-block">
					<h3 class="font-bold text-lg text-[var(--text-primary)] mb-3">Service Times</h3>
					<div class="text-left space-y-2 text-[var(--text-secondary)]">
						<p>🕐 Sunday: 9:00 AM & 11:00 AM</p>
						<p>🕖 Wednesday: 7:00 PM</p>
					</div>
				</div>

				<!-- Past Sermons Link -->
				<div class="mt-8">
					<a
						href="/sermons"
						class="inline-block px-8 py-3 bg-[var(--teal)] text-white font-semibold rounded-lg hover:opacity-90 transition"
					>
						Watch Past Sermons
					</a>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	:global(body) {
		--surface: white;
		--surface-hover: #f9fafb;
		--text-primary: #1a202c;
		--text-secondary: #718096;
		--border: #e2e8f0;
	}

	:global(body.dark) {
		--surface: #1a202c;
		--surface-hover: #2d3748;
		--text-primary: #f7fafc;
		--text-secondary: #a0aec0;
		--border: #4a5568;
	}
</style>
