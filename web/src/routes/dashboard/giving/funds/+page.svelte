<script lang="ts">
	import { onMount } from 'svelte';

	let funds = [];
	let loading = true;
	let showModal = false;
	let editingFund = null;

	let formData = {
		name: '',
		description: '',
		is_default: false,
		is_active: true
	};

	onMount(async () => {
		await loadFunds();
		loading = false;
	});

	async function loadFunds() {
		try {
			const response = await fetch('/api/giving/funds', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});
			if (response.ok) {
				funds = await response.json();
			}
		} catch (error) {
			console.error('Failed to load funds:', error);
		}
	}

	function openModal(fund = null) {
		if (fund) {
			editingFund = fund;
			formData = {
				name: fund.name,
				description: fund.description || '',
				is_default: fund.is_default,
				is_active: fund.is_active
			};
		} else {
			editingFund = null;
			formData = {
				name: '',
				description: '',
				is_default: false,
				is_active: true
			};
		}
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		editingFund = null;
	}

	async function saveFund() {
		try {
			const url = editingFund
				? `/api/giving/funds/${editingFund.id}`
				: '/api/giving/funds';
			
			const method = editingFund ? 'PUT' : 'POST';

			const response = await fetch(url, {
				method,
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`,
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(formData)
			});

			if (response.ok) {
				await loadFunds();
				closeModal();
			} else {
				alert('Failed to save fund');
			}
		} catch (error) {
			console.error('Failed to save fund:', error);
			alert('An error occurred');
		}
	}
</script>

<div class="p-6">
	<div class="mb-6 flex justify-between items-center">
		<div>
			<h1 class="text-3xl font-bold text-[#1B3A4B]">Funds</h1>
			<p class="text-gray-600 mt-1">Manage giving funds and designations</p>
		</div>
		<button
			on:click={() => openModal()}
			class="px-4 py-2 bg-[#4A8B8C] text-white rounded-lg hover:bg-[#3d7576] transition"
		>
			Create Fund
		</button>
	</div>

	{#if loading}
		<div class="flex justify-center items-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[#4A8B8C]"></div>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each funds as fund}
				<div class="bg-white rounded-lg shadow p-6">
					<div class="flex justify-between items-start mb-4">
						<div class="flex-1">
							<h3 class="text-lg font-semibold text-[#1B3A4B] mb-1">{fund.name}</h3>
							{#if fund.description}
								<p class="text-sm text-gray-600">{fund.description}</p>
							{/if}
						</div>
						<button
							on:click={() => openModal(fund)}
							class="text-[#4A8B8C] hover:text-[#3d7576] ml-2"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
							</svg>
						</button>
					</div>
					
					<div class="flex gap-2 flex-wrap">
						{#if fund.is_default}
							<span class="px-2 py-1 bg-[#4A8B8C] text-white text-xs rounded-full">Default</span>
						{/if}
						{#if fund.is_active}
							<span class="px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full">Active</span>
						{:else}
							<span class="px-2 py-1 bg-gray-100 text-gray-800 text-xs rounded-full">Inactive</span>
						{/if}
					</div>
				</div>
			{/each}
		</div>

		{#if funds.length === 0}
			<div class="bg-white rounded-lg shadow p-12 text-center">
				<p class="text-gray-500 mb-4">No funds created yet</p>
				<button
					on:click={() => openModal()}
					class="px-4 py-2 bg-[#4A8B8C] text-white rounded-lg hover:bg-[#3d7576] transition"
				>
					Create Your First Fund
				</button>
			</div>
		{/if}
	{/if}
</div>

<!-- Modal -->
{#if showModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
		<div class="bg-white rounded-lg shadow-xl max-w-lg w-full p-6">
			<h2 class="text-2xl font-bold text-[#1B3A4B] mb-4">
				{editingFund ? 'Edit Fund' : 'Create Fund'}
			</h2>

			<form on:submit|preventDefault={saveFund} class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">
						Fund Name <span class="text-red-500">*</span>
					</label>
					<input
						type="text"
						bind:value={formData.name}
						required
						placeholder="e.g., General Fund, Building Fund, Missions"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">
						Description
					</label>
					<textarea
						bind:value={formData.description}
						rows="3"
						placeholder="Brief description of this fund..."
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent"
					></textarea>
				</div>

				<div class="flex items-center">
					<input
						type="checkbox"
						id="is_default"
						bind:checked={formData.is_default}
						class="h-4 w-4 text-[#4A8B8C] focus:ring-[#4A8B8C] border-gray-300 rounded"
					/>
					<label for="is_default" class="ml-2 text-sm text-gray-700">
						Set as default fund
					</label>
				</div>

				<div class="flex items-center">
					<input
						type="checkbox"
						id="is_active"
						bind:checked={formData.is_active}
						class="h-4 w-4 text-[#4A8B8C] focus:ring-[#4A8B8C] border-gray-300 rounded"
					/>
					<label for="is_active" class="ml-2 text-sm text-gray-700">
						Active (visible for donations)
					</label>
				</div>

				<div class="flex gap-4 pt-4">
					<button
						type="submit"
						class="flex-1 px-4 py-2 bg-[#4A8B8C] text-white rounded-lg hover:bg-[#3d7576] transition"
					>
						{editingFund ? 'Update' : 'Create'}
					</button>
					<button
						type="button"
						on:click={closeModal}
						class="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition"
					>
						Cancel
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
