<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let attendanceReport = null;
	let givingReport = null;
	let membershipReport = null;
	let groupsReport = null;
	let loading = true;
	let error = null;

	// Chart instances
	let attendanceChart = null;
	let givingChart = null;
	let membershipChart = null;
	let groupsChart = null;

	onMount(async () => {
		// Load Chart.js from CDN
		const script = document.createElement('script');
		script.src = 'https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js';
		script.onload = () => {
			loadReports();
		};
		document.head.appendChild(script);
	});

	async function loadReports() {
		try {
			loading = true;
			const [attendance, giving, membership, groups] = await Promise.all([
				api('/api/reports/attendance'),
				api('/api/reports/giving'),
				api('/api/reports/membership'),
				api('/api/reports/groups')
			]);

			attendanceReport = attendance;
			givingReport = giving;
			membershipReport = membership;
			groupsReport = groups;

			// Wait for DOM to update
			setTimeout(() => {
				renderCharts();
				loading = false;
			}, 100);
		} catch (e) {
			error = e.message;
			loading = false;
		}
	}

	function renderCharts() {
		// Attendance Line Chart
		if (attendanceReport && attendanceReport.weekly_data) {
			const ctx = document.getElementById('attendanceChart');
			if (ctx && window.Chart) {
				if (attendanceChart) attendanceChart.destroy();
				attendanceChart = new window.Chart(ctx, {
					type: 'line',
					data: {
						labels: attendanceReport.weekly_data.map(d => d.week_start_date),
						datasets: [{
							label: 'Weekly Attendance',
							data: attendanceReport.weekly_data.map(d => d.attendance_count),
							borderColor: 'rgb(59, 130, 246)',
							backgroundColor: 'rgba(59, 130, 246, 0.1)',
							tension: 0.3,
							fill: true
						}]
					},
					options: {
						responsive: true,
						maintainAspectRatio: false,
						plugins: {
							legend: { display: false }
						},
						scales: {
							y: { beginAtZero: true }
						}
					}
				});
			}
		}

		// Giving Bar Chart
		if (givingReport && givingReport.monthly_data) {
			const ctx = document.getElementById('givingChart');
			if (ctx && window.Chart) {
				if (givingChart) givingChart.destroy();
				givingChart = new window.Chart(ctx, {
					type: 'bar',
					data: {
						labels: givingReport.monthly_data.map(d => d.month),
						datasets: [{
							label: 'Monthly Giving',
							data: givingReport.monthly_data.map(d => d.total_amount),
							backgroundColor: 'rgba(34, 197, 94, 0.8)',
							borderColor: 'rgb(34, 197, 94)',
							borderWidth: 1
						}]
					},
					options: {
						responsive: true,
						maintainAspectRatio: false,
						plugins: {
							legend: { display: false }
						},
						scales: {
							y: { 
								beginAtZero: true,
								ticks: {
									callback: (value) => '$' + value.toLocaleString()
								}
							}
						}
					}
				});
			}
		}

		// Membership Line Chart
		if (membershipReport && membershipReport.monthly_data) {
			const ctx = document.getElementById('membershipChart');
			if (ctx && window.Chart) {
				if (membershipChart) membershipChart.destroy();
				membershipChart = new window.Chart(ctx, {
					type: 'line',
					data: {
						labels: membershipReport.monthly_data.map(d => d.month),
						datasets: [{
							label: 'Total Members',
							data: membershipReport.monthly_data.map(d => d.total_members),
							borderColor: 'rgb(168, 85, 247)',
							backgroundColor: 'rgba(168, 85, 247, 0.1)',
							tension: 0.3,
							fill: true
						}]
					},
					options: {
						responsive: true,
						maintainAspectRatio: false,
						plugins: {
							legend: { display: false }
						},
						scales: {
							y: { beginAtZero: true }
						}
					}
				});
			}
		}

		// Groups Pie Chart
		if (groupsReport) {
			const ctx = document.getElementById('groupsChart');
			if (ctx && window.Chart) {
				if (groupsChart) groupsChart.destroy();
				groupsChart = new window.Chart(ctx, {
					type: 'pie',
					data: {
						labels: ['In Groups', 'Not in Groups'],
						datasets: [{
							data: [groupsReport.members_in_groups, groupsReport.members_not_in_groups],
							backgroundColor: [
								'rgba(59, 130, 246, 0.8)',
								'rgba(203, 213, 225, 0.8)'
							],
							borderColor: [
								'rgb(59, 130, 246)',
								'rgb(203, 213, 225)'
							],
							borderWidth: 1
						}]
					},
					options: {
						responsive: true,
						maintainAspectRatio: false,
						plugins: {
							legend: { position: 'bottom' }
						}
					}
				});
			}
		}
	}
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-3xl font-bold text-[var(--text-primary)]">Reports & Analytics</h1>
		<p class="text-secondary mt-2">Church insights across all modules</p>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<p class="text-secondary">Loading reports...</p>
		</div>
	{:else if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
			Error loading reports: {error}
		</div>
	{:else}
		<!-- Attendance Section -->
		<div class="bg-surface rounded-lg shadow-sm border border-custom p-6">
			<h2 class="text-xl font-semibold text-[var(--text-primary)] mb-4">Attendance Trends</h2>
			<div class="grid grid-cols-3 gap-4 mb-6">
				<div class="bg-blue-50 dark:bg-blue-900/20 p-4 rounded">
					<p class="text-sm text-secondary">Average Attendance</p>
					<p class="text-2xl font-bold text-blue-600 dark:text-blue-400">
						{attendanceReport?.average_attendance?.toFixed(0) || 0}
					</p>
				</div>
				<div class="bg-green-50 dark:bg-green-900/20 p-4 rounded">
					<p class="text-sm text-secondary">Growth Rate</p>
					<p class="text-2xl font-bold" class:text-green-600={attendanceReport?.growth_percentage >= 0} class:text-red-600={attendanceReport?.growth_percentage < 0}>
						{attendanceReport?.growth_percentage?.toFixed(1) || 0}%
					</p>
				</div>
				<div class="bg-purple-50 dark:bg-purple-900/20 p-4 rounded">
					<p class="text-sm text-secondary">Weeks Tracked</p>
					<p class="text-2xl font-bold text-purple-600 dark:text-purple-400">
						{attendanceReport?.weekly_data?.length || 0}
					</p>
				</div>
			</div>
			<div class="h-64">
				<canvas id="attendanceChart"></canvas>
			</div>
		</div>

		<!-- Giving Section -->
		<div class="bg-surface rounded-lg shadow-sm border border-custom p-6">
			<h2 class="text-xl font-semibold text-[var(--text-primary)] mb-4">Giving Trends</h2>
			<div class="grid grid-cols-4 gap-4 mb-6">
				<div class="bg-green-50 dark:bg-green-900/20 p-4 rounded">
					<p class="text-sm text-secondary">Total YTD</p>
					<p class="text-2xl font-bold text-green-600 dark:text-green-400">
						${givingReport?.total_ytd_amount?.toLocaleString() || 0}
					</p>
				</div>
				<div class="bg-blue-50 dark:bg-blue-900/20 p-4 rounded">
					<p class="text-sm text-secondary">Average Gift</p>
					<p class="text-2xl font-bold text-blue-600 dark:text-blue-400">
						${givingReport?.average_gift_amount?.toFixed(2) || 0}
					</p>
				</div>
				<div class="bg-purple-50 dark:bg-purple-900/20 p-4 rounded">
					<p class="text-sm text-secondary">Unique Donors</p>
					<p class="text-2xl font-bold text-purple-600 dark:text-purple-400">
						{givingReport?.donor_count || 0}
					</p>
				</div>
				<div class="bg-yellow-50 dark:bg-yellow-900/20 p-4 rounded">
					<p class="text-sm text-secondary">Months Tracked</p>
					<p class="text-2xl font-bold text-yellow-600 dark:text-yellow-400">
						{givingReport?.monthly_data?.length || 0}
					</p>
				</div>
			</div>
			<div class="h-64">
				<canvas id="givingChart"></canvas>
			</div>
		</div>

		<!-- Membership Section -->
		<div class="bg-surface rounded-lg shadow-sm border border-custom p-6">
			<h2 class="text-xl font-semibold text-[var(--text-primary)] mb-4">Membership Growth</h2>
			<div class="grid grid-cols-4 gap-4 mb-6">
				<div class="bg-purple-50 dark:bg-purple-900/20 p-4 rounded">
					<p class="text-sm text-secondary">New This Month</p>
					<p class="text-2xl font-bold text-purple-600 dark:text-purple-400">
						{membershipReport?.new_members_this_month || 0}
					</p>
				</div>
				<div class="bg-blue-50 dark:bg-blue-900/20 p-4 rounded">
					<p class="text-sm text-secondary">New This Quarter</p>
					<p class="text-2xl font-bold text-blue-600 dark:text-blue-400">
						{membershipReport?.new_members_this_quarter || 0}
					</p>
				</div>
				<div class="bg-green-50 dark:bg-green-900/20 p-4 rounded">
					<p class="text-sm text-secondary">Active Members</p>
					<p class="text-2xl font-bold text-green-600 dark:text-green-400">
						{membershipReport?.active_members || 0}
					</p>
				</div>
				<div class="bg-gray-50 dark:bg-gray-900/20 p-4 rounded">
					<p class="text-sm text-secondary">Inactive Members</p>
					<p class="text-2xl font-bold text-gray-600 dark:text-gray-400">
						{membershipReport?.inactive_members || 0}
					</p>
				</div>
			</div>
			<div class="h-64">
				<canvas id="membershipChart"></canvas>
			</div>
		</div>

		<!-- Groups Section -->
		<div class="bg-surface rounded-lg shadow-sm border border-custom p-6">
			<h2 class="text-xl font-semibold text-[var(--text-primary)] mb-4">Group Participation</h2>
			<div class="grid grid-cols-3 gap-4 mb-6">
				<div class="bg-blue-50 dark:bg-blue-900/20 p-4 rounded">
					<p class="text-sm text-secondary">Participation Rate</p>
					<p class="text-2xl font-bold text-blue-600 dark:text-blue-400">
						{groupsReport?.participation_rate?.toFixed(1) || 0}%
					</p>
				</div>
				<div class="bg-green-50 dark:bg-green-900/20 p-4 rounded">
					<p class="text-sm text-secondary">Active Groups</p>
					<p class="text-2xl font-bold text-green-600 dark:text-green-400">
						{groupsReport?.active_groups || 0}
					</p>
				</div>
				<div class="bg-purple-50 dark:bg-purple-900/20 p-4 rounded">
					<p class="text-sm text-secondary">Avg Members/Group</p>
					<p class="text-2xl font-bold text-purple-600 dark:text-purple-400">
						{groupsReport?.average_members_per_group?.toFixed(1) || 0}
					</p>
				</div>
			</div>
			<div class="h-64 flex justify-center">
				<canvas id="groupsChart" class="max-w-md"></canvas>
			</div>
		</div>
	{/if}
</div>
