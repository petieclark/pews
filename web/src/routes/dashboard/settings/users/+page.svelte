<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import Modal from '$lib/Modal.svelte';

	let users = [];
	let loading = true;
	let error = '';
	let showInvite = false;
	let inviteEmail = '';
	let inviteRole = 'Staff';
	let inviting = false;
	let inviteSuccess = '';

	const roles = ['Admin', 'Staff', 'Volunteer', 'Member'];

	onMount(async () => {
		await loadUsers();
	});

	async function loadUsers() {
		loading = true;
		try {
			users = await api('/api/tenant/users');
		} catch (err) {
			// If endpoint doesn't exist yet, show empty state
			users = [];
		} finally {
			loading = false;
		}
	}

	async function inviteUser() {
		inviting = true;
		inviteSuccess = '';
		error = '';
		try {
			await api('/api/tenant/users/invite', {
				method: 'POST',
				body: JSON.stringify({ email: inviteEmail, role: inviteRole })
			});
			inviteSuccess = `Invitation sent to ${inviteEmail}`;
			inviteEmail = '';
			inviteRole = 'Staff';
			showInvite = false;
			await loadUsers();
		} catch (err) {
			error = err.message;
		} finally {
			inviting = false;
		}
	}

	async function updateRole(userId, newRole) {
		try {
			await api(`/api/tenant/users/${userId}/role`, {
				method: 'PUT',
				body: JSON.stringify({ role: newRole })
			});
			users = users.map(u => u.id === userId ? { ...u, role: newRole } : u);
		} catch (err) {
			error = err.message;
		}
	}

	async function removeUser(userId) {
		if (!confirm('Are you sure you want to remove this user?')) return;
		try {
			await api(`/api/tenant/users/${userId}`, { method: 'DELETE' });
			users = users.filter(u => u.id !== userId);
		} catch (err) {
			error = err.message;
		}
	}

	function getRoleBadge(role) {
		const colors = {
			Admin: 'bg-red-900/30 text-red-300 border-red-700',
			Staff: 'bg-blue-900/30 text-blue-300 border-blue-700',
			Volunteer: 'bg-emerald-900/30 text-emerald-300 border-emerald-700',
			Member: 'bg-gray-700/30 text-gray-300 border-gray-600'
		};
		return colors[role] || colors.Member;
	}
</script>

<div class="max-w-3xl">
	<a href="/dashboard/settings" class="text-sm text-secondary hover:text-[var(--teal)] mb-4 inline-block">← Back to Settings</a>
	
	<div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3 mb-6">
		<h1 class="text-2xl sm:text-3xl font-bold text-primary">Users & Roles</h1>
		<button on:click={() => showInvite = true}
			class="w-full sm:w-auto px-4 py-2 bg-[var(--teal)] text-white rounded-lg font-medium hover:opacity-90">
			Invite User
		</button>
	</div>

	{#if error}
		<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg mb-4">{error}</div>
	{/if}
	{#if inviteSuccess}
		<div class="bg-emerald-900/30 border border-emerald-700 text-emerald-300 px-4 py-3 rounded-lg mb-4">{inviteSuccess}</div>
	{/if}

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else if users.length === 0}
		<div class="bg-surface rounded-lg border border-custom p-8 text-center">
			<p class="text-secondary mb-4">No team members yet. Invite your first user to get started.</p>
			<button on:click={() => showInvite = true}
				class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg font-medium hover:opacity-90">
				Invite User
			</button>
		</div>
	{:else}
		<div class="space-y-3">
			{#each users as user}
				<div class="bg-surface rounded-lg border border-custom p-4 flex flex-col sm:flex-row sm:items-center gap-3">
					<div class="flex-1 min-w-0">
						<div class="font-medium text-primary">{user.name || user.email}</div>
						<div class="text-sm text-secondary truncate">{user.email}</div>
					</div>
					<div class="flex items-center gap-3">
						<select
							value={user.role}
							on:change={(e) => updateRole(user.id, e.target.value)}
							class="px-3 py-1 text-sm border input-border rounded-lg bg-[var(--input-bg)] text-primary focus:ring-2 focus:ring-[var(--teal)]"
						>
							{#each roles as role}
								<option value={role}>{role}</option>
							{/each}
						</select>
						<span class="px-2 py-1 text-xs rounded-full border {getRoleBadge(user.role)}">{user.role}</span>
						<button on:click={() => removeUser(user.id)}
							class="text-secondary hover:text-red-500 p-1" aria-label="Remove user">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
							</svg>
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<Modal show={showInvite} title="Invite User" onClose={() => showInvite = false}>
	<form on:submit|preventDefault={inviteUser} class="space-y-4">
		<div>
			<label for="inv-email" class="block text-sm font-medium text-primary mb-1">Email Address</label>
			<input id="inv-email" type="email" bind:value={inviteEmail} required placeholder="user@example.com"
				class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary" />
		</div>
		<div>
			<label for="inv-role" class="block text-sm font-medium text-primary mb-1">Role</label>
			<select id="inv-role" bind:value={inviteRole}
				class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary">
				{#each roles as role}
					<option value={role}>{role}</option>
				{/each}
			</select>
		</div>
		<div class="flex gap-2 pt-2">
			<button type="button" on:click={() => showInvite = false}
				class="flex-1 px-4 py-2 border border-custom rounded-lg hover:bg-[var(--surface-hover)] text-primary">Cancel</button>
			<button type="submit" disabled={inviting}
				class="flex-1 px-4 py-2 bg-[var(--teal)] text-white rounded-lg font-medium hover:opacity-90 disabled:opacity-50">
				{inviting ? 'Sending...' : 'Send Invite'}
			</button>
		</div>
	</form>
</Modal>
