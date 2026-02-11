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
		const colors = { pending: 'status-pending', praying: 'status-praying', answered: 'status-answered', archived: 'status-archived' };
		return colors[s] || 'status-archived';
	}
</script>

<style>
	.status-pending {
		background-color: #FEF3C7;
		color: #92400E;
	}
	:global(.dark) .status-pending {
		background-color: #78350F;
		color: #FCD34D;
	}
	
	.status-praying {
		background-color: #DBEAFE;
		color: #1E40AF;
	}
	:global(.dark) .status-praying {
		background-color: #1E3A8A;
		color: #93C5FD;
	}
	
	.status-answered {
		background-color: #D1FAE5;
		color: #065F46;
	}
	:global(.dark) .status-answered {
		background-color: #064E3B;
		color: #6EE7B7;
	}
	
	.status-archived {
		background-color: #F3F4F6;
		color: #374151;
	}
	:global(.dark) .status-archived {
		background-color: #1F2937;
		color: #9CA3AF;
	}
</style>

<div class="p-6">
	<h1 class="text-3xl font-bold text-primary mb-6">Prayer Requests</h1>

	<div class="bg-surface rounded-lg shadow p-4 mb-6 border border-custom">
		<select bind:value={statusFilter} on:change={loadPrayerRequests} class="px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent">
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
				<div class="bg-surface rounded-lg shadow p-6 border border-custom">
					<div class="flex justify-between mb-2">
						<h3 class="font-bold text-lg text-primary">{req.name}</h3>
						<span class="px-2 py-1 text-xs rounded-full {getStatusColor(req.status)}">{req.status}</span>
					</div>
					<p class="text-primary mb-4 line-clamp-3">{req.request_text}</p>
					<p class="text-sm text-secondary mb-3">{formatDate(req.submitted_at)}</p>
					<select value={req.status} on:change={(e) => updateStatus(req.id, e.currentTarget.value)} class="w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent">
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
