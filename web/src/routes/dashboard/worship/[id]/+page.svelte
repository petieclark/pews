<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let planId = '';
	let plan = null;
	let items = [];
	let songs = [];
	let users = [];
	let loading = false;
	let showAddModal = false;
	let showExportView = false;
	let draggedItem = null;

	let newItem = {
		item_type: 'song',
		title: '',
		duration_minutes: null,
		notes: '',
		song_id: null,
		assigned_to: null
	};

	const itemTypes = [
		{ value: 'song', label: 'Song' },
		{ value: 'scripture', label: 'Scripture Reading' },
		{ value: 'prayer', label: 'Prayer' },
		{ value: 'announcement', label: 'Announcement' },
		{ value: 'video', label: 'Video' },
		{ value: 'other', label: 'Other' }
	];

	$: planId = $page.params.id;

	onMount(() => {
		loadPlan();
		loadSongs();
		loadUsers();
	});

	async function loadPlan() {
		loading = true;
		try {
			plan = await api(`/api/worship/plans/${planId}`);
			items = plan.items || [];
		} catch (error) {
			console.error('Failed to load plan:', error);
			alert('Plan not found');
			goto('/dashboard/worship');
		} finally {
			loading = false;
		}
	}

	async function loadSongs() {
		try {
			songs = await api('/api/services/songs');
		} catch (error) {
			console.error('Failed to load songs:', error);
		}
	}

	async function loadUsers() {
		try {
			// Assuming there's a users endpoint - adjust if needed
			users = await api('/api/people?limit=100');
		} catch (error) {
			console.error('Failed to load users:', error);
		}
	}

	async function addItem() {
		if (!newItem.title) {
			alert('Title is required');
			return;
		}

		try {
			const itemOrder = items.length > 0 ? Math.max(...items.map(i => i.item_order)) + 1 : 1;
			const created = await api(`/api/worship/plans/${planId}/items`, {
				method: 'POST',
				body: JSON.stringify({
					...newItem,
					item_order: itemOrder
				})
			});

			items = [...items, created];
			showAddModal = false;
			newItem = {
				item_type: 'song',
				title: '',
				duration_minutes: null,
				notes: '',
				song_id: null,
				assigned_to: null
			};
		} catch (error) {
			alert('Failed to add item: ' + error.message);
		}
	}

	async function updateItem(item) {
		try {
			await api(`/api/worship/plans/${planId}/items/${item.id}`, {
				method: 'PUT',
				body: JSON.stringify({
					item_order: item.item_order,
					item_type: item.item_type,
					title: item.title,
					duration_minutes: item.duration_minutes,
					notes: item.notes,
					song_id: item.song_id,
					assigned_to: item.assigned_to
				})
			});
		} catch (error) {
			alert('Failed to update item: ' + error.message);
		}
	}

	async function deleteItem(itemId) {
		if (!confirm('Are you sure you want to delete this item?')) return;

		try {
			await api(`/api/worship/plans/${planId}/items/${itemId}`, {
				method: 'DELETE'
			});
			items = items.filter(i => i.id !== itemId);
		} catch (error) {
			alert('Failed to delete item: ' + error.message);
		}
	}

	async function publishPlan() {
		if (!confirm('Publish this plan? It will be visible to the team.')) return;

		try {
			plan = await api(`/api/worship/plans/${planId}/publish`, {
				method: 'POST'
			});
		} catch (error) {
			alert('Failed to publish plan: ' + error.message);
		}
	}

	async function updatePlanNotes() {
		try {
			plan = await api(`/api/worship/plans/${planId}`, {
				method: 'PUT',
				body: JSON.stringify({
					notes: plan.notes
				})
			});
		} catch (error) {
			alert('Failed to update notes: ' + error.message);
		}
	}

	function onSongSelect(e) {
		const songId = e.target.value;
		if (songId) {
			const song = songs.find(s => s.id === songId);
			if (song) {
				newItem.song_id = songId;
				newItem.title = song.title;
			}
		}
	}

	function getTotalDuration() {
		return items.reduce((sum, item) => sum + (item.duration_minutes || 0), 0);
	}

	// Drag and drop
	function onDragStart(e, item) {
		draggedItem = item;
		e.dataTransfer.effectAllowed = 'move';
	}

	function onDragOver(e) {
		e.preventDefault();
		e.dataTransfer.dropEffect = 'move';
	}

	async function onDrop(e, targetItem) {
		e.preventDefault();
		if (!draggedItem || draggedItem.id === targetItem.id) return;

		const draggedIdx = items.findIndex(i => i.id === draggedItem.id);
		const targetIdx = items.findIndex(i => i.id === targetItem.id);

		// Reorder items
		const newItems = [...items];
		newItems.splice(draggedIdx, 1);
		newItems.splice(targetIdx, 0, draggedItem);

		// Update order numbers
		newItems.forEach((item, idx) => {
			item.item_order = idx + 1;
		});

		items = newItems;

		// Save each item's new order
		for (const item of items) {
			await updateItem(item);
		}

		draggedItem = null;
	}

	function exportPlan() {
		window.open(`/api/worship/plans/${planId}/export`, '_blank');
	}
</script>

