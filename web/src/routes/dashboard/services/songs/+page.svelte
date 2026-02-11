<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	const KEYS = ['C', 'C#', 'D', 'Eb', 'E', 'F', 'F#', 'G', 'Ab', 'A', 'Bb', 'B'];

	let songs = [];
	let total = 0;
	let page = 1;
	let limit = 50;
	let searchQuery = '';
	let loading = false;
	let showModal = false;
	let editingSong = null;
	let formData = {
		title: '',
		artist: '',
		default_key: '',
		tempo: null,
		ccli_number: '',
		lyrics: '',
		notes: '',
		tags: ''
	};

	onMount(() => {
		loadSongs();
	});

	async function loadSongs() {
		loading = true;
		try {
			const params = new URLSearchParams({
				page: page.toString(),
				limit: limit.toString()
			});
			if (searchQuery) {
				params.append('q', searchQuery);
			}
			const response = await api(`/api/services/songs?${params}`);
			songs = response.songs || [];
			total = response.total || 0;
		} catch (error) {
			console.error('Failed to load songs:', error);
		} finally {
			loading = false;
		}
	}

	function handleSearch() {
		page = 1;
		loadSongs();
	}

	function openCreateModal() {
		editingSong = null;
		formData = {
			title: '',
			artist: '',
			default_key: '',
			tempo: null,
			ccli_number: '',
			lyrics: '',
			notes: '',
			tags: ''
		};
		showModal = true;
	}

	function openEditModal(song) {
		editingSong = song;
		formData = {
			title: song.title,
			artist: song.artist || '',
			default_key: song.default_key || '',
			tempo: song.tempo || null,
			ccli_number: song.ccli_number || '',
			lyrics: song.lyrics || '',
			notes: song.notes || '',
			tags: song.tags || ''
		};
		showModal = true;
	}

	async function saveSong() {
		try {
			if (editingSong) {
				await api(`/api/services/songs/${editingSong.id}`, {
					method: 'PUT',
					body: JSON.stringify(formData)
				});
			} else {
				await api('/api/services/songs', {
					method: 'POST',
					body: JSON.stringify(formData)
				});
			}
			showModal = false;
			loadSongs();
		} catch (error) {
			alert('Failed to save song: ' + error.message);
		}
	}

	async function deleteSong(songId, songTitle) {
		if (!confirm(`Delete "${songTitle}"? This will not affect existing service plans.`)) return;
		try {
			await api(`/api/services/songs/${songId}`, {
				method: 'DELETE'
			});
			loadSongs();
		} catch (error) {
			alert('Failed to delete song: ' + error.message);
		}
	}

	function viewSongDetail(songId) {
		goto(`/dashboard/services/songs/${songId}`);
	}

	function formatDate(dateStr) {
		if (!dateStr) return 'Never';
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<div>
			<button
				on:click={() => goto('/dashboard/services')}
				class="text-teal hover:underline mb-2 text-sm"
			>
				← Back to Services
			</button>
			<h1 class="text-3xl font-bold text-navy">Song Library</h1>
			<p class="text-gray-600 mt-1">Manage your worship songs and track usage for CCLI reporting</p>
		</div>
		<button
			on:click={openCreateModal}
			class="px-5 py-2.5 bg-teal text-white rounded-md hover:bg-opacity-90 font-medium shadow-sm"
		>
			+ Add Song
		</button>
	</div>

	<!-- Search -->
	<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
		<div class="flex gap-4">
			<input
				type="text"
				bind:value={searchQuery}
				on:keyup={(e) => e.key === 'Enter' && handleSearch()}
				placeholder="Search by title, artist, or tags..."
				class="flex-1 px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal focus:border-transparent"
			/>
			<button
				on:click={handleSearch}
				class="px-6 py-2 bg-navy text-white rounded-md hover:bg-opacity-90 font-medium"
			>
				Search
			</button>
			{#if searchQuery}
				<button
					on:click={() => {
						searchQuery = '';
						handleSearch();
					}}
					class="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
				>
					Clear
				</button>
			{/if}
		</div>
	</div>

	<!-- Songs table -->
	<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
		{#if loading}
			<div class="p-12 text-center">
				<div class="inline-block animate-spin rounded-full h-8 w-8 border-4 border-teal border-t-transparent"></div>
				<p class="text-gray-500 mt-3">Loading songs...</p>
			</div>
		{:else if songs.length === 0}
			<div class="p-12 text-center">
				<svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
				</svg>
				<h3 class="mt-4 text-lg font-medium text-gray-900">
					{#if searchQuery}
						No songs found
					{:else}
						No songs yet
					{/if}
				</h3>
				<p class="mt-2 text-sm text-gray-500">
					{#if searchQuery}
						Try adjusting your search terms
					{:else}
						Get started by adding your first worship song
					{/if}
				</p>
				{#if !searchQuery}
					<button
						on:click={openCreateModal}
						class="mt-6 px-5 py-2.5 bg-teal text-white rounded-md hover:bg-opacity-90 font-medium"
					>
						Add Your First Song
					</button>
				{/if}
			</div>
		{:else}
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Title
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Artist
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Key
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Tempo
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Last Used
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Usage
						</th>
						<th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
							Actions
						</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each songs as song}
						<tr class="hover:bg-gray-50 transition-colors">
							<td class="px-6 py-4">
								<button
									on:click={() => viewSongDetail(song.id)}
									class="text-left hover:text-teal transition-colors"
								>
									<div class="text-sm font-medium text-gray-900">{song.title}</div>
									{#if song.tags}
										<div class="text-xs text-gray-500 mt-1">
											{#each song.tags.split(',') as tag}
												<span class="inline-block bg-gray-100 rounded px-2 py-0.5 mr-1">{tag.trim()}</span>
											{/each}
										</div>
									{/if}
								</button>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-gray-700">{song.artist || '—'}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-gray-700 font-medium">{song.default_key || '—'}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-gray-700">
									{song.tempo ? `${song.tempo} BPM` : '—'}
								</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-gray-600">{formatDate(song.last_used)}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-teal bg-opacity-10 text-teal">
									{song.times_used} {song.times_used === 1 ? 'time' : 'times'}
								</span>
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-3">
								<button
									on:click={() => viewSongDetail(song.id)}
									class="text-teal hover:text-opacity-80"
								>
									View
								</button>
								<button
									on:click={() => openEditModal(song)}
									class="text-navy hover:text-opacity-80"
								>
									Edit
								</button>
								<button
									on:click={() => deleteSong(song.id, song.title)}
									class="text-red-600 hover:text-red-800"
								>
									Delete
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>

	<!-- Pagination -->
	{#if total > limit}
		<div class="flex justify-between items-center">
			<div class="text-sm text-gray-600">
				Showing {Math.min((page - 1) * limit + 1, total)} to {Math.min(page * limit, total)} of {total} songs
			</div>
			<div class="flex gap-2">
				<button
					on:click={() => {
						page--;
						loadSongs();
					}}
					disabled={page === 1}
					class="px-4 py-2 bg-white border border-gray-300 rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
				>
					Previous
				</button>
				<div class="flex items-center px-4 py-2 bg-white border border-gray-300 rounded-md">
					Page {page} of {Math.ceil(total / limit)}
				</div>
				<button
					on:click={() => {
						page++;
						loadSongs();
					}}
					disabled={page >= Math.ceil(total / limit)}
					class="px-4 py-2 bg-white border border-gray-300 rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
				>
					Next
				</button>
			</div>
		</div>
	{/if}
</div>

<!-- Create/Edit song modal -->
{#if showModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
		<div class="bg-white rounded-lg max-w-3xl w-full p-6 max-h-[90vh] overflow-y-auto">
			<h2 class="text-2xl font-bold text-navy mb-4">
				{editingSong ? 'Edit Song' : 'Add New Song'}
			</h2>
			<form on:submit|preventDefault={saveSong} class="space-y-4">
				<div class="grid grid-cols-2 gap-4">
					<div class="col-span-2">
						<label class="block text-sm font-medium text-gray-700 mb-1">Title *</label>
						<input
							type="text"
							bind:value={formData.title}
							required
							placeholder="e.g., Great Are You Lord"
							class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal focus:border-transparent"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Artist / Author</label>
						<input
							type="text"
							bind:value={formData.artist}
							placeholder="e.g., All Sons & Daughters"
							class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal focus:border-transparent"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Default Key</label>
						<select
							bind:value={formData.default_key}
							class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal focus:border-transparent"
						>
							<option value="">Select a key...</option>
							{#each KEYS as key}
								<option value={key}>{key}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Tempo (BPM)</label>
						<input
							type="number"
							bind:value={formData.tempo}
							min="40"
							max="200"
							placeholder="e.g., 120"
							class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal focus:border-transparent"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">CCLI Number</label>
						<input
							type="text"
							bind:value={formData.ccli_number}
							placeholder="e.g., 6460220"
							class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal focus:border-transparent"
						/>
					</div>
					<div class="col-span-2">
						<label class="block text-sm font-medium text-gray-700 mb-1">Tags</label>
						<input
							type="text"
							bind:value={formData.tags}
							placeholder="e.g., worship, upbeat, opener (comma-separated)"
							class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal focus:border-transparent"
						/>
						<p class="text-xs text-gray-500 mt-1">Use tags to organize and search for songs</p>
					</div>
					<div class="col-span-2">
						<label class="block text-sm font-medium text-gray-700 mb-1">Lyrics</label>
						<textarea
							bind:value={formData.lyrics}
							rows="8"
							placeholder="Paste lyrics here..."
							class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal focus:border-transparent font-mono text-sm"
						></textarea>
					</div>
					<div class="col-span-2">
						<label class="block text-sm font-medium text-gray-700 mb-1">Notes</label>
						<textarea
							bind:value={formData.notes}
							rows="3"
							placeholder="Add any additional notes or instructions..."
							class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal focus:border-transparent"
						></textarea>
					</div>
				</div>
				<div class="flex gap-3 pt-4 border-t">
					<button
						type="button"
						on:click={() => (showModal = false)}
						class="flex-1 px-4 py-2.5 border border-gray-300 rounded-md hover:bg-gray-50 font-medium"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="flex-1 px-4 py-2.5 bg-teal text-white rounded-md hover:bg-opacity-90 font-medium"
					>
						{editingSong ? 'Update Song' : 'Create Song'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

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
