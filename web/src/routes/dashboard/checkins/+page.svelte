<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { Chart, registerables } from 'chart.js';
	Chart.register(...registerables);

	let stats = { total_checkins: 0, first_timers: 0, by_station: [] };
	let events = [];
	let recentCheckins = [];
	let trends = [];
	let loading = true;
	let trendChart = null;
	let chartCanvas;

	// Quick check-in state
	let selectedEvent = '';
	let searchQuery = '';
	let searchResults = [];
	let searching = false;
	let checkingIn = false;
	let searchTimeout;

	onMount(async () => {
		try {
			const [statsData, eventsData, trendsData] = await Promise.all([
				api('/api/checkins/stats'),
				api('/api/checkins/events'),
				api('/api/attendance/trends?period=weekly&limit=12')
			]);
			stats = statsData;
			events = (eventsData || []).filter(e => e.is_active);
			trends = trendsData || [];

			if (events.length > 0) {
				selectedEvent = events[0].id;
				const attendees = await api(`/api/checkins/events/${events[0].id}/attendees`);
				recentCheckins = (attendees || []).slice(0, 10);
			}
		} catch (error) { console.error('Failed to load dashboard:', error); }
		loading = false;
		if (trends.length > 0) renderTrendChart();
	});

	function renderTrendChart() {
		if (!chartCanvas) return;
		const reversed = [...trends].reverse();
		const labels = reversed.map(t => t.period);
		const data = reversed.map(t => t.count);

		if (trendChart) trendChart.destroy();
		trendChart = new Chart(chartCanvas, {
			type: 'line',
			data: {
				labels,
				datasets: [{
					label: 'Attendance',
					data,
					borderColor: '#4A8B8C',
					backgroundColor: 'rgba(74, 139, 140, 0.1)',
					fill: true,
					tension: 0.3,
					pointBackgroundColor: '#4A8B8C',
					pointRadius: 4
				}]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: { legend: { display: false } },
				scales: {
					y: { beginAtZero: true, ticks: { color: '#9CA3AF', stepSize: 1 }, grid: { color: 'rgba(156,163,175,0.1)' } },
					x: { ticks: { color: '#9CA3AF' }, grid: { display: false } }
				}
			}
		});
	}

	function handleSearch() {
		clearTimeout(searchTimeout);
		if (searchQuery.length < 2) { searchResults = []; return; }
		searchTimeout = setTimeout(async () => {
			searching = true;
			try {
				searchResults = await api(`/api/checkins/search?q=${encodeURIComponent(searchQuery)}`);
			} catch { searchResults = []; }
			searching = false;
		}, 300);
	}

	async function quickCheckin(personId) {
		if (!selectedEvent) { alert('Please select an event first'); return; }
		checkingIn = true;
		try {
			const result = await api(`/api/checkins/events/${selectedEvent}/checkin`, {
				method: 'POST',
				body: JSON.stringify({ person_id: personId })
			});
			searchQuery = '';
			searchResults = [];
			// Reload attendees
			const attendees = await api(`/api/checkins/events/${selectedEvent}/attendees`);
			recentCheckins = (attendees || []).slice(0, 10);
			// Reload stats
			stats = await api('/api/checkins/stats');
			if (result.first_time) alert('🎉 First-time visitor checked in!');
		} catch (error) { alert('Check-in failed: ' + error.message); }
		checkingIn = false;
	}

	let showCreateEvent = false;
	let newEvent = { name: '', event_date: new Date().toISOString().split('T')[0], is_active: true };

	async function createEvent() {
		try {
			await api('/api/checkins/events', { method: 'POST', body: JSON.stringify(newEvent) });
			showCreateEvent = false;
			newEvent = { name: '', event_date: new Date().toISOString().split('T')[0], is_active: true };
			location.reload();
		} catch (error) { alert('Failed to create event: ' + error.message); }
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-[var(--text-primary)]">Check-Ins</h1>
		<div class="flex gap-3">
			<button on:click={() => goto('/dashboard/checkins/kiosk')} class="px-4 py-2 border border-custom rounded-md text-secondary hover:bg-[var(--surface-hover)]">🖥️ Kiosk</button>
			<a href="/dashboard/checkins/safety" class="px-4 py-2 border border-custom rounded-md text-secondary hover:bg-[var(--surface-hover)]">🛡️ Safety</a>
			<a href="/dashboard/checkins/stations" class="px-4 py-2 border border-custom rounded-md text-secondary hover:bg-[var(--surface-hover)]">📍 Stations</a>
		</div>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else}
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

		<!-- Quick Check-In -->
		<div class="bg-surface border border-custom rounded-lg shadow p-6">
			<h2 class="text-lg font-semibold text-[var(--text-primary)] mb-4">Quick Check-In</h2>
			<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
				<div>
					<label class="block text-sm font-medium text-secondary mb-1">Service / Event</label>
					<select bind:value={selectedEvent} class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary">
						{#each events as event}
							<option value={event.id}>{event.name} ({event.event_date})</option>
						{/each}
						{#if events.length === 0}
							<option value="">No active events</option>
						{/if}
					</select>
				</div>
				<div class="md:col-span-2 relative">
					<label class="block text-sm font-medium text-secondary mb-1">Search Person</label>
					<input
						type="text"
						bind:value={searchQuery}
						on:input={handleSearch}
						placeholder="Type a name to search..."
						class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary"
					/>
					{#if searchResults.length > 0}
						<div class="absolute z-10 w-full mt-1 bg-surface border border-custom rounded-md shadow-lg max-h-60 overflow-y-auto">
							{#each searchResults as person}
								<button
									class="w-full px-4 py-3 text-left hover:bg-[var(--surface-hover)] flex justify-between items-center border-b border-custom last:border-0"
									on:click={() => quickCheckin(person.id)}
									disabled={checkingIn}
								>
									<div>
										<div class="font-medium text-primary">{person.first_name} {person.last_name}</div>
										{#if person.email}<div class="text-xs text-secondary">{person.email}</div>{/if}
									</div>
									<span class="text-[var(--teal)] text-sm font-medium">Check In →</span>
								</button>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		</div>

		<!-- Attendance Trends Chart -->
		{#if trends.length > 0}
			<div class="bg-surface border border-custom rounded-lg shadow p-6">
				<h2 class="text-lg font-semibold text-[var(--text-primary)] mb-4">Attendance Trends (Last 12 Weeks)</h2>
				<div style="height: 250px;">
					<canvas bind:this={chartCanvas}></canvas>
				</div>
			</div>
		{/if}

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
				<button on:click={() => showCreateEvent = true} class="px-3 py-1 text-sm bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90">+ New Event</button>
			</div>
			{#if events.length === 0}
				<div class="text-center py-8">
					<div class="text-4xl mb-3">📋</div>
					<p class="text-secondary mb-2">No active events today</p>
					<p class="text-sm text-secondary">Create an event to start checking people in.</p>
				</div>
			{:else}
				<div class="space-y-3">
					{#each events as event}
						<div
							class="flex justify-between items-center p-4 border border-custom rounded-lg cursor-pointer hover:bg-[var(--surface-hover)]"
							on:click={() => goto(`/dashboard/checkins/events/${event.id}`)}
							on:keypress={() => goto(`/dashboard/checkins/events/${event.id}`)}
							role="button"
							tabindex="0"
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
				<div class="overflow-x-auto">
					<table class="w-full">
						<thead class="bg-[var(--bg)] border-b border-custom">
							<tr>
								<th class="px-4 py-2 text-left text-xs font-medium text-secondary uppercase">Name</th>
								<th class="px-4 py-2 text-left text-xs font-medium text-secondary uppercase">Station</th>
								<th class="px-4 py-2 text-left text-xs font-medium text-secondary uppercase">Time</th>
								<th class="px-4 py-2 text-left text-xs font-medium text-secondary uppercase">Status</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-custom">
							{#each recentCheckins as checkin}
								<tr class="hover:bg-[var(--surface-hover)]">
									<td class="px-4 py-3 text-sm text-primary font-medium">{checkin.person_name}</td>
									<td class="px-4 py-3 text-sm text-secondary">{checkin.station_name || 'No station'}</td>
									<td class="px-4 py-3 text-sm text-secondary">{new Date(checkin.checked_in_at).toLocaleTimeString()}</td>
									<td class="px-4 py-3">
										{#if checkin.first_time}
											<span class="px-2 py-0.5 text-xs bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100 rounded-full">NEW</span>
										{:else}
											<span class="px-2 py-0.5 text-xs bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100 rounded-full">Regular</span>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}
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
					<input type="text" bind:value={newEvent.name} required class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary" placeholder="Sunday Morning Service" />
				</div>
				<div>
					<label class="block text-sm font-medium text-secondary mb-1">Date</label>
					<input type="date" bind:value={newEvent.event_date} class="w-full px-3 py-2 border border-custom rounded-md bg-surface text-primary" />
				</div>
				<div class="flex justify-end gap-3 pt-2">
					<button type="button" on:click={() => showCreateEvent = false} class="px-4 py-2 border border-custom rounded-md text-secondary">Cancel</button>
					<button type="submit" class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:bg-opacity-90">Create</button>
				</div>
			</form>
		</div>
	</div>
{/if}
