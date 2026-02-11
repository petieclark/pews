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
	<header class="bg-[var(--teal)] text-white p-4 flex justify-between items-center">
		<h1 class="text-2xl font-bold">Check-In Kiosk</h1>
		<div class="flex items-center gap-4">
			{#if events.length > 1}
				<label for="event-select" class="sr-only">Select Event</label>
				<select id="event-select" bind:value={selectedEventId}
					class="px-3 py-2 rounded-md bg-white bg-opacity-20 text-white border-0 text-lg"
					aria-label="Select event for check-in">
					{#each events as event}
						<option value={event.id}>{event.name}</option>
					{/each}
				</select>
			{/if}
			{#if stations.length > 0}
				<label for="station-select" class="sr-only">Select Station</label>
				<select id="station-select" bind:value={selectedStation}
					class="px-3 py-2 rounded-md bg-white bg-opacity-20 text-white border-0 text-lg"
					aria-label="Select check-in station">
					<option value={null}>All Stations</option>
					{#each stations as station}
						<option value={station.id}>{station.name}</option>
					{/each}
				</select>
			{/if}
			<button on:click={exitKiosk} 
				class="px-3 py-2 text-sm bg-white bg-opacity-20 rounded-md hover:bg-opacity-30"
				aria-label="Exit kiosk mode and return to dashboard">
				Exit Kiosk
			</button>
		</div>
	</header>

	<!-- Main Content -->
	<main class="flex-1 flex flex-col items-center justify-center p-8">
		<!-- Success Message with Screen Reader Announcement -->
		<div aria-live="polite" aria-atomic="true" class="sr-only">
			{#if lastCheckin}
				{lastCheckin.firstTime ? `Welcome first-time visitor ${lastCheckin.name}, checked in successfully` : `${lastCheckin.name} checked in at ${lastCheckin.time}`}
			{/if}
		</div>
		
		{#if lastCheckin}
			<div class="mb-8 p-6 rounded-xl text-center {lastCheckin.firstTime ? 'bg-yellow-100 dark:bg-yellow-900 border-2 border-yellow-400' : 'bg-green-100 dark:bg-green-900 border-2 border-green-400'}" style="min-width: 400px;" role="status">
				{#if lastCheckin.firstTime}
					<div class="text-4xl mb-2" aria-hidden="true">🎉</div>
					<div class="text-2xl font-bold text-yellow-800 dark:text-yellow-100">Welcome, {lastCheckin.name}!</div>
					<div class="text-lg text-yellow-700 dark:text-yellow-200">First-time visitor</div>
				{:else}
					<div class="text-4xl mb-2" aria-hidden="true">✅</div>
					<div class="text-2xl font-bold text-green-800 dark:text-green-100">{lastCheckin.name}</div>
					<div class="text-lg text-green-700 dark:text-green-200">Checked in at {lastCheckin.time}</div>
				{/if}
			</div>
		{/if}

		<!-- Giant Search -->
		<div class="w-full max-w-2xl">
			<label for="kiosk-search" class="sr-only">Search for person by name or phone number</label>
			<input
				id="kiosk-search"
				type="search"
				bind:value={searchQuery}
				on:input={handleSearch}
				placeholder="🔍 Search name or phone number..."
				class="w-full px-6 py-5 text-2xl border-2 border-custom rounded-xl bg-surface text-primary focus:ring-4 focus:ring-[var(--teal)] focus:border-transparent shadow-lg"
				aria-describedby="search-instructions"
				autofocus
			/>
			<div id="search-instructions" class="sr-only">
				Type at least 2 characters to search for people. Results will appear below.
			</div>

			<!-- Search Results -->
			{#if searchResults.length > 0}
				<div class="mt-2 bg-surface border border-custom rounded-xl shadow-xl overflow-hidden" role="list" aria-label="Search results">
					{#each searchResults as person, index}
						<button
							on:click={() => checkIn(person)}
							class="w-full flex justify-between items-center p-5 hover:bg-gray-50 dark:hover:bg-gray-800 border-b border-custom last:border-0 text-left focus:ring-2 focus:ring-inset focus:ring-[var(--teal)] focus:outline-none"
							aria-label="Check in {person.first_name} {person.last_name}"
							role="listitem"
						>
							<div>
								<div class="text-xl font-medium text-[var(--text-primary)]">
									{person.first_name} {person.last_name}
								</div>
								{#if person.phone}
									<div class="text-secondary">{person.phone}</div>
								{/if}
							</div>
							<span class="px-6 py-3 bg-[var(--teal)] text-white rounded-lg text-lg font-medium" aria-hidden="true">
								Check In
							</span>
						</button>
					{/each}
				</div>
			{/if}

			{#if searching}
				<div class="mt-2 bg-surface border border-custom rounded-xl shadow p-6 text-center text-secondary text-xl" role="status" aria-live="polite">
					Searching...
				</div>
			{/if}
		</div>
	</main>
</div>

<!-- Medical Alerts Modal -->
{#if showAlerts}
	<div 
		class="fixed inset-0 bg-black bg-opacity-70 flex items-center justify-center z-50"
		on:click={(e) => { if (e.target === e.currentTarget) { showAlerts = false; currentAlerts = []; } }}
		on:keydown={(e) => { if (e.key === 'Escape') { showAlerts = false; currentAlerts = []; } }}
		role="presentation"
	>
		<div 
			class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl max-w-lg w-full mx-4 p-8"
			role="alertdialog"
			aria-modal="true"
			aria-labelledby="alert-dialog-title"
			aria-describedby="alert-dialog-description"
		>
			<h2 id="alert-dialog-title" class="text-2xl font-bold text-red-600 mb-4">
				<span aria-hidden="true">⚠️</span> Medical Alerts
			</h2>
			<div id="alert-dialog-description" class="space-y-4">
				{#each currentAlerts as alert, index}
					<div class="p-4 rounded-lg {alert.severity === 'critical' ? 'bg-red-100 dark:bg-red-900 border-2 border-red-400' : alert.severity === 'high' ? 'bg-orange-100 dark:bg-orange-900 border-2 border-orange-400' : 'bg-yellow-100 dark:bg-yellow-900 border border-yellow-400'}" role="alert">
						<div class="font-semibold text-lg">
							{alert.severity.toUpperCase()} - {alert.alert_type}
						</div>
						<div class="mt-1 text-secondary">{alert.description}</div>
					</div>
				{/each}
			</div>
			<button
				on:click={() => { showAlerts = false; currentAlerts = []; }}
				class="mt-6 w-full py-3 bg-[var(--teal)] text-white rounded-lg text-lg font-medium hover:bg-opacity-90 focus:ring-2 focus:ring-offset-2 focus:ring-[var(--teal)]"
				autofocus
			>
				Acknowledged
			</button>
		</div>
	</div>
{/if}

<style>
	/* Screen reader only class */
	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		white-space: nowrap;
		border-width: 0;
	}
</style>
