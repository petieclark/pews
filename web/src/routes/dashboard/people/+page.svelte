<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let people = [];
	let tags = [];
	let total = 0;
	let page = 1;
	let limit = 50;
	let searchQuery = '';
	let loading = false;
	let showCreateModal = false;
	let newPerson = {
		first_name: '',
		last_name: '',
		email: '',
		phone: '',
		membership_status: 'active'
	};

	onMount(() => {
		loadPeople();
		loadTags();
	});

	async function loadPeople() {
		loading = true;
		try {
			const params = new URLSearchParams({
				page: page.toString(),
				limit: limit.toString()
			});
			if (searchQuery) {
				params.append('q', searchQuery);
			}
			const response = await api(`/api/people?${params}`);
			people = response.people || [];
			total = response.total || 0;
		} catch (error) {
			console.error('Failed to load people:', error);
		} finally {
			loading = false;
		}
	}

	async function loadTags() {
		try {
			tags = await api('/api/tags');
		} catch (error) {
			console.error('Failed to load tags:', error);
		}
	}

	function handleSearch() {
		page = 1;
		loadPeople();
	}

	function viewPerson(id) {
		goto(`/dashboard/people/${id}`);
	}

	async function createPerson() {
		try {
			await api('/api/people', {
				method: 'POST',
				body: JSON.stringify(newPerson)
			});
			showCreateModal = false;
			newPerson = {
				first_name: '',
				last_name: '',
				email: '',
				phone: '',
				membership_status: 'active'
			};
			loadPeople();
		} catch (error) {
			alert('Failed to create person: ' + error.message);
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

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-navy">People</h1>
		<button
			on:click={() => (showCreateModal = true)}
			class="px-4 py-2 bg-teal text-white rounded-md hover:bg-opacity-90"
		>
			Add Person
		</button>
	</div>

	<!-- Search and filters -->
	<div class="bg-white rounded-lg shadow p-4">
		<div class="flex gap-4">
			<input
				type="text"
				bind:value={searchQuery}
				on:keyup={(e) => e.key === 'Enter' && handleSearch()}
				placeholder="Search by name, email, or phone..."
				class="flex-1 px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
			/>
			<button
				on:click={handleSearch}
				class="px-6 py-2 bg-navy text-white rounded-md hover:bg-opacity-90"
			>
				Search
			</button>
		</div>
	</div>

	<!-- People table -->
	<div class="bg-white rounded-lg shadow overflow-hidden">
		{#if loading}
			<div class="p-8 text-center text-gray-500">Loading...</div>
		{:else if people.length === 0}
			<div class="p-8 text-center text-gray-500">
				No people found. {#if searchQuery}Try a different search.{:else}Add your first person to get
					started.{/if}
			</div>
		{:else}
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>Name</th
						>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>Email</th
						>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>Phone</th
						>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>Status</th
						>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>Tags</th
						>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each people as person}
						<tr
							on:click={() => viewPerson(person.id)}
							class="hover:bg-gray-50 cursor-pointer"
						>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm font-medium text-gray-900">
									{person.first_name}
									{person.last_name}
								</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-gray-500">{person.email || '—'}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-gray-500">{person.phone || '—'}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<span
									class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full {getStatusColor(
										person.membership_status
									)}"
								>
									{person.membership_status}
								</span>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="flex gap-1">
									{#if person.tags && person.tags.length > 0}
										{#each person.tags.slice(0, 3) as tag}
											<span
												class="px-2 py-1 text-xs rounded"
												style="background-color: {tag.color}20; color: {tag.color}"
											>
												{tag.name}
											</span>
										{/each}
										{#if person.tags.length > 3}
											<span class="text-xs text-gray-500">+{person.tags.length - 3}</span>
										{/if}
									{/if}
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>

	<!-- Pagination -->
	{#if total > limit}
		<div class="flex justify-center gap-2">
			<button
				on:click={() => {
					page--;
					loadPeople();
				}}
				disabled={page === 1}
				class="px-4 py-2 bg-white border rounded-md disabled:opacity-50"
			>
				Previous
			</button>
			<span class="px-4 py-2">
				Page {page} of {Math.ceil(total / limit)}
			</span>
			<button
				on:click={() => {
					page++;
					loadPeople();
				}}
				disabled={page >= Math.ceil(total / limit)}
				class="px-4 py-2 bg-white border rounded-md disabled:opacity-50"
			>
				Next
			</button>
		</div>
	{/if}
</div>

<!-- Create person modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
		<div class="bg-white rounded-lg max-w-md w-full p-6">
			<h2 class="text-2xl font-bold text-navy mb-4">Add Person</h2>
			<form on:submit|preventDefault={createPerson} class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700">First Name *</label>
					<input
						type="text"
						bind:value={newPerson.first_name}
						required
						class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700">Last Name *</label>
					<input
						type="text"
						bind:value={newPerson.last_name}
						required
						class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700">Email</label>
					<input
						type="email"
						bind:value={newPerson.email}
						class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700">Phone</label>
					<input
						type="tel"
						bind:value={newPerson.phone}
						class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700">Status</label>
					<select
						bind:value={newPerson.membership_status}
						class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					>
						<option value="active">Active</option>
						<option value="inactive">Inactive</option>
						<option value="visitor">Visitor</option>
						<option value="member">Member</option>
					</select>
				</div>
				<div class="flex gap-2 pt-4">
					<button
						type="button"
						on:click={() => (showCreateModal = false)}
						class="flex-1 px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="flex-1 px-4 py-2 bg-teal text-white rounded-md hover:bg-opacity-90"
					>
						Create
					</button>
				</div>
			</form>
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
