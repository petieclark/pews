<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { Chart, registerables } from 'chart.js';
	Chart.register(...registerables);

	let stats = {
		total_this_month: 0,
		total_this_year: 0,
		total_all_time: 0,
		donation_count: 0,
		average_donation: 0,
		fund_breakdown: [] as any[],
		monthly_trend: [] as any[]
	};

	let recentDonations: any[] = [];
	let loading = true;
	let connectStatus = {
		connected: false,
		onboarding_completed: false
	};
	let trendChart: Chart | null = null;
	let chartCanvas: HTMLCanvasElement;

	onMount(async () => {
		await Promise.all([loadStats(), loadRecentDonations(), loadConnectStatus()]);
		loading = false;
		if (stats.monthly_trend.length > 0) {
			renderChart();
		}
	});

	async function loadStats() {
		try {
			stats = await api('/api/giving/stats', { silent: true });
		} catch (error) { console.error('Failed to load stats:', error); }
	}

	async function loadRecentDonations() {
		try {
			const data = await api('/api/giving/donations?per_page=10', { silent: true });
			recentDonations = data.donations || [];
		} catch (error) { console.error('Failed to load donations:', error); }
	}

	async function loadConnectStatus() {
		try {
			connectStatus = await api('/api/giving/connect/status', { silent: true });
		} catch (error) { console.error('Failed to load connect status:', error); }
	}

	function renderChart() {
		if (!chartCanvas) return;
		const trend = stats.monthly_trend;
		const labels = trend.map(t => {
			const [y, m] = t.month.split('-');
			return new Date(parseInt(y), parseInt(m) - 1).toLocaleDateString('en-US', { month: 'short', year: '2-digit' });
		});
		const data = trend.map(t => t.total_cents / 100);

		if (trendChart) trendChart.destroy();
		trendChart = new Chart(chartCanvas, {
			type: 'bar',
			data: {
				labels,
				datasets: [{
					label: 'Monthly Giving',
					data,
					backgroundColor: 'rgba(74, 139, 140, 0.6)',
					borderColor: '#4A8B8C',
					borderWidth: 1,
					borderRadius: 4
				}]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: { display: false },
					tooltip: {
						callbacks: {
							label: (ctx) => formatCurrency(ctx.raw as number * 100)
						}
					}
				},
				scales: {
					y: {
						beginAtZero: true,
						ticks: {
							callback: (v) => '$' + (v as number).toLocaleString(),
							color: '#9CA3AF'
						},
						grid: { color: 'rgba(156, 163, 175, 0.1)' }
					},
					x: {
						ticks: { color: '#9CA3AF' },
						grid: { display: false }
					}
				}
			}
		});
	}

	function formatCurrency(cents: number): string {
		return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(cents / 100);
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
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
					<a href="/dashboard/giving/settings" class="px-4 py-2 bg-yellow-600 text-white rounded-lg hover:bg-yellow-700 transition font-medium whitespace-nowrap">
						Set up now →
					</a>
				</div>
			</div>
		{/if}

		<!-- Empty State -->
		{#if stats.total_all_time === 0 && recentDonations.length === 0}
			<div class="bg-surface rounded-lg shadow border border-custom p-12 text-center mb-8">
				<div class="text-6xl mb-4">💰</div>
				<h2 class="text-2xl font-bold text-primary mb-2">Welcome to Giving</h2>
				<p class="text-secondary mb-6 max-w-md mx-auto">
					Start tracking your church's donations. Record cash and check gifts, manage funds, and generate tax statements.
				</p>
				<div class="flex flex-col sm:flex-row gap-4 justify-center">
					<a href="/dashboard/giving/donations/new" class="px-6 py-3 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 transition font-medium">
						Record First Donation
					</a>
					<a href="/dashboard/giving/funds" class="px-6 py-3 border-2 border-[var(--teal)] text-[var(--teal)] rounded-lg hover:bg-[var(--teal)] hover:text-white transition font-medium">
						Set Up Funds
					</a>
				</div>
			</div>
		{:else}
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
					<p class="text-sm text-secondary mb-1">Average Gift</p>
					<p class="text-3xl font-bold text-primary">{formatCurrency(stats.average_donation)}</p>
				</div>
				<div class="bg-surface rounded-lg shadow p-6 border border-custom">
					<p class="text-sm text-secondary mb-1">Unique Donors</p>
					<p class="text-3xl font-bold text-primary">{stats.donor_count || 0}</p>
				</div>
			</div>

			<!-- Trend Chart -->
			{#if stats.monthly_trend.length > 0}
				<div class="bg-surface rounded-lg shadow border border-custom p-6 mb-8">
					<h2 class="text-xl font-semibold text-primary mb-4">Monthly Giving Trend</h2>
					<div style="height: 280px;">
						<canvas bind:this={chartCanvas}></canvas>
					</div>
				</div>
			{/if}

			<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
				<!-- Recent Donations -->
				<div class="bg-surface rounded-lg shadow border border-custom">
					<div class="p-6 border-b border-custom flex justify-between items-center">
						<h2 class="text-xl font-semibold text-primary">Recent Donations</h2>
						<a href="/dashboard/giving/donations" class="text-sm text-[var(--teal)] hover:underline">View All</a>
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
						{#if !stats.fund_breakdown || stats.fund_breakdown.length === 0}
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
											<div class="bg-[var(--teal)] h-2 rounded-full" style="width: {fund.percentage}%"></div>
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
		{/if}

		<!-- Quick Actions -->
		<div class="bg-surface rounded-lg shadow p-6 border border-custom">
			<h2 class="text-xl font-semibold text-primary mb-4">Quick Actions</h2>
			<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
				<a href="/dashboard/giving/donations/new" class="flex items-center justify-center px-4 py-3 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 transition">
					Record Donation
				</a>
				<a href="/dashboard/giving/funds" class="flex items-center justify-center px-4 py-3 bg-[var(--sage)] text-white rounded-lg hover:opacity-90 transition">
					Manage Funds
				</a>
				<a href="/dashboard/giving/statements" class="flex items-center justify-center px-4 py-3 bg-[var(--navy)] text-white rounded-lg hover:opacity-90 transition">
					Giving Statements
				</a>
				<a href="/dashboard/giving/settings" class="flex items-center justify-center px-4 py-3 border-2 border-[var(--teal)] text-[var(--teal)] rounded-lg hover:bg-[var(--teal)] hover:text-white transition">
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
	.warning-text { color: #78350F; }
	:global(.dark) .warning-text { color: #FDE68A; }
	.warning-subtext { color: #92400E; }
	:global(.dark) .warning-subtext { color: #FCD34D; }
</style>
