<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

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
	let attachments = [];
	let uploadingFile = false;

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

	async function openEditModal(song) {
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
		await loadAttachments(song.id);
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

	async function deleteSong(songId) {
		if (!confirm('Delete this song? This will not affect existing service plans.')) return;
		try {
			await api(`/api/services/songs/${songId}`, {
				method: 'DELETE'
			});
			loadSongs();
		} catch (error) {
			alert('Failed to delete song: ' + error.message);
		}
	}

	function formatDate(dateStr) {
		if (!dateStr) return 'Never';
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	async function loadAttachments(songId) {
		try {
			attachments = await api(`/api/services/songs/${songId}/attachments`);
		} catch (error) {
			console.error('Failed to load attachments:', error);
			attachments = [];
		}
	}

	async function uploadAttachment(event) {
		const file = event.target.files[0];
		if (!file) return;

		if (file.type !== 'application/pdf') {
			alert('Only PDF files are allowed');
			return;
		}

		if (file.size > 10 * 1024 * 1024) {
			alert('File size must be less than 10MB');
			return;
		}

		uploadingFile = true;
		try {
			const formData = new FormData();
			formData.append('file', file);

			const response = await fetch(`/api/services/songs/${editingSong.id}/attachments`, {
				method: 'POST',
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				},
				body: formData
			});

			if (!response.ok) {
				throw new Error('Upload failed');
			}

			await loadAttachments(editingSong.id);
			event.target.value = ''; // Clear file input
		} catch (error) {
			alert('Failed to upload file: ' + error.message);
		} finally {
			uploadingFile = false;
		}
	}

	async function deleteAttachment(attachmentId) {
		if (!confirm('Delete this attachment?')) return;

		try {
			await api(`/api/services/songs/attachments/${attachmentId}`, {
				method: 'DELETE'
			});
			await loadAttachments(editingSong.id);
		} catch (error) {
			alert('Failed to delete attachment: ' + error.message);
		}
	}

	function formatFileSize(bytes) {
		if (bytes < 1024) return bytes + ' B';
		if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
		return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<div>
			<button
				on:click={() => goto('/dashboard/services')}
				class="text-teal hover:underline mb-2"
			>
				← Back to Services
			</button>
			<h1 class="text-3xl font-bold text-[var(--text-primary)]">Song Library</h1>
		</div>
		<button
			on:click={openCreateModal}
			class="px-4 py-2 bg-teal text-white rounded-md hover:bg-opacity-90"
		>
			Add Song
		</button>
	</div>

	<!-- Search -->
	<div class="bg-surface rounded-lg shadow p-4">
		<div class="flex gap-4">
			<input
				type="text"
				bind:value={searchQuery}
				on:keyup={(e) => e.key === 'Enter' && handleSearch()}
				placeholder="Search by title, artist, or tags..."
				class="flex-1 px-4 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
			/>
			<button
				on:click={handleSearch}
				class="px-6 py-2 bg-navy text-white rounded-md hover:bg-opacity-90"
			>
				Search
			</button>
		</div>
	</div>

	<!-- Songs table -->
	<div class="bg-surface rounded-lg shadow overflow-hidden">
		{#if loading}
			<div class="p-8 text-center text-secondary">Loading...</div>
		{:else if songs.length === 0}
			<div class="p-8 text-center text-secondary">
				No songs found. {#if searchQuery}Try a different search.{:else}Add your first song to get
					started.{/if}
			</div>
		{:else}
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-[var(--surface-hover)]">
					<tr>
						<th
							class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider"
							>Title</th
						>
						<th
							class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider"
							>Artist</th
						>
						<th
							class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider"
							>Key</th
						>
						<th
							class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider"
							>Last Used</th
						>
						<th
							class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider"
							>Times Used</th
						>
						<th
							class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider"
							>Actions</th
						>
					</tr>
				</thead>
				<tbody class="bg-surface divide-y divide-gray-200">
					{#each songs as song}
						<tr class="hover:bg-[var(--surface-hover)]">
							<td class="px-6 py-4">
								<div class="text-sm font-medium text-primary">{song.title}</div>
								{#if song.tags}
									<div class="text-xs text-secondary">{song.tags}</div>
								{/if}
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-secondary">{song.artist || '—'}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-secondary">{song.default_key || '—'}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-secondary">{formatDate(song.last_used)}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-secondary">{song.times_used}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm">
								<button
									on:click={() => openEditModal(song)}
									class="text-teal hover:underline mr-3"
								>
									Edit
								</button>
								<button
									on:click={() => deleteSong(song.id)}
									class="text-red-600 hover:underline"
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
		<div class="flex justify-center gap-2">
			<button
				on:click={() => {
					page--;
					loadSongs();
				}}
				disabled={page === 1}
				class="px-4 py-2 bg-surface border rounded-md disabled:opacity-50"
			>
				Previous
			</button>
			<span class="px-4 py-2">
				Page {page} of {Math.ceil(total / limit)}
			</span>
			<button
				on:click={() => {
					page++;
					loadSongs();
				}}
				disabled={page >= Math.ceil(total / limit)}
				class="px-4 py-2 bg-surface border rounded-md disabled:opacity-50"
			>
				Next
			</button>
		</div>
	{/if}
</div>

<!-- Create/Edit song modal -->
{#if showModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
		<div class="bg-surface rounded-lg max-w-2xl w-full p-6 max-h-[90vh] overflow-y-auto">
			<h2 class="text-2xl font-bold text-[var(--text-primary)] mb-4">
				{editingSong ? 'Edit Song' : 'Add Song'}
			</h2>
			<form on:submit|preventDefault={saveSong} class="space-y-4">
				<div class="grid grid-cols-2 gap-4">
					<div class="col-span-2">
						<label class="block text-sm font-medium text-primary">Title *</label>
						<input
							type="text"
							bind:value={formData.title}
							required
							class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-primary">Artist</label>
						<input
							type="text"
							bind:value={formData.artist}
							class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-primary">Default Key</label>
						<input
							type="text"
							bind:value={formData.default_key}
							placeholder="G, C, Bb, etc."
							class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-primary">Tempo (BPM)</label>
						<input
							type="number"
							bind:value={formData.tempo}
							class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-primary">CCLI Number</label>
						<input
							type="text"
							bind:value={formData.ccli_number}
							class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
						/>
					</div>
					<div class="col-span-2">
						<label class="block text-sm font-medium text-primary">Tags</label>
						<input
							type="text"
							bind:value={formData.tags}
							placeholder="worship, fast, opener (comma-separated)"
							class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
						/>
					</div>
					<div class="col-span-2">
						<label class="block text-sm font-medium text-primary">Lyrics</label>
						<textarea
							bind:value={formData.lyrics}
							rows="6"
							class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal font-mono text-sm"
						></textarea>
					</div>
					<div class="col-span-2">
						<label class="block text-sm font-medium text-primary">Notes</label>
						<textarea
							bind:value={formData.notes}
							rows="3"
							class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
						></textarea>
					</div>
					
					{#if editingSong}
						<div class="col-span-2 border-t pt-4">
							<label class="block text-sm font-medium text-primary mb-2">Attachments (PDFs)</label>
							
							<!-- Upload button -->
							<div class="mb-4">
								<label class="cursor-pointer inline-flex items-center px-4 py-2 bg-[var(--surface)] border border-custom text-primary rounded-md hover:bg-[var(--surface-hover)]">
									{#if uploadingFile}
										Uploading...
									{:else}
										📎 Upload PDF
									{/if}
									<input
										type="file"
										accept=".pdf,application/pdf"
										on:change={uploadAttachment}
										disabled={uploadingFile}
										class="hidden"
									/>
								</label>
								<span class="ml-2 text-xs text-secondary">Max 10MB</span>
							</div>

							<!-- Attachments list -->
							{#if attachments.length > 0}
								<div class="space-y-2">
									{#each attachments as attachment}
										<div class="flex items-center justify-between bg-[var(--surface-hover)] p-3 rounded">
											<div class="flex items-center space-x-3">
												<span class="text-2xl">📄</span>
												<div>
													<a
														href={`/api/services/songs/attachments/${attachment.id}`}
														target="_blank"
														class="text-sm font-medium text-primary hover:text-teal"
													>
														{attachment.original_name}
													</a>
													<div class="text-xs text-secondary">
														{formatFileSize(attachment.file_size)} • Uploaded {formatDate(attachment.created_at)}
													</div>
												</div>
											</div>
											<button
												on:click={() => deleteAttachment(attachment.id)}
												class="text-red-600 hover:text-red-800 text-sm"
											>
												Delete
											</button>
										</div>
									{/each}
								</div>
							{:else}
								<p class="text-sm text-secondary italic">No attachments yet. Upload chord charts or sheet music PDFs.</p>
							{/if}
						</div>
					{/if}
				</div>
				<div class="flex gap-2 pt-4">
					<button
						type="button"
						on:click={() => (showModal = false)}
						class="flex-1 px-4 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md hover:bg-[var(--surface-hover)]"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="flex-1 px-4 py-2 bg-teal text-white rounded-md hover:bg-opacity-90"
					>
						{editingSong ? 'Update' : 'Create'}
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
