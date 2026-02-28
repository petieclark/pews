<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	let eventId = '';
	let event = null;
	let attendees = [];
	let searchQuery = '';
	let searchResults = [];
	let searching = false;
	let loading = true;
	let stations = [];
	let selectedStation = null;

	$: eventId = $page.params.id;

	onMount(async () => {
		try {
			const [ev, st] = await Promise.all([
				api(`/api/checkins/events/${eventId}`),
				api('/api/checkins/stations')
			]);
			event = ev;
			stations = (st || []).filter(s => s.is_active);
			await loadAttendees();
		} catch (error) {
			console.error('Failed to load event:', error);
		} finally {
			loading = false;
		}
	});

	async function loadAttendees() {
		try {
			attendees = (await api(`/api/checkins/events/${eventId}/attendees`)) || [];
		} catch (error) {
			console.error('Failed to load attendees:', error);
		}
	}

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

	async function checkInPerson(person) {
		try {
			const result = await api(`/api/checkins/events/${eventId}/checkin`, {
				method: 'POST',
				body: JSON.stringify({
					person_id: person.id,
					station_id: selectedStation || undefined
				})
			});

			// Show medical alerts if any
			if (result.medical_alerts && result.medical_alerts.length > 0) {
				const alertText = result.medical_alerts.map(a =>
					`${a.severity.toUpperCase()}: ${a.alert_type} - ${a.description}`
				).join('\n');
				alert('MEDICAL ALERTS:\n\n' + alertText);
			}

			if (result.first_time) {
				alert(`Welcome! ${person.first_name} ${person.last_name} is a first-time visitor!`);
			}

			searchQuery = '';
			searchResults = [];
			await loadAttendees();
		} catch (error) {
			alert('Check-in failed: ' + error.message);
		}
	}

	async function checkOutPerson(personId) {
		try {
			await api(`/api/checkins/events/${eventId}/checkout`, {
				method: 'POST',
				body: JSON.stringify({ person_id: personId })
			});
			await loadAttendees();
		} catch (error) {
			alert('Check-out failed: ' + error.message);
		}
	}

	function isCheckedIn(personId) {
		return attendees.some(a => a.person_id === personId && !a.checked_out_at);
	}
</script>

<div class="space-y-6">
	{#if loading}
		<div class="text-center py-8 text-secondary">Loading...</div>
	{:else if !event}
		<div class="text-center py-8 text-secondary">Event not found</div>
	{:else}
		<div class="flex justify-between items-center">
			<div>
				<h1 class="text-3xl font-bold text-[var(--text-primary)]">{event.name}</h1>
				<p class="text-secondary">{event.event_date} · {attendees.length} checked in</p>
			</div>
			<a href="/dashboard/checkins" class="px-4 py-2 border border-custom rounded-md text-secondary hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800">
				← Back
			</a>
		</div>

		<!-- Search Bar -->
		<div class="bg-surface border border-custom rounded-lg shadow p-6">
			<div class="flex gap-4">
				<div class="flex-1 relative">
					<input
						type="text"
						bind:value={searchQuery}
						on:input={handleSearch}
						placeholder="🔍 Search by name, email, or phone..."
						class="w-full px-4 py-3 text-lg border border-custom rounded-lg bg-surface text-primary focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent"
						autofocus
					/>

					<!-- Search Results Dropdown -->
					{#if searchResults.length > 0}
						<div class="absolute top-full left-0 right-0 mt-1 bg-surface border border-custom rounded-lg shadow-lg z-10 max-h-80 overflow-y-auto">
							{#each searchResults as person}
								<div class="flex justify-between items-center p-3 hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800 border-b border-custom last:border-0">
									<div>
										<div class="font-medium text-[var(--text-primary)]">
											{person.first_name} {person.last_name}
										</div>
										{#if person.email}
											<div class="text-sm text-secondary">{person.email}</div>
										{/if}
									</div>
									{#if isCheckedIn(person.id)}
										<span class="px-3 py-1 text-sm bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100 rounded-full">
											✓ Checked In
										</span>
									{:else}
										<button
											on:click={() => checkInPerson(person)}
											class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90 text-sm font-medium"
										>
											Check In
										</button>
									{/if}
								</div>
							{/each}
						</div>
					{/if}

					{#if searching}
						<div class="absolute top-full left-0 right-0 mt-1 bg-surface border border-custom rounded-lg shadow-lg p-4 text-center text-secondary">
							Searching...
						</div>
					{/if}
				</div>

				{#if stations.length > 0}
					<select bind:value={selectedStation}
						class="px-3 py-2 border border-custom rounded-lg bg-surface text-primary">
						<option value={null}>All Stations</option>
						{#each stations as station}
							<option value={station.id}>{station.name}</option>
						{/each}
					</select>
				{/if}
			</div>
		</div>

		<!-- Attendees List -->
		<div class="bg-surface border border-custom rounded-lg shadow">
			<div class="p-4 border-b border-custom">
				<h2 class="text-lg font-semibold text-[var(--text-primary)]">Attendees ({attendees.length})</h2>
			</div>

			{#if attendees.length === 0}
				<div class="text-center py-8 text-secondary">No check-ins yet</div>
			{:else}
				<div class="divide-y divide-[var(--border)]">
					{#each attendees as checkin}
						<div class="flex justify-between items-center p-4">
							<div class="flex items-center gap-3">
								{#if checkin.first_time}
									<span class="px-2 py-0.5 text-xs bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100 rounded-full font-semibold">
										🌟 FIRST TIME
									</span>
								{/if}
								<div>
									<div class="font-medium text-[var(--text-primary)]">{checkin.person_name}</div>
									<div class="text-sm text-secondary">
										{checkin.station_name || 'No station'}
										· In: {new Date(checkin.checked_in_at).toLocaleTimeString()}
										{#if checkin.checked_out_at}
											· Out: {new Date(checkin.checked_out_at).toLocaleTimeString()}
										{/if}
									</div>
								</div>
							</div>
							{#if !checkin.checked_out_at}
								<button
									on:click={() => checkOutPerson(checkin.person_id)}
									class="px-3 py-1 text-sm border border-custom rounded-md text-secondary hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800"
								>
									Check Out
								</button>
							{:else}
								<span class="text-sm text-secondary">Checked Out</span>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>
