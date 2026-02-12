<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast';

	let people = [];
	let selectedPersonId = '';
	let assignments = [];
	let loading = false;
	let searchQuery = '';

	onMount(loadPeople);

	async function loadPeople() {
		try {
			const res = await api('/api/people?limit=500');
			people = res.people || [];
		} catch (e) {
			console.error('Failed to load people:', e);
		}
	}

	async function loadSchedule() {
		if (!selectedPersonId) { assignments = []; return; }
		loading = true;
		try {
			const res = await api(`/api/people/${selectedPersonId}/schedule`);
			assignments = res.assignments || [];
		} catch (e) {
			console.error('Failed to load schedule:', e);
		} finally {
			loading = false;
		}
	}

	async function respondToAssignment(assignmentId, serviceId, status) {
		try {
			await api(`/api/services/${serviceId}/team-assignments/${assignmentId}/status`, {
				method: 'PATCH',
				body: JSON.stringify({ status })
			});
			toast.show(status === 'confirmed' ? 'Confirmed!' : 'Declined', status === 'confirmed' ? 'success' : 'info');
			loadSchedule();
		} catch (e) {
			toast.show('Failed to update', 'error');
		}
	}

	function formatDate(dateStr) {
		if (!dateStr) return '';
		const d = new Date(dateStr + 'T00:00:00');
		return d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' });
	}

	$: filteredPeople = searchQuery.length >= 1
		? people.filter(p => `${p.first_name} ${p.last_name}`.toLowerCase().includes(searchQuery.toLowerCase())).slice(0, 15)
		: [];

	$: if (selectedPersonId) loadSchedule();
</script>

<div>
	<div class="mb-8">
		<h1 class="text-2xl font-bold text-[var(--text-primary)]">My Schedule</h1>
		<p class="text-secondary text-sm mt-1">View upcoming service assignments for a team member</p>
	</div>

	<!-- Person Selector -->
	<div class="mb-6 max-w-md">
		<label class="block text-sm text-secondary mb-1">Select Person</label>
		<div class="relative">
			<input bind:value={searchQuery}
				placeholder="Search by name..."
				on:focus={() => {}}
				class="w-full px-3 py-2 bg-[var(--bg)] border border-custom rounded-lg text-[var(--text-primary)]" />
			{#if filteredPeople.length > 0 && searchQuery.length >= 1}
				<div class="absolute z-10 mt-1 w-full bg-surface border border-custom rounded-lg shadow-lg max-h-48 overflow-y-auto">
					{#each filteredPeople as person}
						<button on:click={() => { selectedPersonId = person.id; searchQuery = `${person.first_name} ${person.last_name}`; filteredPeople = []; }}
							class="w-full text-left px-3 py-2 text-sm hover:bg-[var(--surface-hover)] text-[var(--text-primary)]">
							{person.first_name} {person.last_name}
						</button>
					{/each}
				</div>
			{/if}
		</div>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else if selectedPersonId && assignments.length === 0}
		<div class="text-center py-12">
			<div class="text-4xl mb-3">📅</div>
			<h2 class="text-lg font-semibold text-[var(--text-primary)] mb-1">No upcoming assignments</h2>
			<p class="text-secondary text-sm">This person isn't scheduled for any upcoming services.</p>
		</div>
	{:else if assignments.length > 0}
		<div class="space-y-3">
			{#each assignments as a}
				<div class="bg-surface rounded-xl border border-custom p-4 flex items-center gap-4">
					<div class="w-1.5 h-12 rounded-full" style="background-color: {a.team_color || '#4A8B8C'}"></div>
					<div class="flex-1 min-w-0">
						<div class="flex items-center gap-2 mb-1">
							<span class="font-semibold text-sm text-[var(--text-primary)]">{a.service_type_name || 'Service'}</span>
							<span class="text-xs text-secondary">{formatDate(a.service_date)} {a.service_time || ''}</span>
						</div>
						<div class="flex items-center gap-3 text-xs text-secondary">
							<span class="font-medium" style="color: {a.team_color || '#4A8B8C'}">{a.team_name}</span>
							{#if a.position_name}
								<span>·</span>
								<span>{a.position_name}</span>
							{/if}
						</div>
					</div>
					<div class="flex items-center gap-2 shrink-0">
						{#if a.status === 'confirmed'}
							<span class="px-2 py-1 rounded-full text-xs font-medium bg-green-500/20 text-green-400">✅ Confirmed</span>
						{:else if a.status === 'declined'}
							<span class="px-2 py-1 rounded-full text-xs font-medium bg-red-500/20 text-red-400">❌ Declined</span>
						{:else}
							<button on:click={() => respondToAssignment(a.id, a.service_id, 'confirmed')}
								class="px-3 py-1.5 rounded-lg text-xs font-medium bg-green-600 text-white hover:bg-green-500 transition-colors">
								Accept
							</button>
							<button on:click={() => respondToAssignment(a.id, a.service_id, 'declined')}
								class="px-3 py-1.5 rounded-lg text-xs font-medium bg-[var(--bg)] border border-custom text-secondary hover:text-red-400 transition-colors">
								Decline
							</button>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{:else}
		<div class="text-center py-12 text-secondary text-sm">
			Search for a person above to view their schedule.
		</div>
	{/if}
</div>
