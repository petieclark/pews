<script lang="ts">
	import { onMount } from 'svelte';

	interface Event {
		id: string;
		title: string;
		description?: string;
		location?: string;
		start_time: string;
		end_time: string;
		all_day: boolean;
		recurring: string;
		color: string;
	}

	let events: Event[] = [];
	let currentDate = new Date();
	let currentYear = currentDate.getFullYear();
	let currentMonth = currentDate.getMonth();
	let viewMode: 'month' | 'list' = 'month';
	let showModal = false;
	let editingEvent: Event | null = null;

	// Form data
	let formData = {
		title: '',
		description: '',
		location: '',
		start_time: '',
		end_time: '',
		all_day: false,
		recurring: 'none',
		color: '#4A8B8C'
	};

	const monthNames = [
		'January', 'February', 'March', 'April', 'May', 'June',
		'July', 'August', 'September', 'October', 'November', 'December'
	];

	onMount(() => {
		loadEvents();
	});

	async function loadEvents() {
		const firstDay = new Date(currentYear, currentMonth, 1);
		const lastDay = new Date(currentYear, currentMonth + 1, 0);
		
		const from = firstDay.toISOString().split('T')[0];
		const to = lastDay.toISOString().split('T')[0];

		try {
			const response = await fetch(`/api/events?from=${from}&to=${to}`, {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});

			if (response.ok) {
				const data = await response.json();
				events = data.events || [];
			}
		} catch (error) {
			console.error('Failed to load events:', error);
		}
	}

	function getDaysInMonth() {
		const firstDay = new Date(currentYear, currentMonth, 1);
		const lastDay = new Date(currentYear, currentMonth + 1, 0);
		const daysInMonth = lastDay.getDate();
		const startingDayOfWeek = firstDay.getDay();
		
		const days = [];
		
		// Add empty cells for days before the first day
		for (let i = 0; i < startingDayOfWeek; i++) {
			days.push({ day: null, events: [] });
		}
		
		// Add days of the month
		for (let day = 1; day <= daysInMonth; day++) {
			const date = new Date(currentYear, currentMonth, day);
			const dayEvents = events.filter(event => {
				const eventDate = new Date(event.start_time);
				return eventDate.getFullYear() === currentYear &&
				       eventDate.getMonth() === currentMonth &&
				       eventDate.getDate() === day;
			});
			days.push({ day, events: dayEvents });
		}
		
		return days;
	}

	function prevMonth() {
		if (currentMonth === 0) {
			currentMonth = 11;
			currentYear--;
		} else {
			currentMonth--;
		}
		loadEvents();
	}

	function nextMonth() {
		if (currentMonth === 11) {
			currentMonth = 0;
			currentYear++;
		} else {
			currentMonth++;
		}
		loadEvents();
	}

	function openCreateModal() {
		editingEvent = null;
		formData = {
			title: '',
			description: '',
			location: '',
			start_time: new Date().toISOString().slice(0, 16),
			end_time: new Date(Date.now() + 3600000).toISOString().slice(0, 16),
			all_day: false,
			recurring: 'none',
			color: '#4A8B8C'
		};
		showModal = true;
	}

	function openEditModal(event: Event) {
		editingEvent = event;
		formData = {
			title: event.title,
			description: event.description || '',
			location: event.location || '',
			start_time: new Date(event.start_time).toISOString().slice(0, 16),
			end_time: new Date(event.end_time).toISOString().slice(0, 16),
			all_day: event.all_day,
			recurring: event.recurring,
			color: event.color
		};
		showModal = true;
	}

	async function saveEvent() {
		const url = editingEvent ? `/api/events/${editingEvent.id}` : '/api/events';
		const method = editingEvent ? 'PUT' : 'POST';

		try {
			const response = await fetch(url, {
				method,
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				},
				body: JSON.stringify({
					...formData,
					start_time: new Date(formData.start_time).toISOString(),
					end_time: new Date(formData.end_time).toISOString()
				})
			});

			if (response.ok) {
				showModal = false;
				loadEvents();
			} else {
				alert('Failed to save event');
			}
		} catch (error) {
			console.error('Failed to save event:', error);
			alert('Failed to save event');
		}
	}

	async function deleteEvent(id: string) {
		if (!confirm('Delete this event?')) return;

		try {
			const response = await fetch(`/api/events/${id}`, {
				method: 'DELETE',
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});

			if (response.ok) {
				loadEvents();
			} else {
				alert('Failed to delete event');
			}
		} catch (error) {
			console.error('Failed to delete event:', error);
		}
	}

	function formatTime(dateStr: string) {
		const date = new Date(dateStr);
		return date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
	}

	function formatDate(dateStr: string) {
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
	}

	$: days = getDaysInMonth();
