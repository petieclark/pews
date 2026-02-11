<script lang="ts">
	import { onMount } from 'svelte';

	let stats = {
		total_this_month: 0,
		total_this_year: 0,
		total_all_time: 0,
		donation_count: 0,
		average_donation: 0,
		fund_breakdown: [],
		monthly_trend: []
	};

	let recentDonations = [];
	let loading = true;

	onMount(async () => {
		await Promise.all([loadStats(), loadRecentDonations()]);
		loading = false;
	});

	async function loadStats() {
		try {
			const response = await fetch('/api/giving/stats', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});
			if (response.ok) {
				stats = await response.json();
			}
		} catch (error) {
			console.error('Failed to load stats:', error);
		}
	}

	async function loadRecentDonations() {
		try {
			const response = await fetch('/api/giving/donations?per_page=10', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});
			if (response.ok) {
				const data = await response.json();
				recentDonations = data.donations || [];
			}
		} catch (error) {
			console.error('Failed to load donations:', error);
		}
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
</script>

<div class="p-6">
	<div class="mb-6">
		<h1 class="text-3xl font-bold text-[#1B3A4B]">Giving</h1>
		<p class="text-gray-600 mt-1">Track donations and manage funds</p>
	</div>

	{#if loading}
		<div class="flex justify-center items-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[#4A8B8C]"></div>
		</div>
	{:else}
		<!-- Stats Cards -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
			<div class="bg-white rounded-lg shadow p-6">
				<p class="text-sm text-gray-600 mb-1">This Month</p>
				<p class="text-3xl font-bold text-[#1B3A4B]">{formatCurrency(stats.total_this_month)}</p>
			</div>

			<div class="bg-white rounded-lg shadow p-6">
				<p class="text-sm text-gray-600 mb-1">This Year</p>
				<p class="text-3xl font-bold text-[#1B3A4B]">{formatCurrency(stats.total_this_year)}</p>
			</div>

			<div class="bg-white rounded-lg shadow p-6">
				<p class="text-sm text-gray-600 mb-1">All Time</p>
				<p class="text-3xl font-bold text-[#1B3A4B]">{formatCurrency(stats.total_all_time)}</p>
			</div>

			<div class="bg-white rounded-lg shadow p-6">
				<p class="text-sm text-gray-600 mb-1">Average Gift</p>
				<p class="text-3xl font-bold text-[#1B3A4B]">{formatCurrency(stats.average_donation)}</p>
			</div>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
			<!-- Recent Donations -->
			<div class="bg-white rounded-lg shadow">
				<div class="p-6 border-b border-gray-200 flex justify-between items-center">
					<h2 class="text-xl font-semibold text-[#1B3A4B]">Recent Donations</h2>
					<a href="/dashboard/giving/donations" class="text-sm text-[#4A8B8C] hover:underline">
						View All
					</a>
				</div>
				<div class="p-6">
					{#if recentDonations.length === 0}
						<p class="text-gray-500 text-center py-8">No donations yet</p>
					{:else}
						<div class="space-y-4">
							{#each recentDonations as donation}
								<div class="flex justify-between items-center border-b border-gray-100 pb-3">
									<div>
										<p class="font-medium text-[#1B3A4B]">{donation.person_name || 'Anonymous'}</p>
										<p class="text-sm text-gray-600">{donation.fund_name}</p>
										<p class="text-xs text-gray-500">{formatDate(donation.donated_at)}</p>
									</div>
									<p class="font-semibold text-[#4A8B8C]">{formatCurrency(donation.amount_cents)}</p>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>

			<!-- Fund Breakdown -->
			<div class="bg-white rounded-lg shadow">
				<div class="p-6 border-b border-gray-200">
					<h2 class="text-xl font-semibold text-[#1B3A4B]">Fund Breakdown</h2>
				</div>
				<div class="p-6">
					{#if stats.fund_breakdown.length === 0}
						<p class="text-gray-500 text-center py-8">No funds configured</p>
					{:else}
						<div class="space-y-4">
							{#each stats.fund_breakdown as fund}
								<div>
									<div class="flex justify-between items-center mb-2">
										<span class="font-medium text-[#1B3A4B]">{fund.fund_name}</span>
										<span class="text-[#4A8B8C] font-semibold">{formatCurrency(fund.total_cents)}</span>
									</div>
									<div class="w-full bg-gray-200 rounded-full h-2">
										<div
											class="bg-[#4A8B8C] h-2 rounded-full"
											style="width: {fund.percentage}%"
										></div>
									</div>
									<p class="text-xs text-gray-500 mt-1">
										{fund.percentage.toFixed(1)}% • {fund.donor_count} donor{fund.donor_count !== 1 ? 's' : ''}
									</p>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		</div>

		<!-- Quick Actions -->
		<div class="bg-white rounded-lg shadow p-6">
			<h2 class="text-xl font-semibold text-[#1B3A4B] mb-4">Quick Actions</h2>
			<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
				<a
					href="/dashboard/giving/donations/new"
					class="flex items-center justify-center px-4 py-3 bg-[#4A8B8C] text-white rounded-lg hover:bg-[#3d7576] transition"
				>
					Record Donation
				</a>
				<a
					href="/dashboard/giving/funds"
					class="flex items-center justify-center px-4 py-3 bg-[#8FBCB0] text-white rounded-lg hover:bg-[#7aab9f] transition"
				>
					Manage Funds
				</a>
				<a
					href="/dashboard/giving/statements"
					class="flex items-center justify-center px-4 py-3 bg-[#1B3A4B] text-white rounded-lg hover:bg-[#152e3a] transition"
				>
					Giving Statements
				</a>
				<a
					href="/dashboard/giving/settings"
					class="flex items-center justify-center px-4 py-3 border-2 border-[#4A8B8C] text-[#4A8B8C] rounded-lg hover:bg-[#4A8B8C] hover:text-white transition"
				>
					Settings
				</a>
			</div>
		</div>
	{/if}
</div>
