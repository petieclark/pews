<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let settings = { license_number: '', auto_report: false, report_frequency: 'quarterly' };
	let stats = null;
	let report = null;
	let loading = false;
	let saving = false;
	let downloading = false;
	let message = '';
	let messageType = 'success';

	// Date range for report generation
	let startDate = '';
	let endDate = '';

	onMount(async () => {
		// Set default date range to current quarter
		const now = new Date();
		const quarterMonth = Math.floor(now.getMonth() / 3) * 3;
		const qStart = new Date(now.getFullYear(), quarterMonth, 1);
		const qEnd = new Date(now.getFullYear(), quarterMonth + 3, 0);
		startDate = qStart.toISOString().split('T')[0];
		endDate = qEnd.toISOString().split('T')[0];

		await Promise.all([loadSettings(), loadStats()]);
	});

	async function loadSettings() {
		try {
			settings = await api('/api/ccli/settings');
		} catch (e) {
			console.error('Failed to load CCLI settings:', e);
		}
	}

	async function loadStats() {
		try {
			stats = await api('/api/ccli/stats');
		} catch (e) {
			console.error('Failed to load CCLI stats:', e);
		}
	}

	async function saveSettings() {
		saving = true;
		message = '';
		try {
			settings = await api('/api/ccli/settings', {
				method: 'POST',
				body: JSON.stringify({
					license_number: settings.license_number,
					auto_report: settings.auto_report,
					report_frequency: settings.report_frequency
				})
			});
			message = 'Settings saved successfully';
			messageType = 'success';
		} catch (e) {
			message = 'Failed to save settings';
			messageType = 'error';
		}
		saving = false;
		setTimeout(() => message = '', 3000);
	}

	async function generateReport() {
		if (!startDate || !endDate) return;
		loading = true;
		report = null;
		try {
			report = await api(`/api/ccli/report?start=${startDate}&end=${endDate}`);
		} catch (e) {
			message = 'Failed to generate report';
			messageType = 'error';
			setTimeout(() => message = '', 3000);
		}
		loading = false;
	}

	async function downloadCSV() {
		if (!startDate || !endDate) return;
		downloading = true;
		try {
			const token = localStorage.getItem('token');
			const res = await fetch(`/api/ccli/report/download?start=${startDate}&end=${endDate}`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			const blob = await res.blob();
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `ccli-report-${startDate}-to-${endDate}.csv`;
			a.click();
			URL.revokeObjectURL(url);
		} catch (e) {
			message = 'Failed to download report';
			messageType = 'error';
			setTimeout(() => message = '', 3000);
		}
		downloading = false;
	}
</script>

<svelte:head>
	<title>CCLI Settings - Pews</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-primary">CCLI Reporting</h1>
		<p class="text-secondary mt-1">Manage your CCLI license and generate usage reports for copyright compliance.</p>
	</div>

	{#if message}
		<div class="p-4 rounded-lg {messageType === 'success' ? 'bg-green-50 text-green-800 dark:bg-green-900/20 dark:text-green-400' : 'bg-red-50 text-red-800 dark:bg-red-900/20 dark:text-red-400'}">
			{message}
		</div>
	{/if}

	<!-- Stats Cards -->
	{#if stats}
		<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
			<div class="bg-surface rounded-lg border border-custom p-5">
				<div class="text-sm text-secondary">Total Licensed Songs</div>
				<div class="text-3xl font-bold text-primary mt-1">{stats.total_licensed_songs}</div>
			</div>
			<div class="bg-surface rounded-lg border border-custom p-5">
				<div class="text-sm text-secondary">Songs Used This Quarter</div>
				<div class="text-3xl font-bold text-primary mt-1">{stats.songs_used_this_period}</div>
			</div>
			<div class="bg-surface rounded-lg border border-custom p-5">
				<div class="text-sm text-secondary">Reporting Status</div>
				<div class="mt-1">
					{#if stats.reporting_status === 'current'}
						<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-sm font-medium bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400">✓ Current</span>
					{:else if stats.reporting_status === 'overdue'}
						<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-sm font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400">⚠ Overdue</span>
					{:else}
						<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-sm font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300">Not Configured</span>
					{/if}
				</div>
			</div>
		</div>
	{/if}

	<!-- Settings -->
	<div class="bg-surface rounded-lg border border-custom p-6">
		<h2 class="text-lg font-semibold text-primary mb-4">License Settings</h2>
		<div class="space-y-4 max-w-md">
			<div>
				<label for="license" class="block text-sm font-medium text-secondary mb-1">CCLI License Number</label>
				<input
					id="license"
					type="text"
					bind:value={settings.license_number}
					placeholder="e.g., 1234567"
					class="w-full px-3 py-2 bg-[var(--bg)] border border-custom rounded-lg text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)]"
				/>
			</div>

			<div class="flex items-center space-x-3">
				<input
					id="auto-report"
					type="checkbox"
					bind:checked={settings.auto_report}
					class="w-4 h-4 rounded border-custom text-[var(--teal)] focus:ring-[var(--teal)]"
				/>
				<label for="auto-report" class="text-sm text-primary">Enable automatic reporting</label>
			</div>

			<div>
				<label for="frequency" class="block text-sm font-medium text-secondary mb-1">Report Frequency</label>
				<select
					id="frequency"
					bind:value={settings.report_frequency}
					class="w-full px-3 py-2 bg-[var(--bg)] border border-custom rounded-lg text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)]"
				>
					<option value="quarterly">Quarterly</option>
					<option value="semi-annual">Semi-Annual</option>
					<option value="annual">Annual</option>
				</select>
			</div>

			<button
				on:click={saveSettings}
				disabled={saving}
				class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 disabled:opacity-50 transition-opacity"
			>
				{saving ? 'Saving...' : 'Save Settings'}
			</button>
		</div>
	</div>

	<!-- Report Generator -->
	<div class="bg-surface rounded-lg border border-custom p-6">
		<h2 class="text-lg font-semibold text-primary mb-4">Generate Report</h2>
		<div class="flex flex-wrap items-end gap-4 mb-6">
			<div>
				<label for="start" class="block text-sm font-medium text-secondary mb-1">Start Date</label>
				<input
					id="start"
					type="date"
					bind:value={startDate}
					class="px-3 py-2 bg-[var(--bg)] border border-custom rounded-lg text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)]"
				/>
			</div>
			<div>
				<label for="end" class="block text-sm font-medium text-secondary mb-1">End Date</label>
				<input
					id="end"
					type="date"
					bind:value={endDate}
					class="px-3 py-2 bg-[var(--bg)] border border-custom rounded-lg text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)]"
				/>
			</div>
			<button
				on:click={generateReport}
				disabled={loading || !startDate || !endDate}
				class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 disabled:opacity-50 transition-opacity"
			>
				{loading ? 'Generating...' : 'Generate Report'}
			</button>
			{#if report}
				<button
					on:click={downloadCSV}
					disabled={downloading}
					class="px-4 py-2 border border-custom text-primary rounded-lg hover:bg-[var(--surface-hover)] disabled:opacity-50 transition-colors"
				>
					{downloading ? 'Downloading...' : '⬇ Download CSV'}
				</button>
			{/if}
		</div>

		{#if report}
			<div class="text-sm text-secondary mb-3">
				{report.total_songs} songs · {report.total_uses} total uses · {report.period}
			</div>
			<div class="overflow-x-auto">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-custom">
							<th class="text-left py-2 px-3 text-secondary font-medium">Song Title</th>
							<th class="text-left py-2 px-3 text-secondary font-medium">CCLI #</th>
							<th class="text-left py-2 px-3 text-secondary font-medium">Author</th>
							<th class="text-right py-2 px-3 text-secondary font-medium">Times Used</th>
							<th class="text-left py-2 px-3 text-secondary font-medium">Last Used</th>
						</tr>
					</thead>
					<tbody>
						{#each report.songs as song}
							<tr class="border-b border-custom hover:bg-[var(--surface-hover)]">
								<td class="py-2 px-3 text-primary">{song.title}</td>
								<td class="py-2 px-3 text-secondary font-mono text-xs">{song.ccli_number}</td>
								<td class="py-2 px-3 text-secondary">{song.artist}</td>
								<td class="py-2 px-3 text-right text-primary font-medium">{song.times_used}</td>
								<td class="py-2 px-3 text-secondary">{song.last_used}</td>
							</tr>
						{/each}
						{#if report.songs.length === 0}
							<tr>
								<td colspan="5" class="py-8 text-center text-secondary">No CCLI-licensed songs used during this period.</td>
							</tr>
						{/if}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>
