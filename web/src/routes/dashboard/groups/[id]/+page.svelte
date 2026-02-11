<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	let groupId;
	let group = null;
	let members = [];
	let people = [];
	let loading = false;
	let showEditModal = false;
	let showAddMemberModal = false;
	let searchPeople = '';

	let editGroup = {
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

	let newMember = {
		person_id: '',
		role: 'member'
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

	const roles = [
		{ value: 'leader', label: 'Leader' },
		{ value: 'co_leader', label: 'Co-Leader' },
		{ value: 'member', label: 'Member' },
		{ value: 'pending', label: 'Pending' }
	];

	$: groupId = $page.params.id;

	onMount(() => {
		loadGroup();
		loadPeople();
	});

	async function loadGroup() {
		loading = true;
		try {
			group = await api(`/api/groups/${groupId}`);
			members = group.members || [];

			// Populate edit form
			editGroup = {
				name: group.name,
				description: group.description || '',
				group_type: group.group_type,
				meeting_day: group.meeting_day || '',
				meeting_time: group.meeting_time || '',
				meeting_location: group.meeting_location || '',
				is_public: group.is_public,
				is_active: group.is_active,
				max_members: group.max_members,
				photo_url: group.photo_url || ''
			};
		} catch (error) {
			console.error('Failed to load group:', error);
			alert('Group not found');
			goto('/dashboard/groups');
		} finally {
			loading = false;
		}
	}

	async function loadPeople() {
		try {
			const response = await api('/api/people?limit=1000');
			people = response.people || [];
		} catch (error) {
			console.error('Failed to load people:', error);
		}
	}

	async function updateGroup() {
		try {
			// Convert max_members to null if empty
			const payload = { ...editGroup };
			if (!payload.max_members) {
				payload.max_members = null;
			}

			await api(`/api/groups/${groupId}`, {
				method: 'PUT',
				body: JSON.stringify(payload)
			});
			showEditModal = false;
			loadGroup();
		} catch (error) {
			alert('Failed to update group: ' + error.message);
		}
	}

	async function deleteGroup() {
		if (!confirm('Are you sure you want to delete this group? This cannot be undone.')) {
			return;
		}

		try {
			await api(`/api/groups/${groupId}`, {
				method: 'DELETE'
			});
			goto('/dashboard/groups');
		} catch (error) {
			alert('Failed to delete group: ' + error.message);
		}
	}

	async function addMember() {
		try {
			await api(`/api/groups/${groupId}/members`, {
				method: 'POST',
				body: JSON.stringify(newMember)
			});
			showAddMemberModal = false;
			newMember = { person_id: '', role: 'member' };
			loadGroup();
		} catch (error) {
			alert('Failed to add member: ' + error.message);
		}
	}

	async function updateMemberRole(memberId, newRole) {
		try {
			await api(`/api/groups/${groupId}/members/${memberId}`, {
				method: 'PUT',
				body: JSON.stringify({ role: newRole })
			});
			loadGroup();
		} catch (error) {
			alert('Failed to update member role: ' + error.message);
		}
	}

	async function removeMember(memberId) {
		if (!confirm('Remove this member from the group?')) {
			return;
		}

		try {
			await api(`/api/groups/${groupId}/members/${memberId}`, {
				method: 'DELETE'
			});
			loadGroup();
		} catch (error) {
			alert('Failed to remove member: ' + error.message);
		}
	}

	function getGroupTypeLabel(type) {
		const found = groupTypes.find((t) => t.value === type);
		return found ? found.label : type;
	}

	function getMeetingDayLabel(day) {
		if (!day) return '';
		return day.charAt(0).toUpperCase() + day.slice(1);
	}

	function getRoleLabel(role) {
		const found = roles.find((r) => r.value === role);
		return found ? found.label : role;
	}

	function getRoleBadgeClass(role) {
		if (role === 'leader') return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100';
		if (role === 'co_leader') return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-100';
		if (role === 'pending') return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100';
		return 'bg-[var(--surface-hover)] text-primary dark:bg-gray-700 dark:text-gray-300';
	}

	$: filteredPeople = people.filter(
		(p) =>
			!members.some((m) => m.person_id === p.id) &&
			(searchPeople === '' ||
				`${p.first_name} ${p.last_name}`.toLowerCase().includes(searchPeople.toLowerCase()) ||
				(p.email && p.email.toLowerCase().includes(searchPeople.toLowerCase())))
	);
</script>

<div class="space-y-6">
	{#if loading}
		<div class="text-center py-8 text-secondary">Loading group...</div>
	{:else if group}
		<!-- Header -->
		<div class="flex justify-between items-start">
			<div>
				<button
					on:click={() => goto('/dashboard/groups')}
					class="text-sm text-[var(--teal)] hover:underline mb-2"
				>
					← Back to Groups
				</button>
				<h1 class="text-3xl font-bold text-primary">{group.name}</h1>
				<div class="flex gap-2 mt-2">
					<span class="text-sm text-[var(--teal)]">{getGroupTypeLabel(group.group_type)}</span>
					<span
						class={`px-2 py-1 text-xs rounded ${
							group.is_active
								? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100'
								: 'bg-[var(--surface-hover)] text-primary dark:bg-gray-700 dark:text-gray-300'
						}`}
					>
						{group.is_active ? 'Active' : 'Inactive'}
					</span>
					{#if !group.is_public}
						<span class="px-2 py-1 text-xs rounded bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100">
							Private
						</span>
					{/if}
				</div>
			</div>
			<div class="flex gap-2">
				<button
					on:click={() => (showEditModal = true)}
					class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90"
				>
					Edit Group
				</button>
				<button
					on:click={deleteGroup}
					class="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700"
				>
					Delete
				</button>
			</div>
		</div>

		<!-- Group Info -->
		<div class="bg-surface border border-custom rounded-lg shadow p-6">
			<h2 class="text-xl font-bold text-primary mb-4">Group Information</h2>
			<div class="grid grid-cols-2 gap-4 text-sm">
				{#if group.description}
					<div class="col-span-2">
						<span class="font-medium text-secondary">Description:</span>
						<p class="text-primary mt-1">{group.description}</p>
					</div>
				{/if}

				{#if group.meeting_day && group.meeting_time}
					<div>
						<span class="font-medium text-secondary">Meeting Time:</span>
						<p class="text-primary">
							{getMeetingDayLabel(group.meeting_day)}s at {group.meeting_time}
						</p>
					</div>
				{/if}

				{#if group.meeting_location}
					<div>
						<span class="font-medium text-secondary">Location:</span>
						<p class="text-primary">{group.meeting_location}</p>
					</div>
				{/if}

				{#if group.max_members}
					<div>
						<span class="font-medium text-secondary">Capacity:</span>
						<p class="text-primary">
							{members.length} / {group.max_members} members
						</p>
					</div>
				{:else}
					<div>
						<span class="font-medium text-secondary">Members:</span>
						<p class="text-primary">{members.length}</p>
					</div>
				{/if}
			</div>
		</div>

		<!-- Members -->
		<div class="bg-surface border border-custom rounded-lg shadow">
			<div class="p-6 border-b border-custom flex justify-between items-center">
				<h2 class="text-xl font-bold text-primary">Members ({members.length})</h2>
				<button
					on:click={() => (showAddMemberModal = true)}
					class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90"
				>
					Add Member
				</button>
			</div>

			{#if members.length === 0}
				<div class="p-6 text-center text-secondary">No members yet</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full">
						<thead class="bg-[var(--bg)] border-b border-custom">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
									Name
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
									Email
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
									Phone
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
									Role
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
									Joined
								</th>
								<th class="px-6 py-3 text-right text-xs font-medium text-secondary uppercase tracking-wider">
									Actions
								</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-custom">
							{#each members as member}
								<tr class="hover:bg-[var(--bg)]">
									<td class="px-6 py-4 whitespace-nowrap">
										<div class="text-sm font-medium text-primary">
											{member.person?.first_name} {member.person?.last_name}
										</div>
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<div class="text-sm text-secondary">{member.person?.email || '-'}</div>
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<div class="text-sm text-secondary">{member.person?.phone || '-'}</div>
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<select
											value={member.role}
											on:change={(e) => updateMemberRole(member.id, e.target.value)}
											class={`px-2 py-1 text-xs rounded border-0 ${getRoleBadgeClass(member.role)}`}
										>
											{#each roles as role}
												<option value={role.value}>{role.label}</option>
											{/each}
										</select>
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<div class="text-sm text-secondary">
											{new Date(member.joined_at).toLocaleDateString()}
										</div>
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-right text-sm">
										<button
											on:click={() => removeMember(member.id)}
											class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
										>
											Remove
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Edit Group Modal -->
{#if showEditModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<div class="bg-surface rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
			<div class="p-6">
				<h2 class="text-2xl font-bold text-primary mb-4">Edit Group</h2>

				<form on:submit|preventDefault={updateGroup} class="space-y-4">
					<div class="grid grid-cols-2 gap-4">
						<div class="col-span-2">
							<label class="block text-sm font-medium text-secondary mb-1">
								Group Name <span class="text-red-500">*</span>
							</label>
							<input
								type="text"
								bind:value={editGroup.name}
								required
								class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
							/>
						</div>

						<div class="col-span-2">
							<label class="block text-sm font-medium text-secondary mb-1">Description</label>
							<textarea
								bind:value={editGroup.description}
								rows="3"
								class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
							/>
						</div>

						<div>
							<label class="block text-sm font-medium text-secondary mb-1">Type</label>
							<select
								bind:value={editGroup.group_type}
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
								bind:value={editGroup.max_members}
								min="1"
								class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
							/>
						</div>

						<div>
							<label class="block text-sm font-medium text-secondary mb-1">Meeting Day</label>
							<select
								bind:value={editGroup.meeting_day}
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
								bind:value={editGroup.meeting_time}
								class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
							/>
						</div>

						<div class="col-span-2">
							<label class="block text-sm font-medium text-secondary mb-1">Meeting Location</label>
							<input
								type="text"
								bind:value={editGroup.meeting_location}
								class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
							/>
						</div>

						<div class="col-span-2">
							<label class="flex items-center gap-2">
								<input type="checkbox" bind:checked={editGroup.is_public} class="rounded" />
								<span class="text-sm text-secondary">Public (visible in group finder)</span>
							</label>
						</div>

						<div class="col-span-2">
							<label class="flex items-center gap-2">
								<input type="checkbox" bind:checked={editGroup.is_active} class="rounded" />
								<span class="text-sm text-secondary">Active</span>
							</label>
						</div>
					</div>

					<div class="flex justify-end gap-3 pt-4 border-t border-custom">
						<button
							type="button"
							on:click={() => (showEditModal = false)}
							class="px-4 py-2 border border-custom rounded-md text-secondary hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800"
						>
							Cancel
						</button>
						<button
							type="submit"
							class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90"
						>
							Save Changes
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}

<!-- Add Member Modal -->
{#if showAddMemberModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<div class="bg-surface rounded-lg shadow-xl max-w-lg w-full mx-4">
			<div class="p-6">
				<h2 class="text-2xl font-bold text-primary mb-4">Add Member</h2>

				<form on:submit|preventDefault={addMember} class="space-y-4">
					<div>
						<label class="block text-sm font-medium text-secondary mb-1">
							Search Person <span class="text-red-500">*</span>
						</label>
						<input
							type="text"
							bind:value={searchPeople}
							class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary mb-2"
							placeholder="Search by name or email..."
						/>

						<select
							bind:value={newMember.person_id}
							required
							size="6"
							class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
						>
							<option value="">Select a person...</option>
							{#each filteredPeople as person}
								<option value={person.id}>
									{person.first_name} {person.last_name}
									{person.email ? `(${person.email})` : ''}
								</option>
							{/each}
						</select>
					</div>

					<div>
						<label class="block text-sm font-medium text-secondary mb-1">Role</label>
						<select
							bind:value={newMember.role}
							class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
						>
							{#each roles as role}
								<option value={role.value}>{role.label}</option>
							{/each}
						</select>
					</div>

					<div class="flex justify-end gap-3 pt-4 border-t border-custom">
						<button
							type="button"
							on:click={() => {
								showAddMemberModal = false;
								newMember = { person_id: '', role: 'member' };
								searchPeople = '';
							}}
							class="px-4 py-2 border border-custom rounded-md text-secondary hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800"
						>
							Cancel
						</button>
						<button
							type="submit"
							class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90"
						>
							Add Member
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}
