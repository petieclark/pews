<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	$: serviceId = $page.params.id;

	let service = null;
	let items = [];
	let team = [];
	let people = [];
	let songs = [];
	let loading = true;
	let saving = false;
	let showAddItemModal = false;
	let showAddTeamModal = false;
	let showSongSearch = false;
	let showEditModal = false;
	let showDeleteConfirm = false;
	let showSaveTemplateModal = false;
	let editingItemId = null;
	let editingItemNotes = '';
	let draggedIndex = null;
	let dragOverIndex = null;

	let templateName = '';
	let templateDesc = '';

	let newItem = {
		item_type: 'song',
		title: '',
		song_id: null,
		song_key: '',
		position: 0,
		duration_minutes: null,
		notes: '',
		assigned_to: ''
	};

	let newTeamMember = {
		person_id: '',
		role: '',
		status: 'pending'
	};

	let editForm = {
		service_type_id: '',
		service_date: '',
		service_time: '',
		notes: '',
		status: ''
	};

	let songSearch = '';

	const statusFlow = [
		{ value: 'draft', label: 'Draft', icon: '📝', next: 'planning' },
		{ value: 'planning', label: 'Planning', icon: '📋', next: 'rehearsal' },
		{ value: 'rehearsal', label: 'Rehearsal', icon: '🎵', next: 'ready' },
		{ value: 'ready', label: 'Ready', icon: '✅', next: 'completed' },
		{ value: 'completed', label: 'Completed', icon: '🏁', next: null }
	];

	const itemTypeConfig = {
		song: { icon: '🎵', label: 'Song', color: '#4A8B8C' },
		prayer: { icon: '🙏', label: 'Prayer', color: '#8FBCB0' },
		reading: { icon: '📖', label: 'Scripture', color: '#D4A574' },
		sermon: { icon: '📢', label: 'Sermon', color: '#1B3A4B' },
		announcement: { icon: '📣', label: 'Announcement', color: '#6B7280' },
		transition: { icon: '↔️', label: 'Transition', color: '#9CA3AF' },
		other: { icon: '•', label: 'Other', color: '#6B7280' }
	};

	const roles = ['Worship Leader', 'Vocalist', 'Guitarist', 'Bassist', 'Drummer', 'Keys', 'Sound Tech', 'Media/Slides', 'Camera', 'Usher', 'Greeter'];

	onMount(() => {
		loadService();
		loadPeople();
		loadSongs();
	});

	async function loadService() {
		loading = true;
		try {
			service = await api(`/api/services/${serviceId}`);
			items = service.items || [];
			team = service.team || [];
		} catch (error) {
			console.error('Failed to load service:', error);
			goto('/dashboard/services');
		} finally {
			loading = false;
		}
	}

	async function loadPeople() {
		try {
			const response = await api('/api/people?limit=1000');
			people = response.people || [];
		} catch (error) {
			console.error('Failed to load people:', error);
		}
	}

	async function loadSongs() {
		try {
			const response = await api('/api/services/songs?limit=1000');
			songs = response.songs || [];
		} catch (error) {
			console.error('Failed to load songs:', error);
		}
	}

	async function updateStatus(newStatus) {
		saving = true;
		try {
			await api(`/api/services/${serviceId}`, {
				method: 'PUT',
				body: JSON.stringify({
					service_type_id: service.service_type_id,
					service_date: service.service_date.split('T')[0],
					service_time: service.service_time,
					notes: service.notes,
					status: newStatus
				})
			});
			service.status = newStatus;
		} catch (error) {
			alert('Failed to update status');
		} finally {
			saving = false;
		}
	}

	function openEditModal() {
		editForm = {
			service_type_id: service.service_type_id,
			service_date: service.service_date.split('T')[0],
			service_time: service.service_time || '',
			notes: service.notes || '',
			status: service.status
		};
		showEditModal = true;
	}

	async function saveEdit() {
		try {
			await api(`/api/services/${serviceId}`, {
				method: 'PUT',
				body: JSON.stringify(editForm)
			});
			showEditModal = false;
			loadService();
		} catch (error) {
			alert('Failed to update service');
		}
	}

	async function deleteService() {
		try {
			await api(`/api/services/${serviceId}`, { method: 'DELETE' });
			goto('/dashboard/services');
		} catch (error) {
			alert('Failed to delete service');
		}
	}

	// Items
	async function addItem() {
		try {
			newItem.position = items.length + 1;
			await api(`/api/services/${serviceId}/items`, {
				method: 'POST',
				body: JSON.stringify(newItem)
			});
			showAddItemModal = false;
			resetNewItem();
			loadService();
		} catch (error) {
			alert('Failed to add item: ' + error.message);
		}
	}

	function resetNewItem() {
		newItem = {
			item_type: 'song',
			title: '',
			song_id: null,
			song_key: '',
			position: 0,
			duration_minutes: null,
			notes: '',
			assigned_to: ''
		};
	}

	async function deleteItem(itemId) {
		if (!confirm('Remove this item?')) return;
		try {
			await api(`/api/services/${serviceId}/items/${itemId}`, { method: 'DELETE' });
			loadService();
		} catch (error) {
			alert('Failed to delete item');
		}
	}

	async function saveItemNotes(itemId) {
		const item = items.find(i => i.id === itemId);
		if (!item) return;
		try {
			await api(`/api/services/${serviceId}/items/${itemId}`, {
				method: 'PUT',
				body: JSON.stringify({ ...item, notes: editingItemNotes })
			});
			editingItemId = null;
			loadService();
		} catch (error) {
			alert('Failed to save notes');
		}
	}

	// Drag and drop reorder
	function handleDragStart(index) {
		draggedIndex = index;
	}

	function handleDragOver(e, index) {
		e.preventDefault();
		dragOverIndex = index;
	}

	function handleDragEnd() {
		if (draggedIndex !== null && dragOverIndex !== null && draggedIndex !== dragOverIndex) {
			const reordered = [...items];
			const [moved] = reordered.splice(draggedIndex, 1);
			reordered.splice(dragOverIndex, 0, moved);
			items = reordered;

			// Save reorder
			const itemIds = items.map(i => i.id);
			api(`/api/services/${serviceId}/items/reorder`, {
				method: 'PUT',
				body: JSON.stringify({ item_ids: itemIds })
			}).catch(() => loadService());
		}
		draggedIndex = null;
		dragOverIndex = null;
	}

	// Song selection
	function selectSong(song) {
		newItem.song_id = song.id;
		newItem.title = song.title;
		newItem.song_key = song.default_key || '';
		showSongSearch = false;
		songSearch = '';
	}

	// Team
	async function addTeamMember() {
		try {
			await api(`/api/services/${serviceId}/team`, {
				method: 'POST',
				body: JSON.stringify(newTeamMember)
			});
			showAddTeamModal = false;
			newTeamMember = { person_id: '', role: '', status: 'pending' };
			loadService();
		} catch (error) {
			alert('Failed to add team member');
		}
	}

	async function updateTeamStatus(teamId, status) {
		try {
			const member = team.find(t => t.id === teamId);
			await api(`/api/services/${serviceId}/team/${teamId}`, {
				method: 'PUT',
				body: JSON.stringify({ ...member, status })
			});
			loadService();
		} catch (error) {
			alert('Failed to update status');
		}
	}

	async function removeTeamMember(teamId) {
		if (!confirm('Remove this team member?')) return;
		try {
			await api(`/api/services/${serviceId}/team/${teamId}`, { method: 'DELETE' });
			loadService();
		} catch (error) {
			alert('Failed to remove team member');
		}
	}

	// Templates
	async function saveAsTemplate() {
		try {
			await api(`/api/services/${serviceId}/save-template`, {
				method: 'POST',
				body: JSON.stringify({ name: templateName, description: templateDesc })
			});
			showSaveTemplateModal = false;
			templateName = '';
			templateDesc = '';
		} catch (error) {
			alert('Failed to save template');
		}
	}

	function formatDate(dateStr) {
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' });
	}

	function totalDuration() {
		return items.reduce((sum, i) => sum + (i.duration_minutes || 0), 0);
	}

	function getCurrentStatusStep() {
		return statusFlow.findIndex(s => s.value === service?.status);
	}

	$: filteredSongs = songSearch
		? songs.filter(s =>
			s.title.toLowerCase().includes(songSearch.toLowerCase()) ||
			(s.artist && s.artist.toLowerCase().includes(songSearch.toLowerCase())) ||
			(s.ccli_number && s.ccli_number.includes(songSearch))
		).slice(0, 50)
		: songs.slice(0, 50);

	$: currentStep = getCurrentStatusStep();
