<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	let activeTab = 'logs'; // 'logs' or 'security'
	
	// Audit logs state
	let logs = [];
	let totalLogs = 0;
	let currentPage = 1;
	let pageSize = 50;
	let totalPages = 0;
	let loading = false;
	
	// Filters
	let filterUserId = '';
	let filterAction = '';
	let filterEntityType = '';
	
	// Security dashboard state
	let securityData = null;
	let loadingSecurity = false;

	const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

	onMount(() => {
		fetchLogs();
	});

	async function fetchLogs() {
		loading = true;
		const token = localStorage.getItem('token');
		if (!token) {
			goto('/login');
			return;
		}

		const params = new URLSearchParams({
			page: currentPage.toString(),
			page_size: pageSize.toString()
		});

		if (filterUserId) params.append('user_id', filterUserId);
		if (filterAction) params.append('action', filterAction);
		if (filterEntityType) params.append('entity_type', filterEntityType);

		try {
			const res = await fetch(`${API_BASE}/api/audit/logs?${params}`, {
				headers: { Authorization: `Bearer ${token}` }
			});

			if (!res.ok) throw new Error('Failed to fetch logs');

			const data = await res.json();
			logs = data.logs || [];
			totalLogs = data.total;
			totalPages = data.total_pages;
		} catch (err) {
			console.error('Failed to fetch audit logs:', err);
			alert('Failed to load audit logs');
		} finally {
			loading = false;
		}
	}

	async function fetchSecurityDashboard() {
		loadingSecurity = true;
		const token = localStorage.getItem('token');
		if (!token) {
			goto('/login');
			return;
		}

		try {
			const res = await fetch(`${API_BASE}/api/audit/security`, {
				headers: { Authorization: `Bearer ${token}` }
			});

			if (!res.ok) throw new Error('Failed to fetch security dashboard');

			securityData = await res.json();
		} catch (err) {
			console.error('Failed to fetch security dashboard:', err);
			alert('Failed to load security dashboard');
		} finally {
			loadingSecurity = false;
		}
	}

	function handleTabChange(tab) {
		activeTab = tab;
		if (tab === 'security' && !securityData) {
			fetchSecurityDashboard();
		}
	}

	function applyFilters() {
		currentPage = 1;
		fetchLogs();
	}

	function clearFilters() {
		filterUserId = '';
		filterAction = '';
		filterEntityType = '';
		currentPage = 1;
		fetchLogs();
	}

	function changePage(page) {
		currentPage = page;
		fetchLogs();
	}

	async function exportLogs() {
		const token = localStorage.getItem('token');
		if (!token) return;

		try {
			const res = await fetch(`${API_BASE}/api/audit/export`, {
				headers: { Authorization: `Bearer ${token}` }
			});

			if (!res.ok) throw new Error('Failed to export logs');

			const blob = await res.blob();
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `audit_logs_${new Date().toISOString().split('T')[0]}.csv`;
			document.body.appendChild(a);
			a.click();
			a.remove();
		} catch (err) {
			console.error('Failed to export logs:', err);
			alert('Failed to export audit logs');
		}
	}

	function formatTimestamp(ts) {
		return new Date(ts).toLocaleString();
	}

	function formatJSON(json) {
		if (!json) return 'N/A';
		try {
			return JSON.stringify(JSON.parse(json), null, 2);
		} catch {
			return json;
		}
	}
</script>

