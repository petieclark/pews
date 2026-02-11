<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let streams = [];
	let liveStream = null;
	let upcomingStreams = [];
	let pastStreams = [];
	let loading = false;

	onMount(() => {
		loadStreams();
	});

	async function loadStreams() {
		loading = true;
		try {
			// Get live stream
			const live = await api('/api/streaming/live');
			liveStream = live?.id ? live : null;

			// Get all streams
			const response = await api('/api/streaming?limit=50');
			streams = response.streams || [];

			// Separate upcoming and past
			const now = new Date();
			upcomingStreams = streams.filter(s => {
				if (s.status === 'live') return false;
				if (s.status === 'scheduled' && s.scheduled_start) {
					return new Date(s.scheduled_start) > now;
				}
				return false;
			});
			
			pastStreams = streams.filter(s => 
				s.status === 'ended' || s.status === 'archived'
			).slice(0, 10);

		} catch (error) {
			console.error('Failed to load streams:', error);
		} finally {
			loading = false;
		}
	}

	function formatDate(dateStr) {
		if (!dateStr) return 'Not scheduled';
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { 
			weekday: 'short', 
			month: 'short', 
			day: 'numeric', 
			hour: 'numeric',
			minute: '2-digit'
		});
	}

	function getStatusBadge(status) {
		const badges = {
			scheduled: 'bg-blue-100 text-blue-800',
			live: 'bg-red-100 text-red-800 animate-pulse',
			ended: 'bg-[var(--surface-hover)] text-primary',
			archived: 'bg-[var(--surface-hover)] text-secondary'
		};
		return badges[status] || 'bg-[var(--surface-hover)] text-primary';
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold" style="color: var(--navy)">Streaming</h1>
		<button
			on:click={() => goto('/dashboard/streaming/new')}
			class="px-4 py-2 rounded-md text-white hover:opacity-90"
			style="background-color: var(--teal)"
		>
			Schedule Stream
		</button>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<p class="text-secondary">Loading streams...</p>
		</div>
	{:else}
		<!-- Live Stream Card -->
		{#if liveStream}
			<div class="p-6 rounded-lg border-2 border-red-500" style="background-color: var(--surface)">
				<div class="flex items-center gap-2 mb-4">
					<div class="h-3 w-3 bg-red-500 rounded-full animate-pulse"></div>
					<h2 class="text-xl font-bold" style="color: var(--text-primary)">LIVE NOW</h2>
				</div>
				<h3 class="text-2xl font-bold mb-2" style="color: var(--text-primary)">{liveStream.title}</h3>
				<p class="text-secondary mb-4">{liveStream.description || ''}</p>
				<div class="flex gap-4 items-center">
					<span class="text-sm" style="color: var(--text-primary)">👥 {liveStream.viewer_count} watching</span>
					<span class="text-sm text-secondary">Peak: {liveStream.peak_viewers}</span>
					<button
						on:click={() => goto(`/dashboard/streaming/${liveStream.id}`)}
						class="ml-auto px-4 py-2 rounded-md text-white"
						style="background-color: var(--teal)"
					>
						Manage Stream
					</button>
				</div>
			</div>
		{/if}

		<!-- Upcoming Streams -->
		{#if upcomingStreams.length > 0}
			<div>
				<h2 class="text-xl font-bold mb-4" style="color: var(--text-primary)">Upcoming Streams</h2>
				<div class="grid gap-4">
					{#each upcomingStreams as stream}
						<div 
							class="p-4 rounded-lg border cursor-pointer hover:shadow-md transition"
							style="background-color: var(--surface); border-color: var(--border)"
							on:click={() => goto(`/dashboard/streaming/${stream.id}`)}
							on:keydown={(e) => e.key === 'Enter' && goto(`/dashboard/streaming/${stream.id}`)}
							role="button"
							tabindex="0"
						>
							<div class="flex justify-between items-start">
								<div class="flex-1">
									<h3 class="text-lg font-semibold mb-1" style="color: var(--text-primary)">{stream.title}</h3>
									<p class="text-sm text-secondary mb-2">{stream.description || 'No description'}</p>
									<div class="flex gap-4 text-sm text-secondary">
										<span>📅 {formatDate(stream.scheduled_start)}</span>
										<span>📺 {stream.stream_type}</span>
									</div>
								</div>
								<span class="px-2 py-1 rounded text-xs font-medium {getStatusBadge(stream.status)}">
									{stream.status}
								</span>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Past Streams -->
		{#if pastStreams.length > 0}
			<div>
				<h2 class="text-xl font-bold mb-4" style="color: var(--text-primary)">Recent Streams</h2>
				<div class="grid gap-4">
					{#each pastStreams as stream}
						<div 
							class="p-4 rounded-lg border cursor-pointer hover:shadow-md transition"
							style="background-color: var(--surface); border-color: var(--border)"
							on:click={() => goto(`/dashboard/streaming/${stream.id}`)}
							on:keydown={(e) => e.key === 'Enter' && goto(`/dashboard/streaming/${stream.id}`)}
							role="button"
							tabindex="0"
						>
							<div class="flex justify-between items-start">
								<div class="flex-1">
									<h3 class="text-lg font-semibold mb-1" style="color: var(--text-primary)">{stream.title}</h3>
									<p class="text-sm text-secondary mb-2">{stream.description || 'No description'}</p>
									<div class="flex gap-4 text-sm text-secondary">
										<span>📅 {formatDate(stream.actual_start || stream.scheduled_start)}</span>
										<span>👥 Peak viewers: {stream.peak_viewers}</span>
									</div>
								</div>
								<span class="px-2 py-1 rounded text-xs font-medium {getStatusBadge(stream.status)}">
									{stream.status}
								</span>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		{#if !liveStream && upcomingStreams.length === 0 && pastStreams.length === 0}
			<div class="text-center py-12 rounded-lg" style="background-color: var(--surface)">
				<p class="text-secondary mb-4">No streams yet. Schedule your first stream!</p>
				<button
					on:click={() => goto('/dashboard/streaming/new')}
					class="px-4 py-2 rounded-md text-white"
					style="background-color: var(--teal)"
				>
					Schedule Stream
				</button>
			</div>
		{/if}
	{/if}
</div>
