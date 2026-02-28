<script lang="ts">
	import { onMount } from 'svelte';
	import { Music } from 'lucide-svelte';

	let sermons = [];
	let loading = true;
	let searchQuery = '';
	let filterSeries = '';
	let filterSpeaker = '';
	let series = [];
	let speakers = [];

	// Extract tenant_id from subdomain or query param (you'll need to adjust based on your setup)
	let tenantId = '';

	onMount(async () => {
		// TODO: Extract tenant_id from URL/subdomain
		tenantId = new URLSearchParams(window.location.search).get('tenant_id') || '';
		await loadSermons();
		loading = false;
	});

	async function loadSermons() {
		try {
			const params = new URLSearchParams({ tenant_id: tenantId });
			if (filterSeries) params.append('series', filterSeries);
			if (filterSpeaker) params.append('speaker', filterSpeaker);

			const response = await fetch(`/api/sermons/public?${params.toString()}`);
			
			if (response.ok) {
				sermons = await response.json() || [];
				series = [...new Set(sermons.map(s => s.series_name).filter(Boolean))];
				speakers = [...new Set(sermons.map(s => s.speaker).filter(Boolean))];
			}
		} catch (error) {
			console.error('Failed to load sermons:', error);
		}
	}

	function formatDate(dateString) {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	$: filteredSermons = sermons.filter(sermon => {
		return searchQuery === '' || 
			sermon.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
			sermon.speaker.toLowerCase().includes(searchQuery.toLowerCase()) ||
			(sermon.scripture_reference && sermon.scripture_reference.toLowerCase().includes(searchQuery.toLowerCase()));
	});
</script>

<svelte:head>
	<title>Sermons Archive</title>
</svelte:head>

<div class="min-h-screen bg-gray-50">
	<!-- Hero Section -->
	<div class="bg-gradient-to-r from-teal-600 to-teal-800 text-white py-20">
		<div class="container mx-auto px-4 max-w-6xl">
			<h1 class="text-5xl font-bold mb-4">Sermons</h1>
			<p class="text-xl opacity-90">Listen to messages from our church</p>
		</div>
	</div>

	<div class="container mx-auto px-4 max-w-6xl py-12">
		{#if loading}
			<div class="flex justify-center items-center py-12">
				<div class="animate-spin rounded-full h-16 w-16 border-b-2 border-teal-600"></div>
			</div>
		{:else}
			<!-- Search and Filters -->
			<div class="mb-8 grid grid-cols-1 md:grid-cols-3 gap-4">
				<input
					type="text"
					placeholder="Search sermons..."
					bind:value={searchQuery}
					class="px-4 py-3 border rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
				/>
				<select
					bind:value={filterSeries}
					on:change={loadSermons}
					class="px-4 py-3 border rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
				>
					<option value="">All Series</option>
					{#each series as s}<option value={s}>{s}</option>{/each}
				</select>
				<select
					bind:value={filterSpeaker}
					on:change={loadSermons}
					class="px-4 py-3 border rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
				>
					<option value="">All Speakers</option>
					{#each speakers as speaker}<option value={speaker}>{speaker}</option>{/each}
				</select>
			</div>

			<!-- Sermons Grid -->
			{#if filteredSermons.length === 0}
				<div class="text-center py-16">
					<p class="text-2xl text-gray-400">No sermons found</p>
				</div>
			{:else}
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
					{#each filteredSermons as sermon}
						<div class="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-xl transition">
							<div class="p-6">
								{#if sermon.series_name}
									<div class="text-xs font-semibold text-teal-600 uppercase tracking-wide mb-2">
										{sermon.series_name}
									</div>
								{/if}
								<h3 class="text-xl font-bold mb-2 line-clamp-2">{sermon.title}</h3>
								<p class="text-sm text-gray-600 mb-2">{sermon.speaker}</p>
								<p class="text-xs text-gray-500 mb-3">{formatDate(sermon.sermon_date)}</p>
								
								{#if sermon.scripture_reference}
									<p class="text-sm text-gray-700 italic mb-4">{sermon.scripture_reference}</p>
								{/if}

								{#if sermon.audio_url || sermon.video_url}
									<div class="flex gap-2 pt-4 border-t">
										{#if sermon.audio_url}
											<a
												href={sermon.audio_url}
												target="_blank"
												class="flex-1 bg-teal-600 text-white text-center py-2 rounded hover:bg-teal-700 transition text-sm"
											>
												<Music size={14} class="inline" /> Listen
											</a>
										{/if}
										{#if sermon.video_url}
											<a
												href={sermon.video_url}
												target="_blank"
												class="flex-1 bg-gray-600 text-white text-center py-2 rounded hover:bg-gray-700 transition text-sm"
											>
												🎥 Watch
											</a>
										{/if}
									</div>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			{/if}

			<!-- Podcast Feed Link -->
			<div class="mt-12 p-6 bg-teal-50 rounded-lg border border-teal-200">
				<h2 class="text-xl font-bold mb-2">Subscribe to our podcast</h2>
				<p class="text-gray-700 mb-4">Stay updated with our latest sermons</p>
				<div class="flex gap-2">
					<input
						readonly
						value={`${window.location.origin}/api/sermons/feed.xml?tenant_id=${tenantId}`}
						class="flex-1 px-4 py-2 bg-white border rounded"
					/>
					<button
						on:click={() => {
							navigator.clipboard.writeText(`${window.location.origin}/api/sermons/feed.xml?tenant_id=${tenantId}`);
							alert('Feed URL copied to clipboard!');
						}}
						class="px-6 py-2 bg-teal-600 text-white rounded hover:bg-teal-700"
					>
						Copy
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
