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
		const colors = { pending: 'bg-yellow-100 text-yellow-800', praying: 'bg-blue-100 text-blue-800', answered: 'bg-green-100 text-green-800', archived: 'bg-gray-100 text-gray-800' };
		return colors[s] || 'bg-gray-100';
	}
</script>

<div class="p-6">
	<h1 class="text-3xl font-bold text-[#1B3A4B] mb-6">Prayer Requests</h1>

	<div class="bg-white rounded-lg shadow p-4 mb-6">
		<select bind:value={statusFilter} on:change={loadPrayerRequests} class="px-3 py-2 border rounded-lg">
			<option value="">All</option>
			<option value="pending">Pending</option>
			<option value="praying">Praying</option>
			<option value="answered">Answered</option>
			<option value="archived">Archived</option>
		</select>
	</div>

	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
		{#if loading}
			<div class="col-span-full text-center py-12"><div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[#4A8B8C] mx-auto"></div></div>
		{:else if prayerRequests.length === 0}
			<p class="col-span-full text-center text-gray-500">No requests</p>
		{:else}
			{#each prayerRequests as req}
				<div class="bg-white rounded-lg shadow p-6">
					<div class="flex justify-between mb-2">
						<h3 class="font-bold text-lg">{req.name}</h3>
						<span class="px-2 py-1 text-xs rounded-full {getStatusColor(req.status)}">{req.status}</span>
					</div>
					<p class="text-gray-700 mb-4 line-clamp-3">{req.request_text}</p>
					<p class="text-sm text-gray-600 mb-3">{formatDate(req.submitted_at)}</p>
					<select value={req.status} on:change={(e) => updateStatus(req.id, e.currentTarget.value)} class="w-full px-3 py-2 border rounded-lg">
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
