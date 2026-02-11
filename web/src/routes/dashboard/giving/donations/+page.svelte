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
			case 'completed': return 'status-completed';
			case 'pending': return 'status-pending';
			case 'failed': return 'status-failed';
			case 'refunded': return 'status-refunded';
			default: return 'status-refunded';
		}
	}
</script>

<div class="p-6">
	<div class="mb-6 flex justify-between items-center">
		<div>
			<h1 class="text-3xl font-bold text-primary">Donations</h1>
			<p class="text-secondary mt-1">View and manage all donations</p>
		</div>
		<a
			href="/dashboard/giving/donations/new"
			class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 transition"
		>
			Record Donation
		</a>
	</div>

	<!-- Filters -->
	<div class="bg-surface rounded-lg shadow p-6 mb-6 border border-custom">
		<h2 class="text-lg font-semibold text-primary mb-4">Filters</h2>
		<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
			<div>
				<label class="block text-sm font-medium text-primary mb-1">Fund</label>
				<select
					bind:value={fundFilter}
					class="w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
				>
					<option value="">All Funds</option>
					{#each funds as fund}
						<option value={fund.id}>{fund.name}</option>
					{/each}
				</select>
			</div>

			<div>
				<label class="block text-sm font-medium text-primary mb-1">From Date</label>
				<input
					type="date"
					bind:value={fromDate}
					class="w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
				/>
			</div>

			<div>
				<label class="block text-sm font-medium text-primary mb-1">To Date</label>
				<input
					type="date"
					bind:value={toDate}
					class="w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
				/>
			</div>

			<div class="flex items-end gap-2">
				<button
					on:click={applyFilters}
					class="flex-1 px-4 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 transition"
				>
					Apply
				</button>
				<button
					on:click={clearFilters}
					class="flex-1 px-4 py-2 border border-custom text-primary rounded-lg hover:bg-[var(--surface-hover)] transition"
				>
					Clear
				</button>
			</div>
		</div>
	</div>

	<!-- Donations Table -->
	<div class="bg-surface rounded-lg shadow overflow-hidden border border-custom">
		{#if loading}
			<div class="flex justify-center items-center py-12">
				<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--teal)]"></div>
			</div>
		{:else if donations.length === 0}
			<div class="text-center py-12">
				<p class="text-secondary">No donations found</p>
			</div>
		{:else}
			<table class="min-w-full divide-y divide-[var(--border)]">
				<thead class="bg-[var(--surface-hover)]">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
							Date
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
							Donor
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
							Fund
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
							Amount
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
							Method
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
							Status
						</th>
					</tr>
				</thead>
				<tbody class="bg-surface divide-y divide-[var(--border)]">
					{#each donations as donation}
						<tr class="hover:bg-[var(--surface-hover)]">
							<td class="px-6 py-4 whitespace-nowrap text-sm text-primary">
								{formatDate(donation.donated_at)}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-primary">
								{donation.person_name || 'Anonymous'}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-primary">
								{donation.fund_name}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm font-semibold text-[var(--teal)]">
								{formatCurrency(donation.amount_cents)}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-secondary">
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
			<div class="px-6 py-4 border-t border-custom flex justify-between items-center">
				<p class="text-sm text-secondary">
					Showing {((page - 1) * perPage) + 1} to {Math.min(page * perPage, total)} of {total} donations
				</p>
				<div class="flex gap-2">
					<button
						on:click={prevPage}
						disabled={page === 1}
						class="px-4 py-2 border border-custom rounded-lg hover:bg-[var(--surface-hover)] disabled:opacity-50 disabled:cursor-not-allowed text-primary"
					>
						Previous
					</button>
					<button
						on:click={nextPage}
						disabled={page * perPage >= total}
						class="px-4 py-2 border border-custom rounded-lg hover:bg-[var(--surface-hover)] disabled:opacity-50 disabled:cursor-not-allowed text-primary"
					>
						Next
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.status-completed {
		background-color: #D1FAE5;
		color: #065F46;
	}
	:global(.dark) .status-completed {
		background-color: #064E3B;
		color: #6EE7B7;
	}
	
	.status-pending {
		background-color: #FEF3C7;
		color: #92400E;
	}
	:global(.dark) .status-pending {
		background-color: #78350F;
		color: #FCD34D;
	}
	
	.status-failed {
		background-color: #FEE2E2;
		color: #991B1B;
	}
	:global(.dark) .status-failed {
		background-color: #7F1D1D;
		color: #FCA5A5;
	}
	
	.status-refunded {
		background-color: #F3F4F6;
		color: #374151;
	}
	:global(.dark) .status-refunded {
		background-color: #1F2937;
		color: #9CA3AF;
	}
</style>
