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
			active: 'status-active',
			inactive: 'status-inactive',
			visitor: 'status-visitor',
			member: 'status-member'
		};
		return colors[status] || 'status-inactive';
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-primary">People</h1>
		<button
			on:click={() => (showCreateModal = true)}
			class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:opacity-90"
		>
			Add Person
		</button>
	</div>

	<!-- Search and filters -->
	<div class="bg-surface rounded-lg shadow p-4 border border-custom">
		<div class="flex gap-4">
			<input
				type="text"
				bind:value={searchQuery}
				on:keyup={(e) => e.key === 'Enter' && handleSearch()}
				placeholder="Search by name, email, or phone..."
				class="flex-1 px-4 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
			/>
			<button
				on:click={handleSearch}
				class="px-6 py-2 bg-[var(--navy)] text-white rounded-md hover:opacity-90"
			>
				Search
			</button>
		</div>
	</div>

	<!-- People table -->
	<div class="bg-surface rounded-lg shadow overflow-hidden border border-custom">
		{#if loading}
			<div class="p-8 text-center text-secondary">Loading...</div>
		{:else if people.length === 0}
			<div class="p-8 text-center text-secondary">
				No people found. {#if searchQuery}Try a different search.{:else}Add your first person to get
					started.{/if}
			</div>
		{:else}
			<table class="min-w-full divide-y divide-[var(--border)]">
				<thead class="bg-[var(--surface-hover)]">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider"
							>Name</th
						>
						<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider"
							>Email</th
						>
						<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider"
							>Phone</th
						>
						<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider"
							>Status</th
						>
						<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider"
							>Tags</th
						>
					</tr>
				</thead>
				<tbody class="bg-surface divide-y divide-[var(--border)]">
					{#each people as person}
						<tr
							on:click={() => viewPerson(person.id)}
							class="hover:bg-[var(--surface-hover)] cursor-pointer"
						>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm font-medium text-primary">
									{person.first_name}
									{person.last_name}
								</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-secondary">{person.email || '—'}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-secondary">{person.phone || '—'}</div>
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
											<span class="text-xs text-secondary">+{person.tags.length - 3}</span>
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
				class="px-4 py-2 bg-surface border border-custom rounded-md disabled:opacity-50 text-primary"
			>
				Previous
			</button>
			<span class="px-4 py-2 text-primary">
				Page {page} of {Math.ceil(total / limit)}
			</span>
			<button
				on:click={() => {
					page++;
					loadPeople();
				}}
				disabled={page >= Math.ceil(total / limit)}
				class="px-4 py-2 bg-surface border border-custom rounded-md disabled:opacity-50 text-primary"
			>
				Next
			</button>
		</div>
	{/if}
</div>

<!-- Create person modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
		<div class="bg-surface rounded-lg max-w-md w-full p-6 border border-custom">
			<h2 class="text-2xl font-bold text-primary mb-4">Add Person</h2>
			<form on:submit|preventDefault={createPerson} class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-primary">First Name *</label>
					<input
						type="text"
						bind:value={newPerson.first_name}
						required
						class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-primary">Last Name *</label>
					<input
						type="text"
						bind:value={newPerson.last_name}
						required
						class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-primary">Email</label>
					<input
						type="email"
						bind:value={newPerson.email}
						class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-primary">Phone</label>
					<input
						type="tel"
						bind:value={newPerson.phone}
						class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-primary">Status</label>
					<select
						bind:value={newPerson.membership_status}
						class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"
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
						class="flex-1 px-4 py-2 border border-custom rounded-md hover:bg-[var(--surface-hover)] text-primary"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="flex-1 px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:opacity-90"
					>
						Create
					</button>
				</div>
			</form>
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
