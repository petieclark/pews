<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let households = [];
	let loading = false;
	let showCreateModal = false;
	let newHousehold = {
		name: '',
		address_line1: '',
		address_line2: '',
		city: '',
		state: '',
		zip: ''
	};

	onMount(() => {
		loadHouseholds();
	});

	async function loadHouseholds() {
		loading = true;
		try {
			households = await api('/api/households');
		} catch (error) {
			console.error('Failed to load households:', error);
		} finally {
			loading = false;
		}
	}

	async function createHousehold() {
		try {
			await api('/api/households', {
				method: 'POST',
				body: JSON.stringify(newHousehold)
			});
			showCreateModal = false;
			newHousehold = {
				name: '',
				address_line1: '',
				address_line2: '',
				city: '',
				state: '',
				zip: ''
			};
			loadHouseholds();
		} catch (error) {
			alert('Failed to create household: ' + error.message);
		}
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<div>
			<a href="/dashboard/people" class="text-teal hover:underline mb-2">← Back to People</a>
			<h1 class="text-3xl font-bold text-navy">Households</h1>
		</div>
		<button
			on:click={() => (showCreateModal = true)}
			class="px-4 py-2 bg-teal text-white rounded-md hover:bg-opacity-90"
		>
			Add Household
		</button>
	</div>

	<!-- Households list -->
	<div class="bg-surface rounded-lg shadow overflow-hidden">
		{#if loading}
			<div class="p-8 text-center text-secondary">Loading...</div>
		{:else if households.length === 0}
			<div class="p-8 text-center text-secondary">
				No households found. Create your first household to get started.
			</div>
		{:else}
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-4">
				{#each households as household}
					<div class="border rounded-lg p-4 hover:shadow-md transition-shadow">
						<h3 class="text-lg font-semibold text-navy mb-2">{household.name}</h3>
						{#if household.address_line1}
							<div class="text-sm text-secondary">
								<p>{household.address_line1}</p>
								{#if household.address_line2}<p>{household.address_line2}</p>{/if}
								{#if household.city || household.state || household.zip}
									<p>
										{household.city}, {household.state}
										{household.zip}
									</p>
								{/if}
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- Create household modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
		<div class="bg-surface rounded-lg max-w-md w-full p-6">
			<h2 class="text-2xl font-bold text-navy mb-4">Add Household</h2>
			<form on:submit|preventDefault={createHousehold} class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-primary">Household Name *</label>
					<input
						type="text"
						bind:value={newHousehold.name}
						required
						placeholder="e.g., Smith Family"
						class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-primary">Address</label>
					<input
						type="text"
						bind:value={newHousehold.address_line1}
						placeholder="Street address"
						class="mt-1 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					/>
					<input
						type="text"
						bind:value={newHousehold.address_line2}
						placeholder="Apt, suite, etc."
						class="mt-2 block w-full px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					/>
					<div class="grid grid-cols-3 gap-2 mt-2">
						<input
							type="text"
							bind:value={newHousehold.city}
							placeholder="City"
							class="px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
						/>
						<input
							type="text"
							bind:value={newHousehold.state}
							placeholder="State"
							class="px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
						/>
						<input
							type="text"
							bind:value={newHousehold.zip}
							placeholder="ZIP"
							class="px-3 py-2 border input-border rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
						/>
					</div>
				</div>
				<div class="flex gap-2 pt-4">
					<button
						type="button"
						on:click={() => (showCreateModal = false)}
						class="flex-1 px-4 py-2 border input-border rounded-md hover:bg-[var(--surface-hover)]"
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