</script>

{#if loading}
	<div class="flex items-center justify-center h-64">
		<div class="inline-block w-8 h-8 border-2 border-[var(--teal)] border-t-transparent rounded-full animate-spin"></div>
	</div>
{:else if service}
	<div class="space-y-6 max-w-6xl mx-auto">
		<!-- Breadcrumb & Actions -->
		<div class="flex items-center justify-between">
			<button on:click={() => goto('/dashboard/services')} class="inline-flex items-center gap-1 text-sm text-[var(--teal)] hover:underline">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
				Services
			</button>
			<div class="flex items-center gap-2">
				<button on:click={() => (showSaveTemplateModal = true)} class="btn-ghost" title="Save as Template">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z"/></svg>
					Save as Template
				</button>
				<button on:click={openEditModal} class="btn-ghost">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
					Edit
				</button>
				<button on:click={() => (showDeleteConfirm = true)} class="btn-ghost text-red-400 hover:text-red-300">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
					Delete
				</button>
			</div>
		</div>

		<!-- Header Card -->
		<div class="card p-6">
			<div class="flex items-start justify-between">
				<div class="flex items-start gap-4">
					<div class="w-1.5 h-16 rounded-full flex-shrink-0" style="background-color: {service.service_type?.color || '#4A8B8C'}"></div>
					<div>
						<h1 class="text-2xl font-bold text-[var(--text-primary)]">
							{service.service_type?.name || 'Service'}
						</h1>
						<p class="text-[var(--text-secondary)] mt-1">
							{formatDate(service.service_date)}{service.service_time ? ` · ${service.service_time}` : ''}
						</p>
						{#if service.notes}
							<p class="text-[var(--text-secondary)] text-sm mt-2">{service.notes}</p>
						{/if}
					</div>
				</div>
			</div>

			<!-- Status Workflow -->
			<div class="mt-6 flex items-center gap-2">
				{#each statusFlow as step, i}
					<button
						on:click={() => updateStatus(step.value)}
						class="status-step {service.status === step.value ? 'active' : ''} {i <= currentStep ? 'passed' : ''}"
						disabled={saving}
					>
						<span class="text-sm">{step.icon}</span>
						<span class="text-xs font-medium">{step.label}</span>
					</button>
					{#if i < statusFlow.length - 1}
						<div class="flex-1 h-0.5 {i < currentStep ? 'bg-[var(--teal)]' : 'bg-[var(--border)]'} rounded"></div>
					{/if}
				{/each}
			</div>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Service Order -->
			<div class="lg:col-span-2 card">
				<div class="flex items-center justify-between p-5 border-b border-[var(--border)]">
					<div>
						<h2 class="text-lg font-semibold text-[var(--text-primary)]">Order of Service</h2>
						<p class="text-sm text-[var(--text-secondary)]">
							{items.length} item{items.length !== 1 ? 's' : ''}{totalDuration() > 0 ? ` · ~${totalDuration()} min` : ''}
						</p>
					</div>
					<button on:click={() => (showAddItemModal = true)} class="btn-primary text-sm">
						<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
						Add Item
					</button>
				</div>

				{#if items.length === 0}
					<div class="p-12 text-center">
						<p class="text-[var(--text-secondary)]">No items yet. Build your service order.</p>
					</div>
				{:else}
					<div class="divide-y divide-[var(--border)]">
						{#each items as item, index}
							<div
								class="item-row {dragOverIndex === index ? 'drag-over' : ''}"
								draggable="true"
								on:dragstart={() => handleDragStart(index)}
								on:dragover={(e) => handleDragOver(e, index)}
								on:dragend={handleDragEnd}
								on:drop={handleDragEnd}
							>
								<div class="flex items-start gap-3 px-5 py-3">
									<!-- Drag Handle -->
									<div class="flex-shrink-0 pt-1 cursor-grab text-[var(--text-secondary)] opacity-40 hover:opacity-100">
										<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><circle cx="9" cy="6" r="1.5"/><circle cx="15" cy="6" r="1.5"/><circle cx="9" cy="12" r="1.5"/><circle cx="15" cy="12" r="1.5"/><circle cx="9" cy="18" r="1.5"/><circle cx="15" cy="18" r="1.5"/></svg>
									</div>

									<!-- Position Number -->
									<span class="flex-shrink-0 w-6 h-6 rounded-full bg-[var(--surface-hover)] text-[var(--text-secondary)] text-xs flex items-center justify-center font-medium">
										{index + 1}
									</span>

									<!-- Type Icon -->
									<span class="text-lg flex-shrink-0" title={itemTypeConfig[item.item_type]?.label || item.item_type}>
										{itemTypeConfig[item.item_type]?.icon || '•'}
									</span>

									<!-- Content -->
									<div class="flex-1 min-w-0">
										<div class="flex items-center gap-2">
											<span class="font-medium text-[var(--text-primary)]">{item.title}</span>
											{#if item.item_type === 'song' && item.song}
												{#if item.song_key}
													<span class="text-xs px-1.5 py-0.5 rounded bg-[var(--surface-hover)] text-[var(--text-secondary)]">
														{item.song_key}
													</span>
												{/if}
												{#if item.song.tempo}
													<span class="text-xs text-[var(--text-secondary)]">{item.song.tempo} BPM</span>
												{/if}
												{#if item.song.ccli_number}
													<span class="text-xs text-[var(--text-secondary)]">CCLI {item.song.ccli_number}</span>
												{/if}
											{/if}
										</div>
										{#if item.assigned_to}
											<p class="text-xs text-[var(--text-secondary)] mt-0.5">👤 {item.assigned_to}</p>
										{/if}
										{#if item.duration_minutes}
											<p class="text-xs text-[var(--text-secondary)] mt-0.5">⏱ {item.duration_minutes} min</p>
										{/if}

										<!-- Notes -->
										{#if editingItemId === item.id}
											<div class="mt-2 flex gap-2">
												<input
													type="text"
													bind:value={editingItemNotes}
													on:keydown={(e) => e.key === 'Enter' && saveItemNotes(item.id)}
													class="input-field text-sm flex-1"
													placeholder="Add a note..."
												/>
												<button on:click={() => saveItemNotes(item.id)} class="btn-primary text-xs px-3">Save</button>
												<button on:click={() => (editingItemId = null)} class="btn-ghost text-xs">Cancel</button>
											</div>
										{:else if item.notes}
											<button
												on:click={() => { editingItemId = item.id; editingItemNotes = item.notes; }}
												class="text-xs text-[var(--text-secondary)] mt-1 hover:text-[var(--text-primary)] italic"
											>
												📝 {item.notes}
											</button>
										{:else}
											<button
												on:click={() => { editingItemId = item.id; editingItemNotes = ''; }}
												class="text-xs text-[var(--text-secondary)] mt-1 hover:text-[var(--teal)] opacity-0 group-hover:opacity-100"
											>
												+ Add note
											</button>
										{/if}
									</div>

									<!-- Actions -->
									<button
										on:click={() => deleteItem(item.id)}
										class="flex-shrink-0 p-1 rounded text-[var(--text-secondary)] hover:text-red-400 hover:bg-red-400/10"
										title="Remove item"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
									</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Sidebar: Team -->
			<div class="space-y-6">
				<div class="card">
					<div class="flex items-center justify-between p-4 border-b border-[var(--border)]">
						<h2 class="font-semibold text-[var(--text-primary)]">Team</h2>
						<button on:click={() => (showAddTeamModal = true)} class="btn-primary text-xs px-3 py-1.5">
							+ Add
						</button>
					</div>

					{#if team.length === 0}
						<div class="p-6 text-center text-sm text-[var(--text-secondary)]">
							No team members assigned
						</div>
					{:else}
						<div class="divide-y divide-[var(--border)]">
							{#each team as member}
								<div class="p-3">
									<div class="flex items-center justify-between">
										<div>
											<div class="font-medium text-sm text-[var(--text-primary)]">
												{member.person_first_name} {member.person_last_name}
											</div>
											<div class="text-xs text-[var(--text-secondary)]">{member.role}</div>
										</div>
										<div class="flex items-center gap-1">
											<button
												on:click={() => updateTeamStatus(member.id, 'accepted')}
												class="status-btn {member.status === 'accepted' ? 'status-accepted' : ''}"
												title="Accepted"
											>✓</button>
											<button
												on:click={() => updateTeamStatus(member.id, 'pending')}
												class="status-btn {member.status === 'pending' ? 'status-pending' : ''}"
												title="Pending"
											>?</button>
											<button
												on:click={() => updateTeamStatus(member.id, 'declined')}
												class="status-btn {member.status === 'declined' ? 'status-declined' : ''}"
												title="Declined"
											>✕</button>
											<button
												on:click={() => removeTeamMember(member.id)}
												class="ml-1 p-1 rounded text-[var(--text-secondary)] hover:text-red-400"
												title="Remove"
											>
												<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
											</button>
										</div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Quick Stats -->
				<div class="card p-4 space-y-3">
					<h3 class="font-semibold text-sm text-[var(--text-primary)]">Summary</h3>
					<div class="flex justify-between text-sm">
						<span class="text-[var(--text-secondary)]">Songs</span>
						<span class="text-[var(--text-primary)]">{items.filter(i => i.item_type === 'song').length}</span>
					</div>
					<div class="flex justify-between text-sm">
						<span class="text-[var(--text-secondary)]">Total Items</span>
						<span class="text-[var(--text-primary)]">{items.length}</span>
					</div>
					<div class="flex justify-between text-sm">
						<span class="text-[var(--text-secondary)]">Est. Duration</span>
						<span class="text-[var(--text-primary)]">{totalDuration() > 0 ? `${totalDuration()} min` : '—'}</span>
					</div>
					<div class="flex justify-between text-sm">
						<span class="text-[var(--text-secondary)]">Team Size</span>
						<span class="text-[var(--text-primary)]">{team.length}</span>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}

<!-- Add Item Modal -->
{#if showAddItemModal}
	<div class="modal-overlay" on:click|self={() => (showAddItemModal = false)}>
		<div class="modal-content max-w-lg">
			<div class="flex items-center justify-between mb-5">
				<h2 class="text-lg font-bold text-[var(--text-primary)]">Add Item</h2>
				<button on:click={() => (showAddItemModal = false)} class="close-btn">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
				</button>
			</div>

			<!-- Quick Type Selector -->
			<div class="grid grid-cols-4 gap-2 mb-5">
				{#each Object.entries(itemTypeConfig) as [key, config]}
					<button
						on:click={() => { newItem.item_type = key; if (key !== 'song') { newItem.song_id = null; newItem.song_key = ''; } }}
						class="type-select-btn {newItem.item_type === key ? 'active' : ''}"
					>
						<span class="text-lg">{config.icon}</span>
						<span class="text-xs">{config.label}</span>
					</button>
				{/each}
			</div>

			<form on:submit|preventDefault={addItem} class="space-y-4">
				{#if newItem.item_type === 'song'}
					<div>
						<label class="label">Song *</label>
						<div class="flex gap-2 mt-1">
							<input type="text" value={newItem.title} readonly placeholder="Select a song..." class="input-field flex-1 cursor-pointer" on:click={() => (showSongSearch = true)} />
							<button type="button" on:click={() => (showSongSearch = true)} class="btn-secondary">Browse</button>
						</div>
					</div>
					{#if newItem.song_id}
						<div>
							<label class="label">Key</label>
							<input type="text" bind:value={newItem.song_key} placeholder="G, C, Bb..." class="input-field" />
						</div>
					{/if}
				{:else}
					<div>
						<label class="label">Title *</label>
						<input type="text" bind:value={newItem.title} required class="input-field" placeholder="e.g. Opening Prayer, John 3:16..." />
					</div>
				{/if}

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="label">Assigned To</label>
						<input type="text" bind:value={newItem.assigned_to} placeholder="Person or role" class="input-field" />
					</div>
					<div>
						<label class="label">Duration (min)</label>
						<input type="number" bind:value={newItem.duration_minutes} min="1" class="input-field" />
					</div>
				</div>

				<div>
					<label class="label">Notes</label>
					<textarea bind:value={newItem.notes} rows="2" class="input-field" placeholder="Key of G, capo 2..."></textarea>
				</div>

				<div class="flex gap-3 pt-2">
					<button type="button" on:click={() => (showAddItemModal = false)} class="btn-secondary flex-1">Cancel</button>
					<button type="submit" class="btn-primary flex-1" disabled={newItem.item_type === 'song' && !newItem.song_id}>Add Item</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Song Search Modal -->
{#if showSongSearch}
	<div class="modal-overlay" on:click|self={() => { showSongSearch = false; songSearch = ''; }}>
		<div class="modal-content max-w-2xl">
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-lg font-bold text-[var(--text-primary)]">Select a Song</h2>
				<button on:click={() => { showSongSearch = false; songSearch = ''; }} class="close-btn">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
				</button>
			</div>
			<div class="relative mb-4">
				<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-secondary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>
				<input
					type="text"
					bind:value={songSearch}
					placeholder="Search by title, artist, or CCLI..."
					class="input-field pl-10"
					autofocus
				/>
			</div>
			<div class="space-y-1 max-h-96 overflow-y-auto">
				{#each filteredSongs as song}
					<button
						on:click={() => selectSong(song)}
						class="w-full text-left p-3 rounded-lg hover:bg-[var(--surface-hover)] transition-colors"
					>
						<div class="flex items-center justify-between">
							<div>
								<div class="font-medium text-[var(--text-primary)]">{song.title}</div>
								{#if song.artist}
									<div class="text-sm text-[var(--text-secondary)]">{song.artist}</div>
								{/if}
							</div>
							<div class="flex items-center gap-3 text-xs text-[var(--text-secondary)]">
								{#if song.default_key}
									<span class="px-1.5 py-0.5 rounded bg-[var(--surface-hover)]">{song.default_key}</span>
								{/if}
								{#if song.tempo}
									<span>{song.tempo} BPM</span>
								{/if}
								{#if song.ccli_number}
									<span>CCLI {song.ccli_number}</span>
								{/if}
							</div>
						</div>
					</button>
				{/each}
				{#if filteredSongs.length === 0}
					<p class="text-center text-[var(--text-secondary)] py-8">No songs found</p>
				{/if}
			</div>
		</div>
	</div>
{/if}

<!-- Add Team Member Modal -->
{#if showAddTeamModal}
	<div class="modal-overlay" on:click|self={() => (showAddTeamModal = false)}>
		<div class="modal-content max-w-md">
			<div class="flex items-center justify-between mb-5">
				<h2 class="text-lg font-bold text-[var(--text-primary)]">Add Team Member</h2>
				<button on:click={() => (showAddTeamModal = false)} class="close-btn">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
				</button>
			</div>
			<form on:submit|preventDefault={addTeamMember} class="space-y-4">
				<div>
					<label class="label">Person *</label>
					<select bind:value={newTeamMember.person_id} required class="input-field">
						<option value="">Select a person</option>
						{#each people as person}
							<option value={person.id}>{person.first_name} {person.last_name}</option>
						{/each}
					</select>
				</div>
				<div>
					<label class="label">Role *</label>
					<select bind:value={newTeamMember.role} required class="input-field">
						<option value="">Select a role</option>
						{#each roles as role}
							<option value={role}>{role}</option>
						{/each}
						<option value="__custom">Custom...</option>
					</select>
					{#if newTeamMember.role === '__custom'}
						<input
							type="text"
							bind:value={newTeamMember.role}
							placeholder="Enter custom role"
							class="input-field mt-2"
						/>
					{/if}
				</div>
				<div class="flex gap-3 pt-2">
					<button type="button" on:click={() => (showAddTeamModal = false)} class="btn-secondary flex-1">Cancel</button>
					<button type="submit" class="btn-primary flex-1">Add Member</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Edit Service Modal -->
{#if showEditModal}
	<div class="modal-overlay" on:click|self={() => (showEditModal = false)}>
		<div class="modal-content max-w-md">
			<h2 class="text-lg font-bold text-[var(--text-primary)] mb-5">Edit Service</h2>
			<form on:submit|preventDefault={saveEdit} class="space-y-4">
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="label">Date</label>
						<input type="date" bind:value={editForm.service_date} required class="input-field" />
					</div>
					<div>
						<label class="label">Time</label>
						<input type="text" bind:value={editForm.service_time} placeholder="10:30 AM" class="input-field" />
					</div>
				</div>
				<div>
					<label class="label">Notes</label>
					<textarea bind:value={editForm.notes} rows="3" class="input-field"></textarea>
				</div>
				<div class="flex gap-3 pt-2">
					<button type="button" on:click={() => (showEditModal = false)} class="btn-secondary flex-1">Cancel</button>
					<button type="submit" class="btn-primary flex-1">Save Changes</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Delete Confirmation -->
{#if showDeleteConfirm}
	<div class="modal-overlay" on:click|self={() => (showDeleteConfirm = false)}>
		<div class="modal-content max-w-sm text-center">
			<div class="text-4xl mb-3">⚠️</div>
			<h2 class="text-lg font-bold text-[var(--text-primary)] mb-2">Delete Service?</h2>
			<p class="text-sm text-[var(--text-secondary)] mb-5">This will permanently delete this service and all its items. This cannot be undone.</p>
			<div class="flex gap-3">
				<button on:click={() => (showDeleteConfirm = false)} class="btn-secondary flex-1">Cancel</button>
				<button on:click={deleteService} class="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 text-sm font-medium">Delete</button>
			</div>
		</div>
	</div>
{/if}

<!-- Save as Template Modal -->
{#if showSaveTemplateModal}
	<div class="modal-overlay" on:click|self={() => (showSaveTemplateModal = false)}>
		<div class="modal-content max-w-md">
			<h2 class="text-lg font-bold text-[var(--text-primary)] mb-5">Save as Template</h2>
			<form on:submit|preventDefault={saveAsTemplate} class="space-y-4">
				<div>
					<label class="label">Template Name *</label>
					<input type="text" bind:value={templateName} required class="input-field" placeholder="e.g. Standard Sunday Morning" />
				</div>
				<div>
					<label class="label">Description</label>
					<textarea bind:value={templateDesc} rows="2" class="input-field" placeholder="Optional description..."></textarea>
				</div>
				<div class="flex gap-3 pt-2">
					<button type="button" on:click={() => (showSaveTemplateModal = false)} class="btn-secondary flex-1">Cancel</button>
					<button type="submit" class="btn-primary flex-1">Save Template</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
	.card {
		background: var(--surface);
		border-radius: 0.75rem;
		border: 1px solid var(--border);
	}

	.btn-primary {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 0.5rem 1rem;
		background: var(--teal);
		color: white;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		transition: opacity 0.15s;
	}
	.btn-primary:hover { opacity: 0.9; }
	.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

	.btn-secondary {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 0.5rem 1rem;
		background: var(--surface);
		color: var(--text-primary);
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
	}
	.btn-secondary:hover { background: var(--surface-hover); }

	.btn-ghost {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.375rem 0.75rem;
		color: var(--text-secondary);
		border-radius: 0.375rem;
		font-size: 0.8125rem;
		font-weight: 500;
		transition: all 0.15s;
	}
	.btn-ghost:hover { background: var(--surface-hover); color: var(--text-primary); }

	.close-btn {
		padding: 0.25rem;
		border-radius: 0.375rem;
		color: var(--text-secondary);
	}
	.close-btn:hover { background: var(--surface-hover); }

	.input-field {
		display: block;
		width: 100%;
		padding: 0.5rem 0.75rem;
		background: var(--input-bg);
		color: var(--text-primary);
		border: 1px solid var(--input-border);
		border-radius: 0.5rem;
		font-size: 0.875rem;
		margin-top: 0.25rem;
	}
	.input-field:focus {
		outline: none;
		border-color: var(--teal);
		box-shadow: 0 0 0 2px rgba(74, 139, 140, 0.2);
	}

	.label {
		display: block;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text-primary);
	}

	.status-step {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.25rem;
		padding: 0.5rem 0.75rem;
		border-radius: 0.5rem;
		color: var(--text-secondary);
		transition: all 0.15s;
	}
	.status-step:hover { background: var(--surface-hover); }
	.status-step.active {
		background: rgba(74, 139, 140, 0.15);
		color: var(--teal);
	}
	.status-step.passed { color: var(--text-primary); }

	.item-row {
		transition: background 0.1s;
	}
	.item-row:hover {
		background: var(--surface-hover);
	}
	.item-row.drag-over {
		border-top: 2px solid var(--teal);
	}

	.type-select-btn {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.25rem;
		padding: 0.5rem;
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		color: var(--text-secondary);
		transition: all 0.15s;
	}
	.type-select-btn:hover { border-color: var(--teal); color: var(--text-primary); }
	.type-select-btn.active {
		border-color: var(--teal);
		background: rgba(74, 139, 140, 0.1);
		color: var(--teal);
	}

	.status-btn {
		width: 1.5rem;
		height: 1.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 0.25rem;
		font-size: 0.625rem;
		border: 1px solid var(--border);
		color: var(--text-secondary);
		transition: all 0.15s;
	}
	.status-btn:hover { border-color: var(--teal); }
	.status-accepted { background: rgba(34, 197, 94, 0.15); border-color: #22c55e; color: #4ADE80; }
	.status-pending { background: rgba(234, 179, 8, 0.15); border-color: #eab308; color: #FACC15; }
	.status-declined { background: rgba(239, 68, 68, 0.15); border-color: #ef4444; color: #F87171; }

	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.6);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
		z-index: 50;
		backdrop-filter: blur(4px);
	}
	.modal-content {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: 0.75rem;
		padding: 1.5rem;
		width: 100%;
		max-height: 90vh;
		overflow-y: auto;
	}
</style>
