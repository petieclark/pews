<script lang="ts">
	import { onMount } from 'svelte';

	let prayerRequests = [];
	let loading = true;
	let statusFilter = '';
	let selectedRequest: any = null;
	let showModal = false;

	onMount(async () => {
		await loadPrayerRequests();
	});

	async function loadPrayerRequests() {
		loading = true;
		try {
			let url = '/api/prayer-requests?';
			if (statusFilter) url += `status=${statusFilter}`;

			const response = await fetch(url, {
				headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
			});
			if (response.ok) prayerRequests = await response.json();
		} catch (error) {
			console.error('Failed to load:', error);
		}
		loading = false;
	}

	async function updateStatus(id: string, status: string) {
		try {
			await fetch(`/api/prayer-requests/${id}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				},
				body: JSON.stringify({ status })
			});
			await loadPrayerRequests();
		} catch (error) {
			console.error('Failed to update:', error);
		}
	}

	function formatDate(d: string) {
		return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}

	function getStatusColor(s: string) {
		const colors = { 
			pending: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100', 
			praying: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100', 
			answered: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100', 
			archived: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200' 
		};
		return colors[s] || 'bg-gray-100 dark:bg-gray-700';
	}
</script>

<div class="p-6">
	<h1 class="text-3xl font-bold text-[var(--text-primary)] mb-6">Prayer Requests</h1>

	<div class="bg-surface border border-custom rounded-lg shadow p-4 mb-6">
		<select bind:value={statusFilter} on:change={loadPrayerRequests} class="px-3 py-2 border border-custom rounded-lg bg-surface text-primary">
			<option value="">All</option>
			<option value="pending">Pending</option>
			<option value="praying">Praying</option>
			<option value="answered">Answered</option>
			<option value="archived">Archived</option>
		</select>
	</div>

	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
		{#if loading}
			<div class="col-span-full text-center py-12"><div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--teal)] mx-auto"></div></div>
		{:else if prayerRequests.length === 0}
			<p class="col-span-full text-center text-secondary">No requests</p>
		{:else}
			{#each prayerRequests as req}
				<div class="bg-surface border border-custom rounded-lg shadow p-6">
					<div class="flex justify-between mb-2">
						<h3 class="font-bold text-lg text-[var(--text-primary)]">{req.name}</h3>
						<span class="px-2 py-1 text-xs rounded-full {getStatusColor(req.status)}">{req.status}</span>
					</div>
					<p class="text-secondary mb-4 line-clamp-3">{req.request_text}</p>
					<p class="text-sm text-secondary mb-3">{formatDate(req.submitted_at)}</p>
					{#if req.follower_count !== undefined}
						<p class="text-xs text-secondary mb-3">👥 {req.follower_count} {req.follower_count === 1 ? 'follower' : 'followers'}</p>
					{/if}
					<select value={req.status} on:change={(e) => updateStatus(req.id, e.currentTarget.value)} class="w-full px-3 py-2 border border-custom rounded-lg bg-surface text-primary">
						<option value="pending">Pending</option>
						<option value="praying">Praying</option>
						<option value="answered">Answered</option>
						<option value="archived">Archived</option>
					</select>
				</div>
			{/each}
		{/if}
	</div>
</div>
