<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let groups = [];
	let total = 0;
	let page = 1;
	let limit = 50;
	let loading = false;
	let showCreateModal = false;
	let filterType = '';
	let filterActive = '';

	let newGroup = {
		name: '',
		description: '',
		group_type: 'small_group',
		meeting_day: '',
		meeting_time: '',
		meeting_location: '',
		is_public: true,
		is_active: true,
		max_members: null,
		photo_url: ''
	};

	let leaderSearch = '';
	let leaderResults = [];
	let selectedLeader = null;
	let leaderSearchTimeout;

	const groupTypes = [
		{ value: 'small_group', label: 'Small Group' },
		{ value: 'ministry_team', label: 'Ministry Team' },
		{ value: 'class', label: 'Class' },
		{ value: 'committee', label: 'Committee' },
		{ value: 'bible_study', label: 'Bible Study' },
		{ value: 'team', label: 'Team' }
	];

	const meetingDays = [
		{ value: 'monday', label: 'Monday' },
		{ value: 'tuesday', label: 'Tuesday' },
		{ value: 'wednesday', label: 'Wednesday' },
		{ value: 'thursday', label: 'Thursday' },
		{ value: 'friday', label: 'Friday' },
		{ value: 'saturday', label: 'Saturday' },
		{ value: 'sunday', label: 'Sunday' }
	];

	onMount(() => {
		loadGroups();
	});

	async function loadGroups() {
		loading = true;
		try {
			const params = new URLSearchParams({ page: page.toString(), limit: limit.toString() });
			if (filterType) params.append('type', filterType);
			if (filterActive) params.append('active', filterActive);
			const response = await api(`/api/groups?${params}`);
			groups = response.groups || [];
			total = response.total || 0;

			// Load members for each group to find leaders
			for (let g of groups) {
				try {
					const members = await api(`/api/groups/${g.id}/members`);
					const leader = (members || []).find(m => m.role === 'leader');
					g.leader_name = leader ? `${leader.person?.first_name || ''} ${leader.person?.last_name || ''}`.trim() : null;
				} catch { g.leader_name = null; }
			}
			groups = groups; // trigger reactivity
		} catch (error) {
			console.error('Failed to load groups:', error);
		} finally {
			loading = false;
		}
	}

	function handleFilter() { page = 1; loadGroups(); }
	function viewGroup(id) { goto(`/dashboard/groups/${id}`); }

	function handleLeaderSearch() {
		clearTimeout(leaderSearchTimeout);
		if (leaderSearch.length < 2) { leaderResults = []; return; }
		leaderSearchTimeout = setTimeout(async () => {
			try {
				leaderResults = await api(`/api/checkins/search?q=${encodeURIComponent(leaderSearch)}`);
			} catch { leaderResults = []; }
		}, 300);
	}

	function selectLeader(person) {
		selectedLeader = person;
		leaderSearch = `${person.first_name} ${person.last_name}`;
		leaderResults = [];
	}

	async function createGroup() {
		try {
			const payload = { ...newGroup };
			if (!payload.max_members) payload.max_members = null;
			const created = await api('/api/groups', { method: 'POST', body: JSON.stringify(payload) });
			if (selectedLeader && created.id) {
				try {
					await api(`/api/groups/${created.id}/members`, { method: 'POST', body: JSON.stringify({ person_id: selectedLeader.id, role: 'leader' }) });
				} catch (e) { console.error('Failed to add leader:', e); }
			}
			showCreateModal = false;
			resetNewGroup();
			loadGroups();
		} catch (error) { alert('Failed to create group: ' + error.message); }
	}

	function resetNewGroup() {
		newGroup = { name: '', description: '', group_type: 'small_group', meeting_day: '', meeting_time: '', meeting_location: '', is_public: true, is_active: true, max_members: null, photo_url: '' };
		leaderSearch = '';
		leaderResults = [];
		selectedLeader = null;
	}

	function getGroupTypeLabel(type) {
		return groupTypes.find(t => t.value === type)?.label || type;
	}

	function getGroupTypeIcon(type) {
		const icons = { small_group: '👥', ministry_team: '⚡', class: '📖', committee: '📋', bible_study: '📚', team: '🏆' };
		return icons[type] || '👥';
	}

	function getMeetingDayLabel(day) {
		return day ? day.charAt(0).toUpperCase() + day.slice(1) : '';
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-[var(--text-primary)]">Groups</h1>
		<button on:click={() => (showCreateModal = true)} class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90">
			Create Group
		</button>
	</div>

	<!-- Filters -->
	<div class="bg-surface p-4 rounded-lg shadow border border-custom flex gap-4">
		<div class="flex-1">
			<label class="block text-sm font-medium text-secondary mb-1">Type</label>
			<select bind:value={filterType} on:change={handleFilter} class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary">
				<option value="">All Types</option>
				{#each groupTypes as type}
					<option value={type.value}>{type.label}</option>
				{/each}
			</select>
		</div>
		<div class="flex-1">
			<label class="block text-sm font-medium text-secondary mb-1">Status</label>
			<select bind:value={filterActive} on:change={handleFilter} class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary">
				<option value="">All Statuses</option>
				<option value="true">Active</option>
				<option value="false">Inactive</option>
			</select>
		</div>
	</div>

	{#if loading}
		<div class="flex justify-center items-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else if groups.length === 0 && !filterType && !filterActive}
		<!-- Empty State -->
		<div class="bg-surface rounded-lg shadow border border-custom p-12 text-center">
			<div class="text-6xl mb-4">👥</div>
			<h2 class="text-2xl font-bold text-primary mb-2">No groups yet</h2>
			<p class="text-secondary mb-6 max-w-md mx-auto">
				Create your first group to get started. Organize small groups, ministry teams, classes, and committees.
			</p>
			<button on:click={() => (showCreateModal = true)} class="px-6 py-3 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 transition font-medium">
				Create Your First Group
			</button>
		</div>
	{:else if groups.length === 0}
		<div class="text-center py-8 text-secondary">No groups match your filters</div>
	{:else}
		<!-- Groups Grid -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each groups as group}
				<div
					class="bg-surface border border-custom rounded-lg shadow hover:shadow-md transition-shadow cursor-pointer overflow-hidden"
					on:click={() => viewGroup(group.id)}
					on:keypress={() => viewGroup(group.id)}
					role="button"
					tabindex="0"
				>
					{#if group.photo_url}
						<img src={group.photo_url} alt={group.name} class="w-full h-40 object-cover" />
					{:else}
						<div class="w-full h-40 bg-gradient-to-br from-[var(--teal)] to-[var(--sage)] flex items-center justify-center">
							<span class="text-5xl">{getGroupTypeIcon(group.group_type)}</span>
						</div>
					{/if}

					<div class="p-5">
						<div class="flex justify-between items-start mb-2">
							<h3 class="text-lg font-bold text-primary leading-tight">{group.name}</h3>
							<span class={`px-2 py-0.5 text-xs rounded-full shrink-0 ml-2 ${group.is_active ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100' : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300'}`}>
								{group.is_active ? 'Active' : 'Inactive'}
							</span>
						</div>

						<p class="text-sm text-[var(--teal)] font-medium mb-2">{getGroupTypeLabel(group.group_type)}</p>

						{#if group.description}
							<p class="text-sm text-secondary mb-3 line-clamp-2">{group.description}</p>
						{/if}

						<div class="space-y-1.5 text-sm text-secondary">
							{#if group.leader_name}
								<div class="flex items-center gap-2">
									<span>👤</span>
									<span class="font-medium">{group.leader_name}</span>
								</div>
							{/if}
							{#if group.meeting_day && group.meeting_time}
								<div class="flex items-center gap-2">
									<span>🕐</span>
									<span>{getMeetingDayLabel(group.meeting_day)}s at {group.meeting_time}</span>
								</div>
							{/if}
							{#if group.meeting_location}
								<div class="flex items-center gap-2">
									<span>📍</span>
									<span>{group.meeting_location}</span>
								</div>
							{/if}
							<div class="flex items-center justify-between pt-2 border-t border-custom mt-2">
								<span class="text-xs">{group.member_count || 0} member{group.member_count !== 1 ? 's' : ''}{#if group.max_members} / {group.max_members} max{/if}</span>
								{#if !group.is_public}
									<span class="px-2 py-0.5 text-xs rounded bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100">Private</span>
								{/if}
							</div>
						</div>
					</div>
				</div>
			{/each}
		</div>

		<div class="text-sm text-secondary">Showing {groups.length} of {total} groups</div>
	{/if}
</div>

<!-- Create Group Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<div class="bg-surface rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
			<div class="p-6">
				<h2 class="text-2xl font-bold text-primary mb-4">Create New Group</h2>
				<form on:submit|preventDefault={createGroup} class="space-y-4">
					<div class="grid grid-cols-2 gap-4">
						<div class="col-span-2">
							<label class="block text-sm font-medium text-secondary mb-1">Group Name <span class="text-red-500">*</span></label>
							<input type="text" bind:value={newGroup.name} required class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary" placeholder="Sunday Morning Bible Study" />
						</div>
						<div class="col-span-2">
							<label class="block text-sm font-medium text-secondary mb-1">Description</label>
							<textarea bind:value={newGroup.description} rows="3" class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary" placeholder="A brief description of your group..."></textarea>
						</div>
						<div>
							<label class="block text-sm font-medium text-secondary mb-1">Type</label>
							<select bind:value={newGroup.group_type} class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary">
								{#each groupTypes as type}<option value={type.value}>{type.label}</option>{/each}
							</select>
						</div>
						<div>
							<label class="block text-sm font-medium text-secondary mb-1">Max Members</label>
							<input type="number" bind:value={newGroup.max_members} min="1" class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary" placeholder="Optional" />
						</div>
						<div class="col-span-2 relative">
							<label class="block text-sm font-medium text-secondary mb-1">Leader</label>
							<input type="text" bind:value={leaderSearch} on:input={handleLeaderSearch} class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary" placeholder="Search for a leader..." />
							{#if selectedLeader}
								<div class="mt-1 text-sm text-[var(--teal)]">✓ {selectedLeader.first_name} {selectedLeader.last_name}
									<button type="button" on:click={() => { selectedLeader = null; leaderSearch = ''; }} class="ml-2 text-red-500 text-xs">clear</button>
								</div>
							{/if}
							{#if leaderResults.length > 0 && !selectedLeader}
								<div class="absolute z-10 w-full mt-1 bg-surface border border-custom rounded-md shadow-lg max-h-40 overflow-y-auto">
									{#each leaderResults as person}
										<button type="button" class="w-full px-3 py-2 text-left hover:bg-[var(--surface-hover)] text-sm" on:click={() => selectLeader(person)}>
											{person.first_name} {person.last_name} {person.email ? `(${person.email})` : ''}
										</button>
									{/each}
								</div>
							{/if}
						</div>
						<div>
							<label class="block text-sm font-medium text-secondary mb-1">Meeting Day</label>
							<select bind:value={newGroup.meeting_day} class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary">
								<option value="">Select a day...</option>
								{#each meetingDays as day}<option value={day.value}>{day.label}</option>{/each}
							</select>
						</div>
						<div>
							<label class="block text-sm font-medium text-secondary mb-1">Meeting Time</label>
							<input type="text" bind:value={newGroup.meeting_time} class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary" placeholder="7:00 PM" />
						</div>
						<div class="col-span-2">
							<label class="block text-sm font-medium text-secondary mb-1">Meeting Location</label>
							<input type="text" bind:value={newGroup.meeting_location} class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary" placeholder="Room 101, Main Building" />
						</div>
						<div class="col-span-2">
							<label class="flex items-center gap-2">
								<input type="checkbox" bind:checked={newGroup.is_public} class="rounded" />
								<span class="text-sm text-secondary">Public (visible in group finder)</span>
							</label>
						</div>
						<div class="col-span-2">
							<label class="flex items-center gap-2">
								<input type="checkbox" bind:checked={newGroup.is_active} class="rounded" />
								<span class="text-sm text-secondary">Active</span>
							</label>
						</div>
					</div>
					<div class="flex justify-end gap-3 pt-4 border-t border-custom">
						<button type="button" on:click={() => { showCreateModal = false; resetNewGroup(); }} class="px-4 py-2 border border-custom rounded-md text-secondary hover:bg-[var(--surface-hover)]">Cancel</button>
						<button type="submit" class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90">Create Group</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}
