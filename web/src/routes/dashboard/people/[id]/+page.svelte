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
			active: 'bg-green-100 text-green-800',
			inactive: 'bg-gray-100 text-gray-800',
			visitor: 'bg-blue-100 text-blue-800',
			member: 'bg-teal-100 text-teal-800'
		};
		return colors[status] || 'bg-gray-100 text-gray-800';
	}
</script>

{#if loading}
	<div class="p-8 text-center text-gray-500">Loading...</div>
{:else if person}
	<div class="space-y-6">
		<!-- Header -->
		<div class="flex justify-between items-start">
			<div>
				<button
					on:click={() => goto('/dashboard/people')}
					class="text-teal hover:underline mb-2"
				>
					← Back to People
				</button>
				<h1 class="text-3xl font-bold text-navy">
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
						class="px-4 py-2 bg-teal text-white rounded-md hover:bg-opacity-90"
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
			<div class="md:col-span-2 bg-white rounded-lg shadow p-6">
				{#if editing}
					<form on:submit|preventDefault={savePerson} class="space-y-4">
						<div class="grid grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700">First Name</label>
								<input
									type="text"
									bind:value={editForm.first_name}
									required
									class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
								/>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700">Last Name</label>
								<input
									type="text"
									bind:value={editForm.last_name}
									required
									class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
								/>
							</div>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700">Email</label>
							<input
								type="email"
								bind:value={editForm.email}
								class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
							/>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700">Phone</label>
							<input
								type="tel"
								bind:value={editForm.phone}
								class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
							/>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700">Address</label>
							<input
								type="text"
								bind:value={editForm.address_line1}
								placeholder="Street address"
								class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
							/>
							<input
								type="text"
								bind:value={editForm.address_line2}
								placeholder="Apt, suite, etc."
								class="mt-2 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
							/>
							<div class="grid grid-cols-3 gap-2 mt-2">
								<input
									type="text"
									bind:value={editForm.city}
									placeholder="City"
									class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
								/>
								<input
									type="text"
									bind:value={editForm.state}
									placeholder="State"
									class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
								/>
								<input
									type="text"
									bind:value={editForm.zip}
									placeholder="ZIP"
									class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
								/>
							</div>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700">Gender</label>
							<input
								type="text"
								bind:value={editForm.gender}
								class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
							/>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700">Status</label>
							<select
								bind:value={editForm.membership_status}
								class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
							>
								<option value="active">Active</option>
								<option value="inactive">Inactive</option>
								<option value="visitor">Visitor</option>
								<option value="member">Member</option>
							</select>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700">Notes</label>
							<textarea
								bind:value={editForm.notes}
								rows="4"
								class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
							/>
						</div>
						<div class="flex gap-2 pt-4">
							<button
								type="button"
								on:click={() => (editing = false)}
								class="flex-1 px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
							>
								Cancel
							</button>
							<button
								type="submit"
								class="flex-1 px-4 py-2 bg-teal text-white rounded-md hover:bg-opacity-90"
							>
								Save
							</button>
						</div>
					</form>
				{:else}
					<dl class="space-y-4">
						<div>
							<dt class="text-sm font-medium text-gray-500">Email</dt>
							<dd class="mt-1 text-sm text-gray-900">{person.email || '—'}</dd>
						</div>
						<div>
							<dt class="text-sm font-medium text-gray-500">Phone</dt>
							<dd class="mt-1 text-sm text-gray-900">{person.phone || '—'}</dd>
						</div>
						{#if person.address_line1}
							<div>
								<dt class="text-sm font-medium text-gray-500">Address</dt>
								<dd class="mt-1 text-sm text-gray-900">
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
								<dt class="text-sm font-medium text-gray-500">Gender</dt>
								<dd class="mt-1 text-sm text-gray-900">{person.gender}</dd>
							</div>
						{/if}
						{#if person.household}
							<div>
								<dt class="text-sm font-medium text-gray-500">Household</dt>
								<dd class="mt-1 text-sm text-gray-900">{person.household.name}</dd>
							</div>
						{/if}
						{#if person.notes}
							<div>
								<dt class="text-sm font-medium text-gray-500">Notes</dt>
								<dd class="mt-1 text-sm text-gray-900 whitespace-pre-wrap">{person.notes}</dd>
							</div>
						{/if}
					</dl>
				{/if}
			</div>

			<!-- Tags sidebar -->
			<div class="space-y-6">
				<div class="bg-white rounded-lg shadow p-6">
					<div class="flex justify-between items-center mb-4">
						<h2 class="text-lg font-semibold text-navy">Tags</h2>
						<button
							on:click={() => (showAddTag = !showAddTag)}
							class="text-teal hover:underline text-sm"
						>
							+ Add
						</button>
					</div>
					{#if showAddTag}
						<div class="mb-4 space-y-2">
							<select
								bind:value={selectedTagId}
								class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal text-sm"
							>
								<option value="">Select tag...</option>
								{#each availableTags as tag}
									<option value={tag.id}>{tag.name}</option>
								{/each}
							</select>
							<button
								on:click={addTag}
								class="w-full px-3 py-2 bg-teal text-white rounded-md hover:bg-opacity-90 text-sm"
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
							<p class="text-sm text-gray-500">No tags</p>
						{/if}
					</div>
				</div>

				<div class="bg-white rounded-lg shadow p-6">
					<h2 class="text-lg font-semibold text-navy mb-4">Activity Timeline</h2>
					<p class="text-sm text-gray-500">Coming soon...</p>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	:global(.bg-navy) {
		background-color: #1b3a4b;
	}
	:global(.text-navy) {
		color: #1b3a4b;
	}
	:global(.bg-teal) {
		background-color: #4a8b8c;
	}
	:global(.text-teal) {
		color: #4a8b8c;
	}
	:global(.ring-teal) {
		--tw-ring-color: #4a8b8c;
	}
</style>
