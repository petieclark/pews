<script lang="ts">
	import { onMount } from 'svelte';

	let prayerRequests = [];
	let loading = true;

	onMount(async () => {
		try {
			const response = await fetch('/api/prayer-requests/public?limit=50', {
				headers: { 'X-Tenant-ID': getTenantId() }
			});
			if (response.ok) prayerRequests = await response.json();
		} catch (error) {
			console.error('Failed to load:', error);
		}
		loading = false;
	});

	function getTenantId() {
		return localStorage.getItem('tenant_id') || window.location.hostname.split('.')[0];
	}

	function formatDate(d: string) {
		const days = Math.floor((Date.now() - new Date(d).getTime()) / 86400000);
		if (days === 0) return 'Today';
		if (days === 1) return 'Yesterday';
		if (days < 7) return `${days} days ago`;
		return new Date(d).toLocaleDateString();
	}
</script>

<svelte:head><title>Prayer Wall</title></svelte:head>

<div class="min-h-screen bg-gradient-to-b from-[#1B3A4B] to-[#4A8B8C]">
	<div class="text-center py-16 px-4">
		<h1 class="text-5xl font-bold text-white mb-4">Prayer Wall</h1>
		<p class="text-xl text-white/90 mb-8">Join us in praying for our community</p>
		<a href="/prayer/submit" class="inline-block px-8 py-4 bg-white text-[#1B3A4B] font-semibold rounded-lg shadow-lg hover:shadow-xl transform hover:scale-105 transition">
			Submit a Prayer Request
		</a>
	</div>

	<div class="max-w-6xl mx-auto px-4 pb-16">
		{#if loading}
			<div class="flex justify-center py-12"><div class="animate-spin rounded-full h-16 w-16 border-b-4 border-white"></div></div>
		{:else if prayerRequests.length === 0}
			<p class="text-center text-white text-lg">No public prayer requests yet.</p>
		{:else}
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
				{#each prayerRequests as req}
					<div class="bg-white rounded-lg shadow-lg overflow-hidden transform hover:scale-105 transition">
						<div class="bg-[#4A8B8C] px-6 py-4 flex justify-between items-center">
							<h3 class="font-semibold text-white">{req.name}</h3>
							<span class="text-white/80 text-sm">{formatDate(req.submitted_at)}</span>
						</div>
						<div class="p-6">
							<p class="text-gray-700">{req.request_text}</p>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
