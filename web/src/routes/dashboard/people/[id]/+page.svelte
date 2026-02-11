<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let person = null;
	let tags = [];
	let loading = true;
	let editing = false;
	let editForm = {};
	let availableTags = [];
	let showAddTag = false;
	let selectedTagId = '';

	$: personId = $page.params.id;

	onMount(() => {
		loadPerson();
		loadAvailableTags();
		loadEngagementScore();
	});

	async function loadPerson() {
		loading = true;
		try {
			person = await api(`/api/people/${personId}`);
			editForm = { ...person };
		} catch (error) {
			console.error('Failed to load person:', error);
			alert('Person not found');
			goto('/dashboard/people');
		} finally {
			loading = false;
		}
	}

	async function loadAvailableTags() {
		try {
			availableTags = await api('/api/tags');
		} catch (error) {
			console.error('Failed to load tags:', error);
		}
	}

	async function loadEngagementScore() {
		try {
			engagementScore = await api(`/api/engagement/scores/${personId}`);
		} catch (error) {
			console.log('Engagement score not available:', error);
			engagementScore = null;
		}
	}

	async function recalculateEngagement() {
		try {
			engagementScore = await api(`/api/engagement/scores/${personId}/calculate`, {
				method: 'POST'
			});
		} catch (error) {
			alert('Failed to calculate engagement score: ' + error.message);
		}
	}

	function getEngagementColor(score) {
		if (score >= 75) return 'bg-[var(--teal)]';
		if (score >= 50) return 'bg-[var(--sage)]';
		if (score >= 25) return 'bg-[#f59e0b]';
		return 'bg-[#ef4444]';
	}

	function getEngagementLabel(score) {
		if (score >= 75) return 'High';
		if (score >= 50) return 'Medium';
		if (score >= 25) return 'Low';
		return 'Inactive';
	}

	async function savePerson() {
		try {
			await api(`/api/people/${personId}`, {
				method: 'PUT',
				body: JSON.stringify(editForm)
			});
			editing = false;
			loadPerson();
		} catch (error) {
			alert('Failed to save person: ' + error.message);
		}
	}

	async function deletePerson() {
		if (!confirm('Are you sure you want to delete this person?')) return;
		try {
			await api(`/api/people/${personId}`, {
				method: 'DELETE'
			});
			goto('/dashboard/people');
		} catch (error) {
			alert('Failed to delete person: ' + error.message);
		}
	}

	async function addTag() {
		if (!selectedTagId) return;
		try {
			await api(`/api/people/${personId}/tags`, {
				method: 'POST',
				body: JSON.stringify({ tag_id: selectedTagId })
			});
			showAddTag = false;
			selectedTagId = '';
			loadPerson();
		} catch (error) {
			alert('Failed to add tag: ' + error.message);
		}
	}

	async function removeTag(tagId) {
		try {
			await api(`/api/people/${personId}/tags/${tagId}`, {
				method: 'DELETE'
			});
			loadPerson();
		} catch (error) {
			alert('Failed to remove tag: ' + error.message);
		}
	}

	function getStatusColor(status) {
		const colors = {
			active: 'status-active',
			inactive: 'status-inactive',
			visitor: 'status-visitor',
			member: 'status-member'
		};
		return colors[status] || 'status-inactive';
	}
</script>

