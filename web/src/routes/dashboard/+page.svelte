<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let kpis = null;
	let atRisk = [];
	let loading = true;
	let error = '';

	onMount(async () => {
		try {
			// Fetch dashboard KPIs
			kpis = await api('/api/dashboard/kpis');
			
			// Fetch at-risk people
			atRisk = await api('/api/engagement/at-risk');
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	});

	function formatCurrency(cents) {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(cents / 100);
	}

	function formatPercent(value) {
		return (value >= 0 ? '+' : '') + value.toFixed(1) + '%';
	}
</script>

<div>
	<h1 class="text-3xl font-bold text-primary mb-6">Dashboard</h1>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else if error}
		<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg">
			{error}
		</div>
	{:else if kpis}
		<!-- KPI Cards -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
			<!-- Total Active Members -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-secondary text-sm uppercase tracking-wide">Active Members</p>
						<p class="text-3xl font-bold text-primary mt-2">{kpis.total_active_members}</p>
					</div>
					<div class="bg-[var(--teal)] bg-opacity-10 rounded-full p-3">
						<svg class="w-6 h-6 text-[var(--teal)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
						</svg>
					</div>
				</div>
			</div>

			<!-- Average Attendance -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-secondary text-sm uppercase tracking-wide">Avg Attendance (4wk)</p>
						<p class="text-3xl font-bold text-primary mt-2">{Math.round(kpis.average_attendance)}</p>
						{#if kpis.attendance_trend && kpis.attendance_trend.length > 0}
							<p class="text-xs text-secondary mt-1">
								<svg class="inline w-16 h-4" viewBox="0 0 100 20">
									{#each kpis.attendance_trend as value, i}
										<circle 
											cx={i * (100 / (kpis.attendance_trend.length - 1))} 
											cy={20 - (value / Math.max(...kpis.attendance_trend) * 15)} 
											r="1.5" 
											fill="var(--teal)" 
										/>
										{#if i > 0}
											<line 
												x1={(i - 1) * (100 / (kpis.attendance_trend.length - 1))} 
												y1={20 - (kpis.attendance_trend[i - 1] / Math.max(...kpis.attendance_trend) * 15)} 
												x2={i * (100 / (kpis.attendance_trend.length - 1))} 
												y2={20 - (value / Math.max(...kpis.attendance_trend) * 15)} 
												stroke="var(--teal)" 
												stroke-width="2" 
											/>
										{/if}
									{/each}
								</svg>
							</p>
						{/if}
					</div>
					<div class="bg-[var(--sage)] bg-opacity-10 rounded-full p-3">
						<svg class="w-6 h-6 text-[var(--sage)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
						</svg>
					</div>
				</div>
			</div>

			<!-- Giving This Month -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-secondary text-sm uppercase tracking-wide">Giving This Month</p>
						<p class="text-3xl font-bold text-primary mt-2">{formatCurrency(kpis.giving_this_month_cents)}</p>
						<p class="text-xs mt-1" class:text-green-600={kpis.giving_percent_change >= 0} class:text-red-600={kpis.giving_percent_change < 0}>
							{formatPercent(kpis.giving_percent_change)} vs last month
						</p>
					</div>
					<div class="bg-green-100 rounded-full p-3">
						<svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
				</div>
			</div>

			<!-- New Visitors -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-secondary text-sm uppercase tracking-wide">New Visitors</p>
						<p class="text-3xl font-bold text-primary mt-2">{kpis.new_visitors_this_month}</p>
						<p class="text-xs text-secondary mt-1">This month</p>
					</div>
					<div class="bg-blue-100 rounded-full p-3">
						<svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
						</svg>
					</div>
				</div>
			</div>
		</div>

		<!-- Engagement & At-Risk Widgets -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
			<!-- Engagement Distribution -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<h2 class="text-xl font-semibold text-primary mb-4">Engagement Distribution</h2>
				<div class="flex items-center justify-center mb-4">
					<svg viewBox="0 0 100 100" class="w-48 h-48">
						{#if kpis.engagement_distribution}
							{@const total = kpis.engagement_distribution.high + kpis.engagement_distribution.medium + kpis.engagement_distribution.low + kpis.engagement_distribution.inactive}
							{@const high = (kpis.engagement_distribution.high / total) * 100}
							{@const medium = (kpis.engagement_distribution.medium / total) * 100}
							{@const low = (kpis.engagement_distribution.low / total) * 100}
							{@const inactive = (kpis.engagement_distribution.inactive / total) * 100}
							
							<!-- High (75-100) - Teal -->
							<circle cx="50" cy="50" r="40" fill="none" stroke="var(--teal)" stroke-width="20" 
								stroke-dasharray="{high * 2.51} 251.2" 
								transform="rotate(-90 50 50)" />
							
							<!-- Medium (50-74) - Sage -->
							<circle cx="50" cy="50" r="40" fill="none" stroke="var(--sage)" stroke-width="20" 
								stroke-dasharray="{medium * 2.51} 251.2" 
								stroke-dashoffset="{-high * 2.51}"
								transform="rotate(-90 50 50)" />
							
							<!-- Low (25-49) - Orange -->
							<circle cx="50" cy="50" r="40" fill="none" stroke="#f59e0b" stroke-width="20" 
								stroke-dasharray="{low * 2.51} 251.2" 
								stroke-dashoffset="{-(high + medium) * 2.51}"
								transform="rotate(-90 50 50)" />
							
							<!-- Inactive (0-24) - Red -->
							<circle cx="50" cy="50" r="40" fill="none" stroke="#ef4444" stroke-width="20" 
								stroke-dasharray="{inactive * 2.51} 251.2" 
								stroke-dashoffset="{-(high + medium + low) * 2.51}"
								transform="rotate(-90 50 50)" />
						{/if}
					</svg>
				</div>
				{#if kpis.engagement_distribution}
					<div class="grid grid-cols-2 gap-3">
						<div class="flex items-center">
							<div class="w-3 h-3 rounded-full bg-[var(--teal)] mr-2"></div>
							<span class="text-sm text-secondary">High (75+): {kpis.engagement_distribution.high}</span>
						</div>
						<div class="flex items-center">
							<div class="w-3 h-3 rounded-full bg-[var(--sage)] mr-2"></div>
							<span class="text-sm text-secondary">Medium (50-74): {kpis.engagement_distribution.medium}</span>
						</div>
						<div class="flex items-center">
							<div class="w-3 h-3 rounded-full bg-[#f59e0b] mr-2"></div>
							<span class="text-sm text-secondary">Low (25-49): {kpis.engagement_distribution.low}</span>
						</div>
						<div class="flex items-center">
							<div class="w-3 h-3 rounded-full bg-[#ef4444] mr-2"></div>
							<span class="text-sm text-secondary">Inactive (0-24): {kpis.engagement_distribution.inactive}</span>
						</div>
					</div>
				{/if}
			</div>

			<!-- At-Risk Members -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-xl font-semibold text-primary">At-Risk Members</h2>
					<span class="bg-red-100 text-red-700 text-sm px-3 py-1 rounded-full font-medium">
						{atRisk.length} people
					</span>
				</div>
				<p class="text-secondary text-sm mb-4">Engagement declined >20% in last 30 days</p>
				{#if atRisk.length === 0}
					<div class="text-center py-8 text-secondary">
						<svg class="w-12 h-12 mx-auto mb-2 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<p>No at-risk members detected</p>
					</div>
				{:else}
					<div class="space-y-2 max-h-64 overflow-y-auto">
						{#each atRisk.slice(0, 5) as person}
							<div class="flex items-center justify-between p-3 bg-[var(--surface-hover)] rounded-lg">
								<div class="flex-1">
									<p class="font-medium text-primary">{person.person_name}</p>
									<p class="text-xs text-secondary">{person.person_email || 'No email'}</p>
								</div>
								<div class="text-right">
									<p class="text-sm font-semibold text-red-600">{person.percent_change.toFixed(1)}%</p>
									<p class="text-xs text-secondary">{person.current_score} → {person.previous_score}</p>
								</div>
							</div>
						{/each}
						{#if atRisk.length > 5}
							<a href="/dashboard/engagement/at-risk" class="block text-center text-sm text-[var(--teal)] hover:underline mt-2">
								View all {atRisk.length} at-risk members →
							</a>
						{/if}
					</div>
				{/if}
			</div>
		</div>

		<!-- Action Items Widget -->
		<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
			<h2 class="text-xl font-semibold text-primary mb-4">Action Items</h2>
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
				<div class="p-4 bg-[var(--surface-hover)] rounded-lg">
					<p class="text-2xl font-bold text-primary">0</p>
					<p class="text-sm text-secondary">Overdue Follow-ups</p>
				</div>
				<div class="p-4 bg-[var(--surface-hover)] rounded-lg">
					<p class="text-2xl font-bold text-primary">0</p>
					<p class="text-sm text-secondary">Unread Prayer Requests</p>
				</div>
				<div class="p-4 bg-[var(--surface-hover)] rounded-lg">
					<p class="text-2xl font-bold text-primary">0</p>
					<p class="text-sm text-secondary">Pending Volunteers</p>
				</div>
				<div class="p-4 bg-[var(--surface-hover)] rounded-lg">
					<p class="text-2xl font-bold text-primary">{atRisk.length}</p>
					<p class="text-sm text-secondary">At-Risk Members</p>
				</div>
			</div>
		</div>

		<!-- Quick Links -->
		<div class="mt-8 grid grid-cols-1 md:grid-cols-3 gap-4">
			<a href="/dashboard/people" class="bg-surface border border-custom rounded-lg p-6 hover:bg-[var(--surface-hover)] transition-colors">
				<h3 class="text-lg font-semibold text-primary mb-2">People</h3>
				<p class="text-sm text-secondary">Manage members and view engagement scores</p>
			</a>
			<a href="/dashboard/giving" class="bg-surface border border-custom rounded-lg p-6 hover:bg-[var(--surface-hover)] transition-colors">
				<h3 class="text-lg font-semibold text-primary mb-2">Giving</h3>
				<p class="text-sm text-secondary">View donations and giving trends</p>
			</a>
			<a href="/dashboard/checkins" class="bg-surface border border-custom rounded-lg p-6 hover:bg-[var(--surface-hover)] transition-colors">
				<h3 class="text-lg font-semibold text-primary mb-2">Check-ins</h3>
				<p class="text-sm text-secondary">Track attendance and check-in history</p>
			</a>
		</div>
	{/if}
</div>
