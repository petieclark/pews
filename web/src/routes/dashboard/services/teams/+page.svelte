<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast';

	let teams = [];
	let loading = true;
	let showCreateModal = false;
	let showDetailView = false;
	let selectedTeam = null;
	let editingTeam = null;
	let people = [];
	let searchQuery = '';
	let showAddMember = false;
	let newPositionName = '';

	let newTeam = { name: '', description: '', color: '#4A8B8C' };

	const presetColors = ['#4A8B8C', '#1B3A4B', '#8FBCB0', '#E74C3C', '#F39C12', '#9B59B6', '#2ECC71', '#3498DB'];

	onMount(() => {
		loadTeams();
		loadPeople();
	});

	async function loadTeams() {
		loading = true;
		try {
			const res = await api('/api/teams');
			teams = res.teams || [];
		} catch (e) {
			console.error('Failed to load teams:', e);
		} finally {
			loading = false;
		}
	}

	async function loadPeople() {
		try {
			const res = await api('/api/people?limit=500');
			people = res.people || [];
		} catch (e) {
			console.error('Failed to load people:', e);
		}
	}

	async function createTeam() {
		try {
			await api('/api/teams', { method: 'POST', body: JSON.stringify(newTeam) });
			toast.show('Team created!', 'success');
			showCreateModal = false;
			newTeam = { name: '', description: '', color: '#4A8B8C' };
			loadTeams();
		} catch (e) {
			toast.show('Failed to create team', 'error');
		}
	}

	async function openTeam(team) {
		try {
			selectedTeam = await api(`/api/teams/${team.id}`);
			showDetailView = true;
		} catch (e) {
			toast.show('Failed to load team', 'error');
		}
	}

	async function saveTeam() {
		try {
			await api(`/api/teams/${editingTeam.id}`, {
				method: 'PUT',
				body: JSON.stringify(editingTeam)
			});
			toast.show('Team updated!', 'success');
			editingTeam = null;
			loadTeams();
			if (selectedTeam && selectedTeam.id === editingTeam?.id) {
				openTeam({ id: selectedTeam.id });
			}
		} catch (e) {
			toast.show('Failed to update team', 'error');
		}
	}

	async function deleteTeam(team) {
		if (!confirm(`Delete "${team.name}"? This cannot be undone.`)) return;
		try {
			await api(`/api/teams/${team.id}`, { method: 'DELETE' });
			toast.show('Team deleted', 'success');
			if (selectedTeam?.id === team.id) {
				showDetailView = false;
				selectedTeam = null;
			}
			loadTeams();
		} catch (e) {
			toast.show('Failed to delete team', 'error');
		}
	}

	async function addPosition() {
		if (!newPositionName.trim()) return;
		try {
			await api(`/api/teams/${selectedTeam.id}/positions`, {
				method: 'POST',
				body: JSON.stringify({ name: newPositionName, sort_order: (selectedTeam.positions?.length || 0) })
			});
			newPositionName = '';
			openTeam({ id: selectedTeam.id });
		} catch (e) {
			toast.show('Failed to add position', 'error');
		}
	}

	async function deletePosition(posId) {
		try {
			await api(`/api/teams/${selectedTeam.id}/positions/${posId}`, { method: 'DELETE' });
			openTeam({ id: selectedTeam.id });
		} catch (e) {
			toast.show('Failed to delete position', 'error');
		}
	}

	async function addMember(personId, positionId = null) {
		try {
			const body = { person_id: personId };
			if (positionId) body.position_id = positionId;
			await api(`/api/teams/${selectedTeam.id}/members`, {
				method: 'POST',
				body: JSON.stringify(body)
			});
			showAddMember = false;
			searchQuery = '';
			openTeam({ id: selectedTeam.id });
		} catch (e) {
			toast.show('Failed to add member', 'error');
		}
	}

	async function updateMemberStatus(memberId, status) {
		try {
			await api(`/api/teams/${selectedTeam.id}/members/${memberId}/status`, {
				method: 'PATCH',
				body: JSON.stringify({ status })
			});
			openTeam({ id: selectedTeam.id });
		} catch (e) {
			toast.show('Failed to update member status', 'error');
		}
	}

	async function updateMemberPosition(memberId, positionId) {
		try {
			await api(`/api/teams/${selectedTeam.id}/members/${memberId}`, {
				method: 'PUT',
				body: JSON.stringify({ position_id: positionId || null })
			});
			openTeam({ id: selectedTeam.id });
		} catch (e) {
			toast.show('Failed to update member', 'error');
		}
	}

	async function removeMember(memberId) {
		try {
			await api(`/api/teams/${selectedTeam.id}/members/${memberId}`, { method: 'DELETE' });
			openTeam({ id: selectedTeam.id });
		} catch (e) {
			toast.show('Failed to remove member', 'error');
		}
	}

	function getInitials(first, last) {
		return ((first?.[0] || '') + (last?.[0] || '')).toUpperCase();
	}

	$: filteredPeople = searchQuery.length >= 2
		? people.filter(p => {
			const name = `${p.first_name} ${p.last_name}`.toLowerCase();
			const alreadyMember = selectedTeam?.members?.some(m => m.person_id === p.id);
			return name.includes(searchQuery.toLowerCase()) && !alreadyMember;
		}).slice(0, 10)
		: [];
