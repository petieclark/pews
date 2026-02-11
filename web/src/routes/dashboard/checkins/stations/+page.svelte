<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let stations = [];
	let loading = true;
	let showModal = false;
	let editingStation = null;
	let form = { name: '', location: '', is_active: true };

	onMount(async () => {
		await loadStations();
	});

	async function loadStations() {
		loading = true;
		try {
			stations = (await api('/api/checkins/stations')) || [];
		} catch (error) {
			console.error('Failed to load stations:', error);
		} finally {
			loading = false;
		}
	}

	function openCreate() {
		editingStation = null;
		form = { name: '', location: '', is_active: true };
		showModal = true;
	}

	function openEdit(station) {
		editingStation = station;
		form = { name: station.name, location: station.location || '', is_active: station.is_active };
		showModal = true;
	}

	async function saveStation() {
		try {
			if (editingStation) {
				await api(`/api/checkins/stations/${editingStation.id}`, {
					method: 'PUT',
					body: JSON.stringify(form)
				});
			} else {
				await api('/api/checkins/stations', {
					method: 'POST',
					body: JSON.stringify(form)
				});
			}
			showModal = false;
			await loadStations();
		} catch (error) {
			alert('Failed to save station: ' + error.message);
		}
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<div>
			<h1 class="text-3xl font-bold text-[var(--text-primary)]">Check-In Stations</h1>
			<p class="text-secondary mt-1">Manage physical check-in locations</p>
		</div>
		<div class="flex gap-3">
			<a href="/dashboard/checkins" class="px-4 py-2 border border-custom rounded-md text-secondary hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800">
				← Back
			</a>
			<button on:click={openCreate}
				class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90">
				+ Add Station
			</button>
		</div>
	</div>

	{#if loading}
		<div class="text-center py-8 text-secondary">Loading stations...</div>
	{:else if stations.length === 0}
		<div class="text-center py-8 text-secondary">No stations yet. Create your first station.</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each stations as station}
				<div class="bg-surface border border-custom rounded-lg shadow p-5 cursor-pointer hover:shadow-md transition-shadow"
					on:click={() => openEdit(station)}>
					<div class="flex justify-between items-start">
						<h3 class="text-lg font-semibold text-[var(--text-primary)]">{station.name}</h3>
						<span class={`px-2 py-0.5 text-xs rounded ${station.is_active
							? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100'
							: 'bg-[var(--surface-hover)] text-primary dark:bg-gray-700 dark:text-gray-300'}`}>
							{station.is_active ? 'Active' : 'Inactive'}
						</span>
					</div>
					{#if station.location}
						<p class="text-sm text-secondary mt-2">📍 {station.location}</p>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>

{#if showModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<div class="bg-surface rounded-lg shadow-xl max-w-md w-full mx-4 p-6">
			<h2 class="text-2xl font-bold text-primary mb-4">
				{editingStation ? 'Edit Station' : 'New Station'}
			</h2>
			<form on:submit|preventDefault={saveStation} class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-secondary mb-1">Station Name *</label>
					<input type="text" bind:value={form.name} required
						class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
						placeholder="Main Lobby" />
				</div>
				<div>
					<label class="block text-sm font-medium text-secondary mb-1">Location</label>
					<input type="text" bind:value={form.location}
						class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
						placeholder="Building A, First Floor" />
				</div>
				<div>
					<label class="flex items-center gap-2">
						<input type="checkbox" bind:checked={form.is_active} class="rounded" />
						<span class="text-sm text-secondary">Active</span>
					</label>
				</div>
				<div class="flex justify-end gap-3 pt-2">
					<button type="button" on:click={() => showModal = false}
						class="px-4 py-2 border border-custom rounded-md text-secondary">Cancel</button>
					<button type="submit"
						class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90">
						{editingStation ? 'Update' : 'Create'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
