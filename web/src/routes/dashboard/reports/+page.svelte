<script>
	import { onMount, onDestroy } from 'svelte';
	import { api } from '$lib/api';

	let activeTab = 'attendance';
	const tabs = [
		{ id: 'attendance', label: 'Attendance', icon: '👥' },
		{ id: 'giving', label: 'Giving', icon: '💰' },
		{ id: 'growth', label: 'Growth', icon: '📈' },
		{ id: 'songs', label: 'Songs', icon: '🎵' },
		{ id: 'engagement', label: 'Engagement', icon: '⚡' }
	];

	let dateRange = '12m';
	const ranges = [
		{ value: '1m', label: 'This Month' },
		{ value: '3m', label: 'Last 3 Months' },
		{ value: '6m', label: 'Last 6 Months' },
		{ value: '12m', label: 'Last 12 Months' },
		{ value: '12w', label: 'Last 12 Weeks' }
	];

	let loading = false;
	let error = null;
	let data = {};
	let charts = {};
	let chartReady = false;

	onMount(() => {
		const script = document.createElement('script');
		script.src = 'https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js';
		script.onload = () => {
			chartReady = true;
		};
		document.head.appendChild(script);
	});

	onDestroy(() => {
		Object.values(charts).forEach(c => c?.destroy?.());
	});

	let prevTab = '';
	let prevRange = '';
	$: if (chartReady && (activeTab !== prevTab || dateRange !== prevRange)) {
		prevTab = activeTab;
		prevRange = dateRange;
		loadTab();
	}

	async function loadTab() {
		if (!chartReady) return;
		loading = true;
		error = null;
		try {
			const rangeParam = (activeTab === 'songs' || activeTab === 'engagement') ? '' : `?range=${dateRange}`;
			data[activeTab] = await api(`/api/reports/${activeTab}${rangeParam}`);
			setTimeout(() => renderCharts(), 50);
		} catch (e) {
			error = e.message;
		}
		loading = false;
	}

	function destroyCharts() {
		Object.values(charts).forEach(c => c?.destroy?.());
		charts = {};
	}

	function makeChart(id, config) {
		const el = document.getElementById(id);
		if (!el || !window.Chart) return;
		if (charts[id]) charts[id].destroy();
		const darkMode = document.documentElement.classList.contains('dark') || 
			getComputedStyle(document.documentElement).getPropertyValue('--bg').trim().startsWith('#1');
		const textColor = darkMode ? '#e2e8f0' : '#334155';
		const gridColor = darkMode ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.06)';

		// Apply theme to scales
		if (config.options?.scales) {
			Object.values(config.options.scales).forEach(s => {
				if (!s.ticks) s.ticks = {};
				s.ticks.color = textColor;
				if (!s.grid) s.grid = {};
				s.grid.color = gridColor;
			});
		}
		if (config.options?.plugins?.legend) {
			if (!config.options.plugins.legend.labels) config.options.plugins.legend.labels = {};
			config.options.plugins.legend.labels.color = textColor;
		}

		charts[id] = new window.Chart(el, config);
	}

	function renderCharts() {
		destroyCharts();
		const d = data[activeTab];
		if (!d) return;

		if (activeTab === 'attendance') renderAttendance(d);
		else if (activeTab === 'giving') renderGiving(d);
		else if (activeTab === 'growth') renderGrowth(d);
		else if (activeTab === 'songs') renderSongs(d);
		else if (activeTab === 'engagement') renderEngagement(d);
	}

	function baseOpts(maintainAspect = false) {
		return {
			responsive: true,
			maintainAspectRatio: maintainAspect,
			plugins: { legend: { display: false } },
			scales: { y: { beginAtZero: true } }
		};
	}

	function renderAttendance(d) {
		if (d.weekly_trend?.labels?.length) {
			makeChart('att-weekly', {
				type: 'line', data: d.weekly_trend,
				options: baseOpts()
			});
		}
		if (d.by_service_type?.labels?.length) {
			makeChart('att-type', {
				type: 'bar', data: d.by_service_type,
				options: baseOpts()
			});
		}
	}

	function renderGiving(d) {
		if (d.monthly_totals?.labels?.length) {
			makeChart('giv-monthly', {
				type: 'bar', data: d.monthly_totals,
				options: { ...baseOpts(), scales: { y: { beginAtZero: true, ticks: { callback: v => '$' + v.toLocaleString() } } } }
			});
		}
		if (d.by_fund?.labels?.length) {
			makeChart('giv-fund', {
				type: 'pie', data: d.by_fund,
				options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { position: 'bottom' } } }
			});
		}
		if (d.donor_trend?.labels?.length) {
			makeChart('giv-donors', {
				type: 'line', data: d.donor_trend,
				options: baseOpts()
			});
		}
	}

	function renderGrowth(d) {
		if (d.membership_growth?.labels?.length) {
			makeChart('grw-cumulative', {
				type: 'line', data: d.membership_growth,
				options: baseOpts()
			});
		}
		if (d.new_by_month?.labels?.length) {
			makeChart('grw-new', {
				type: 'bar', data: d.new_by_month,
				options: baseOpts()
			});
		}
	}

	function renderSongs(d) {
		if (d.top_songs?.labels?.length) {
			makeChart('song-top', {
				type: 'bar', data: d.top_songs,
				options: { ...baseOpts(), indexAxis: 'y' }
			});
		}
		if (d.by_key?.labels?.length) {
			makeChart('song-key', {
				type: 'pie', data: d.by_key,
				options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { position: 'bottom' } } }
			});
		}
	}

	function renderEngagement(d) {
		if (d.distribution?.labels?.length) {
			makeChart('eng-dist', {
				type: 'bar', data: d.distribution,
				options: baseOpts()
			});
		}
		if (d.trend?.labels?.length) {
			makeChart('eng-trend', {
				type: 'line', data: d.trend,
				options: baseOpts()
			});
		}
	}

	function printReport() {
		window.print();
	}
