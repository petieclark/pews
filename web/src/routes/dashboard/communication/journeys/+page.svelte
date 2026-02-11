<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let journeys = [];
	let loading = true;
	let error = '';

	onMount(async () => {
		await loadJourneys();
	});

	async function loadJourneys() {
		try {
			loading = true;
			journeys = await api('/api/communication/journeys');
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	async function toggleActive(journey) {
		try {
			await api(`/api/communication/journeys/${journey.id}`, {
				method: 'PUT',
				body: JSON.stringify({ ...journey, is_active: !journey.is_active })
			});
			await loadJourneys();
		} catch (err) {
			error = err.message;
		}
	}
</script>

<div>
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-3xl font-bold" style="color: var(--text-primary)">Automated Journeys</h1>
		<button
			on:click={() => goto('/dashboard/communication/journeys/new')}
			class="px-4 py-2 rounded-lg font-medium"
			style="background: var(--teal); color: white"
		>
			Create Journey
		</button>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2" style="border-color: var(--teal)"></div>
		</div>
	{:else if error}
		<div class="px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">
			{error}
		</div>
	{:else if journeys.length === 0}
		<div class="rounded-lg shadow border p-12 text-center" style="background: var(--surface); border-color: var(--border)">
			<div class="text-5xl mb-4">🗺️</div>
			<h2 class="text-xl font-semibold mb-2" style="color: var(--text-primary)">No journeys yet</h2>
			<p class="mb-6" style="color: var(--text-secondary)">Create automated message sequences to nurture visitors and members</p>
			<button
				on:click={() => goto('/dashboard/communication/journeys/new')}
				class="px-6 py-2 rounded-lg font-medium"
				style="background: var(--teal); color: white"
			>
				Create First Journey
			</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each journeys as journey}
				<div class="rounded-lg shadow border p-6" style="background: var(--surface); border-color: var(--border)">
					<div class="flex items-start justify-between mb-4">
						<div class="flex-1">
							<h3 class="font-semibold text-lg mb-1" style="color: var(--text-primary)">{journey.name}</h3>
							<p class="text-sm" style="color: var(--text-secondary)">{journey.description || 'No description'}</p>
						</div>
						<button
							on:click={() => toggleActive(journey)}
							class="ml-2 px-3 py-1 rounded-full text-xs font-medium"
							style="background: {journey.is_active ? 'var(--teal)' : 'var(--surface-hover)'}; color: {journey.is_active ? 'white' : 'var(--text-secondary)'}"
						>
							{journey.is_active ? 'Active' : 'Paused'}
						</button>
					</div>

					<div class="text-sm space-y-2 mb-4" style="color: var(--text-secondary)">
						<div>
							<span class="font-medium">Trigger:</span>
							<span class="capitalize">{journey.trigger_type.replace('_', ' ')}</span>
						</div>
						<div>
							<span class="font-medium">Enrolled:</span>
							<span>{journey.enrollment_count || 0} people</span>
						</div>
					</div>

					<button
						on:click={() => goto(`/dashboard/communication/journeys/${journey.id}`)}
						class="w-full px-4 py-2 rounded-lg font-medium border"
						style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
					>
						Edit Journey
					</button>
				</div>
			{/each}
		</div>
	{/if}
</div>
