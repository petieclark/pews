<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { Music } from 'lucide-svelte';

	let sermons = [];
	let loading = true;
	let searchQuery = '';
	let filterSeries = '';
	let filterSpeaker = '';
	let filterPublished = '';
	let series = [];
	let speakers = [];
	let viewMode = 'cards';

	onMount(async () => {
		await loadSermons();
		loading = false;
	});

	async function loadSermons() {
		try {
			const params = new URLSearchParams();
			if (searchQuery) params.append('q', searchQuery);
			if (filterSeries) params.append('series', filterSeries);
			if (filterSpeaker) params.append('speaker', filterSpeaker);
			if (filterPublished) params.append('published', filterPublished);

			const data = await api(`/api/sermons?${params.toString()}`);
			sermons = data || [];
			series = [...new Set(sermons.map(s => s.series_name).filter(Boolean))];
			speakers = [...new Set(sermons.map(s => s.speaker).filter(Boolean))];
		} catch (error) {
			console.error('Failed to load sermons:', error);
			sermons = [];
		}
	}

	async function deleteSermon(id) {
		if (!confirm('Are you sure you want to delete this sermon?')) return;
		try {
			await api(`/api/sermons/${id}`, { method: 'DELETE' });
			await loadSermons();
		} catch (error) {
			console.error('Failed to delete sermon:', error);
		}
	}

	async function togglePublished(sermon) {
		try {
			await api(`/api/sermons/${sermon.id}/publish`, {
				method: 'PUT',
				body: JSON.stringify({ published: !sermon.published })
			});
			await loadSermons();
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

	function formatShortDate(dateString) {
		return new Date(dateString).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric'
		});
	}

	$: filteredSermons = sermons.filter(sermon => {
		return searchQuery === '' ||
			sermon.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
			sermon.speaker.toLowerCase().includes(searchQuery.toLowerCase()) ||
			(sermon.scripture_reference && sermon.scripture_reference.toLowerCase().includes(searchQuery.toLowerCase()));
	});

	// Group by series for card view
	$: groupedSermons = (() => {
		const groups = {};
		const standalone = [];
		for (const s of filteredSermons) {
			if (s.series_name) {
				if (!groups[s.series_name]) groups[s.series_name] = [];
				groups[s.series_name].push(s);
			} else {
				standalone.push(s);
			}
		}
		return { groups, standalone };
	})();
</script>