</script>

<div class="calendar-page">
	<header>
		<h1>📅 Calendar</h1>
		<div class="actions">
			<button on:click={openCreateModal} class="btn-primary">+ New Event</button>
		</div>
	</header>

	<div class="toolbar">
		<div class="view-toggle">
			<button class:active={viewMode === 'month'} on:click={() => viewMode = 'month'}>Month</button>
			<button class:active={viewMode === 'list'} on:click={() => viewMode = 'list'}>List</button>
		</div>
		
		<div class="month-nav">
			<button on:click={prevMonth}>◀</button>
			<span>{monthNames[currentMonth]} {currentYear}</span>
			<button on:click={nextMonth}>▶</button>
		</div>
	</div>

	{#if viewMode === 'month'}
		<div class="calendar-grid">
			<div class="day-header">Sun</div>
			<div class="day-header">Mon</div>
			<div class="day-header">Tue</div>
			<div class="day-header">Wed</div>
			<div class="day-header">Thu</div>
			<div class="day-header">Fri</div>
			<div class="day-header">Sat</div>

			{#each days as dayData}
				<div class="day-cell" class:empty={!dayData.day}>
					{#if dayData.day}
						<div class="day-number">{dayData.day}</div>
						{#each dayData.events as event}
							<button class="event-pill" style="background-color: {event.color}" on:click={() => openEditModal(event)}>
								{event.title}
							</button>
						{/each}
					{/if}
				</div>
			{/each}
		</div>
	{:else}
		<div class="list-view">
			{#if events.length === 0}
				<p class="empty">No events this month</p>
			{:else}
				{#each events as event}
					<div class="event-card">
						<div class="event-color" style="background-color: {event.color}"></div>
						<div class="event-info">
							<h3>{event.title}</h3>
							<p class="event-time">
								{formatDate(event.start_time)} at {formatTime(event.start_time)}
							</p>
							{#if event.location}
								<p class="event-location">📍 {event.location}</p>
							{/if}
							{#if event.description}
								<p class="event-desc">{event.description}</p>
							{/if}
						</div>
						<div class="event-actions">
							<button on:click={() => openEditModal(event)}>Edit</button>
							<button on:click={() => deleteEvent(event.id)} class="delete">Delete</button>
						</div>
					</div>
				{/each}
			{/if}
		</div>
	{/if}
</div>

{#if showModal}
	<div class="modal-overlay" on:click={() => showModal = false}>
		<div class="modal" on:click|stopPropagation>
			<h2>{editingEvent ? 'Edit Event' : 'New Event'}</h2>
			<form on:submit|preventDefault={saveEvent}>
				<label>
					Title *
					<input type="text" bind:value={formData.title} required />
				</label>

				<label>
					Description
					<textarea bind:value={formData.description} rows="3"></textarea>
				</label>

				<label>
					Location
					<input type="text" bind:value={formData.location} />
				</label>

				<label>
					Start Time *
					<input type="datetime-local" bind:value={formData.start_time} required />
				</label>

				<label>
					End Time *
					<input type="datetime-local" bind:value={formData.end_time} required />
				</label>

				<label>
					<input type="checkbox" bind:checked={formData.all_day} />
					All Day Event
				</label>

				<label>
					Recurring
					<select bind:value={formData.recurring}>
						<option value="none">None</option>
						<option value="weekly">Weekly</option>
						<option value="monthly">Monthly</option>
					</select>
				</label>

				<label>
					Color
					<input type="color" bind:value={formData.color} />
				</label>

				<div class="modal-actions">
					<button type="button" on:click={() => showModal = false}>Cancel</button>
					<button type="submit" class="btn-primary">Save</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
	.calendar-page {
		padding: 2rem;
	}

	header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
	}

	h1 {
		margin: 0;
		font-size: 2rem;
	}

	.toolbar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
	}

	.view-toggle button {
		padding: 0.5rem 1rem;
		border: 1px solid #ddd;
		background: white;
		cursor: pointer;
	}

	.view-toggle button.active {
		background: #4A8B8C;
		color: white;
		border-color: #4A8B8C;
	}

	.month-nav {
		display: flex;
		gap: 1rem;
		align-items: center;
	}

	.month-nav span {
		font-weight: 600;
		min-width: 150px;
		text-align: center;
	}

	.calendar-grid {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		gap: 1px;
		background: #ddd;
		border: 1px solid #ddd;
	}

	.day-header {
		background: #f5f5f5;
		padding: 0.75rem;
		font-weight: 600;
		text-align: center;
	}

	.day-cell {
		background: white;
		min-height: 100px;
		padding: 0.5rem;
	}

	.day-cell.empty {
		background: #fafafa;
	}

	.day-number {
		font-weight: 600;
		margin-bottom: 0.25rem;
	}

	.event-pill {
		display: block;
		width: 100%;
		padding: 0.25rem;
		margin-bottom: 0.25rem;
		border: none;
		border-radius: 3px;
		color: white;
		font-size: 0.75rem;
		text-align: left;
		cursor: pointer;
	}

	.list-view {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.event-card {
		display: flex;
		gap: 1rem;
		padding: 1rem;
		background: white;
		border: 1px solid #ddd;
		border-radius: 8px;
	}

	.event-color {
		width: 4px;
		border-radius: 2px;
	}

	.event-info {
		flex: 1;
	}

	.event-info h3 {
		margin: 0 0 0.5rem 0;
	}

	.event-time, .event-location, .event-desc {
		margin: 0.25rem 0;
		color: #666;
		font-size: 0.9rem;
	}

	.event-actions {
		display: flex;
		gap: 0.5rem;
		align-items: flex-start;
	}

	.btn-primary {
		background: #4A8B8C;
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 4px;
		cursor: pointer;
	}

	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal {
		background: white;
		padding: 2rem;
		border-radius: 8px;
		max-width: 500px;
		width: 90%;
		max-height: 90vh;
		overflow-y: auto;
	}

	.modal h2 {
		margin: 0 0 1.5rem 0;
	}

	form label {
		display: block;
		margin-bottom: 1rem;
		font-weight: 500;
	}

	form input[type="text"],
	form input[type="datetime-local"],
	form textarea,
	form select {
		display: block;
		width: 100%;
		padding: 0.5rem;
		margin-top: 0.25rem;
		border: 1px solid #ddd;
		border-radius: 4px;
	}

	form input[type="checkbox"] {
		margin-right: 0.5rem;
	}

	form input[type="color"] {
		width: 60px;
		height: 40px;
		border: 1px solid #ddd;
		border-radius: 4px;
		cursor: pointer;
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
		margin-top: 1.5rem;
	}

	button {
		padding: 0.5rem 1rem;
		border: 1px solid #ddd;
		background: white;
		border-radius: 4px;
		cursor: pointer;
	}

	button:hover {
		background: #f5f5f5;
	}

	button.delete {
		color: #e53e3e;
	}

	.empty {
		text-align: center;
		color: #999;
		padding: 2rem;
	}
</style>