<div class="space-y-6">
	{#if loading}
		<div class="flex justify-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-navy"></div>
		</div>
	{:else if plan}
		<div class="flex justify-between items-center">
			<div>
				<button
					on:click={() => goto('/dashboard/worship')}
					class="text-gray-600 hover:text-gray-900 mb-2"
				>
					← Back to Plans
				</button>
				<h1 class="text-3xl font-bold text-navy">Service Plan Builder</h1>
			</div>
			<div class="flex gap-2">
				<button
					on:click={exportPlan}
					class="px-4 py-2 bg-white border border-navy text-navy rounded-md hover:bg-gray-50"
				>
					Export / Print
				</button>
				{#if plan.status === 'draft'}
					<button
						on:click={publishPlan}
						class="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
					>
						Publish Plan
					</button>
				{:else}
					<span class="px-4 py-2 bg-green-100 text-green-800 rounded-md font-medium">
						✓ Published
					</span>
				{/if}
			</div>
		</div>

		<!-- Plan Info -->
		<div class="bg-white rounded-lg p-6 shadow-sm">
			<div class="flex justify-between items-start mb-4">
				<div class="flex-1">
					<label class="block text-sm font-medium text-gray-700 mb-1">Plan Notes</label>
					<textarea
						bind:value={plan.notes}
						on:blur={updatePlanNotes}
						rows="2"
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-navy focus:border-navy"
						placeholder="Add overall notes for this service plan..."
					></textarea>
				</div>
				<div class="ml-6 text-right">
					<div class="text-sm text-gray-500">Total Duration</div>
					<div class="text-3xl font-bold text-navy">{getTotalDuration()}</div>
					<div class="text-sm text-gray-500">minutes</div>
				</div>
			</div>
		</div>

		<!-- Items List -->
		<div class="bg-white rounded-lg shadow-sm">
			<div class="p-6 border-b border-gray-200 flex justify-between items-center">
				<h2 class="text-xl font-bold">Service Order</h2>
				<button
					on:click={() => (showAddModal = true)}
					class="px-4 py-2 bg-navy text-white rounded-md hover:bg-navy-dark"
				>
					+ Add Item
				</button>
			</div>

			{#if items.length === 0}
				<div class="p-12 text-center text-gray-500">
					<p class="mb-4">No items in this plan yet.</p>
					<button
						on:click={() => (showAddModal = true)}
						class="px-4 py-2 bg-navy text-white rounded-md hover:bg-navy-dark"
					>
						Add First Item
					</button>
				</div>
			{:else}
				<div class="divide-y divide-gray-200">
					{#each items as item (item.id)}
						<div
							draggable="true"
							on:dragstart={(e) => onDragStart(e, item)}
							on:dragover={onDragOver}
							on:drop={(e) => onDrop(e, item)}
							class="p-6 hover:bg-gray-50 cursor-move"
						>
							<div class="flex items-start gap-4">
								<div class="text-2xl font-bold text-gray-400 w-8">
									{item.item_order}
								</div>
								<div class="flex-1">
									<div class="flex items-start justify-between">
										<div class="flex-1">
											<div class="flex items-center gap-2 mb-2">
												<span class="px-2 py-1 bg-blue-100 text-blue-800 text-xs font-medium rounded">
													{item.item_type}
												</span>
												<h3 class="text-lg font-semibold">{item.title}</h3>
											</div>
											{#if item.notes}
												<p class="text-sm text-gray-600 mb-2">{item.notes}</p>
											{/if}
											<div class="flex gap-4 text-sm text-gray-500">
												{#if item.duration_minutes}
													<span>⏱️ {item.duration_minutes} min</span>
												{/if}
												{#if item.assigned_to_name}
													<span>👤 {item.assigned_to_name}</span>
												{/if}
												{#if item.song_title}
													<span>🎵 {item.song_title}</span>
												{/if}
											</div>
										</div>
										<button
											on:click={() => deleteItem(item.id)}
											class="text-red-600 hover:text-red-800 ml-4"
										>
											Delete
										</button>
									</div>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Add Item Modal -->
{#if showAddModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 overflow-y-auto">
		<div class="bg-white rounded-lg p-6 max-w-2xl w-full m-4">
			<h2 class="text-xl font-bold mb-4">Add Service Item</h2>

			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Item Type</label>
					<select
						bind:value={newItem.item_type}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-navy focus:border-navy"
					>
						{#each itemTypes as type}
							<option value={type.value}>{type.label}</option>
						{/each}
					</select>
				</div>

				{#if newItem.item_type === 'song'}
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Select Song (Optional)</label>
						<select
							on:change={onSongSelect}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-navy focus:border-navy"
						>
							<option value="">Select from song library...</option>
							{#each songs as song}
								<option value={song.id}>{song.title} - {song.artist || 'Unknown Artist'}</option>
							{/each}
						</select>
					</div>
				{/if}

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Title</label>
					<input
						type="text"
						bind:value={newItem.title}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-navy focus:border-navy"
						placeholder="Enter title..."
					/>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Duration (minutes)</label>
						<input
							type="number"
							bind:value={newItem.duration_minutes}
							min="0"
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-navy focus:border-navy"
							placeholder="0"
						/>
					</div>

					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Assigned To</label>
						<select
							bind:value={newItem.assigned_to}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-navy focus:border-navy"
						>
							<option value={null}>Not assigned</option>
							{#each users as user}
								<option value={user.id}>{user.first_name} {user.last_name}</option>
							{/each}
						</select>
					</div>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Notes</label>
					<textarea
						bind:value={newItem.notes}
						rows="3"
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-navy focus:border-navy"
						placeholder="e.g., key of G, start soft, repeat chorus..."
					></textarea>
				</div>
			</div>

			<div class="flex justify-end gap-2 mt-6">
				<button
					on:click={() => (showAddModal = false)}
					class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
				>
					Cancel
				</button>
				<button
					on:click={addItem}
					class="px-4 py-2 bg-navy text-white rounded-md hover:bg-navy-dark"
				>
					Add Item
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.text-navy {
		color: #1e3a8a;
	}
	.bg-navy {
		background-color: #1e3a8a;
	}
	.hover\:bg-navy-dark:hover {
		background-color: #1e40af;
	}
	.border-navy {
		border-color: #1e3a8a;
	}
	.ring-navy {
		--tw-ring-color: #1e3a8a;
	}
</style>
