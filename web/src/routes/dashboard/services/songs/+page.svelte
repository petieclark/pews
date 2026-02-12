<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let songs = [];
	let total = 0;
	let page = 1;
	let limit = 20;
	let searchQuery = '';
	let loading = false;
	let viewMode = 'list'; // 'list' or 'card'

	// Filters
	let filterKey = '';
	let filterTag = '';
	let filterHasLyrics = '';
	let sortBy = 'title';

	// Stats
	let stats = null;
	let statsLoading = true;

	// Available filter options (from stats)
	let allKeys = [];
	let allTags = [];

	onMount(() => {
		loadStats();
		loadSongs();
	});

	async function loadStats() {
		statsLoading = true;
		try {
			stats = await api('/api/services/songs/stats');
			allKeys = stats.all_keys || [];
			allTags = stats.all_tags || [];
		} catch (error) {
			console.error('Failed to load stats:', error);
		} finally {
			statsLoading = false;
		}
	}

	async function loadSongs() {
		loading = true;
		try {
			const params = new URLSearchParams({
				page: page.toString(),
				limit: limit.toString()
			});
			if (searchQuery) params.append('q', searchQuery);
			if (filterKey) params.append('key', filterKey);
			if (filterTag) params.append('tag', filterTag);
			if (filterHasLyrics) params.append('has_lyrics', filterHasLyrics);
			if (sortBy) params.append('sort', sortBy);

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

	function applyFilters() {
		page = 1;
		loadSongs();
	}

	function clearFilters() {
		filterKey = '';
		filterTag = '';
		filterHasLyrics = '';
		sortBy = 'title';
		searchQuery = '';
		page = 1;
		loadSongs();
	}

	function formatDate(dateStr) {
		if (!dateStr) return 'Never';
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	function getTagColor(tag) {
		const colors = [
			'bg-teal/20 text-teal',
			'bg-blue-500/20 text-blue-400',
			'bg-purple-500/20 text-purple-400',
			'bg-amber-500/20 text-amber-400',
			'bg-rose-500/20 text-rose-400',
			'bg-green-500/20 text-green-400',
			'bg-cyan-500/20 text-cyan-400',
			'bg-orange-500/20 text-orange-400',
		];
		let hash = 0;
		for (let i = 0; i < tag.length; i++) {
			hash = tag.charCodeAt(i) + ((hash << 5) - hash);
		}
		return colors[Math.abs(hash) % colors.length];
	}

	$: totalPages = Math.ceil(total / limit);
	$: hasFilters = filterKey || filterTag || filterHasLyrics || searchQuery;
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex justify-between items-center">
		<div>
			<button
				on:click={() => goto('/dashboard/services')}
				class="text-teal hover:underline mb-2 text-sm flex items-center gap-1"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
				Back to Services
			</button>
			<h1 class="text-3xl font-bold text-[var(--text-primary)]">Song Library</h1>
			{#if !statsLoading && stats}
				<p class="text-sm text-[var(--text-secondary)] mt-1">{stats.total_songs} songs in library</p>
			{/if}
		</div>
		<button
			on:click={() => goto('/dashboard/services/songs/new')}
			class="px-4 py-2.5 bg-teal text-white rounded-lg hover:bg-opacity-90 font-medium flex items-center gap-2"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
			Add Song
		</button>
	</div>

	<!-- Stats Dashboard -->
	{#if !statsLoading && stats}
		<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
			<div class="bg-[var(--surface)] rounded-lg p-4 border border-[var(--border)]">
				<div class="text-2xl font-bold text-[var(--text-primary)]">{stats.total_songs}</div>
				<div class="text-sm text-[var(--text-secondary)]">Total Songs</div>
			</div>
			<div class="bg-[var(--surface)] rounded-lg p-4 border border-[var(--border)]">
				<div class="text-2xl font-bold text-[var(--text-primary)]">{stats.with_lyrics}</div>
				<div class="text-sm text-[var(--text-secondary)]">With Lyrics/Chords</div>
			</div>
			<div class="bg-[var(--surface)] rounded-lg p-4 border border-[var(--border)]">
				<div class="text-2xl font-bold text-[var(--text-primary)]">{stats.with_attachments}</div>
				<div class="text-sm text-[var(--text-secondary)]">With PDFs</div>
			</div>
			<div class="bg-[var(--surface)] rounded-lg p-4 border border-[var(--border)]">
				<div class="text-2xl font-bold text-teal">{allKeys.length}</div>
				<div class="text-sm text-[var(--text-secondary)]">Unique Keys</div>
			</div>
		</div>

		<!-- Most Used & Recently Added -->
		{#if (stats.most_used && stats.most_used.length > 0) || (stats.recently_added && stats.recently_added.length > 0)}
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
				{#if stats.most_used && stats.most_used.length > 0}
					<div class="bg-[var(--surface)] rounded-lg p-4 border border-[var(--border)]">
						<h3 class="text-sm font-semibold text-[var(--text-secondary)] uppercase tracking-wider mb-3">Most Used</h3>
						<div class="space-y-2">
							{#each stats.most_used as song, i}
								<button
									on:click={() => goto(`/dashboard/services/songs/${song.id}`)}
									class="w-full flex items-center gap-3 p-2 rounded hover:bg-[var(--surface-hover)] transition-colors text-left"
								>
									<span class="text-xs font-mono text-[var(--text-secondary)] w-4">{i + 1}.</span>
									<div class="flex-1 min-w-0">
										<div class="text-sm font-medium text-[var(--text-primary)] truncate">{song.title}</div>
										<div class="text-xs text-[var(--text-secondary)]">{song.artist || 'Unknown'}</div>
									</div>
									<span class="text-xs font-medium text-teal">{song.times_used}×</span>
								</button>
							{/each}
						</div>
					</div>
				{/if}
				{#if stats.recently_added && stats.recently_added.length > 0}
					<div class="bg-[var(--surface)] rounded-lg p-4 border border-[var(--border)]">
						<h3 class="text-sm font-semibold text-[var(--text-secondary)] uppercase tracking-wider mb-3">Recently Added</h3>
						<div class="space-y-2">
							{#each stats.recently_added as song}
								<button
									on:click={() => goto(`/dashboard/services/songs/${song.id}`)}
									class="w-full flex items-center gap-3 p-2 rounded hover:bg-[var(--surface-hover)] transition-colors text-left"
								>
									<div class="flex-1 min-w-0">
										<div class="text-sm font-medium text-[var(--text-primary)] truncate">{song.title}</div>
										<div class="text-xs text-[var(--text-secondary)]">{song.artist || 'Unknown'}</div>
									</div>
									<span class="text-xs text-[var(--text-secondary)]">{formatDate(song.created_at)}</span>
								</button>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{/if}
	{/if}

	<!-- Search & Filters -->
	<div class="bg-[var(--surface)] rounded-lg border border-[var(--border)] p-4 space-y-4">
		<!-- Search bar + view toggle -->
		<div class="flex gap-3 items-center">
			<div class="relative flex-1">
				<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-secondary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>
				<input
					type="text"
					bind:value={searchQuery}
					on:input={() => { page = 1; loadSongs(); }}
					placeholder="Search by title, artist, CCLI number..."
					class="w-full pl-10 pr-4 py-2.5 border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)] rounded-lg focus:outline-none focus:ring-2 focus:ring-teal/50 focus:border-teal"
				/>
			</div>
			<!-- View toggle -->
			<div class="flex rounded-lg border border-[var(--border)] overflow-hidden">
				<button
					on:click={() => viewMode = 'list'}
					class="p-2.5 {viewMode === 'list' ? 'bg-teal text-white' : 'bg-[var(--bg)] text-[var(--text-secondary)] hover:text-[var(--text-primary)]'}"
					title="List view"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/></svg>
				</button>
				<button
					on:click={() => viewMode = 'card'}
					class="p-2.5 {viewMode === 'card' ? 'bg-teal text-white' : 'bg-[var(--bg)] text-[var(--text-secondary)] hover:text-[var(--text-primary)]'}"
					title="Card view"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm10 0a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zM4 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4zm10 0a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"/></svg>
				</button>
			</div>
		</div>

		<!-- Filter row -->
		<div class="flex flex-wrap gap-3 items-center">
			<select
				bind:value={filterKey}
				on:change={applyFilters}
				class="px-3 py-2 border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)] rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-teal/50"
			>
				<option value="">All Keys</option>
				{#each allKeys as key}
					<option value={key}>{key}</option>
				{/each}
			</select>

			<select
				bind:value={filterTag}
				on:change={applyFilters}
				class="px-3 py-2 border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)] rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-teal/50"
			>
				<option value="">All Tags</option>
				{#each allTags as tag}
					<option value={tag}>{tag}</option>
				{/each}
			</select>

			<select
				bind:value={filterHasLyrics}
				on:change={applyFilters}
				class="px-3 py-2 border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)] rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-teal/50"
			>
				<option value="">Lyrics: Any</option>
				<option value="yes">Has Lyrics</option>
				<option value="no">No Lyrics</option>
			</select>

			<select
				bind:value={sortBy}
				on:change={applyFilters}
				class="px-3 py-2 border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)] rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-teal/50"
			>
				<option value="title">Sort: Title A-Z</option>
				<option value="title_desc">Sort: Title Z-A</option>
				<option value="artist">Sort: Artist</option>
				<option value="last_used">Sort: Last Used</option>
				<option value="times_used">Sort: Most Used</option>
				<option value="recently_added">Sort: Recently Added</option>
			</select>

			{#if hasFilters}
				<button
					on:click={clearFilters}
					class="text-sm text-teal hover:underline"
				>
					Clear filters
				</button>
			{/if}

			<div class="ml-auto text-sm text-[var(--text-secondary)]">
				{total} result{total !== 1 ? 's' : ''}
			</div>
		</div>
	</div>

	<!-- Song List / Cards -->
	{#if loading}
		<div class="flex justify-center py-12">
			<div class="animate-spin rounded-full h-8 w-8 border-4 border-teal border-t-transparent"></div>
		</div>
	{:else if songs.length === 0}
		<div class="bg-[var(--surface)] rounded-lg border border-[var(--border)] p-12 text-center">
			<svg class="mx-auto h-12 w-12 text-[var(--text-secondary)] mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3"/></svg>
			<h3 class="text-lg font-medium text-[var(--text-primary)]">No songs found</h3>
			<p class="text-[var(--text-secondary)] mt-1">
				{#if hasFilters}Try adjusting your filters or search.{:else}Add your first song to get started.{/if}
			</p>
		</div>
	{:else if viewMode === 'list'}
		<!-- List View -->
		<div class="bg-[var(--surface)] rounded-lg border border-[var(--border)] overflow-hidden">
			<table class="min-w-full">
				<thead>
					<tr class="border-b border-[var(--border)]">
						<th class="px-5 py-3 text-left text-xs font-semibold text-[var(--text-secondary)] uppercase tracking-wider">Title</th>
						<th class="px-5 py-3 text-left text-xs font-semibold text-[var(--text-secondary)] uppercase tracking-wider hidden md:table-cell">Artist</th>
						<th class="px-5 py-3 text-left text-xs font-semibold text-[var(--text-secondary)] uppercase tracking-wider hidden lg:table-cell">Key</th>
						<th class="px-5 py-3 text-left text-xs font-semibold text-[var(--text-secondary)] uppercase tracking-wider hidden lg:table-cell">Tempo</th>
						<th class="px-5 py-3 text-left text-xs font-semibold text-[var(--text-secondary)] uppercase tracking-wider hidden md:table-cell">Tags</th>
						<th class="px-5 py-3 text-left text-xs font-semibold text-[var(--text-secondary)] uppercase tracking-wider hidden sm:table-cell">Last Used</th>
						<th class="px-5 py-3 text-right text-xs font-semibold text-[var(--text-secondary)] uppercase tracking-wider hidden sm:table-cell">Used</th>
					</tr>
				</thead>
				<tbody>
					{#each songs as song}
						<tr
							on:click={() => goto(`/dashboard/services/songs/${song.id}`)}
							class="border-b border-[var(--border)] hover:bg-[var(--surface-hover)] cursor-pointer transition-colors"
						>
							<td class="px-5 py-3.5">
								<div class="flex items-center gap-2">
									<span class="text-sm font-medium text-[var(--text-primary)]">{song.title}</span>
									{#if song.youtube_url}<span class="text-xs text-red-400" title="YouTube">▶</span>{/if}
									{#if song.spotify_url}<span class="text-xs text-green-400" title="Spotify">🟢</span>{/if}
									{#if song.apple_music_url}<span class="text-xs text-pink-400" title="Apple Music">🍎</span>{/if}
									{#if song.rehearsal_url}<span class="text-xs text-amber-400" title="Rehearsal">🎧</span>{/if}
								</div>
								{#if song.ccli_number}
									<div class="text-xs text-[var(--text-secondary)] font-mono">CCLI {song.ccli_number}</div>
								{/if}
							</td>
							<td class="px-5 py-3.5 hidden md:table-cell">
								<div class="text-sm text-[var(--text-secondary)]">{song.artist || '—'}</div>
							</td>
							<td class="px-5 py-3.5 hidden lg:table-cell">
								{#if song.default_key}
									<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-teal/20 text-teal">
										{song.default_key}
									</span>
								{:else}
									<span class="text-sm text-[var(--text-secondary)]">—</span>
								{/if}
							</td>
							<td class="px-5 py-3.5 hidden lg:table-cell">
								<div class="text-sm text-[var(--text-secondary)]">{song.tempo ? `${song.tempo} BPM` : '—'}</div>
							</td>
							<td class="px-5 py-3.5 hidden md:table-cell">
								<div class="flex flex-wrap gap-1">
									{#if song.tags}
										{#each song.tags.split(',').slice(0, 3) as tag}
											<span class="inline-flex items-center px-1.5 py-0.5 rounded text-xs {getTagColor(tag.trim())}">
												{tag.trim()}
											</span>
										{/each}
										{#if song.tags.split(',').length > 3}
											<span class="text-xs text-[var(--text-secondary)]">+{song.tags.split(',').length - 3}</span>
										{/if}
									{/if}
								</div>
							</td>
							<td class="px-5 py-3.5 hidden sm:table-cell">
								<div class="text-sm text-[var(--text-secondary)]">{formatDate(song.last_used)}</div>
							</td>
							<td class="px-5 py-3.5 text-right hidden sm:table-cell">
								<div class="text-sm font-medium text-[var(--text-primary)]">{song.times_used}</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{:else}
		<!-- Card View -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
			{#each songs as song}
				<button
					on:click={() => goto(`/dashboard/services/songs/${song.id}`)}
					class="bg-[var(--surface)] rounded-lg border border-[var(--border)] p-4 hover:border-teal/50 hover:shadow-lg transition-all text-left group"
				>
					<div class="flex items-start justify-between mb-2">
						<h3 class="text-sm font-semibold text-[var(--text-primary)] group-hover:text-teal transition-colors line-clamp-2">{song.title}</h3>
						{#if song.default_key}
							<span class="ml-2 shrink-0 inline-flex items-center px-2 py-0.5 rounded text-xs font-bold bg-teal/20 text-teal">
								{song.default_key}
							</span>
						{/if}
					</div>
					<div class="text-xs text-[var(--text-secondary)] mb-3">{song.artist || 'Unknown Artist'}</div>

					<div class="flex items-center gap-3 text-xs text-[var(--text-secondary)] mb-3">
						{#if song.tempo}
							<span>♩ {song.tempo}</span>
						{/if}
						{#if song.ccli_number}
							<span class="font-mono">CCLI {song.ccli_number}</span>
						{/if}
					</div>

					{#if song.youtube_url || song.spotify_url || song.apple_music_url || song.rehearsal_url}
						<div class="flex items-center gap-2 mb-3">
							{#if song.youtube_url}
								<span class="text-xs px-1.5 py-0.5 rounded bg-red-500/15 text-red-400" title="YouTube">▶</span>
							{/if}
							{#if song.spotify_url}
								<span class="text-xs px-1.5 py-0.5 rounded bg-green-500/15 text-green-400" title="Spotify">🟢</span>
							{/if}
							{#if song.apple_music_url}
								<span class="text-xs px-1.5 py-0.5 rounded bg-pink-500/15 text-pink-400" title="Apple Music">🍎</span>
							{/if}
							{#if song.rehearsal_url}
								<span class="text-xs px-1.5 py-0.5 rounded bg-amber-500/15 text-amber-400" title="Rehearsal Track">🎧</span>
							{/if}
						</div>
					{/if}

					{#if song.tags}
						<div class="flex flex-wrap gap-1 mb-3">
							{#each song.tags.split(',').slice(0, 3) as tag}
								<span class="inline-flex items-center px-1.5 py-0.5 rounded text-xs {getTagColor(tag.trim())}">
									{tag.trim()}
								</span>
							{/each}
						</div>
					{/if}

					<div class="flex items-center justify-between text-xs text-[var(--text-secondary)] pt-2 border-t border-[var(--border)]">
						<span>{formatDate(song.last_used)}</span>
						<span class="font-medium">{song.times_used}× used</span>
					</div>
				</button>
			{/each}
		</div>
	{/if}

	<!-- Pagination -->
	{#if totalPages > 1}
		<div class="flex items-center justify-center gap-2">
			<button
				on:click={() => { page = 1; loadSongs(); }}
				disabled={page === 1}
				class="px-3 py-2 text-sm bg-[var(--surface)] border border-[var(--border)] rounded-lg disabled:opacity-30 hover:bg-[var(--surface-hover)] text-[var(--text-primary)]"
			>
				First
			</button>
			<button
				on:click={() => { page--; loadSongs(); }}
				disabled={page === 1}
				class="px-3 py-2 text-sm bg-[var(--surface)] border border-[var(--border)] rounded-lg disabled:opacity-30 hover:bg-[var(--surface-hover)] text-[var(--text-primary)]"
			>
				← Prev
			</button>
			
			{#each Array.from({length: Math.min(5, totalPages)}, (_, i) => {
				let start = Math.max(1, Math.min(page - 2, totalPages - 4));
				return start + i;
			}).filter(p => p <= totalPages) as p}
				<button
					on:click={() => { page = p; loadSongs(); }}
					class="px-3 py-2 text-sm rounded-lg {p === page ? 'bg-teal text-white' : 'bg-[var(--surface)] border border-[var(--border)] hover:bg-[var(--surface-hover)] text-[var(--text-primary)]'}"
				>
					{p}
				</button>
			{/each}

			<button
				on:click={() => { page++; loadSongs(); }}
				disabled={page >= totalPages}
				class="px-3 py-2 text-sm bg-[var(--surface)] border border-[var(--border)] rounded-lg disabled:opacity-30 hover:bg-[var(--surface-hover)] text-[var(--text-primary)]"
			>
				Next →
			</button>
			<button
				on:click={() => { page = totalPages; loadSongs(); }}
				disabled={page >= totalPages}
				class="px-3 py-2 text-sm bg-[var(--surface)] border border-[var(--border)] rounded-lg disabled:opacity-30 hover:bg-[var(--surface-hover)] text-[var(--text-primary)]"
			>
				Last
			</button>
		</div>
	{/if}
</div>

<style>
	.line-clamp-2 {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>
