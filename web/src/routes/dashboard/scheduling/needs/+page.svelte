<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast';
	import { CalendarDays, Search, UserPlus, AlertCircle, CheckCircle } from 'lucide-svelte';

	// State
	let schedulingData: any = null;
	let loading = false;
	let searchQuery = '';
	let searchResults: any[] = [];
	let showingSearch = false;
	let selectedServiceTeamId = '';
	let assigningPersonId = '';
	let dateFilterStart = '';
	let dateFilterEnd = '';

	// Load scheduling needs on mount
	onMount(async () => {
		await loadSchedulingNeeds();
	});

	async function loadSchedulingNeeds() {
		loading = true;
		try {
			const params = new URLSearchParams();
			if (dateFilterStart) params.append('start', dateFilterStart);
			if (dateFilterEnd) params.append('end', dateFilterEnd);
			
			const res = await api(`/api/scheduling/needs?${params.toString()}`);
			schedulingData = res;
		} catch (e) {
			console.error('Failed to load scheduling needs:', e);
			toast.show('Failed to load scheduling data', 'error');
		} finally {
			loading = false;
		}
	}

	async function searchPeople() {
		if (!searchQuery.trim()) return;
		
		try {
			const res = await api(`/api/people/search?q=${encodeURIComponent(searchQuery)}`);
			searchResults = res.results || [];
			showingSearch = true;
		} catch (e) {
			console.error('Failed to search people:', e);
			toast.show('Failed to search', 'error');
		}
	}

	async function assignPerson(serviceTeamId: string, personId: string) {
		try {
			await api(`/api/services/service-teams/${serviceTeamId}/assign`, {
				method: 'PUT',
				body: JSON.stringify({ person_id: personId })
			});
			
			toast.show('Volunteer assigned!', 'success');
			showingSearch = false;
			searchQuery = '';
			searchResults = [];
			await loadSchedulingNeeds(); // Refresh data
		} catch (e) {
			console.error('Failed to assign:', e);
			toast.show('Failed to assign volunteer', 'error');
		}
	}

	function closeSearch() {
		showingSearch = false;
		searchQuery = '';
		searchResults = [];
	}

	function formatDate(dateStr: string) {
		const d = new Date(dateStr + 'T00:00:00');
		return d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' });
	}

	function getDaysUntil(dateStr: string) {
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		const serviceDate = new Date(dateStr + 'T00:00:00');
		const diffTime = serviceDate.getTime() - today.getTime();
		return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
	}

	function getUrgencyColor(daysUntil: number) {
		if (daysUntil <= 3) return 'bg-red-500/20 text-red-400 border-red-500/50';
		if (daysUntil <= 7) return 'bg-orange-500/20 text-orange-400 border-orange-500/50';
		return '';
	}

	function getUrgencyLabel(daysUntil: number) {
		if (daysUntil <= 3) return 'Critical';
		if (daysUntil <= 7) return 'Urgent';
		if (daysUntil <= 14) return 'Soon';
		return '';
	}

	// Group positions by date for display
	$: groupedByDate = schedulingData?.by_date || {};
	
	// Sort dates ascending
	$: sortedDates = Object.keys(groupedByDate).sort((a, b) => new Date(a).getTime() - new Date(b).getTime());
</script>

