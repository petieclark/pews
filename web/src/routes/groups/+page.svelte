<script>
	import { onMount } from 'svelte';

	let groups = [];
	let loading = true;
	let filterType = '';

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8190';

	const groupTypes = [
		{ value: 'small_group', label: 'Small Group' },
		{ value: 'ministry_team', label: 'Ministry Team' },
		{ value: 'class', label: 'Class' },
		{ value: 'committee', label: 'Committee' },
		{ value: 'bible_study', label: 'Bible Study' },
		{ value: 'team', label: 'Team' }
	];

	onMount(async () => {
		// Get tenant_id from URL or config
		const params = new URLSearchParams(window.location.search);
		const tenantId = params.get('tenant_id') || '';
		if (!tenantId) {
			loading = false;
			return;
		}
		try {
			const res = await fetch(`${API_URL}/api/groups/public?tenant_id=${tenantId}`);
			if (res.ok) groups = await res.json();
		} catch (e) { console.error(e); }
		loading = false;
	});

	function getGroupTypeLabel(type) {
		return groupTypes.find(t => t.value === type)?.label || type;
	}

	function getGroupTypeIcon(type) {
		const icons = { small_group: '👥', ministry_team: '⚡', class: '📖', committee: '📋', bible_study: '📚', team: '🏆' };
		return icons[type] || '👥';
	}

	$: filteredGroups = filterType ? groups.filter(g => g.group_type === filterType) : groups;
</script>

<svelte:head>
	<title>Find a Group</title>
</svelte:head>

<div class="min-h-screen bg-[var(--bg)]">
	<div class="max-w-6xl mx-auto px-6 py-12">
		<div class="text-center mb-10">
			<h1 class="text-4xl font-bold text-primary mb-3">Find a Group</h1>
			<p class="text-lg text-secondary max-w-2xl mx-auto">
				Join a group to connect with others, grow in your faith, and find community.
			</p>
		</div>

		{#if loading}
			<div class="flex justify-center py-12">
				<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--teal)]"></div>
			</div>
		{:else if groups.length === 0}
			<div class="text-center py-12 text-secondary">
				<p class="text-xl">No groups available right now.</p>
				<p class="mt-2">Check back soon!</p>
			</div>
		{:else}
			<!-- Filter -->
			<div class="flex justify-center mb-8">
				<div class="flex flex-wrap gap-2">
					<button
						class="px-4 py-2 rounded-full text-sm font-medium transition {!filterType ? 'bg-[var(--teal)] text-white' : 'bg-surface border border-custom text-secondary hover:bg-[var(--surface-hover)]'}"
						on:click={() => filterType = ''}
					>All</button>
					{#each groupTypes as type}
						{#if groups.some(g => g.group_type === type.value)}
							<button
								class="px-4 py-2 rounded-full text-sm font-medium transition {filterType === type.value ? 'bg-[var(--teal)] text-white' : 'bg-surface border border-custom text-secondary hover:bg-[var(--surface-hover)]'}"
								on:click={() => filterType = type.value}
							>{type.label}</button>
						{/if}
					{/each}
				</div>
			</div>

			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
				{#each filteredGroups as group}
					<div class="bg-surface border border-custom rounded-lg shadow overflow-hidden">
						<div class="w-full h-40 bg-gradient-to-br from-[var(--teal)] to-[var(--sage)] flex items-center justify-center">
							<span class="text-5xl">{getGroupTypeIcon(group.group_type)}</span>
						</div>
						<div class="p-5">
							<h3 class="text-lg font-bold text-primary mb-1">{group.name}</h3>
							<p class="text-sm text-[var(--teal)] font-medium mb-2">{getGroupTypeLabel(group.group_type)}</p>
							{#if group.description}
								<p class="text-sm text-secondary mb-3">{group.description}</p>
							{/if}
							<div class="space-y-1.5 text-sm text-secondary">
								{#if group.meeting_day && group.meeting_time}
									<div>🕐 {group.meeting_day.charAt(0).toUpperCase() + group.meeting_day.slice(1)}s at {group.meeting_time}</div>
								{/if}
								{#if group.meeting_location}
									<div>📍 {group.meeting_location}</div>
								{/if}
								<div>{group.member_count || 0} member{group.member_count !== 1 ? 's' : ''}{#if group.max_members} · {group.max_members - (group.member_count || 0)} spots left{/if}</div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
