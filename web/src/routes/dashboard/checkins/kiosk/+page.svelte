<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let events = [];
	let selectedEventId = '';
	let searchQuery = '';
	let searchResults = [];
	let searching = false;
	let stations = [];
	let selectedStation = null;
	let lastCheckin = null;
	let showAlerts = false;
	let currentAlerts = [];

	onMount(async () => {
		try {
			const [eventsData, stationsData] = await Promise.all([
				api('/api/checkins/events'),
				api('/api/checkins/stations')
			]);
			events = (eventsData || []).filter(e => e.is_active);
			stations = (stationsData || []).filter(s => s.is_active);
			if (events.length > 0) selectedEventId = events[0].id;
		} catch (error) {
			console.error('Failed to load:', error);
		}
	});

	let searchTimeout;
	function handleSearch() {
		clearTimeout(searchTimeout);
		if (searchQuery.length < 2) {
			searchResults = [];
			return;
		}
		searchTimeout = setTimeout(async () => {
			searching = true;
			try {
				searchResults = (await api(`/api/checkins/search?q=${encodeURIComponent(searchQuery)}`)) || [];
			} catch (error) {
				console.error('Search failed:', error);
			} finally {
				searching = false;
			}
		}, 300);
	}

	async function checkIn(person) {
		if (!selectedEventId) {
			alert('Please select an event first');
			return;
		}

		try {
			const result = await api(`/api/checkins/events/${selectedEventId}/checkin`, {
				method: 'POST',
				body: JSON.stringify({
					person_id: person.id,
					station_id: selectedStation || undefined
				})
			});

			lastCheckin = {
				name: `${person.first_name} ${person.last_name}`,
				firstTime: result.first_time,
				time: new Date().toLocaleTimeString()
			};

			if (result.medical_alerts && result.medical_alerts.length > 0) {
				currentAlerts = result.medical_alerts;
				showAlerts = true;
			}

			searchQuery = '';
			searchResults = [];

			// Auto-clear success message
			setTimeout(() => { lastCheckin = null; }, 5000);
		} catch (error) {
			alert('Check-in failed: ' + error.message);
		}
	}

	function exitKiosk() {
		window.location.href = '/dashboard/checkins';
	}
</script>

<svelte:head>
	<style>
		nav, footer { display: none !important; }
		main { max-width: 100% !important; padding: 0 !important; }
	</style>
</svelte:head>

<div class="min-h-screen bg-[var(--bg)] flex flex-col">
	<!-- Header -->
	<div class="bg-[var(--teal)] text-white p-4 flex justify-between items-center">
		<h1 class="text-2xl font-bold">Check-In Kiosk</h1>
		<div class="flex items-center gap-4">
			{#if events.length > 1}
				<select bind:value={selectedEventId}
					class="px-3 py-2 rounded-md bg-surface bg-opacity-20 text-white border-0 text-lg">
					{#each events as event}
						<option value={event.id}>{event.name}</option>
					{/each}
				</select>
			{/if}
			{#if stations.length > 0}
				<select bind:value={selectedStation}
					class="px-3 py-2 rounded-md bg-surface bg-opacity-20 text-white border-0 text-lg">
					<option value={null}>All Stations</option>
					{#each stations as station}
						<option value={station.id}>{station.name}</option>
					{/each}
				</select>
			{/if}
			<button on:click={exitKiosk} class="px-3 py-2 text-sm bg-surface bg-opacity-20 rounded-md hover:bg-opacity-30">
				Exit Kiosk
			</button>
		</div>
	</div>

	<!-- Main Content -->
	<div class="flex-1 flex flex-col items-center justify-center p-8">
		<!-- Success Message -->
		{#if lastCheckin}
			<div class="mb-8 p-6 rounded-xl text-center {lastCheckin.firstTime ? 'bg-yellow-100 dark:bg-yellow-900 border-2 border-yellow-400' : 'bg-green-100 dark:bg-green-900 border-2 border-green-400'}" style="min-width: 400px;">
				{#if lastCheckin.firstTime}
					<div class="text-4xl mb-2">🎉</div>
					<div class="text-2xl font-bold text-yellow-800 dark:text-yellow-100">Welcome, {lastCheckin.name}!</div>
					<div class="text-lg text-yellow-700 dark:text-yellow-200">First-time visitor</div>
				{:else}
					<div class="text-4xl mb-2">✅</div>
					<div class="text-2xl font-bold text-green-800 dark:text-green-100">{lastCheckin.name}</div>
					<div class="text-lg text-green-700 dark:text-green-200">Checked in at {lastCheckin.time}</div>
				{/if}
			</div>
		{/if}

		<!-- Giant Search -->
		<div class="w-full max-w-2xl">
			<input
				type="text"
				bind:value={searchQuery}
				on:input={handleSearch}
				placeholder="🔍 Search name or phone number..."
				class="w-full px-6 py-5 text-2xl border-2 border-custom rounded-xl bg-surface text-primary focus:ring-4 focus:ring-[var(--teal)] focus:border-transparent shadow-lg"
				autofocus
			/>

			<!-- Search Results -->
			{#if searchResults.length > 0}
				<div class="mt-2 bg-surface border border-custom rounded-xl shadow-xl overflow-hidden">
					{#each searchResults as person}
						<button
							on:click={() => checkIn(person)}
							class="w-full flex justify-between items-center p-5 hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800 border-b border-custom last:border-0 text-left"
						>
							<div>
								<div class="text-xl font-medium text-[var(--text-primary)]">
									{person.first_name} {person.last_name}
								</div>
								{#if person.phone}
									<div class="text-secondary">{person.phone}</div>
								{/if}
							</div>
							<span class="px-6 py-3 bg-[var(--teal)] text-white rounded-lg text-lg font-medium">
								Check In
							</span>
						</button>
					{/each}
				</div>
			{/if}

			{#if searching}
				<div class="mt-2 bg-surface border border-custom rounded-xl shadow p-6 text-center text-secondary text-xl">
					Searching...
				</div>
			{/if}
		</div>
	</div>
</div>

<!-- Medical Alerts Modal -->
{#if showAlerts}
	<div class="fixed inset-0 bg-black bg-opacity-70 flex items-center justify-center z-50">
		<div class="bg-surface dark:bg-gray-800 rounded-xl shadow-2xl max-w-lg w-full mx-4 p-8">
			<h2 class="text-2xl font-bold text-red-600 mb-4">⚠️ Medical Alerts</h2>
			<div class="space-y-4">
				{#each currentAlerts as alert}
					<div class="p-4 rounded-lg {alert.severity === 'critical' ? 'bg-red-100 dark:bg-red-900 border-2 border-red-400' : alert.severity === 'high' ? 'bg-orange-100 dark:bg-orange-900 border-2 border-orange-400' : 'bg-yellow-100 dark:bg-yellow-900 border border-yellow-400'}">
						<div class="font-semibold text-lg">
							{alert.severity.toUpperCase()} - {alert.alert_type}
						</div>
						<div class="mt-1 text-secondary">{alert.description}</div>
					</div>
				{/each}
			</div>
			<button
				on:click={() => { showAlerts = false; currentAlerts = []; }}
				class="mt-6 w-full py-3 bg-[var(--teal)] text-white rounded-lg text-lg font-medium hover:bg-opacity-90"
			>
				Acknowledged
			</button>
		</div>
	</div>
{/if}
