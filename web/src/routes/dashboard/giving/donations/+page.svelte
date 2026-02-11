<script lang="ts">
	import { onMount } from 'svelte';

	let donations = [];
	let loading = true;
	let total = 0;
	let page = 1;
	let perPage = 20;

	// Filters
	let fundFilter = '';
	let personFilter = '';
	let fromDate = '';
	let toDate = '';

	let funds = [];

	onMount(async () => {
		await loadFunds();
		await loadDonations();
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

	async function loadDonations() {
		loading = true;
		try {
			let url = `/api/giving/donations?page=${page}&per_page=${perPage}`;
			if (fundFilter) url += `&fund_id=${fundFilter}`;
			if (personFilter) url += `&person_id=${personFilter}`;
			if (fromDate) url += `&from=${fromDate}`;
			if (toDate) url += `&to=${toDate}`;

			const response = await fetch(url, {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});
			if (response.ok) {
				const data = await response.json();
				donations = data.donations || [];
				total = data.total || 0;
			}
		} catch (error) {
			console.error('Failed to load donations:', error);
		} finally {
			loading = false;
		}
	}

	async function applyFilters() {
		page = 1;
		await loadDonations();
	}

	async function clearFilters() {
		fundFilter = '';
		personFilter = '';
		fromDate = '';
		toDate = '';
		page = 1;
		await loadDonations();
	}

	function formatCurrency(cents: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(cents / 100);
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function nextPage() {
		if ((page * perPage) < total) {
			page++;
			loadDonations();
		}
	}

	function prevPage() {
		if (page > 1) {
			page--;
			loadDonations();
		}
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'completed': return 'bg-green-100 text-green-800';
			case 'pending': return 'bg-yellow-100 text-yellow-800';
			case 'failed': return 'bg-red-100 text-red-800';
			case 'refunded': return 'bg-gray-100 text-gray-800';
			default: return 'bg-gray-100 text-gray-800';
		}
	}
</script>

<div class="p-6">
	<div class="mb-6 flex justify-between items-center">
		<div>
			<h1 class="text-3xl font-bold text-[#1B3A4B]">Donations</h1>
			<p class="text-gray-600 mt-1">View and manage all donations</p>
		</div>
		<a
			href="/dashboard/giving/donations/new"
			class="px-4 py-2 bg-[#4A8B8C] text-white rounded-lg hover:bg-[#3d7576] transition"
		>
			Record Donation
		</a>
	</div>

	<!-- Filters -->
	<div class="bg-white rounded-lg shadow p-6 mb-6">
		<h2 class="text-lg font-semibold text-[#1B3A4B] mb-4">Filters</h2>
		<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Fund</label>
				<select
					bind:value={fundFilter}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent"
				>
					<option value="">All Funds</option>
					{#each funds as fund}
						<option value={fund.id}>{fund.name}</option>
					{/each}
				</select>
			</div>

			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">From Date</label>
				<input
					type="date"
					bind:value={fromDate}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent"
				/>
			</div>

			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">To Date</label>
				<input
					type="date"
					bind:value={toDate}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent"
				/>
			</div>

			<div class="flex items-end gap-2">
				<button
					on:click={applyFilters}
					class="flex-1 px-4 py-2 bg-[#4A8B8C] text-white rounded-lg hover:bg-[#3d7576] transition"
				>
					Apply
				</button>
				<button
					on:click={clearFilters}
					class="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition"
				>
					Clear
				</button>
			</div>
		</div>
	</div>

	<!-- Donations Table -->
	<div class="bg-white rounded-lg shadow overflow-hidden">
		{#if loading}
			<div class="flex justify-center items-center py-12">
				<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[#4A8B8C]"></div>
			</div>
		{:else if donations.length === 0}
			<div class="text-center py-12">
				<p class="text-gray-500">No donations found</p>
			</div>
		{:else}
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Date
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Donor
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Fund
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Amount
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Method
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Status
						</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each donations as donation}
						<tr class="hover:bg-gray-50">
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
								{formatDate(donation.donated_at)}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
								{donation.person_name || 'Anonymous'}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
								{donation.fund_name}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm font-semibold text-[#4A8B8C]">
								{formatCurrency(donation.amount_cents)}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
								{donation.payment_method || 'N/A'}
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<span class="px-2 py-1 text-xs rounded-full {getStatusColor(donation.status)}">
									{donation.status}
								</span>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>

			<!-- Pagination -->
			<div class="px-6 py-4 border-t border-gray-200 flex justify-between items-center">
				<p class="text-sm text-gray-700">
					Showing {((page - 1) * perPage) + 1} to {Math.min(page * perPage, total)} of {total} donations
				</p>
				<div class="flex gap-2">
					<button
						on:click={prevPage}
						disabled={page === 1}
						class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Previous
					</button>
					<button
						on:click={nextPage}
						disabled={page * perPage >= total}
						class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Next
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