</script>

<div class="space-y-6 reports-page">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h1 class="text-3xl font-bold text-[var(--text-primary)]">Reports & Analytics</h1>
			<p class="text-secondary mt-1">Insights across your ministry</p>
		</div>
		<div class="flex items-center gap-3">
			<select
				bind:value={dateRange}
				class="bg-surface border border-custom rounded-lg px-3 py-2 text-sm text-[var(--text-primary)] focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent"
			>
				{#each ranges as r}
					<option value={r.value}>{r.label}</option>
				{/each}
			</select>
			<button
				on:click={printReport}
				class="hidden sm:flex items-center gap-2 px-4 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 text-sm font-medium print:hidden"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z"/></svg>
				Print Report
			</button>
		</div>
	</div>

	<!-- Tabs -->
	<div class="flex gap-1 bg-surface rounded-xl p-1 border border-custom overflow-x-auto print:hidden">
		{#each tabs as tab}
			<button
				on:click={() => activeTab = tab.id}
				class="flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium whitespace-nowrap transition-all
					{activeTab === tab.id 
						? 'bg-[var(--teal)] text-white shadow-sm' 
						: 'text-secondary hover:text-[var(--text-primary)] hover:bg-[var(--surface-hover)]'}"
			>
				<span>{tab.icon}</span>
				{tab.label}
			</button>
		{/each}
	</div>

	<!-- Loading / Error -->
	{#if loading}
		<div class="text-center py-16">
			<div class="inline-block w-8 h-8 border-4 border-[var(--teal)] border-t-transparent rounded-full animate-spin"></div>
			<p class="text-secondary mt-3">Loading report...</p>
		</div>
	{:else if error}
		<div class="bg-red-500/10 border border-red-500/30 text-red-400 px-4 py-3 rounded-lg">
			Error: {error}
		</div>
	{:else}

		<!-- ===== ATTENDANCE ===== -->
		{#if activeTab === 'attendance'}
			{@const d = data.attendance}
			{#if d}
				<!-- KPIs -->
				{#if d.kpis?.length}
					<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
						{#each d.kpis as kpi}
							<div class="report-card">
								<p class="text-sm text-secondary">{kpi.label}</p>
								<p class="text-2xl font-bold text-[var(--text-primary)] mt-1">{kpi.value}</p>
								{#if kpi.trend}
									<span class="text-xs mt-1 {kpi.trend === 'up' ? 'text-green-400' : kpi.trend === 'down' ? 'text-red-400' : 'text-secondary'}">
										{kpi.trend === 'up' ? '↑' : kpi.trend === 'down' ? '↓' : '→'} {kpi.change?.toFixed(1)}%
									</span>
								{/if}
							</div>
						{/each}
					</div>
				{/if}

				<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
					<div class="report-card">
						<h3 class="chart-title">Weekly Attendance Trend</h3>
						{#if d.weekly_trend?.labels?.length}
							<div class="chart-container"><canvas id="att-weekly"></canvas></div>
						{:else}
							<div class="empty-state">No attendance data yet. Check in people at services to see trends here.</div>
						{/if}
					</div>
					<div class="report-card">
						<h3 class="chart-title">Attendance by Service Type</h3>
						{#if d.by_service_type?.labels?.length}
							<div class="chart-container"><canvas id="att-type"></canvas></div>
						{:else}
							<div class="empty-state">No service type data yet. Set up service types and check people in.</div>
						{/if}
					</div>
				</div>
			{/if}

		<!-- ===== GIVING ===== -->
		{:else if activeTab === 'giving'}
			{@const d = data.giving}
			{#if d}
				{#if d.kpis?.length}
					<div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
						{#each d.kpis as kpi}
							<div class="report-card">
								<p class="text-sm text-secondary">{kpi.label}</p>
								<p class="text-2xl font-bold text-[var(--text-primary)] mt-1">{kpi.value}</p>
							</div>
						{/each}
					</div>
				{/if}

				<div class="report-card">
					<h3 class="chart-title">Monthly Giving</h3>
					{#if d.monthly_totals?.labels?.length}
						<div class="chart-container"><canvas id="giv-monthly"></canvas></div>
					{:else}
						<div class="empty-state">No giving data yet. Record donations to see monthly trends.</div>
					{/if}
				</div>

				<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
					<div class="report-card">
						<h3 class="chart-title">Giving by Fund</h3>
						{#if d.by_fund?.labels?.length}
							<div class="chart-container"><canvas id="giv-fund"></canvas></div>
						{:else}
							<div class="empty-state">No fund data yet. Create funds and record donations.</div>
						{/if}
					</div>
					<div class="report-card">
						<h3 class="chart-title">Donor Count Trend</h3>
						{#if d.donor_trend?.labels?.length}
							<div class="chart-container"><canvas id="giv-donors"></canvas></div>
						{:else}
							<div class="empty-state">No donor data yet.</div>
						{/if}
					</div>
				</div>
			{/if}

		<!-- ===== GROWTH ===== -->
		{:else if activeTab === 'growth'}
			{@const d = data.growth}
			{#if d}
				{#if d.kpis?.length}
					<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
						{#each d.kpis as kpi}
							<div class="report-card">
								<p class="text-sm text-secondary">{kpi.label}</p>
								<p class="text-2xl font-bold text-[var(--text-primary)] mt-1">{kpi.value}</p>
							</div>
						{/each}
					</div>
				{/if}

				<!-- Funnel -->
				{#if d.funnel?.length}
					<div class="report-card">
						<h3 class="chart-title">Membership Funnel</h3>
						<div class="flex flex-col items-center gap-2 py-4">
							{#each d.funnel as step, i}
								{@const width = 100 - (i * 20)}
								<div
									class="rounded-lg py-3 px-4 text-center text-white font-medium transition-all"
									style="width: {width}%; background: {['#1B3A4B','#4A8B8C','#8FBCB0'][i] || '#4A8B8C'}"
								>
									{step.label}: {step.count} ({step.pct.toFixed(0)}%)
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
					<div class="report-card">
						<h3 class="chart-title">Membership Growth</h3>
						{#if d.membership_growth?.labels?.length}
							<div class="chart-container"><canvas id="grw-cumulative"></canvas></div>
						{:else}
							<div class="empty-state">No membership data yet. Add people to track growth.</div>
						{/if}
					</div>
					<div class="report-card">
						<h3 class="chart-title">New Members by Month</h3>
						{#if d.new_by_month?.labels?.length}
							<div class="chart-container"><canvas id="grw-new"></canvas></div>
						{:else}
							<div class="empty-state">No new member data yet.</div>
						{/if}
					</div>
				</div>
			{/if}

		<!-- ===== SONGS ===== -->
		{:else if activeTab === 'songs'}
			{@const d = data.songs}
			{#if d}
				{#if d.kpis?.length}
					<div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
						{#each d.kpis as kpi}
							<div class="report-card">
								<p class="text-sm text-secondary">{kpi.label}</p>
								<p class="text-2xl font-bold text-[var(--text-primary)] mt-1">{kpi.value}</p>
							</div>
						{/each}
					</div>
				{/if}

				<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
					<div class="report-card">
						<h3 class="chart-title">Top 20 Songs</h3>
						{#if d.top_songs?.labels?.length}
							<div class="chart-container-tall"><canvas id="song-top"></canvas></div>
						{:else}
							<div class="empty-state">No song usage data. Add songs to services to see which are most popular.</div>
						{/if}
					</div>
					<div class="report-card">
						<h3 class="chart-title">Songs by Key</h3>
						{#if d.by_key?.labels?.length}
							<div class="chart-container"><canvas id="song-key"></canvas></div>
						{:else}
							<div class="empty-state">No key data. Set keys on songs to see distribution.</div>
						{/if}
					</div>
				</div>

				<!-- Unused songs table -->
				{#if d.unused_songs?.length}
					<div class="report-card">
						<h3 class="chart-title">Songs Not Used in 6+ Months</h3>
						<div class="overflow-x-auto">
							<table class="w-full text-sm">
								<thead>
									<tr class="border-b border-custom">
										<th class="text-left py-2 px-3 text-secondary font-medium">Title</th>
										<th class="text-left py-2 px-3 text-secondary font-medium">Artist</th>
										<th class="text-left py-2 px-3 text-secondary font-medium">Last Used</th>
									</tr>
								</thead>
								<tbody>
									{#each d.unused_songs as song}
										<tr class="border-b border-custom/50 hover:bg-[var(--surface-hover)]">
											<td class="py-2 px-3 text-[var(--text-primary)]">{song.title}</td>
											<td class="py-2 px-3 text-secondary">{song.artist || '—'}</td>
											<td class="py-2 px-3 text-secondary">{song.last_used === 'Never' ? 'Never' : new Date(song.last_used).toLocaleDateString()}</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</div>
				{/if}
			{/if}

		<!-- ===== ENGAGEMENT ===== -->
		{:else if activeTab === 'engagement'}
			{@const d = data.engagement}
			{#if d}
				{#if d.kpis?.length}
					<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
						{#each d.kpis as kpi}
							<div class="report-card">
								<p class="text-sm text-secondary">{kpi.label}</p>
								<p class="text-2xl font-bold text-[var(--text-primary)] mt-1">{kpi.value}</p>
							</div>
						{/each}
					</div>
				{/if}

				<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
					<div class="report-card">
						<h3 class="chart-title">Engagement Distribution</h3>
						{#if d.distribution?.labels?.length}
							<div class="chart-container"><canvas id="eng-dist"></canvas></div>
						{:else}
							<div class="empty-state">No engagement scores yet. Run engagement calculation to see distribution.</div>
						{/if}
					</div>
					<div class="report-card">
						<h3 class="chart-title">Engagement Trend</h3>
						{#if d.trend?.labels?.length}
							<div class="chart-container"><canvas id="eng-trend"></canvas></div>
						{:else}
							<div class="empty-state">Not enough data for trend. Engagement scores update over time.</div>
						{/if}
					</div>
				</div>
			{/if}
		{/if}
	{/if}
</div>

<style>
	.report-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: 0.75rem;
		padding: 1.5rem;
		box-shadow: 0 1px 3px rgba(0,0,0,0.08);
	}
	.chart-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}
	.chart-container {
		height: 280px;
		position: relative;
	}
	.chart-container-tall {
		height: 500px;
		position: relative;
	}
	.empty-state {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 200px;
		color: var(--text-secondary, #94a3b8);
		font-size: 0.875rem;
		text-align: center;
		padding: 1rem;
		background: var(--surface-hover, rgba(0,0,0,0.03));
		border-radius: 0.5rem;
	}

	/* Print styles */
	@media print {
		.reports-page {
			color: #000 !important;
		}
		.report-card {
			break-inside: avoid;
			border: 1px solid #ddd;
			box-shadow: none;
		}
		:global(nav), :global(.print\\:hidden) {
			display: none !important;
		}
	}
</style>