<div class="p-6">
	<div class="mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
		<div>
			<h1 class="text-3xl font-bold text-[var(--text-primary)]">Sermons</h1>
			<p class="mt-1 text-secondary">{filteredSermons.length} sermon{filteredSermons.length !== 1 ? 's' : ''} · Manage notes and podcast feed</p>
		</div>
		<div class="flex gap-2">
			<div class="flex bg-surface border border-custom rounded-lg overflow-hidden">
				<button on:click={() => viewMode = 'cards'} class="px-3 py-1.5 text-sm {viewMode === 'cards' ? 'bg-[var(--teal)] text-white' : 'text-secondary hover:text-primary'}">
					Cards
				</button>
				<button on:click={() => viewMode = 'table'} class="px-3 py-1.5 text-sm {viewMode === 'table' ? 'bg-[var(--teal)] text-white' : 'text-secondary hover:text-primary'}">
					Table
				</button>
			</div>
			<button on:click={() => goto('/dashboard/sermons/new')} class="bg-[var(--teal)] text-white px-4 py-2 rounded-lg hover:opacity-90 text-sm font-medium">
				+ New Sermon
			</button>
		</div>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else}
		<!-- Filters -->
		<div class="mb-6 grid grid-cols-1 md:grid-cols-4 gap-3">
			<input type="text" placeholder="Search sermons..." bind:value={searchQuery} class="px-4 py-2 border border-custom rounded-lg bg-surface text-primary focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
			<select bind:value={filterSeries} on:change={loadSermons} class="px-4 py-2 border border-custom rounded-lg bg-surface text-primary">
				<option value="">All Series</option>
				{#each series as s}<option value={s}>{s}</option>{/each}
			</select>
			<select bind:value={filterSpeaker} on:change={loadSermons} class="px-4 py-2 border border-custom rounded-lg bg-surface text-primary">
				<option value="">All Speakers</option>
				{#each speakers as speaker}<option value={speaker}>{speaker}</option>{/each}
			</select>
			<select bind:value={filterPublished} on:change={loadSermons} class="px-4 py-2 border border-custom rounded-lg bg-surface text-primary">
				<option value="">All Status</option>
				<option value="true">Published</option>
				<option value="false">Drafts</option>
			</select>
		</div>

		{#if filteredSermons.length === 0}
			<div class="bg-surface border border-custom rounded-xl p-12 text-center">
				<svg class="w-16 h-16 mx-auto mb-4 text-secondary opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
				</svg>
				<p class="text-secondary text-lg">No sermons found</p>
				<button on:click={() => goto('/dashboard/sermons/new')} class="mt-4 bg-[var(--teal)] text-white px-4 py-2 rounded-lg hover:opacity-90 text-sm">
					Create your first sermon
				</button>
			</div>
		{:else if viewMode === 'cards'}
			<!-- Card View grouped by series -->
			{#each Object.entries(groupedSermons.groups) as [seriesName, seriesSermons]}
				<div class="mb-8">
					<div class="flex items-center gap-2 mb-3">
						<span class="px-3 py-1 text-sm font-semibold rounded-full bg-[var(--teal)] bg-opacity-15 text-[var(--teal)]">{seriesName}</span>
						<span class="text-xs text-secondary">{seriesSermons.length} sermon{seriesSermons.length !== 1 ? 's' : ''}</span>
					</div>
					<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
						{#each seriesSermons as sermon}
							<button on:click={() => goto(`/dashboard/sermons/${sermon.id}`)} class="bg-surface border border-custom rounded-xl p-5 text-left hover:border-[var(--teal)] transition-colors group">
								<div class="flex items-start justify-between mb-2">
									<span class="text-xs text-secondary">{formatShortDate(sermon.sermon_date)}</span>
									<span class="px-2 py-0.5 text-xs rounded-full {sermon.published ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100' : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100'}">
										{sermon.published ? 'Published' : 'Draft'}
									</span>
								</div>
								<h3 class="font-semibold text-[var(--text-primary)] group-hover:text-[var(--teal)] transition-colors">{sermon.title}</h3>
								<p class="text-sm text-secondary mt-1">{sermon.speaker}</p>
								{#if sermon.scripture_reference}
									<p class="text-xs text-[var(--teal)] mt-2 font-medium">{sermon.scripture_reference}</p>
								{/if}
								<div class="flex items-center gap-2 mt-3 text-xs text-secondary">
									{#if sermon.audio_url}<span><Music size={14} class="inline" /> Audio</span>{/if}
									{#if sermon.video_url}<span>🎥 Video</span>{/if}
								</div>
							</button>
						{/each}
					</div>
				</div>
			{/each}

			{#if groupedSermons.standalone.length > 0}
				<div class="mb-8">
					{#if Object.keys(groupedSermons.groups).length > 0}
						<div class="flex items-center gap-2 mb-3">
							<span class="text-sm font-medium text-secondary">Standalone Sermons</span>
						</div>
					{/if}
					<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
						{#each groupedSermons.standalone as sermon}
							<button on:click={() => goto(`/dashboard/sermons/${sermon.id}`)} class="bg-surface border border-custom rounded-xl p-5 text-left hover:border-[var(--teal)] transition-colors group">
								<div class="flex items-start justify-between mb-2">
									<span class="text-xs text-secondary">{formatShortDate(sermon.sermon_date)}</span>
									<span class="px-2 py-0.5 text-xs rounded-full {sermon.published ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100' : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100'}">
										{sermon.published ? 'Published' : 'Draft'}
									</span>
								</div>
								<h3 class="font-semibold text-[var(--text-primary)] group-hover:text-[var(--teal)] transition-colors">{sermon.title}</h3>
								<p class="text-sm text-secondary mt-1">{sermon.speaker}</p>
								{#if sermon.scripture_reference}
									<p class="text-xs text-[var(--teal)] mt-2 font-medium">{sermon.scripture_reference}</p>
								{/if}
								<div class="flex items-center gap-2 mt-3 text-xs text-secondary">
									{#if sermon.audio_url}<span><Music size={14} class="inline" /> Audio</span>{/if}
									{#if sermon.video_url}<span>🎥 Video</span>{/if}
								</div>
							</button>
						{/each}
					</div>
				</div>
			{/if}

		{:else}
			<!-- Table View -->
			<div class="bg-surface border border-custom rounded-lg shadow overflow-hidden">
				<table class="w-full">
					<thead class="bg-[var(--surface-hover)] border-b border-custom">
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
							<tr class="hover:bg-[var(--surface-hover)]">
								<td class="px-6 py-4 whitespace-nowrap text-sm text-[var(--text-primary)]">{formatDate(sermon.sermon_date)}</td>
								<td class="px-6 py-4 text-sm">
									<div class="font-medium text-[var(--text-primary)]">{sermon.title}</div>
									{#if sermon.scripture_reference}
										<div class="text-xs text-[var(--teal)]">{sermon.scripture_reference}</div>
									{/if}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-[var(--text-primary)]">{sermon.speaker}</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									{#if sermon.series_name}
										<span class="px-2 py-0.5 text-xs rounded-full bg-[var(--teal)] bg-opacity-15 text-[var(--teal)]">{sermon.series_name}</span>
									{:else}
										<span class="text-secondary">—</span>
									{/if}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-[var(--text-primary)]">
									{#if sermon.audio_url}<Music size={14} class="inline" />{/if}
									{#if sermon.video_url}🎥{/if}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<button on:click|stopPropagation={() => togglePublished(sermon)} class="px-2 py-1 rounded text-xs {sermon.published ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100' : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100'}">
										{sermon.published ? 'Published' : 'Draft'}
									</button>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-right text-sm">
									<button on:click={() => goto(`/dashboard/sermons/${sermon.id}`)} class="text-[var(--teal)] hover:opacity-80 mr-3">View</button>
									<button on:click|stopPropagation={() => deleteSermon(sermon.id)} class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300">Delete</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	{/if}
</div>