<div class="min-h-screen">
	<!-- Header -->
	<div class="mb-8">
		<h1 class="text-2xl font-bold text-[var(--text-primary)]">Scheduling Needs</h1>
		<p class="text-secondary text-sm mt-1">View unfilled volunteer positions and assign team members quickly</p>
		
		<!-- Date Filter -->
		<div class="flex items-center gap-3 mt-4">
			<label class="text-sm text-secondary">Date Range:</label>
			<input type="date" bind:value={dateFilterStart} 
				class="px-2 py-1 bg-[var(--bg)] border border-custom rounded text-sm text-[var(--text-primary)]" />
			<span class="text-secondary">to</span>
			<input type="date" bind:value={dateFilterEnd} 
				class="px-2 py-1 bg-[var(--bg)] border border-custom rounded text-sm text-[var(--text-primary)]" />
			<button on:click={loadSchedulingNeeds}
				class="px-3 py-1 text-xs font-medium bg-[var(--teal)] text-white rounded hover:bg-teal-600 transition-colors">
				Update
			</button>
		</div>
	</div>

	<!-- Loading State -->
	{#if loading}
		<div class="flex justify-center py-12">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
		div>
	{:else if !schedulingData || sortedDates.length === 0}
		<!-- Empty State -->
		<div class="text-center py-16">
			<div class="mb-4"><CalendarDays size={64} class="mx-auto text-secondary opacity-50" /></div>
			<h2 class="text-xl font-semibold text-[var(--text-primary)] mb-2">No scheduling needs found</h2>
			<p class="text-secondary text-sm">All volunteer positions are filled for the selected date range.</p>
		</div>
	{:else}

	<!-- Scheduling Needs by Date -->
	{#each sortedDates as dateStr}
		<div class="mb-8">
			<!-- Date Header with Urgency Badge -->
			<h2 class="text-lg font-semibold text-[var(--text-primary)] mb-3 flex items-center gap-2">
				<span>{formatDate(dateStr)}</span>
				{#if getDaysUntil(dateStr) <= 7}
					<span class="px-2 py-0.5 rounded-full text-xs font-medium bg-red-500/20 text-red-400 border border-red-500/50">
						{getUrgencyLabel(getDaysUntil(dateStr))}
					</span>
				{/if}
			</h2>

			<!-- Positions Grouped by Team -->
			<div class="space-y-3">
				{#each groupingByTeam(schedulingData.by_date[dateStr]) as teamPositions, teamName}
					<div class="bg-surface rounded-xl border border-custom p-4">
						{#if teamName && teamName !== 'null'}
							<h3 class="text-sm font-semibold text-[var(--text-primary)] mb-2 flex items-center gap-2" 
								style="color: {teamPositions[0].team_color || '#4A8B8C'}">
								<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
									<circle cx="9" cy="7" r="4"></circle>
									<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
									<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
								</svg>
								{teamName}
							</h3>
						{/if}

						<!-- Unfilled Positions -->
						<div class="space-y-2">
							{#each teamPositions as pos}
								<div class="bg-[var(--bg)] rounded-lg p-3 border-l-4 {getUrgencyColor(getDaysUntil(pos.service_date))}"
									style="border-left-color: {pos.team_color || '#4A8B8C'}">
									<div class="flex items-start justify-between gap-3">
										<div class="flex-1 min-w-0">
											<div class="flex items-center gap-2 mb-1">
												<span class="font-medium text-sm text-[var(--text-primary)]">{pos.role}</span>
												{#if pos.is_urgent}
													<span class="px-1.5 py-0.5 rounded text-xs bg-red-500/30 text-red-400">!</span>
												{/if}
											</div>
											<div class="text-xs text-secondary">
												Service: {pos.service_name || 'Worship Service'} · {pos.service_time || ''}
											</div>
										</div>

										<!-- Quick Assign Button -->
										<button on:click={() => { selectedServiceTeamId = pos.id; showingSearch = true; assignPerson = null; }}
											class="flex items-center gap-1 px-3 py-1.5 rounded-lg text-xs font-medium bg-[var(--teal)] text-white hover:bg-teal-600 transition-colors shrink-0">
											<UserPlus size={14} />
											Assign
										</button>
									</div>
								</div>
							{/each}

							{#if teamPositions.every(p => p.person_id)}
								<div class="text-center py-2 px-3 bg-green-500/10 border border-green-500/30 rounded-lg">
									<span class="text-sm text-green-400 flex items-center justify-center gap-1">
										<CheckCircle size={14} /> All positions filled!
									</span>
								</div>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/each}

	<!-- Quick Assign Modal/Search Overlay -->
	{#if showingSearch && selectedServiceTeamId}
		<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" on:click={closeSearch}>
			<div class="bg-surface rounded-xl border border-custom max-w-md w-full shadow-2xl" on:click|stop>
				<div class="p-4 border-b border-custom">
					<h3 class="text-lg font-semibold text-[var(--text-primary)] flex items-center gap-2">
						<Search size={18} />
						Assign Volunteer
					</h3>
					<p class="text-xs text-secondary mt-1">Search to find and assign a volunteer</p>
				</div>

				<!-- Search Input -->
				<div class="p-4 border-b border-custom">
					<input type="text" bind:value={searchQuery} 
						placeholder="Search by name or email..."
						on:input={searchPeople}
						class="w-full px-3 py-2 bg-[var(--bg)] border border-custom rounded-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--teal)]"
						focus />
				</div>

				<!-- Results List -->
				<div class="p-4 max-h-64 overflow-y-auto">
					{#if searchResults.length === 0 && searchQuery}
						<p class="text-center text-secondary text-sm py-4">No matches found.</p>
					{:else if searchResults.length > 0}
						<div class="space-y-2">
							{#each searchResults as person, idx}
								<button on:click={() => assignPerson(selectedServiceTeamId, person.id)}
									class="w-full text-left p-3 rounded-lg hover:bg-[var(--surface-hover)] transition-colors group"
									style={idx === 0 ? 'background-color: var(--surface-hover)' : ''}>
									<div class="flex items-center justify-between">
										<div>
											<div class="font-medium text-sm text-[var(--text-primary)]">
												{person.first_name} {person.last_name}
											</div>
											{#if person.email}
												<div class="text-xs text-secondary">{person.email}</div>
											{/if}
										</div>
										{#if person.is_available}
											<span class="px-2 py-0.5 rounded-full text-xs bg-green-500/20 text-green-400">Available</span>
										{:else}
											<span class="px-2 py-0.5 rounded-full text-xs bg-orange-500/20 text-orange-400">Busy</span>
										{/if}
									</div>
								</button>
							{/each}
						</div>
					{:else if !searchQuery}
						<div class="text-center py-6 text-secondary text-sm">
							Start typing to search for volunteers.
						</div>
					{/if}
				</div>

				<!-- Close Button -->
				<div class="p-4 border-t border-custom flex justify-end">
					<button on:click={closeSearch}
						class="px-4 py-2 text-sm font-medium bg-[var(--bg)] text-[var(--text-primary)] rounded-lg hover:bg-[var(--surface-hover)] transition-colors">
						Cancel
					</button>
				</div>
			</div>
		</div>
	{/if}

	{/if}
</div>

<style>
	/* Custom scrollbar for search results */
	div.max-h-64.overflow-y-auto::-webkit-scrollbar {
		width: 6px;
	}
	div.max-h-64.overflow-y-auto::-webkit-scrollbar-track {
		background: var(--bg);
		border-radius: 3px;
	}
	div.max-h-64.overflow-y-auto::-webkit-scrollbar-thumb {
		background: var(--teal);
		border-radius: 3px;
	}
</style>

<!-- Helper function for grouping -->
<script>
	function groupByTeam(positions) {
		const grouped = {};
		for (const pos of positions) {
			const teamName = pos.team_name || 'Unassigned';
			if (!grouped[teamName]) grouped[teamName] = [];
			grouped[teamName].push(pos);
		}
		return grouped;
	}
</script>
