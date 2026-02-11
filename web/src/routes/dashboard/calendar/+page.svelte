<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	interface CalendarEvent {
		id: string;
		title: string;
		description?: string;
		location?: string;
		start_time: string;
		end_time: string;
		all_day: boolean;
		recurring: string;
		event_type: string;
		color: string;
		room_id?: string;
		room_name?: string;
	}

	let events: CalendarEvent[] = [];
	let currentDate = new Date();
	let currentYear = currentDate.getFullYear();
	let currentMonth = currentDate.getMonth();
	let viewMode: 'month' | 'list' = 'month';
	let showCreateModal = false;
	let showDetailModal = false;
	let selectedEvent: CalendarEvent | null = null;
	let editingEvent: CalendarEvent | null = null;
	let availableRooms: any[] = [];
	let typeFilter = '';

	const eventTypes = [
		{ value: 'service', label: 'Service', color: '#4A8B8C' },
		{ value: 'meeting', label: 'Meeting', color: '#1B3A4B' },
		{ value: 'class', label: 'Class', color: '#8B5CF6' },
		{ value: 'social', label: 'Social', color: '#F59E0B' },
		{ value: 'outreach', label: 'Outreach', color: '#10B981' },
		{ value: 'other', label: 'Other', color: '#6B7280' }
	];

	let formData = {
		title: '',
		description: '',
		location: '',
		start_time: '',
		end_time: '',
		all_day: false,
		recurring: 'none',
		event_type: 'other',
		color: '',
		room_id: ''
	};

	const monthNames = [
		'January', 'February', 'March', 'April', 'May', 'June',
		'July', 'August', 'September', 'October', 'November', 'December'
	];

	onMount(() => { loadEvents(); });

	async function loadEvents() {
		const firstDay = new Date(currentYear, currentMonth, 1);
		const lastDay = new Date(currentYear, currentMonth + 1, 0, 23, 59, 59);
		const from = firstDay.toISOString();
		const to = lastDay.toISOString();

		try {
			let url = `/api/events?from=${from}&to=${to}&limit=100`;
			if (typeFilter) url += `&type=${typeFilter}`;
			const data = await api(url);
			events = data.events || [];
		} catch (error) {
			console.error('Failed to load events:', error);
		}
	}

	function getTypeColor(type: string) {
		return eventTypes.find(t => t.value === type)?.color || '#6B7280';
	}

	function getTypeLabel(type: string) {
		return eventTypes.find(t => t.value === type)?.label || 'Other';
	}

	function getDaysInMonth() {
		const firstDay = new Date(currentYear, currentMonth, 1);
		const lastDay = new Date(currentYear, currentMonth + 1, 0);
		const daysInMonth = lastDay.getDate();
		const startingDayOfWeek = firstDay.getDay();
		const days: { day: number | null; events: CalendarEvent[]; isToday: boolean }[] = [];

		for (let i = 0; i < startingDayOfWeek; i++) {
			days.push({ day: null, events: [], isToday: false });
		}

		const today = new Date();
		for (let day = 1; day <= daysInMonth; day++) {
			const isToday = today.getFullYear() === currentYear && today.getMonth() === currentMonth && today.getDate() === day;
			const dayEvents = events.filter(event => {
				const d = new Date(event.start_time);
				return d.getFullYear() === currentYear && d.getMonth() === currentMonth && d.getDate() === day;
			});
			days.push({ day, events: dayEvents, isToday });
		}

		return days;
	}

	function prevMonth() {
		if (currentMonth === 0) { currentMonth = 11; currentYear--; }
		else { currentMonth--; }
		loadEvents();
	}

	function nextMonth() {
		if (currentMonth === 11) { currentMonth = 0; currentYear++; }
		else { currentMonth++; }
		loadEvents();
	}

	function goToToday() {
		const now = new Date();
		currentYear = now.getFullYear();
		currentMonth = now.getMonth();
		loadEvents();
	}

	function openCreateModal() {
		editingEvent = null;
		const now = new Date();
		const later = new Date(now.getTime() + 3600000);
		formData = {
			title: '', description: '', location: '',
			start_time: now.toISOString().slice(0, 16),
			end_time: later.toISOString().slice(0, 16),
			all_day: false, recurring: 'none', event_type: 'other', color: '', room_id: ''
		};
		showCreateModal = true;
		loadRooms();
	}

	function openEditModal(event: CalendarEvent) {
		editingEvent = event;
		formData = {
			title: event.title,
			description: event.description || '',
			location: event.location || '',
			start_time: new Date(event.start_time).toISOString().slice(0, 16),
			end_time: new Date(event.end_time).toISOString().slice(0, 16),
			all_day: event.all_day,
			recurring: event.recurring || 'none',
			event_type: event.event_type || 'other',
			color: event.color || '',
			room_id: event.room_id || ''
		};
		showDetailModal = false;
		showCreateModal = true;
		loadRooms();
	}

	function openDetailModal(event: CalendarEvent) {
		selectedEvent = event;
		showDetailModal = true;
	}

	async function loadRooms() {
		if (!formData.start_time || !formData.end_time) return;
		try {
			const start = new Date(formData.start_time).toISOString();
			const end = new Date(formData.end_time).toISOString();
			availableRooms = await api(`/api/events/available-rooms?start=${start}&end=${end}`) || [];
		} catch { availableRooms = []; }
	}

	function onTypeChange() {
		const t = eventTypes.find(t => t.value === formData.event_type);
		if (t) formData.color = t.color;
	}

	async function saveEvent() {
		const url = editingEvent ? `/api/events/${editingEvent.id}` : '/api/events';
		const method = editingEvent ? 'PUT' : 'POST';

		try {
			await api(url, {
				method,
				body: JSON.stringify({
					...formData,
					start_time: new Date(formData.start_time).toISOString(),
					end_time: new Date(formData.end_time).toISOString(),
					color: formData.color || getTypeColor(formData.event_type),
					room_id: formData.room_id || null
				})
			});
			showCreateModal = false;
			loadEvents();
		} catch (error) {
			console.error('Failed to save event:', error);
		}
	}

	async function deleteEvent(id: string) {
		if (!confirm('Delete this event?')) return;
		try {
			await api(`/api/events/${id}`, { method: 'DELETE' });
			showDetailModal = false;
			showCreateModal = false;
			loadEvents();
		} catch (error) {
			console.error('Failed to delete event:', error);
		}
	}

	function formatTime(dateStr: string) {
		return new Date(dateStr).toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
	}

	function formatFullDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' });
	}

	function recurringLabel(r: string) {
		return { weekly: 'Weekly', monthly: 'Monthly', none: 'One-time' }[r] || r;
	}

	$: days = getDaysInMonth();
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-[var(--text-primary)]">📅 Calendar</h1>
		<button on:click={openCreateModal} class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 font-medium">
			+ New Event
		</button>
	</div>

	<!-- Toolbar -->
	<div class="bg-[var(--surface)] border border-[var(--border)] rounded-lg p-4 flex flex-wrap justify-between items-center gap-4">
		<div class="flex items-center gap-2">
			<!-- View toggle -->
			<div class="flex rounded-lg overflow-hidden border border-[var(--border)]">
				<button class="px-4 py-2 text-sm font-medium transition-colors {viewMode === 'month' ? 'bg-[var(--teal)] text-white' : 'bg-[var(--surface)] text-[var(--text-secondary)] hover:text-[var(--text-primary)]'}" on:click={() => viewMode = 'month'}>
					Month
				</button>
				<button class="px-4 py-2 text-sm font-medium transition-colors {viewMode === 'list' ? 'bg-[var(--teal)] text-white' : 'bg-[var(--surface)] text-[var(--text-secondary)] hover:text-[var(--text-primary)]'}" on:click={() => viewMode = 'list'}>
					List
				</button>
			</div>

			<!-- Type filter -->
			<select bind:value={typeFilter} on:change={loadEvents} class="px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--surface)] text-[var(--text-primary)] text-sm">
				<option value="">All Types</option>
				{#each eventTypes as t}
					<option value={t.value}>{t.label}</option>
				{/each}
			</select>
		</div>

		<!-- Month navigation -->
		<div class="flex items-center gap-3">
			<button on:click={prevMonth} class="p-2 rounded-lg border border-[var(--border)] bg-[var(--surface)] text-[var(--text-primary)] hover:bg-[var(--surface-hover)]">◀</button>
			<span class="font-semibold text-[var(--text-primary)] min-w-[160px] text-center">{monthNames[currentMonth]} {currentYear}</span>
			<button on:click={nextMonth} class="p-2 rounded-lg border border-[var(--border)] bg-[var(--surface)] text-[var(--text-primary)] hover:bg-[var(--surface-hover)]">▶</button>
			<button on:click={goToToday} class="px-3 py-2 text-sm rounded-lg border border-[var(--border)] bg-[var(--surface)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--surface-hover)]">Today</button>
		</div>

		<!-- Type legend -->
		<div class="flex items-center gap-3 text-xs">
			{#each eventTypes as t}
				<span class="flex items-center gap-1">
					<span class="w-3 h-3 rounded-full inline-block" style="background-color: {t.color}"></span>
					{t.label}
				</span>
			{/each}
		</div>
	</div>

	<!-- Calendar Grid -->
	{#if viewMode === 'month'}
		<div class="bg-[var(--surface)] border border-[var(--border)] rounded-lg overflow-hidden">
			<div class="grid grid-cols-7">
				{#each ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'] as dayName}
					<div class="px-3 py-2 text-center text-sm font-semibold text-[var(--text-secondary)] bg-[var(--surface-hover)] border-b border-[var(--border)]">{dayName}</div>
				{/each}

				{#each days as dayData, i}
					<div class="min-h-[110px] p-1.5 border-b border-r border-[var(--border)] {!dayData.day ? 'bg-[var(--bg)]' : ''} {dayData.isToday ? 'bg-[var(--teal)]/5' : ''}">
						{#if dayData.day}
							<div class="text-sm font-medium mb-1 {dayData.isToday ? 'text-[var(--teal)] font-bold' : 'text-[var(--text-primary)]'}">{dayData.day}</div>
							{#each dayData.events.slice(0, 3) as event}
								<button class="block w-full text-left px-1.5 py-0.5 mb-0.5 rounded text-xs text-white truncate cursor-pointer hover:opacity-80 transition-opacity" style="background-color: {event.color || getTypeColor(event.event_type)}" on:click={() => openDetailModal(event)}>
									{event.title}
								</button>
							{/each}
							{#if dayData.events.length > 3}
								<div class="text-xs text-[var(--text-secondary)] pl-1">+{dayData.events.length - 3} more</div>
							{/if}
						{/if}
					</div>
				{/each}
			</div>
		</div>

	{:else}
		<!-- List View -->
		<div class="space-y-3">
			{#if events.length === 0}
				<div class="bg-[var(--surface)] border border-[var(--border)] rounded-lg p-12 text-center">
					<div class="text-4xl mb-3">📅</div>
					<h3 class="text-lg font-semibold text-[var(--text-primary)] mb-2">No events this month</h3>
					<p class="text-[var(--text-secondary)] mb-4">Create your first event to get started.</p>
					<button on:click={openCreateModal} class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90">+ New Event</button>
				</div>
			{:else}
				{#each events as event}
					<div class="bg-[var(--surface)] border border-[var(--border)] rounded-lg p-4 flex gap-4 hover:border-[var(--teal)] transition-colors cursor-pointer" on:click={() => openDetailModal(event)}>
						<div class="w-1 rounded-full flex-shrink-0" style="background-color: {event.color || getTypeColor(event.event_type)}"></div>
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2 mb-1">
								<h3 class="font-semibold text-[var(--text-primary)] truncate">{event.title}</h3>
								<span class="px-2 py-0.5 text-xs rounded-full text-white flex-shrink-0" style="background-color: {getTypeColor(event.event_type)}">{getTypeLabel(event.event_type)}</span>
								{#if event.recurring !== 'none'}
									<span class="text-xs text-[var(--text-secondary)]">🔄 {recurringLabel(event.recurring)}</span>
								{/if}
							</div>
							<p class="text-sm text-[var(--text-secondary)]">
								{formatDate(event.start_time)} · {formatTime(event.start_time)} – {formatTime(event.end_time)}
							</p>
							{#if event.location}
								<p class="text-sm text-[var(--text-secondary)] mt-0.5">📍 {event.location}</p>
							{/if}
							{#if event.room_name}
								<p class="text-sm text-[var(--text-secondary)] mt-0.5">🏠 {event.room_name}</p>
							{/if}
						</div>
					</div>
				{/each}
			{/if}
		</div>
	{/if}
</div>

<!-- Event Detail Modal -->
{#if showDetailModal && selectedEvent}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" on:click={() => showDetailModal = false}>
		<div class="bg-[var(--surface)] rounded-xl p-6 w-full max-w-lg border border-[var(--border)] shadow-xl" on:click|stopPropagation>
			<div class="flex items-start justify-between mb-4">
				<div>
					<div class="flex items-center gap-2 mb-1">
						<span class="w-3 h-3 rounded-full" style="background-color: {selectedEvent.color || getTypeColor(selectedEvent.event_type)}"></span>
						<span class="text-xs font-medium px-2 py-0.5 rounded-full text-white" style="background-color: {getTypeColor(selectedEvent.event_type)}">{getTypeLabel(selectedEvent.event_type)}</span>
						{#if selectedEvent.recurring !== 'none'}
							<span class="text-xs text-[var(--text-secondary)]">🔄 {recurringLabel(selectedEvent.recurring)}</span>
						{/if}
					</div>
					<h2 class="text-xl font-bold text-[var(--text-primary)]">{selectedEvent.title}</h2>
				</div>
				<button on:click={() => showDetailModal = false} class="text-[var(--text-secondary)] hover:text-[var(--text-primary)] text-xl">✕</button>
			</div>

			<div class="space-y-3 text-sm">
				<div class="flex items-center gap-2 text-[var(--text-secondary)]">
					<span>📅</span>
					<span>{formatFullDate(selectedEvent.start_time)}</span>
				</div>
				<div class="flex items-center gap-2 text-[var(--text-secondary)]">
					<span>🕐</span>
					<span>{formatTime(selectedEvent.start_time)} – {formatTime(selectedEvent.end_time)}</span>
				</div>
				{#if selectedEvent.location}
					<div class="flex items-center gap-2 text-[var(--text-secondary)]">
						<span>📍</span>
						<span>{selectedEvent.location}</span>
					</div>
				{/if}
				{#if selectedEvent.room_name}
					<div class="flex items-center gap-2 text-[var(--text-secondary)]">
						<span>🏠</span>
						<span>{selectedEvent.room_name}</span>
					</div>
				{/if}
				{#if selectedEvent.description}
					<div class="pt-3 border-t border-[var(--border)]">
						<p class="text-[var(--text-primary)] whitespace-pre-wrap">{selectedEvent.description}</p>
					</div>
				{/if}
			</div>

			<div class="flex gap-2 justify-end mt-6 pt-4 border-t border-[var(--border)]">
				<button on:click={() => deleteEvent(selectedEvent.id)} class="px-4 py-2 text-sm text-red-500 hover:bg-red-500/10 rounded-lg transition-colors">Delete</button>
				<button on:click={() => openEditModal(selectedEvent)} class="px-4 py-2 text-sm bg-[var(--teal)] text-white rounded-lg hover:opacity-90">Edit</button>
			</div>
		</div>
	</div>
{/if}

<!-- Create/Edit Event Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" on:click={() => showCreateModal = false}>
		<div class="bg-[var(--surface)] rounded-xl p-6 w-full max-w-lg border border-[var(--border)] shadow-xl max-h-[90vh] overflow-y-auto" on:click|stopPropagation>
			<h2 class="text-xl font-bold text-[var(--text-primary)] mb-4">{editingEvent ? 'Edit Event' : 'New Event'}</h2>
			<form on:submit|preventDefault={saveEvent} class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Title *</label>
					<input type="text" bind:value={formData.title} required class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]" />
				</div>

				<div>
					<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Event Type</label>
					<select bind:value={formData.event_type} on:change={onTypeChange} class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]">
						{#each eventTypes as t}
							<option value={t.value}>{t.label}</option>
						{/each}
					</select>
				</div>

				<div>
					<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Description</label>
					<textarea bind:value={formData.description} rows="3" class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]"></textarea>
				</div>

				<div>
					<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Location</label>
					<input type="text" bind:value={formData.location} class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]" />
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Start *</label>
						<input type="datetime-local" bind:value={formData.start_time} on:change={loadRooms} required class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]" />
					</div>
					<div>
						<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">End *</label>
						<input type="datetime-local" bind:value={formData.end_time} on:change={loadRooms} required class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]" />
					</div>
				</div>

				<div class="flex items-center gap-4">
					<label class="flex items-center gap-2 text-sm text-[var(--text-primary)]">
						<input type="checkbox" bind:checked={formData.all_day} class="rounded" />
						All Day
					</label>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Recurring</label>
						<select bind:value={formData.recurring} class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]">
							<option value="none">None</option>
							<option value="weekly">Weekly</option>
							<option value="monthly">Monthly</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Color</label>
						<input type="color" bind:value={formData.color} class="w-full h-10 rounded-lg border border-[var(--border)] cursor-pointer" />
					</div>
				</div>

				{#if availableRooms && availableRooms.length > 0}
					<div>
						<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Room</label>
						<select bind:value={formData.room_id} class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]">
							<option value="">No room</option>
							{#each availableRooms as room}
								<option value={room.id}>{room.name}{room.capacity ? ` (capacity: ${room.capacity})` : ''}</option>
							{/each}
						</select>
					</div>
				{/if}

				<div class="flex gap-2 justify-end pt-4 border-t border-[var(--border)]">
					{#if editingEvent}
						<button type="button" on:click={() => deleteEvent(editingEvent.id)} class="px-4 py-2 text-sm text-red-500 hover:bg-red-500/10 rounded-lg mr-auto">Delete</button>
					{/if}
					<button type="button" on:click={() => showCreateModal = false} class="px-4 py-2 text-sm border border-[var(--border)] rounded-lg text-[var(--text-secondary)] hover:text-[var(--text-primary)]">Cancel</button>
					<button type="submit" class="px-6 py-2 text-sm bg-[var(--teal)] text-white rounded-lg hover:opacity-90 font-medium">Save</button>
				</div>
			</form>
		</div>
	</div>
{/if}
