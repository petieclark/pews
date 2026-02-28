<script>
	import { onMount } from 'svelte';
	import { ShieldCheck, Users, AlertTriangle } from 'lucide-svelte';
	import { api } from '$lib/api';

	let searchQuery = '';
	let searchResults = [];
	let selectedPerson = null;
	let alerts = [];
	let pickups = [];
	let loading = false;

	// Alert form
	let showAlertModal = false;
	let alertForm = { alert_type: 'allergy', severity: 'medium', description: '' };

	// Pickup form
	let showPickupModal = false;
	let pickupSearch = '';
	let pickupSearchResults = [];
	let pickupForm = { pickup_person_id: '', relationship: '', is_active: true };

	let searchTimeout;
	function handleSearch() {
		clearTimeout(searchTimeout);
		if (searchQuery.length < 2) { searchResults = []; return; }
		searchTimeout = setTimeout(async () => {
			try {
				searchResults = (await api(`/api/checkins/search?q=${encodeURIComponent(searchQuery)}`)) || [];
			} catch (e) { console.error(e); }
		}, 300);
	}

	async function selectPerson(person) {
		selectedPerson = person;
		searchQuery = '';
		searchResults = [];
		loading = true;
		try {
			const [a, p] = await Promise.all([
				api(`/api/checkins/person/${person.id}/alerts`),
				api(`/api/checkins/person/${person.id}/pickups`)
			]);
			alerts = a || [];
			pickups = p || [];
		} catch (e) { console.error(e); }
		finally { loading = false; }
	}

	async function createAlert() {
		try {
			await api(`/api/checkins/person/${selectedPerson.id}/alerts`, {
				method: 'POST',
				body: JSON.stringify(alertForm)
			});
			showAlertModal = false;
			alertForm = { alert_type: 'allergy', severity: 'medium', description: '' };
			alerts = (await api(`/api/checkins/person/${selectedPerson.id}/alerts`)) || [];
		} catch (e) { alert('Failed: ' + e.message); }
	}

	async function deleteAlert(alertId) {
		if (!confirm('Remove this alert?')) return;
		try {
			await api(`/api/checkins/person/${selectedPerson.id}/alerts/${alertId}`, { method: 'DELETE' });
			alerts = alerts.filter(a => a.id !== alertId);
		} catch (e) { alert('Failed: ' + e.message); }
	}

	let pickupTimeout;
	function handlePickupSearch() {
		clearTimeout(pickupTimeout);
		if (pickupSearch.length < 2) { pickupSearchResults = []; return; }
		pickupTimeout = setTimeout(async () => {
			try {
				pickupSearchResults = (await api(`/api/checkins/search?q=${encodeURIComponent(pickupSearch)}`)) || [];
			} catch (e) { console.error(e); }
		}, 300);
	}

	function selectPickupPerson(person) {
		pickupForm.pickup_person_id = person.id;
		pickupSearch = `${person.first_name} ${person.last_name}`;
		pickupSearchResults = [];
	}

	async function createPickup() {
		try {
			await api(`/api/checkins/person/${selectedPerson.id}/pickups`, {
				method: 'POST',
				body: JSON.stringify(pickupForm)
			});
			showPickupModal = false;
			pickupForm = { pickup_person_id: '', relationship: '', is_active: true };
			pickupSearch = '';
			pickups = (await api(`/api/checkins/person/${selectedPerson.id}/pickups`)) || [];
		} catch (e) { alert('Failed: ' + e.message); }
	}

	async function deletePickup(pickupId) {
		if (!confirm('Remove this authorized pickup?')) return;
		try {
			await api(`/api/checkins/person/${selectedPerson.id}/pickups/${pickupId}`, { method: 'DELETE' });
			pickups = pickups.filter(p => p.id !== pickupId);
		} catch (e) { alert('Failed: ' + e.message); }
	}

	const alertTypes = ['allergy', 'medical', 'dietary'];
	const severityLevels = ['low', 'medium', 'high', 'critical'];
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<div>
			<h1 class="text-3xl font-bold text-[var(--text-primary)]"><ShieldCheck size={24} class="inline" /> Child Safety</h1>
			<p class="text-secondary mt-1">Manage medical alerts and authorized pickups</p>
		</div>
		<a href="/dashboard/checkins" class="px-4 py-2 border border-custom rounded-md text-secondary hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800">
			← Back
		</a>
	</div>

	<!-- Search -->
	<div class="bg-surface border border-custom rounded-lg shadow p-6">
		<label class="block text-sm font-medium text-secondary mb-2">Search Child</label>
		<div class="relative">
			<input type="text" bind:value={searchQuery} on:input={handleSearch}
				placeholder="🔍 Search by name..."
				class="w-full px-4 py-3 text-lg border border-custom rounded-lg bg-surface text-primary" />
			{#if searchResults.length > 0}
				<div class="absolute top-full left-0 right-0 mt-1 bg-surface border border-custom rounded-lg shadow-lg z-10">
					{#each searchResults as person}
						<button on:click={() => selectPerson(person)}
							class="w-full text-left p-3 hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800 border-b border-custom last:border-0">
							<span class="font-medium text-[var(--text-primary)]">{person.first_name} {person.last_name}</span>
							{#if person.email}<span class="text-sm text-secondary ml-2">{person.email}</span>{/if}
						</button>
					{/each}
				</div>
			{/if}
		</div>
	</div>

	{#if selectedPerson}
		<div class="text-xl font-semibold text-[var(--text-primary)]">
			{selectedPerson.first_name} {selectedPerson.last_name}
		</div>

		{#if loading}
			<div class="text-center py-4 text-secondary">Loading...</div>
		{:else}
			<!-- Medical Alerts -->
			<div class="bg-surface border border-custom rounded-lg shadow p-6">
				<div class="flex justify-between items-center mb-4">
					<h2 class="text-lg font-semibold text-[var(--text-primary)]"><AlertTriangle size={20} class="inline" /> Medical Alerts</h2>
					<button on:click={() => showAlertModal = true}
						class="px-3 py-1 text-sm bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90">
						+ Add Alert
					</button>
				</div>

				{#if alerts.length === 0}
					<p class="text-secondary">No medical alerts</p>
				{:else}
					<div class="space-y-3">
						{#each alerts as alert}
							<div class="flex justify-between items-start p-3 rounded-lg
								{alert.severity === 'critical' ? 'bg-red-50 dark:bg-red-900/30 border border-red-300 dark:border-red-700' :
								 alert.severity === 'high' ? 'bg-orange-50 dark:bg-orange-900/30 border border-orange-300 dark:border-orange-700' :
								 'bg-yellow-50 dark:bg-yellow-900/30 border border-yellow-300 dark:border-yellow-700'}">
								<div>
									<div class="font-medium text-[var(--text-primary)]">
										<span class="uppercase text-xs font-bold mr-2 {alert.severity === 'critical' ? 'text-red-600' : alert.severity === 'high' ? 'text-orange-600' : 'text-yellow-600'}">
											{alert.severity}
										</span>
										{alert.alert_type}
									</div>
									<div class="text-sm text-secondary mt-1">{alert.description}</div>
								</div>
								<button on:click={() => deleteAlert(alert.id)}
									class="text-red-500 hover:text-red-700 text-sm ml-4">✕</button>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Authorized Pickups -->
			<div class="bg-surface border border-custom rounded-lg shadow p-6">
				<div class="flex justify-between items-center mb-4">
					<h2 class="text-lg font-semibold text-[var(--text-primary)]"><Users size={20} class="inline" /> Authorized Pickups</h2>
					<button on:click={() => showPickupModal = true}
						class="px-3 py-1 text-sm bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90">
						+ Add Pickup
					</button>
				</div>

				{#if pickups.length === 0}
					<p class="text-secondary">No authorized pickups</p>
				{:else}
					<div class="space-y-2">
						{#each pickups as pickup}
							<div class="flex justify-between items-center p-3 border border-custom rounded-lg">
								<div>
									<div class="font-medium text-[var(--text-primary)]">{pickup.pickup_person_name}</div>
									<div class="text-sm text-secondary">{pickup.relationship}</div>
								</div>
								<div class="flex items-center gap-3">
									<span class={`px-2 py-0.5 text-xs rounded ${pickup.is_active
										? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100'
										: 'bg-[var(--surface-hover)] text-primary dark:bg-gray-700 dark:text-gray-300'}`}>
										{pickup.is_active ? 'Active' : 'Inactive'}
									</span>
									<button on:click={() => deletePickup(pickup.id)}
										class="text-red-500 hover:text-red-700 text-sm">✕</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	{/if}
</div>

<!-- Add Alert Modal -->
{#if showAlertModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<div class="bg-surface rounded-lg shadow-xl max-w-md w-full mx-4 p-6">
			<h2 class="text-2xl font-bold text-primary mb-4">Add Medical Alert</h2>
			<form on:submit|preventDefault={createAlert} class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-secondary mb-1">Type</label>
					<select bind:value={alertForm.alert_type}
						class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary">
						{#each alertTypes as t}<option value={t}>{t.charAt(0).toUpperCase() + t.slice(1)}</option>{/each}
					</select>
				</div>
				<div>
					<label class="block text-sm font-medium text-secondary mb-1">Severity</label>
					<select bind:value={alertForm.severity}
						class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary">
						{#each severityLevels as s}<option value={s}>{s.charAt(0).toUpperCase() + s.slice(1)}</option>{/each}
					</select>
				</div>
				<div>
					<label class="block text-sm font-medium text-secondary mb-1">Description *</label>
					<textarea bind:value={alertForm.description} required rows="3"
						class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
						placeholder="Peanut allergy - carries EpiPen" />
				</div>
				<div class="flex justify-end gap-3 pt-2">
					<button type="button" on:click={() => showAlertModal = false}
						class="px-4 py-2 border border-custom rounded-md text-secondary">Cancel</button>
					<button type="submit"
						class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90">Add Alert</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Add Pickup Modal -->
{#if showPickupModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<div class="bg-surface rounded-lg shadow-xl max-w-md w-full mx-4 p-6">
			<h2 class="text-2xl font-bold text-primary mb-4">Add Authorized Pickup</h2>
			<form on:submit|preventDefault={createPickup} class="space-y-4">
				<div class="relative">
					<label class="block text-sm font-medium text-secondary mb-1">Person *</label>
					<input type="text" bind:value={pickupSearch} on:input={handlePickupSearch}
						class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
						placeholder="Search person..." />
					{#if pickupSearchResults.length > 0}
						<div class="absolute top-full left-0 right-0 mt-1 bg-surface border border-custom rounded-lg shadow-lg z-20 max-h-40 overflow-y-auto">
							{#each pickupSearchResults as person}
								<button type="button" on:click={() => selectPickupPerson(person)}
									class="w-full text-left p-2 hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800 text-sm border-b border-custom last:border-0">
									{person.first_name} {person.last_name}
								</button>
							{/each}
						</div>
					{/if}
				</div>
				<div>
					<label class="block text-sm font-medium text-secondary mb-1">Relationship *</label>
					<input type="text" bind:value={pickupForm.relationship} required
						class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
						placeholder="Parent, Grandparent, Guardian..." />
				</div>
				<div class="flex justify-end gap-3 pt-2">
					<button type="button" on:click={() => { showPickupModal = false; pickupSearch = ''; pickupSearchResults = []; }}
						class="px-4 py-2 border border-custom rounded-md text-secondary">Cancel</button>
					<button type="submit"
						class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90">Add Pickup</button>
				</div>
			</form>
		</div>
	</div>
{/if}