</script>

<div>
	{#if showDetailView && selectedTeam}
		<!-- Team Detail View -->
		<div class="mb-6">
			<button on:click={() => { showDetailView = false; selectedTeam = null; loadTeams(); }}
				class="text-sm text-secondary hover:text-primary flex items-center gap-1">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
				Back to Teams
			</button>
		</div>

		<div class="flex items-center gap-4 mb-8">
			<div class="w-12 h-12 rounded-xl flex items-center justify-center text-white font-bold text-lg"
				style="background-color: {selectedTeam.color}">
				{selectedTeam.name[0]}
			</div>
			<div class="flex-1">
				<h1 class="text-2xl font-bold text-[var(--text-primary)]">{selectedTeam.name}</h1>
				{#if selectedTeam.description}
					<p class="text-secondary text-sm mt-1">{selectedTeam.description}</p>
				{/if}
			</div>
			<span class="text-sm text-secondary">{selectedTeam.members?.length || 0} members</span>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Positions -->
			<div class="bg-surface rounded-xl border border-custom p-6">
				<h2 class="text-lg font-semibold text-[var(--text-primary)] mb-4">Positions</h2>
				<div class="space-y-2 mb-4">
					{#each selectedTeam.positions || [] as pos}
						<div class="flex items-center justify-between py-2 px-3 bg-[var(--bg)] rounded-lg">
							<span class="text-sm text-[var(--text-primary)]">{pos.name}</span>
							<button on:click={() => deletePosition(pos.id)} class="text-red-400 hover:text-red-300 text-xs">Remove</button>
						</div>
					{/each}
					{#if !selectedTeam.positions?.length}
						<p class="text-secondary text-sm">No positions yet</p>
					{/if}
				</div>
				<div class="flex gap-2">
					<input bind:value={newPositionName} placeholder="New position..."
						on:keydown={e => e.key === 'Enter' && addPosition()}
						class="flex-1 px-3 py-2 bg-[var(--bg)] border border-custom rounded-lg text-sm text-[var(--text-primary)]" />
					<button on:click={addPosition}
						class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg text-sm hover:opacity-90">Add</button>
				</div>
			</div>

			<!-- Members -->
			<div class="bg-surface rounded-xl border border-custom p-6">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold text-[var(--text-primary)]">Members</h2>
					<button on:click={() => showAddMember = !showAddMember}
						class="px-3 py-1.5 bg-[var(--teal)] text-white rounded-lg text-sm hover:opacity-90">
						+ Add Member
					</button>
				</div>

				{#if showAddMember}
					<div class="mb-4 relative">
						<input bind:value={searchQuery} placeholder="Search people..."
							class="w-full px-3 py-2 bg-[var(--bg)] border border-custom rounded-lg text-sm text-[var(--text-primary)]" />
						{#if filteredPeople.length > 0}
							<div class="absolute z-10 mt-1 w-full bg-surface border border-custom rounded-lg shadow-lg max-h-48 overflow-y-auto">
								{#each filteredPeople as person}
									<button on:click={() => addMember(person.id)}
										class="w-full text-left px-3 py-2 text-sm hover:bg-[var(--surface-hover)] text-[var(--text-primary)]">
										{person.first_name} {person.last_name}
										{#if person.email}<span class="text-secondary ml-2">{person.email}</span>{/if}
									</button>
								{/each}
							</div>
						{/if}
					</div>
				{/if}

				<div class="space-y-2">
					{#each selectedTeam.members || [] as member}
						<div class="flex items-center gap-3 py-2 px-3 bg-[var(--bg)] rounded-lg">
							<div class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-semibold"
								style="background-color: {selectedTeam.color}">
								{getInitials(member.first_name, member.last_name)}
							</div>
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2">
									<span class="text-sm font-medium text-[var(--text-primary)]">{member.first_name} {member.last_name}</span>
									<span class="text-[10px] px-1.5 py-0.5 rounded-full font-medium {member.status === 'active' ? 'bg-green-500/20 text-green-400' : member.status === 'on-break' ? 'bg-yellow-500/20 text-yellow-400' : 'bg-gray-500/20 text-gray-400'}">{member.status}</span>
								</div>
								<div class="flex items-center gap-2">
									<select value={member.position_id || ''}
										on:change={e => updateMemberPosition(member.id, e.target.value || null)}
										class="text-xs bg-transparent text-secondary border-none p-0 cursor-pointer">
										<option value="">No position</option>
										{#each selectedTeam.positions || [] as pos}
											<option value={pos.id}>{pos.name}</option>
										{/each}
									</select>
									{#if member.email}
										<span class="text-[10px] text-secondary">{member.email}</span>
									{/if}
								</div>
							</div>
							<div class="flex items-center gap-1">
								<select value={member.status}
									on:change={e => updateMemberStatus(member.id, e.target.value)}
									class="text-xs bg-[var(--bg)] text-secondary border border-custom rounded px-1 py-0.5">
									<option value="active">Active</option>
									<option value="inactive">Inactive</option>
									<option value="on-break">On Break</option>
								</select>
								<button on:click={() => removeMember(member.id)} class="text-red-400 hover:text-red-300 text-xs">Remove</button>
							</div>
						</div>
					{/each}
					{#if !selectedTeam.members?.length}
						<p class="text-secondary text-sm">No members yet</p>
					{/if}
				</div>
			</div>
		</div>

	{:else}
		<!-- Teams List -->
		<div class="flex items-center justify-between mb-8">
			<div>
				<h1 class="text-2xl font-bold text-[var(--text-primary)]">Volunteer Teams</h1>
				<p class="text-secondary text-sm mt-1">Manage your ministry teams and volunteers</p>
			</div>
			<button on:click={() => showCreateModal = true}
				class="px-4 py-2.5 bg-[var(--teal)] text-white rounded-xl text-sm font-medium hover:opacity-90 transition-opacity">
				+ New Team
			</button>
		</div>

		{#if loading}
			<div class="flex justify-center py-20">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
			</div>
		{:else if teams.length === 0}
			<div class="text-center py-20">
				<div class="text-6xl mb-4">👥</div>
				<h2 class="text-xl font-semibold text-[var(--text-primary)] mb-2">No teams yet</h2>
				<p class="text-secondary mb-6">Create your first volunteer team to get started</p>
				<button on:click={() => showCreateModal = true}
					class="px-6 py-2.5 bg-[var(--teal)] text-white rounded-xl text-sm font-medium hover:opacity-90">
					Create Team
				</button>
			</div>
		{:else}
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each teams as team}
					<div class="bg-surface rounded-xl border border-custom overflow-hidden hover:shadow-lg transition-shadow cursor-pointer"
						on:click={() => openTeam(team)} on:keydown={e => e.key === 'Enter' && openTeam(team)} role="button" tabindex="0">
						<div class="h-2" style="background-color: {team.color}"></div>
						<div class="p-5">
							<div class="flex items-center gap-3 mb-3">
								<div class="w-10 h-10 rounded-lg flex items-center justify-center text-white font-bold"
									style="background-color: {team.color}">
									{team.name[0]}
								</div>
								<div class="flex-1">
									<h3 class="font-semibold text-[var(--text-primary)]">{team.name}</h3>
									<span class="text-xs text-secondary">{team.member_count} member{team.member_count !== 1 ? 's' : ''} · {team.position_count || 0} position{team.position_count !== 1 ? 's' : ''}</span>
								</div>
								<div class="flex gap-1">
									<button on:click|stopPropagation={() => { editingTeam = { ...team, is_active: true }; }}
										class="p-1.5 text-secondary hover:text-primary rounded-lg hover:bg-[var(--surface-hover)]">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
									</button>
									<button on:click|stopPropagation={() => deleteTeam(team)}
										class="p-1.5 text-secondary hover:text-red-400 rounded-lg hover:bg-[var(--surface-hover)]">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
									</button>
								</div>
							</div>
							{#if team.description}
								<p class="text-sm text-secondary line-clamp-2">{team.description}</p>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	{/if}
</div>

<!-- Create Team Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" on:click|self={() => showCreateModal = false}>
		<div class="bg-surface rounded-xl border border-custom p-6 w-full max-w-md">
			<h2 class="text-lg font-semibold text-[var(--text-primary)] mb-4">Create Team</h2>
			<div class="space-y-4">
				<div>
					<label class="block text-sm text-secondary mb-1">Team Name</label>
					<input bind:value={newTeam.name} placeholder="e.g. Worship Team"
						class="w-full px-3 py-2 bg-[var(--bg)] border border-custom rounded-lg text-[var(--text-primary)]" />
				</div>
				<div>
					<label class="block text-sm text-secondary mb-1">Description</label>
					<textarea bind:value={newTeam.description} rows="2" placeholder="What does this team do?"
						class="w-full px-3 py-2 bg-[var(--bg)] border border-custom rounded-lg text-[var(--text-primary)]"></textarea>
				</div>
				<div>
					<label class="block text-sm text-secondary mb-1">Color</label>
					<div class="flex gap-2">
						{#each presetColors as color}
							<button on:click={() => newTeam.color = color}
								class="w-8 h-8 rounded-full border-2 transition-transform {newTeam.color === color ? 'border-white scale-110' : 'border-transparent'}"
								style="background-color: {color}"></button>
						{/each}
					</div>
				</div>
			</div>
			<div class="flex justify-end gap-3 mt-6">
				<button on:click={() => showCreateModal = false} class="px-4 py-2 text-sm text-secondary hover:text-primary">Cancel</button>
				<button on:click={createTeam} disabled={!newTeam.name}
					class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg text-sm hover:opacity-90 disabled:opacity-50">Create</button>
			</div>
		</div>
	</div>
{/if}

<!-- Edit Team Modal -->
{#if editingTeam}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" on:click|self={() => editingTeam = null}>
		<div class="bg-surface rounded-xl border border-custom p-6 w-full max-w-md">
			<h2 class="text-lg font-semibold text-[var(--text-primary)] mb-4">Edit Team</h2>
			<div class="space-y-4">
				<div>
					<label class="block text-sm text-secondary mb-1">Team Name</label>
					<input bind:value={editingTeam.name}
						class="w-full px-3 py-2 bg-[var(--bg)] border border-custom rounded-lg text-[var(--text-primary)]" />
				</div>
				<div>
					<label class="block text-sm text-secondary mb-1">Description</label>
					<textarea bind:value={editingTeam.description} rows="2"
						class="w-full px-3 py-2 bg-[var(--bg)] border border-custom rounded-lg text-[var(--text-primary)]"></textarea>
				</div>
				<div>
					<label class="block text-sm text-secondary mb-1">Color</label>
					<div class="flex gap-2">
						{#each presetColors as color}
							<button on:click={() => editingTeam.color = color}
								class="w-8 h-8 rounded-full border-2 transition-transform {editingTeam.color === color ? 'border-white scale-110' : 'border-transparent'}"
								style="background-color: {color}"></button>
						{/each}
					</div>
				</div>
			</div>
			<div class="flex justify-end gap-3 mt-6">
				<button on:click={() => editingTeam = null} class="px-4 py-2 text-sm text-secondary hover:text-primary">Cancel</button>
				<button on:click={saveTeam} disabled={!editingTeam.name}
					class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg text-sm hover:opacity-90 disabled:opacity-50">Save</button>
			</div>
		</div>
	</div>
{/if}