{#if loading}
	<div class="p-8 text-center text-secondary">Loading...</div>
{:else if person}
	<div class="space-y-6">
		<!-- Header -->
		<div class="flex justify-between items-start">
			<div>
				<button
					on:click={() => goto('/dashboard/people')}
					class="text-[var(--teal)] hover:underline mb-2"
				>
					← Back to People
				</button>
				<h1 class="text-3xl font-bold text-primary">
					{person.first_name}
					{person.last_name}
				</h1>
				<span
					class="mt-2 px-2 inline-flex text-xs leading-5 font-semibold rounded-full {getStatusColor(
						person.membership_status
					)}"
				>
					{person.membership_status}
				</span>
			</div>
			<div class="flex gap-2">
				{#if !editing}
					<button
						on:click={() => {
							editing = true;
							editForm = { ...person };
						}}
						class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:opacity-90"
					>
						Edit
					</button>
				{/if}
				<button
					on:click={deletePerson}
					class="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700"
				>
					Delete
				</button>
			</div>
		</div>

		<!-- Main content -->
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
			<!-- Info card -->
			<div class="md:col-span-2 bg-surface rounded-lg shadow p-6 border border-custom">
				{#if editing}
					<form on:submit|preventDefault={savePerson} class="space-y-4">
						<div class="grid grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-primary">First Name</label>
								<input
									type="text"
									bind:value={editForm.first_name}
									required
									class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
								/>
							</div>
							<div>
								<label class="block text-sm font-medium text-primary">Last Name</label>
								<input
									type="text"
									bind:value={editForm.last_name}
									required
									class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
								/>
							</div>
						</div>
						<div>
							<label class="block text-sm font-medium text-primary">Email</label>
							<input
								type="email"
								bind:value={editForm.email}
								class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
							/>
						</div>
						<div>
							<label class="block text-sm font-medium text-primary">Phone</label>
							<input
								type="tel"
								bind:value={editForm.phone}
								class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
							/>
						</div>
						<div>
							<label class="block text-sm font-medium text-primary">Address</label>
							<input
								type="text"
								bind:value={editForm.address_line1}
								placeholder="Street address"
								class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
							/>
							<input
								type="text"
								bind:value={editForm.address_line2}
								placeholder="Apt, suite, etc."
								class="mt-2 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
							/>
							<div class="grid grid-cols-3 gap-2 mt-2">
								<input
									type="text"
									bind:value={editForm.city}
									placeholder="City"
									class="px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
								/>
								<input
									type="text"
									bind:value={editForm.state}
									placeholder="State"
									class="px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
								/>
								<input
									type="text"
									bind:value={editForm.zip}
									placeholder="ZIP"
									class="px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
								/>
							</div>
						</div>
						<div>
							<label class="block text-sm font-medium text-primary">Gender</label>
							<input
								type="text"
								bind:value={editForm.gender}
								class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
							/>
						</div>
						<div>
							<label class="block text-sm font-medium text-primary">Status</label>
							<select
								bind:value={editForm.membership_status}
								class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
							>
								<option value="active">Active</option>
								<option value="inactive">Inactive</option>
								<option value="visitor">Visitor</option>
								<option value="member">Member</option>
							</select>
						</div>
						<div>
							<label class="block text-sm font-medium text-primary">Notes</label>
							<textarea
								bind:value={editForm.notes}
								rows="4"
								class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
							/>
						</div>
						<div class="flex gap-2 pt-4">
							<button
								type="button"
								on:click={() => (editing = false)}
								class="flex-1 px-4 py-2 border border-custom rounded-md hover:bg-[var(--surface-hover)] text-primary"
							>
								Cancel
							</button>
							<button
								type="submit"
								class="flex-1 px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:opacity-90"
							>
								Save
							</button>
						</div>
					</form>
				{:else}
					<dl class="space-y-4">
						<div>
							<dt class="text-sm font-medium text-secondary">Email</dt>
							<dd class="mt-1 text-sm text-primary">{person.email || '—'}</dd>
						</div>
						<div>
							<dt class="text-sm font-medium text-secondary">Phone</dt>
							<dd class="mt-1 text-sm text-primary">{person.phone || '—'}</dd>
						</div>
						{#if person.address_line1}
							<div>
								<dt class="text-sm font-medium text-secondary">Address</dt>
								<dd class="mt-1 text-sm text-primary">
									{person.address_line1}<br />
									{#if person.address_line2}{person.address_line2}<br />{/if}
									{#if person.city || person.state || person.zip}
										{person.city}, {person.state}
										{person.zip}
									{/if}
								</dd>
							</div>
						{/if}
						{#if person.gender}
							<div>
								<dt class="text-sm font-medium text-secondary">Gender</dt>
								<dd class="mt-1 text-sm text-primary">{person.gender}</dd>
							</div>
						{/if}
						{#if person.household}
							<div>
								<dt class="text-sm font-medium text-secondary">Household</dt>
								<dd class="mt-1 text-sm text-primary">{person.household.name}</dd>
							</div>
						{/if}
						{#if person.notes}
							<div>
								<dt class="text-sm font-medium text-secondary">Notes</dt>
								<dd class="mt-1 text-sm text-primary whitespace-pre-wrap">{person.notes}</dd>
							</div>
						{/if}
					</dl>
				{/if}
			</div>

			<!-- Sidebar -->
			<div class="space-y-6">
				<!-- Engagement Score -->
				{#if engagementScore}
					<div class="bg-surface rounded-lg shadow p-6 border border-custom">
						<div class="flex justify-between items-center mb-4">
							<h2 class="text-lg font-semibold text-primary">Engagement Score</h2>
							<button
								on:click={recalculateEngagement}
								class="text-[var(--teal)] hover:underline text-sm"
							>
								↻ Recalculate
							</button>
						</div>
						<div class="text-center mb-4">
							<div class="inline-flex items-center justify-center w-24 h-24 rounded-full {getEngagementColor(engagementScore.score)} text-white text-3xl font-bold">
								{engagementScore.score}
							</div>
							<p class="mt-2 text-sm text-secondary">{getEngagementLabel(engagementScore.score)} Engagement</p>
						</div>
						<div class="space-y-2">
							<div>
								<div class="flex justify-between text-sm mb-1">
									<span class="text-secondary">Attendance</span>
									<span class="text-primary font-medium">{engagementScore.attendance_score}/100</span>
								</div>
								<div class="w-full bg-gray-200 rounded-full h-2">
									<div class="bg-[var(--teal)] h-2 rounded-full" style="width: {engagementScore.attendance_score}%"></div>
								</div>
							</div>
							<div>
								<div class="flex justify-between text-sm mb-1">
									<span class="text-secondary">Giving</span>
									<span class="text-primary font-medium">{engagementScore.giving_score}/100</span>
								</div>
								<div class="w-full bg-gray-200 rounded-full h-2">
									<div class="bg-[var(--teal)] h-2 rounded-full" style="width: {engagementScore.giving_score}%"></div>
								</div>
							</div>
							<div>
								<div class="flex justify-between text-sm mb-1">
									<span class="text-secondary">Groups</span>
									<span class="text-primary font-medium">{engagementScore.group_score}/100</span>
								</div>
								<div class="w-full bg-gray-200 rounded-full h-2">
									<div class="bg-[var(--teal)] h-2 rounded-full" style="width: {engagementScore.group_score}%"></div>
								</div>
							</div>
							<div>
								<div class="flex justify-between text-sm mb-1">
									<span class="text-secondary">Volunteering</span>
									<span class="text-primary font-medium">{engagementScore.volunteer_score}/100</span>
								</div>
								<div class="w-full bg-gray-200 rounded-full h-2">
									<div class="bg-[var(--teal)] h-2 rounded-full" style="width: {engagementScore.volunteer_score}%"></div>
								</div>
							</div>
							<div>
								<div class="flex justify-between text-sm mb-1">
									<span class="text-secondary">Connection</span>
									<span class="text-primary font-medium">{engagementScore.connection_score}/100</span>
								</div>
								<div class="w-full bg-gray-200 rounded-full h-2">
									<div class="bg-[var(--teal)] h-2 rounded-full" style="width: {engagementScore.connection_score}%"></div>
								</div>
							</div>
						</div>
					</div>
				{/if}

				<!-- Tags -->
				<div class="bg-surface rounded-lg shadow p-6 border border-custom">
					<div class="flex justify-between items-center mb-4">
						<h2 class="text-lg font-semibold text-primary">Tags</h2>
						<button
							on:click={() => (showAddTag = !showAddTag)}
							class="text-[var(--teal)] hover:underline text-sm"
						>
							+ Add
						</button>
					</div>
					{#if showAddTag}
						<div class="mb-4 space-y-2">
							<select
								bind:value={selectedTagId}
								class="block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] text-sm bg-[var(--input-bg)] text-primary"
							>
								<option value="">Select tag...</option>
								{#each availableTags as tag}
									<option value={tag.id}>{tag.name}</option>
								{/each}
							</select>
							<button
								on:click={addTag}
								class="w-full px-3 py-2 bg-[var(--teal)] text-white rounded-md hover:opacity-90 text-sm"
							>
								Add Tag
							</button>
						</div>
					{/if}
					<div class="space-y-2">
						{#if person.tags && person.tags.length > 0}
							{#each person.tags as tag}
								<div
									class="flex items-center justify-between px-3 py-2 rounded-md"
									style="background-color: {tag.color}20"
								>
									<span class="text-sm" style="color: {tag.color}">{tag.name}</span>
									<button
										on:click={() => removeTag(tag.id)}
										class="text-red-600 hover:text-red-800 text-xs"
									>
										×
									</button>
								</div>
							{/each}
						{:else}
							<p class="text-sm text-secondary">No tags</p>
						{/if}
					</div>
				</div>

				<div class="bg-surface rounded-lg shadow p-6 border border-custom">
					<h2 class="text-lg font-semibold text-primary mb-4">Activity Timeline</h2>
					<p class="text-sm text-secondary">Coming soon...</p>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	.status-active {
		background-color: #D1FAE5;
		color: #065F46;
	}
	:global(.dark) .status-active {
		background-color: #064E3B;
		color: #6EE7B7;
	}
	
	.status-inactive {
		background-color: #F3F4F6;
		color: #374151;
	}
	:global(.dark) .status-inactive {
		background-color: #1F2937;
		color: #9CA3AF;
	}
	
	.status-visitor {
		background-color: #DBEAFE;
		color: #1E40AF;
	}
	:global(.dark) .status-visitor {
		background-color: #1E3A8A;
		color: #93C5FD;
	}
	
	.status-member {
		background-color: #CCFBF1;
		color: #115E59;
	}
	:global(.dark) .status-member {
		background-color: #134E4A;
		color: #5EEAD4;
	}
</style>
