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
	let connectStatus = {
		connected: false,
		onboarding_completed: false
	};

	onMount(async () => {
		await Promise.all([loadStats(), loadRecentDonations(), loadConnectStatus()]);
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

	async function loadConnectStatus() {
		try {
			const response = await fetch('/api/giving/connect/status', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});
			if (response.ok) {
				connectStatus = await response.json();
			}
		} catch (error) {
			console.error('Failed to load connect status:', error);
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
		<h1 class="text-3xl font-bold text-primary">Giving</h1>
		<p class="text-secondary mt-1">Track donations and manage funds</p>
	</div>

	{#if loading}
		<div class="flex justify-center items-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else}
		<!-- Stripe Not Connected Banner -->
		{#if !connectStatus.connected}
			<div class="warning-banner mb-6 rounded-lg shadow border border-custom">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<span class="text-2xl">⚠️</span>
						<div>
							<p class="font-semibold warning-text">Online giving is not set up yet.</p>
							<p class="text-sm warning-subtext">Connect with Stripe to start accepting online donations.</p>
						</div>
					</div>
					<a
						href="/dashboard/giving/settings"
						class="px-4 py-2 bg-yellow-600 text-white rounded-lg hover:bg-yellow-700 transition font-medium whitespace-nowrap"
					>
						Set up now →
					</a>
				</div>
			</div>
		{/if}
		<!-- Stats Cards -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
			<div class="bg-surface rounded-lg shadow p-6 border border-custom">
				<p class="text-sm text-secondary mb-1">This Month</p>
				<p class="text-3xl font-bold text-primary">{formatCurrency(stats.total_this_month)}</p>
			</div>

			<div class="bg-surface rounded-lg shadow p-6 border border-custom">
				<p class="text-sm text-secondary mb-1">This Year</p>
				<p class="text-3xl font-bold text-primary">{formatCurrency(stats.total_this_year)}</p>
			</div>

			<div class="bg-surface rounded-lg shadow p-6 border border-custom">
				<p class="text-sm text-secondary mb-1">All Time</p>
				<p class="text-3xl font-bold text-primary">{formatCurrency(stats.total_all_time)}</p>
			</div>

			<div class="bg-surface rounded-lg shadow p-6 border border-custom">
				<p class="text-sm text-secondary mb-1">Average Gift</p>
				<p class="text-3xl font-bold text-primary">{formatCurrency(stats.average_donation)}</p>
			</div>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
			<!-- Recent Donations -->
			<div class="bg-surface rounded-lg shadow border border-custom">
				<div class="p-6 border-b border-custom flex justify-between items-center">
					<h2 class="text-xl font-semibold text-primary">Recent Donations</h2>
					<a href="/dashboard/giving/donations" class="text-sm text-[var(--teal)] hover:underline">
						View All
					</a>
				</div>
				<div class="p-6">
					{#if recentDonations.length === 0}
						<p class="text-secondary text-center py-8">No donations yet</p>
					{:else}
						<div class="space-y-4">
							{#each recentDonations as donation}
								<div class="flex justify-between items-center border-b border-custom pb-3">
									<div>
										<p class="font-medium text-primary">{donation.person_name || 'Anonymous'}</p>
										<p class="text-sm text-secondary">{donation.fund_name}</p>
										<p class="text-xs text-secondary">{formatDate(donation.donated_at)}</p>
									</div>
									<p class="font-semibold text-[var(--teal)]">{formatCurrency(donation.amount_cents)}</p>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>

			<!-- Fund Breakdown -->
			<div class="bg-surface rounded-lg shadow border border-custom">
				<div class="p-6 border-b border-custom">
					<h2 class="text-xl font-semibold text-primary">Fund Breakdown</h2>
				</div>
				<div class="p-6">
					{#if stats.fund_breakdown.length === 0}
						<p class="text-secondary text-center py-8">No funds configured</p>
					{:else}
						<div class="space-y-4">
							{#each stats.fund_breakdown as fund}
								<div>
									<div class="flex justify-between items-center mb-2">
										<span class="font-medium text-primary">{fund.fund_name}</span>
										<span class="text-[var(--teal)] font-semibold">{formatCurrency(fund.total_cents)}</span>
									</div>
									<div class="w-full bg-[var(--surface-hover)] rounded-full h-2">
										<div
											class="bg-[var(--teal)] h-2 rounded-full"
											style="width: {fund.percentage}%"
										></div>
									</div>
									<p class="text-xs text-secondary mt-1">
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
		<div class="bg-surface rounded-lg shadow p-6 border border-custom">
			<h2 class="text-xl font-semibold text-primary mb-4">Quick Actions</h2>
			<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
				<a
					href="/dashboard/giving/donations/new"
					class="flex items-center justify-center px-4 py-3 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 transition"
				>
					Record Donation
				</a>
				<a
					href="/dashboard/giving/funds"
					class="flex items-center justify-center px-4 py-3 bg-[var(--sage)] text-white rounded-lg hover:opacity-90 transition"
				>
					Manage Funds
				</a>
				<a
					href="/dashboard/giving/statements"
					class="flex items-center justify-center px-4 py-3 bg-[var(--navy)] text-white rounded-lg hover:opacity-90 transition"
				>
					Giving Statements
				</a>
				<a
					href="/dashboard/giving/settings"
					class="flex items-center justify-center px-4 py-3 border-2 border-[var(--teal)] text-[var(--teal)] rounded-lg hover:bg-[var(--teal)] hover:text-white transition"
				>
					Settings
				</a>
			</div>
		</div>
	{/if}
</div>

<style>
	.warning-banner {
		background-color: #FEF3C7;
		border-left: 4px solid #F59E0B;
		padding: 1rem;
	}
	:global(.dark) .warning-banner {
		background-color: #451A03;
		border-left-color: #F59E0B;
	}
	
	.warning-text {
		color: #78350F;
	}
	:global(.dark) .warning-text {
		color: #FDE68A;
	}
	
	.warning-subtext {
		color: #92400E;
	}
	:global(.dark) .warning-subtext {
		color: #FCD34D;
	}
</style>
