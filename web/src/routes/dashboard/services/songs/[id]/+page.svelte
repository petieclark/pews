<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	let songId = '';
	let song = null;
	let usage = [];
	let loading = true;
	let usageLoading = true;

	$: songId = $page.params.id;

	onMount(() => {
		loadSong();
		loadUsage();
	});

	async function loadSong() {
		loading = true;
		try {
			song = await api(`/api/services/songs/${songId}`);
		} catch (error) {
			console.error('Failed to load song:', error);
			alert('Song not found');
			goto('/dashboard/services/songs');
		} finally {
			loading = false;
		}
	}

	async function loadUsage() {
		usageLoading = true;
		try {
			usage = await api(`/api/services/songs/${songId}/usage`);
		} catch (error) {
			console.error('Failed to load song usage:', error);
			usage = [];
		} finally {
			usageLoading = false;
		}
	}

	function formatDate(dateStr) {
		if (!dateStr) return 'Never';
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { 
			weekday: 'short',
			month: 'short', 
			day: 'numeric', 
			year: 'numeric' 
		});
	}

	function formatTime(timeStr) {
		if (!timeStr) return '';
		return timeStr;
	}

	function goToService(serviceId) {
		goto(`/dashboard/services/${serviceId}`);
	}
</script>

<div class="space-y-6">
	{#if loading}
		<div class="p-12 text-center">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-4 border-teal border-t-transparent"></div>
			<p class="text-gray-500 mt-3">Loading song...</p>
		</div>
	{:else if song}
		<!-- Header -->
		<div class="flex justify-between items-start">
			<div>
				<button
					on:click={() => goto('/dashboard/services/songs')}
					class="text-teal hover:underline mb-2 text-sm"
				>
					← Back to Song Library
				</button>
				<h1 class="text-3xl font-bold text-navy">{song.title}</h1>
				{#if song.artist}
					<p class="text-lg text-gray-600 mt-1">{song.artist}</p>
				{/if}
			</div>
			<button
				on:click={() => goto(`/dashboard/services/songs?edit=${songId}`)}
				class="px-5 py-2.5 bg-navy text-white rounded-md hover:bg-opacity-90 font-medium"
			>
				Edit Song
			</button>
		</div>

		<!-- Song Details -->
		<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
			<!-- Basic Info Card -->
			<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
				<h2 class="text-lg font-semibold text-navy mb-4">Song Details</h2>
				<dl class="space-y-3">
					{#if song.default_key}
						<div class="flex justify-between">
							<dt class="text-sm font-medium text-gray-500">Default Key</dt>
							<dd class="text-sm text-gray-900 font-semibold">{song.default_key}</dd>
						</div>
					{/if}
					{#if song.tempo}
						<div class="flex justify-between">
							<dt class="text-sm font-medium text-gray-500">Tempo</dt>
							<dd class="text-sm text-gray-900">{song.tempo} BPM</dd>
						</div>
					{/if}
					{#if song.ccli_number}
						<div class="flex justify-between">
							<dt class="text-sm font-medium text-gray-500">CCLI Number</dt>
							<dd class="text-sm text-gray-900 font-mono">{song.ccli_number}</dd>
						</div>
					{/if}
					{#if song.tags}
						<div class="flex justify-between">
							<dt class="text-sm font-medium text-gray-500">Tags</dt>
							<dd class="text-sm text-gray-900">
								{#each song.tags.split(',') as tag}
									<span class="inline-block bg-gray-100 rounded px-2 py-1 mr-1 mb-1">{tag.trim()}</span>
								{/each}
							</dd>
						</div>
					{/if}
				</dl>
			</div>

			<!-- Usage Stats Card -->
			<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
				<h2 class="text-lg font-semibold text-navy mb-4">Usage Statistics</h2>
				<dl class="space-y-3">
					<div class="flex justify-between">
						<dt class="text-sm font-medium text-gray-500">Times Used</dt>
						<dd class="text-sm text-gray-900 font-semibold">{song.times_used}</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-sm font-medium text-gray-500">Last Used</dt>
						<dd class="text-sm text-gray-900">{formatDate(song.last_used)}</dd>
					</div>
					<div class="flex justify-between">
						<dt class="text-sm font-medium text-gray-500">Added</dt>
						<dd class="text-sm text-gray-900">{formatDate(song.created_at)}</dd>
					</div>
				</dl>
			</div>
		</div>

		<!-- Notes -->
		{#if song.notes}
			<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
				<h2 class="text-lg font-semibold text-navy mb-3">Notes</h2>
				<p class="text-gray-700 whitespace-pre-wrap">{song.notes}</p>
			</div>
		{/if}

		<!-- Lyrics -->
		{#if song.lyrics}
			<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
				<h2 class="text-lg font-semibold text-navy mb-3">Lyrics</h2>
				<pre class="text-sm text-gray-700 whitespace-pre-wrap font-sans">{song.lyrics}</pre>
			</div>
		{/if}

		<!-- Usage History -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			<div class="px-6 py-4 border-b border-gray-200">
				<h2 class="text-lg font-semibold text-navy">Service History</h2>
				<p class="text-sm text-gray-600 mt-1">
					Track where and when this song has been used (helpful for CCLI reporting)
				</p>
			</div>
			
			{#if usageLoading}
				<div class="p-8 text-center">
					<div class="inline-block animate-spin rounded-full h-6 w-6 border-4 border-teal border-t-transparent"></div>
					<p class="text-gray-500 text-sm mt-2">Loading usage history...</p>
				</div>
			{:else if usage.length === 0}
				<div class="p-8 text-center">
					<svg class="mx-auto h-10 w-10 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
					</svg>
					<h3 class="mt-3 text-sm font-medium text-gray-900">Not used yet</h3>
					<p class="mt-1 text-sm text-gray-500">
						This song hasn't been added to any service yet
					</p>
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Service Date
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Service Name
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Key Used
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Position
								</th>
								<th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
									Actions
								</th>
							</tr>
						</thead>
						<tbody class="bg-white divide-y divide-gray-200">
							{#each usage as item}
								<tr class="hover:bg-gray-50 transition-colors">
									<td class="px-6 py-4 whitespace-nowrap">
										<div class="text-sm text-gray-900">{formatDate(item.service_date)}</div>
										{#if item.service_time}
											<div class="text-xs text-gray-500">{formatTime(item.service_time)}</div>
										{/if}
									</td>
									<td class="px-6 py-4">
										<div class="text-sm text-gray-900">{item.service_name || 'Service'}</div>
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<div class="text-sm text-gray-700 font-medium">{item.song_key || song.default_key || '—'}</div>
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
											#{item.position}
										</span>
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-right text-sm">
										<button
											on:click={() => goToService(item.service_id)}
											class="text-teal hover:text-opacity-80 font-medium"
										>
											View Service →
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	:global(.bg-navy) {
		background-color: #1b3a4b;
	}
	:global(.text-navy) {
		color: #1b3a4b;
	}
	:global(.bg-teal) {
		background-color: #4a8b8c;
	}
	:global(.text-teal) {
		color: #4a8b8c;
	}
	:global(.ring-teal) {
		--tw-ring-color: #4a8b8c;
	}
</style>
