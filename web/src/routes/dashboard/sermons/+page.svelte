<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	let sermons = [];
	let loading = true;
	let searchQuery = '';
	let filterSeries = '';
	let filterSpeaker = '';
	let series = [];
	let speakers = [];

	onMount(async () => {
		await loadSermons();
		loading = false;
	});

	async function loadSermons() {
		try {
			const params = new URLSearchParams();
			if (filterSeries) params.append('series', filterSeries);
			if (filterSpeaker) params.append('speaker', filterSpeaker);
			
			const response = await fetch(`/api/sermons?${params.toString()}`, {
				headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
			});
			
			if (response.ok) {
				sermons = await response.json() || [];
				series = [...new Set(sermons.map(s => s.series_name).filter(Boolean))];
				speakers = [...new Set(sermons.map(s => s.speaker).filter(Boolean))];
			}
		} catch (error) {
			console.error('Failed to load sermons:', error);
		}
	}

	async function deleteSermon(id) {
		if (!confirm('Are you sure?')) return;
		try {
			const response = await fetch(`/api/sermons/${id}`, {
				method: 'DELETE',
				headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
			});
			if (response.ok) await loadSermons();
		} catch (error) {
			console.error('Failed to delete sermon:', error);
		}
	}

	async function togglePublished(sermon) {
		try {
			const response = await fetch(`/api/sermons/${sermon.id}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				},
				body: JSON.stringify({
					...sermon,
					published: !sermon.published,
					sermon_date: sermon.sermon_date.split('T')[0]
				})
			});
			if (response.ok) await loadSermons();
		} catch (error) {
			console.error('Failed to update sermon:', error);
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

<div class="p-6">
	<div class="mb-6 flex justify-between items-center">
		<div>
			<h1 class="text-3xl font-bold text-[var(--text-primary)]">Sermons</h1>
			<p class="mt-1 text-secondary">Manage sermon notes and podcast feed</p>
		</div>
		<button on:click={() => goto('/dashboard/sermons/new')} class="bg-[var(--teal)] text-white px-4 py-2 rounded hover:opacity-90">
			+ New Sermon
		</button>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else}
		<div class="mb-6 grid grid-cols-1 md:grid-cols-3 gap-4">
			<input type="text" placeholder="Search..." bind:value={searchQuery} class="px-4 py-2 border border-custom rounded bg-surface text-primary" />
			<select bind:value={filterSeries} on:change={loadSermons} class="px-4 py-2 border border-custom rounded bg-surface text-primary">
				<option value="">All Series</option>
				{#each series as s}<option value={s}>{s}</option>{/each}
			</select>
			<select bind:value={filterSpeaker} on:change={loadSermons} class="px-4 py-2 border border-custom rounded bg-surface text-primary">
				<option value="">All Speakers</option>
				{#each speakers as speaker}<option value={speaker}>{speaker}</option>{/each}
			</select>
		</div>

		<div class="bg-surface border border-custom rounded-lg shadow">
			{#if filteredSermons.length === 0}
				<div class="p-8 text-center text-secondary">No sermons found</div>
			{:else}
				<table class="w-full">
					<thead class="bg-[var(--surface-hover)] dark:bg-gray-800 border-b border-custom">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase">Date</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase">Title</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase">Speaker</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase">Series</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase">Media</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase">Status</th>
							<th class="px-6 py-3 text-right text-xs font-medium text-secondary uppercase">Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-[var(--border)]">
						{#each filteredSermons as sermon}
							<tr class="hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800">
								<td class="px-6 py-4 whitespace-nowrap text-sm text-[var(--text-primary)]">{formatDate(sermon.sermon_date)}</td>
								<td class="px-6 py-4 text-sm">
									<div class="font-medium text-[var(--text-primary)]">{sermon.title}</div>
									{#if sermon.scripture_reference}
										<div class="text-xs text-secondary">{sermon.scripture_reference}</div>
									{/if}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-[var(--text-primary)]">{sermon.speaker}</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-[var(--text-primary)]">{sermon.series_name || '-'}</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-[var(--text-primary)]">
									{#if sermon.audio_url}🎵{/if}
									{#if sermon.video_url}🎥{/if}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<button on:click={() => togglePublished(sermon)} class="px-2 py-1 rounded text-xs {sermon.published ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100' : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100'}">
										{sermon.published ? 'Published' : 'Draft'}
									</button>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-right text-sm">
									<button on:click={() => goto(`/dashboard/sermons/${sermon.id}`)} class="text-[var(--teal)] hover:opacity-80 mr-3">Edit</button>
									<button on:click={() => deleteSermon(sermon.id)} class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300">Delete</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>
	{/if}
</div>