<div class="container mx-auto p-6 max-w-7xl">
	<div class="mb-6">
		<h1 class="text-3xl font-bold text-gray-900">Audit Log & Security</h1>
		<p class="text-gray-600 mt-2">Monitor all admin actions and security events</p>
	</div>

	<!-- Tabs -->
	<div class="border-b border-gray-200 mb-6">
		<nav class="-mb-px flex space-x-8">
			<button
				class="py-2 px-1 border-b-2 font-medium text-sm {activeTab === 'logs'
					? 'border-blue-500 text-blue-600'
					: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
				on:click={() => handleTabChange('logs')}
			>
				Audit Logs
			</button>
			<button
				class="py-2 px-1 border-b-2 font-medium text-sm {activeTab === 'security'
					? 'border-blue-500 text-blue-600'
					: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
				on:click={() => handleTabChange('security')}
			>
				Security Dashboard
			</button>
		</nav>
	</div>

	{#if activeTab === 'logs'}
		<!-- Filters -->
		<div class="bg-white rounded-lg shadow p-6 mb-6">
			<h2 class="text-lg font-semibold mb-4">Filters</h2>
			<div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">User ID</label>
					<input
						type="text"
						bind:value={filterUserId}
						class="w-full border border-gray-300 rounded px-3 py-2"
						placeholder="Filter by user ID"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Action</label>
					<select bind:value={filterAction} class="w-full border border-gray-300 rounded px-3 py-2">
						<option value="">All actions</option>
						<option value="auth.login">Login</option>
						<option value="auth.logout">Logout</option>
						<option value="create">Create</option>
						<option value="update">Update</option>
						<option value="delete">Delete</option>
						<option value="settings.change">Settings Change</option>
						<option value="module.enable">Module Enable</option>
						<option value="module.disable">Module Disable</option>
					</select>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Entity Type</label>
					<input
						type="text"
						bind:value={filterEntityType}
						class="w-full border border-gray-300 rounded px-3 py-2"
						placeholder="e.g., people, groups"
					/>
				</div>
			</div>
			<div class="flex gap-2">
				<button
					on:click={applyFilters}
					class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
				>
					Apply Filters
				</button>
				<button
					on:click={clearFilters}
					class="bg-gray-300 text-gray-700 px-4 py-2 rounded hover:bg-gray-400"
				>
					Clear
				</button>
				<button
					on:click={exportLogs}
					class="ml-auto bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
				>
					Export CSV
				</button>
			</div>
		</div>

		<!-- Logs Table -->
		<div class="bg-white rounded-lg shadow overflow-hidden">
			{#if loading}
				<div class="p-12 text-center text-gray-500">Loading...</div>
			{:else if logs.length === 0}
				<div class="p-12 text-center text-gray-500">No audit logs found</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Timestamp</th>
								<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Action</th>
								<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">User</th>
								<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Entity</th>
								<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">IP</th>
								<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Details</th>
							</tr>
						</thead>
						<tbody class="bg-white divide-y divide-gray-200">
							{#each logs as log}
								<tr class="hover:bg-gray-50">
									<td class="px-4 py-3 text-sm text-gray-900 whitespace-nowrap">
										{formatTimestamp(log.timestamp)}
									</td>
									<td class="px-4 py-3 text-sm">
										<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
											{log.action}
										</span>
									</td>
									<td class="px-4 py-3 text-sm text-gray-500 font-mono text-xs">
										{log.user_id || 'System'}
									</td>
									<td class="px-4 py-3 text-sm text-gray-500">
										{#if log.entity_type}
											<div class="font-mono text-xs">
												{log.entity_type}
												{#if log.entity_id}
													<br />
													<span class="text-gray-400">{log.entity_id.substring(0, 8)}...</span>
												{/if}
											</div>
										{:else}
											N/A
										{/if}
									</td>
									<td class="px-4 py-3 text-sm text-gray-500 font-mono text-xs">
										{log.ip_address || 'N/A'}
									</td>
									<td class="px-4 py-3 text-sm">
										<button
											class="text-blue-600 hover:text-blue-800"
											on:click={() => {
												alert(`Old: ${formatJSON(log.old_value)}\n\nNew: ${formatJSON(log.new_value)}`);
											}}
										>
											View
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>

				<!-- Pagination -->
				<div class="bg-gray-50 px-4 py-3 flex items-center justify-between border-t border-gray-200">
					<div class="text-sm text-gray-700">
						Showing {(currentPage - 1) * pageSize + 1} to {Math.min(
							currentPage * pageSize,
							totalLogs
						)} of {totalLogs} results
					</div>
					<div class="flex gap-2">
						<button
							on:click={() => changePage(currentPage - 1)}
							disabled={currentPage === 1}
							class="px-3 py-1 bg-white border border-gray-300 rounded disabled:opacity-50 hover:bg-gray-50"
						>
							Previous
						</button>
						<span class="px-3 py-1">Page {currentPage} of {totalPages}</span>
						<button
							on:click={() => changePage(currentPage + 1)}
							disabled={currentPage === totalPages}
							class="px-3 py-1 bg-white border border-gray-300 rounded disabled:opacity-50 hover:bg-gray-50"
						>
							Next
						</button>
					</div>
				</div>
			{/if}
		</div>
	{:else if activeTab === 'security'}
		<!-- Security Dashboard -->
		{#if loadingSecurity}
			<div class="p-12 text-center text-gray-500">Loading security data...</div>
		{:else if securityData}
			<!-- Stats Cards -->
			<div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
				<div class="bg-white rounded-lg shadow p-6">
					<h3 class="text-sm font-medium text-gray-500 uppercase">Active Sessions</h3>
					<p class="text-3xl font-bold text-gray-900 mt-2">{securityData.active_sessions_count}</p>
				</div>
				<div class="bg-white rounded-lg shadow p-6">
					<h3 class="text-sm font-medium text-gray-500 uppercase">Failed Logins (24h)</h3>
					<p class="text-3xl font-bold text-red-600 mt-2">{securityData.failed_logins_last_24h}</p>
				</div>
				<div class="bg-white rounded-lg shadow p-6">
					<h3 class="text-sm font-medium text-gray-500 uppercase">Users Without 2FA</h3>
					<p class="text-3xl font-bold text-yellow-600 mt-2">{securityData.users_without_2fa}</p>
				</div>
			</div>

			<!-- Recent Failed Logins -->
			<div class="bg-white rounded-lg shadow p-6 mb-6">
				<h2 class="text-lg font-semibold mb-4">Recent Failed Login Attempts</h2>
				{#if securityData.recent_failed_logins?.length > 0}
					<div class="overflow-x-auto">
						<table class="min-w-full">
							<thead class="bg-gray-50">
								<tr>
									<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Email</th>
									<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">IP Address</th>
									<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Time</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-gray-200">
								{#each securityData.recent_failed_logins as attempt}
									<tr>
										<td class="px-4 py-2 text-sm">{attempt.email}</td>
										<td class="px-4 py-2 text-sm font-mono">{attempt.ip_address || 'N/A'}</td>
										<td class="px-4 py-2 text-sm">{formatTimestamp(attempt.attempted_at)}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{:else}
					<p class="text-gray-500">No recent failed login attempts</p>
				{/if}
			</div>

			<!-- Unusual Activities -->
			<div class="bg-white rounded-lg shadow p-6 mb-6">
				<h2 class="text-lg font-semibold mb-4">Unusual Activity</h2>
				{#if securityData.unusual_activities?.length > 0}
					<div class="space-y-3">
						{#each securityData.unusual_activities as activity}
							<div class="border-l-4 border-yellow-500 bg-yellow-50 p-4">
								<div class="flex justify-between items-start">
									<div>
										<p class="font-semibold text-gray-900">{activity.email}</p>
										<p class="text-sm text-gray-600">{activity.reason}</p>
										<p class="text-xs text-gray-500 mt-1 font-mono">IP: {activity.ip_address}</p>
									</div>
									<span class="text-xs text-gray-500">{formatTimestamp(activity.timestamp)}</span>
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<p class="text-gray-500">No unusual activity detected</p>
				{/if}
			</div>

			<!-- Password Changes -->
			<div class="bg-white rounded-lg shadow p-6">
				<h2 class="text-lg font-semibold mb-4">User Password Status</h2>
				<div class="overflow-x-auto">
					<table class="min-w-full">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Email</th>
								<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Last Changed</th>
								<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Days Ago</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-200">
							{#each securityData.user_password_changes || [] as user}
								<tr class="{user.days_since_change > 90 ? 'bg-red-50' : ''}">
									<td class="px-4 py-2 text-sm">{user.email}</td>
									<td class="px-4 py-2 text-sm">
										{user.password_changed_at ? formatTimestamp(user.password_changed_at) : 'Never'}
									</td>
									<td class="px-4 py-2 text-sm">
										{#if user.days_since_change !== null}
											<span class="{user.days_since_change > 90 ? 'text-red-600 font-semibold' : ''}">
												{user.days_since_change} days
											</span>
										{:else}
											N/A
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}
	{/if}
</div>
