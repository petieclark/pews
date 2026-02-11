<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let stats = { total_checkins: 0, first_timers: 0, by_station: [] };
	let events = [];
	let recentCheckins = [];
	let loading = true;

	onMount(async () => {
		try {
			const [statsData, eventsData] = await Promise.all([
				api('/api/checkins/stats'),
				api('/api/checkins/events')
			]);
			stats = statsData;
			events = (eventsData || []).filter(e => e.is_active);

			// Load recent checkins from active events
			if (events.length > 0) {
				const attendees = await api(`/api/checkins/events/${events[0].id}/attendees`);
				recentCheckins = (attendees || []).slice(0, 10);
			}
		} catch (error) {
			console.error('Failed to load dashboard:', error);
		} finally {
			loading = false;
		}
	});

	let showCreateEvent = false;
	let newEvent = { name: '', event_date: new Date().toISOString().split('T')[0], is_active: true };

	async function createEvent() {
		try {
			await api('/api/checkins/events', {
				method: 'POST',
				body: JSON.stringify(newEvent)
			});
			showCreateEvent = false;
			newEvent = { name: '', event_date: new Date().toISOString().split('T')[0], is_active: true };
			location.reload();
		} catch (error) {
			alert('Failed to create event: ' + error.message);
		}
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-[var(--text-primary)]">Check-Ins</h1>
		<div class="flex gap-3">
			<button
				on:click={() => goto('/dashboard/checkins/kiosk')}
				class="px-4 py-2 border border-custom rounded-md text-secondary hover:bg-gray-50 dark:hover:bg-gray-800"
			>
				🖥️ Kiosk Mode
			</button>
			<a href="/dashboard/checkins/safety" class="px-4 py-2 border border-custom rounded-md text-secondary hover:bg-gray-50 dark:hover:bg-gray-800">
				🛡️ Child Safety
			</a>
			<a href="/dashboard/checkins/stations" class="px-4 py-2 border border-custom rounded-md text-secondary hover:bg-gray-50 dark:hover:bg-gray-800">
				📍 Stations
			</a>
		</div>
	</div>

	<!-- Stats Cards -->
	<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
		<div class="bg-surface border border-custom rounded-lg shadow p-6">
			<div class="text-sm text-secondary mb-1">Total Check-Ins Today</div>
			<div class="text-4xl font-bold text-[var(--teal)]">{stats.total_checkins}</div>
		</div>
		<div class="bg-surface border border-custom rounded-lg shadow p-6">
			<div class="text-sm text-secondary mb-1">First-Timers</div>
			<div class="text-4xl font-bold text-[var(--navy)]">{stats.first_timers}</div>
		</div>
		<div class="bg-surface border border-custom rounded-lg shadow p-6">
			<div class="text-sm text-secondary mb-1">Active Stations</div>
			<div class="text-4xl font-bold text-[var(--text-primary)]">{stats.by_station?.length || 0}</div>
		</div>
	</div>

	<!-- By Station Breakdown -->
	{#if stats.by_station && stats.by_station.length > 0}
		<div class="bg-surface border border-custom rounded-lg shadow p-6">
			<h2 class="text-lg font-semibold text-[var(--text-primary)] mb-4">By Station</h2>
			<div class="space-y-3">
				{#each stats.by_station as station}
					<div class="flex justify-between items-center">
						<span class="text-secondary">{station.station_name}</span>
						<span class="font-semibold text-[var(--text-primary)]">{station.count}</span>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Active Events -->
	<div class="bg-surface border border-custom rounded-lg shadow p-6">
		<div class="flex justify-between items-center mb-4">
			<h2 class="text-lg font-semibold text-[var(--text-primary)]">Active Events</h2>
			<button
				on:click={() => showCreateEvent = true}
				class="px-3 py-1 text-sm bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90"
			>
				+ New Event
			</button>
		</div>

		{#if loading}
			<div class="text-center py-4 text-secondary">Loading...</div>
		{:else if events.length === 0}
			<div class="text-center py-4 text-secondary">No active events today</div>
		{:else}
			<div class="space-y-3">
				{#each events as event}
					<div
						class="flex justify-between items-center p-4 border border-custom rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800"
						on:click={() => goto(`/dashboard/checkins/events/${event.id}`)}
					>
						<div>
							<div class="font-medium text-[var(--text-primary)]">{event.name}</div>
							<div class="text-sm text-secondary">{event.event_date}</div>
						</div>
						<div class="text-right">
							<div class="text-2xl font-bold text-[var(--teal)]">{event.checkin_count}</div>
							<div class="text-xs text-secondary">checked in</div>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Recent Check-ins -->
	{#if recentCheckins.length > 0}
		<div class="bg-surface border border-custom rounded-lg shadow p-6">
			<h2 class="text-lg font-semibold text-[var(--text-primary)] mb-4">Recent Check-Ins</h2>
			<div class="space-y-2">
				{#each recentCheckins as checkin}
					<div class="flex justify-between items-center py-2 border-b border-custom last:border-0">
						<div class="flex items-center gap-3">
							{#if checkin.first_time}
								<span class="px-2 py-0.5 text-xs bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100 rounded-full">NEW</span>
							{/if}
							<span class="text-[var(--text-primary)]">{checkin.person_name}</span>
						</div>
						<div class="text-sm text-secondary">
							{checkin.station_name || 'No station'}
							· {new Date(checkin.checked_in_at).toLocaleTimeString()}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

<!-- Create Event Modal -->

{#if showCreateEvent}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<div class="bg-surface rounded-lg shadow-xl max-w-md w-full mx-4 p-6">
			<h2 class="text-2xl font-bold text-primary mb-4">New Check-In Event</h2>
			<form on:submit|preventDefault={createEvent} class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-secondary mb-1">Event Name</label>
					<input type="text" bind:value={newEvent.name} required
						class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
						placeholder="Sunday Morning Service" />
				</div>
				<div>
					<label class="block text-sm font-medium text-secondary mb-1">Date</label>
					<input type="date" bind:value={newEvent.event_date}
						class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary" />
				</div>
				<div class="flex justify-end gap-3 pt-2">
					<button type="button" on:click={() => showCreateEvent = false}
						class="px-4 py-2 border border-custom rounded-md text-secondary">Cancel</button>
					<button type="submit"
						class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90">Create</button>
				</div>
			</form>
		</div>
	</div>
{/if}
