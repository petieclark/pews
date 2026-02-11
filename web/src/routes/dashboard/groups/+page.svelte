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

	const groupTypes = [
		{ value: 'small_group', label: 'Small Group' },
		{ value: 'bible_study', label: 'Bible Study' },
		{ value: 'team', label: 'Team' },
		{ value: 'ministry_team', label: 'Ministry Team' },
		{ value: 'class', label: 'Class' },
		{ value: 'committee', label: 'Committee' }
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
			const params = new URLSearchParams({
				page: page.toString(),
				limit: limit.toString()
			});
			if (filterType) {
				params.append('type', filterType);
			}
			if (filterActive) {
				params.append('active', filterActive);
			}
			const response = await api(`/api/groups?${params}`);
			groups = response.groups || [];
			total = response.total || 0;
		} catch (error) {
			console.error('Failed to load groups:', error);
		} finally {
			loading = false;
		}
	}

	function handleFilter() {
		page = 1;
		loadGroups();
	}

	function viewGroup(id) {
		goto(`/dashboard/groups/${id}`);
	}

	async function createGroup() {
		try {
			// Convert max_members to null if empty
			const payload = { ...newGroup };
			if (!payload.max_members) {
				payload.max_members = null;
			}

			await api('/api/groups', {
				method: 'POST',
				body: JSON.stringify(payload)
			});
			showCreateModal = false;
			resetNewGroup();
			loadGroups();
		} catch (error) {
			alert('Failed to create group: ' + error.message);
		}
	}

	function resetNewGroup() {
		newGroup = {
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
	}

	function getGroupTypeLabel(type) {
		const found = groupTypes.find((t) => t.value === type);
		return found ? found.label : type;
	}

	function getMeetingDayLabel(day) {
		if (!day) return '';
		return day.charAt(0).toUpperCase() + day.slice(1);
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-[var(--text-primary)]">Groups</h1>
		<button
			on:click={() => (showCreateModal = true)}
			class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90"
		>
			Create Group
		</button>
	</div>

	<!-- Filters -->
	<div class="bg-surface p-4 rounded-lg shadow border border-custom flex gap-4">
		<div class="flex-1">
			<label class="block text-sm font-medium text-secondary mb-1">Type</label>
			<select
				bind:value={filterType}
				on:change={handleFilter}
				class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
			>
				<option value="">All Types</option>
				{#each groupTypes as type}
					<option value={type.value}>{type.label}</option>
				{/each}
			</select>
		</div>
		<div class="flex-1">
			<label class="block text-sm font-medium text-secondary mb-1">Status</label>
			<select
				bind:value={filterActive}
				on:change={handleFilter}
				class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
			>
				<option value="">All Statuses</option>
				<option value="true">Active</option>
				<option value="false">Inactive</option>
			</select>
		</div>
	</div>

	{#if loading}
		<div class="text-center py-8 text-secondary">Loading groups...</div>
	{:else if groups.length === 0}
		<div class="text-center py-8 text-secondary">No groups found</div>
	{:else}
		<!-- Groups Grid -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each groups as group}
				<div
					class="bg-surface border border-custom rounded-lg shadow hover:shadow-md transition-shadow cursor-pointer"
					on:click={() => viewGroup(group.id)}
				>
					{#if group.photo_url}
						<img
							src={group.photo_url}
							alt={group.name}
							class="w-full h-48 object-cover rounded-t-lg"
						/>
					{:else}
						<div class="w-full h-48 bg-[#8FBCB0] rounded-t-lg flex items-center justify-center">
							<span class="text-6xl text-white">
								{group.name.charAt(0).toUpperCase()}
							</span>
						</div>
					{/if}

					<div class="p-4">
						<div class="flex justify-between items-start mb-2">
							<h3 class="text-xl font-bold text-primary">{group.name}</h3>
							<span
								class={`px-2 py-1 text-xs rounded ${
									group.is_active
										? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100'
										: 'bg-[var(--surface-hover)] text-primary dark:bg-gray-700 dark:text-gray-300'
								}`}
							>
								{group.is_active ? 'Active' : 'Inactive'}
							</span>
						</div>

						<p class="text-sm text-[var(--teal)] mb-2">{getGroupTypeLabel(group.group_type)}</p>

						{#if group.description}
							<p class="text-sm text-secondary mb-3 line-clamp-2">{group.description}</p>
						{/if}

						<div class="space-y-1 text-sm text-secondary">
							{#if group.meeting_day && group.meeting_time}
								<div class="flex items-center gap-2">
									<span class="font-medium">
										{getMeetingDayLabel(group.meeting_day)}s at {group.meeting_time}
									</span>
								</div>
							{/if}

							{#if group.meeting_location}
								<div class="flex items-center gap-2">
									<span>📍 {group.meeting_location}</span>
								</div>
							{/if}

							<div class="flex items-center gap-2 text-xs">
								<span>{group.member_count || 0} member{group.member_count !== 1 ? 's' : ''}</span>
								{#if group.max_members}
									<span>/ {group.max_members} max</span>
								{/if}
								{#if !group.is_public}
									<span class="ml-auto bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100 px-2 py-0.5 rounded">
										Private
									</span>
								{/if}
							</div>
						</div>
					</div>
				</div>
			{/each}
		</div>

		<!-- Pagination info -->
		<div class="text-sm text-secondary">
			Showing {groups.length} of {total} groups
		</div>
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
							<label class="block text-sm font-medium text-secondary mb-1">
								Group Name <span class="text-red-500">*</span>
							</label>
							<input
								type="text"
								bind:value={newGroup.name}
								required
								class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
								placeholder="Sunday Morning Bible Study"
							/>
						</div>

						<div class="col-span-2">
							<label class="block text-sm font-medium text-secondary mb-1">Description</label>
							<textarea
								bind:value={newGroup.description}
								rows="3"
								class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
								placeholder="A brief description of your group..."
							></textarea>
						</div>

						<div>
							<label class="block text-sm font-medium text-secondary mb-1">Type</label>
							<select
								bind:value={newGroup.group_type}
								class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
							>
								{#each groupTypes as type}
									<option value={type.value}>{type.label}</option>
								{/each}
							</select>
						</div>

						<div>
							<label class="block text-sm font-medium text-secondary mb-1">Max Members</label>
							<input
								type="number"
								bind:value={newGroup.max_members}
								min="1"
								class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
								placeholder="Optional"
							/>
						</div>

						<div>
							<label class="block text-sm font-medium text-secondary mb-1">Meeting Day</label>
							<select
								bind:value={newGroup.meeting_day}
								class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
							>
								<option value="">Select a day...</option>
								{#each meetingDays as day}
									<option value={day.value}>{day.label}</option>
								{/each}
							</select>
						</div>

						<div>
							<label class="block text-sm font-medium text-secondary mb-1">Meeting Time</label>
							<input
								type="text"
								bind:value={newGroup.meeting_time}
								class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
								placeholder="7:00 PM"
							/>
						</div>

						<div class="col-span-2">
							<label class="block text-sm font-medium text-secondary mb-1">Meeting Location</label>
							<input
								type="text"
								bind:value={newGroup.meeting_location}
								class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
								placeholder="Room 101, Main Building"
							/>
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
						<button
							type="button"
							on:click={() => {
								showCreateModal = false;
								resetNewGroup();
							}}
							class="px-4 py-2 border border-custom rounded-md text-secondary hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800"
						>
							Cancel
						</button>
						<button
							type="submit"
							class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90"
						>
							Create Group
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}
