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
			<h1 class="text-3xl font-bold">Sermons</h1>
			<p class="mt-1 text-gray-600">Manage sermon notes and podcast feed</p>
		</div>
		<button on:click={() => goto('/dashboard/sermons/new')} class="bg-teal-600 text-white px-4 py-2 rounded hover:bg-teal-700">
			+ New Sermon
		</button>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-teal-600"></div>
		</div>
	{:else}
		<div class="mb-6 grid grid-cols-1 md:grid-cols-3 gap-4">
			<input type="text" placeholder="Search..." bind:value={searchQuery} class="px-4 py-2 border rounded" />
			<select bind:value={filterSeries} on:change={loadSermons} class="px-4 py-2 border rounded">
				<option value="">All Series</option>
				{#each series as s}<option value={s}>{s}</option>{/each}
			</select>
			<select bind:value={filterSpeaker} on:change={loadSermons} class="px-4 py-2 border rounded">
				<option value="">All Speakers</option>
				{#each speakers as speaker}<option value={speaker}>{speaker}</option>{/each}
			</select>
		</div>

		<div class="bg-white rounded-lg shadow">
			{#if filteredSermons.length === 0}
				<div class="p-8 text-center text-gray-500">No sermons found</div>
			{:else}
				<table class="w-full">
					<thead class="bg-gray-50 border-b">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Title</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Speaker</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Series</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Media</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
							<th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y">
						{#each filteredSermons as sermon}
							<tr class="hover:bg-gray-50">
								<td class="px-6 py-4 whitespace-nowrap text-sm">{formatDate(sermon.sermon_date)}</td>
								<td class="px-6 py-4 text-sm">
									<div class="font-medium">{sermon.title}</div>
									{#if sermon.scripture_reference}
										<div class="text-xs text-gray-500">{sermon.scripture_reference}</div>
									{/if}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">{sermon.speaker}</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">{sermon.series_name || '-'}</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									{#if sermon.audio_url}🎵{/if}
									{#if sermon.video_url}🎥{/if}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<button on:click={() => togglePublished(sermon)} class="px-2 py-1 rounded text-xs {sermon.published ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'}">
										{sermon.published ? 'Published' : 'Draft'}
									</button>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-right text-sm">
									<button on:click={() => goto(`/dashboard/sermons/${sermon.id}`)} class="text-teal-600 hover:text-teal-900 mr-3">Edit</button>
									<button on:click={() => deleteSermon(sermon.id)} class="text-red-600 hover:text-red-900">Delete</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>
	{/if}
</div>
