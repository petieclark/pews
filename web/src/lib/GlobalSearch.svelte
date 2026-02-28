<script>
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { User, Music, Building2, Users, ClipboardList } from 'lucide-svelte';

	let query = '';
	let results = null;
	let showResults = false;
	let loading = false;
	let searchTimeout;
	let selectedIndex = -1;
	let inputEl;

	$: flatResults = results ? [
		...(results.people || []).map(r => ({ ...r, type: 'person', href: `/dashboard/people/${r.id}` })),
		...(results.songs || []).map(r => ({ ...r, type: 'song', href: `/dashboard/services/songs/${r.id}` })),
		...(results.services || []).map(r => ({ ...r, type: 'service', href: `/dashboard/services/${r.id}` })),
		...(results.groups || []).map(r => ({ ...r, type: 'group', href: `/dashboard/groups/${r.id}` })),
	] : [];

	async function handleInput() {
		clearTimeout(searchTimeout);
		if (!query.trim()) {
			results = null;
			showResults = false;
			return;
		}
		searchTimeout = setTimeout(async () => {
			loading = true;
			try {
				results = await api(`/api/search?q=${encodeURIComponent(query)}`);
				showResults = true;
				selectedIndex = -1;
			} catch (e) {
				results = null;
			} finally {
				loading = false;
			}
		}, 250);
	}

	function handleKeydown(e) {
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			selectedIndex = Math.min(selectedIndex + 1, flatResults.length - 1);
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			selectedIndex = Math.max(selectedIndex - 1, -1);
		} else if (e.key === 'Enter' && selectedIndex >= 0) {
			e.preventDefault();
			navigateTo(flatResults[selectedIndex].href);
		} else if (e.key === 'Escape') {
			showResults = false;
			inputEl?.blur();
		}
	}

	function navigateTo(href) {
		showResults = false;
		query = '';
		results = null;
		goto(href);
	}

	function getTypeIcon(type) {
		const icons = { person: User, song: Music, service: Building2, group: Users };
		return icons[type] || ClipboardList;
	}

	function getTypeLabel(type) {
		const labels = { person: 'People', song: 'Songs', service: 'Services', group: 'Groups' };
		return labels[type] || type;
	}

	function getResultLabel(item) {
		if (item.type === 'person') return item.name;
		if (item.type === 'song') return item.title;
		if (item.type === 'service') return `${item.name} — ${item.date}`;
		if (item.type === 'group') return item.name;
		return item.name || item.title || '';
	}

	function getResultSub(item) {
		if (item.type === 'person') return item.email || '';
		if (item.type === 'song') return item.artist || '';
		return '';
	}

	function handleBlur() {
		setTimeout(() => { showResults = false; }, 200);
	}
</script>

<div class="relative">
	<div class="relative">
		<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
		</svg>
		<input
			bind:this={inputEl}
			type="text"
			bind:value={query}
			on:input={handleInput}
			on:keydown={handleKeydown}
			on:focus={() => { if (results) showResults = true; }}
			on:blur={handleBlur}
			placeholder="Search people, songs, groups..."
			class="w-64 pl-9 pr-3 py-1.5 text-sm border rounded-lg bg-[var(--bg)] border-[var(--border)] text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent transition-all focus:w-80"
		/>
		{#if loading}
			<div class="absolute right-3 top-1/2 -translate-y-1/2">
				<div class="w-3 h-3 border-2 border-[var(--teal)] border-t-transparent rounded-full animate-spin"></div>
			</div>
		{/if}
	</div>

	{#if showResults && results}
		<div class="absolute top-full mt-2 left-0 right-0 w-80 bg-surface border border-custom rounded-xl shadow-xl z-50 overflow-hidden max-h-96 overflow-y-auto">
			{#if flatResults.length === 0}
				<div class="p-4 text-center text-sm text-secondary">No results found for "{query}"</div>
			{:else}
				{#each ['person', 'song', 'service', 'group'] as type}
					{@const items = flatResults.filter(r => r.type === type)}
					{#if items.length > 0}
						<div class="px-3 py-1.5 text-xs font-semibold text-secondary uppercase tracking-wider bg-[var(--surface-hover)]">
							{getTypeLabel(type)}
						</div>
						{#each items as item, i}
							{@const globalIndex = flatResults.indexOf(item)}
							<button
								class="w-full text-left px-3 py-2.5 flex items-center gap-3 hover:bg-[var(--surface-hover)] transition-colors
									{globalIndex === selectedIndex ? 'bg-[var(--surface-hover)]' : ''}"
								on:mousedown|preventDefault={() => navigateTo(item.href)}
							>
								<span class="flex-shrink-0"><svelte:component this={getTypeIcon(type)} size={16} /></span>
								<div class="min-w-0 flex-1">
									<p class="text-sm font-medium text-primary truncate">{getResultLabel(item)}</p>
									{#if getResultSub(item)}
										<p class="text-xs text-secondary truncate">{getResultSub(item)}</p>
									{/if}
								</div>
							</button>
						{/each}
					{/if}
				{/each}
			{/if}
		</div>
	{/if}
</div>

<style>
	input:focus {
		width: 20rem;
	}
</style>
