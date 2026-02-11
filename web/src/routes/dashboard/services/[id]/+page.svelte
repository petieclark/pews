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
	let showAddItemModal = false;
	let showAddTeamModal = false;
	let showSongSearchModal = false;
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
	let searchQuery = '';
	let searchResults = [];

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
			alert('Failed to load service');
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

	async function addItem() {
		try {
			newItem.position = items.length + 1;
			await api(`/api/services/${serviceId}/items`, {
				method: 'POST',
				body: JSON.stringify(newItem)
			});
			showAddItemModal = false;
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
			loadService();
		} catch (error) {
			alert('Failed to add item: ' + error.message);
		}
	}

	async function deleteItem(itemId) {
		if (!confirm('Delete this item?')) return;
		try {
			await api(`/api/services/${serviceId}/items/${itemId}`, {
				method: 'DELETE'
			});
			loadService();
		} catch (error) {
			alert('Failed to delete item: ' + error.message);
		}
	}

	async function moveItem(itemId, direction) {
		const index = items.findIndex((i) => i.id === itemId);
		if (index === -1) return;

		const newIndex = direction === 'up' ? index - 1 : index + 1;
		if (newIndex < 0 || newIndex >= items.length) return;

		const item = items[index];
		const otherItem = items[newIndex];

		// Update positions
		try {
			await api(`/api/services/${serviceId}/items/${item.id}`, {
				method: 'PUT',
				body: JSON.stringify({ ...item, position: newIndex + 1 })
			});
			await api(`/api/services/${serviceId}/items/${otherItem.id}`, {
				method: 'PUT',
				body: JSON.stringify({ ...otherItem, position: index + 1 })
			});
			loadService();
		} catch (error) {
			alert('Failed to reorder items: ' + error.message);
		}
	}

	async function addTeamMember() {
		try {
			await api(`/api/services/${serviceId}/team`, {
				method: 'POST',
				body: JSON.stringify(newTeamMember)
			});
			showAddTeamModal = false;
			newTeamMember = {
				person_id: '',
				role: '',
				status: 'pending'
			};
			loadService();
		} catch (error) {
			alert('Failed to add team member: ' + error.message);
		}
	}

	async function updateTeamStatus(teamId, status) {
		try {
			const member = team.find((t) => t.id === teamId);
			await api(`/api/services/${serviceId}/team/${teamId}`, {
				method: 'PUT',
				body: JSON.stringify({ ...member, status })
			});
			loadService();
		} catch (error) {
			alert('Failed to update team member: ' + error.message);
		}
	}

	async function removeTeamMember(teamId) {
		if (!confirm('Remove this team member?')) return;
		try {
			await api(`/api/services/${serviceId}/team/${teamId}`, {
				method: 'DELETE'
			});
			loadService();
		} catch (error) {
			alert('Failed to remove team member: ' + error.message);
		}
	}

	function selectSong(song) {
		newItem.song_id = song.id;
		newItem.title = song.title;
		newItem.song_key = song.default_key || '';
		showSongSearchModal = false;
		searchQuery = '';
	}

	function getItemTypeIcon(type) {
		const icons = {
			song: '♫',
			prayer: '🙏',
			reading: '📖',
			sermon: '📢',
			announcement: '📣',
			other: '•'
		};
		return icons[type] || '•';
	}

	function getStatusColor(status) {
		const colors = {
			pending: 'bg-yellow-100 text-yellow-800',
			accepted: 'bg-green-100 text-green-800',
			declined: 'bg-red-100 text-red-800'
		};
		return colors[status] || 'bg-[var(--surface-hover)] text-primary';
	}

	function formatDate(dateStr) {
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' });
	}

	$: filteredSongs = searchQuery
		? songs.filter((s) =>
				s.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
				(s.artist && s.artist.toLowerCase().includes(searchQuery.toLowerCase()))
		  )
		: songs;
</script>

{#if loading}
	<div class="p-8 text-center text-secondary">Loading...</div>
{:else if service}
	<div class="space-y-6">
		<!-- Header -->
		<div class="flex justify-between items-start">
			<div>
				<button
					on:click={() => goto('/dashboard/services')}
					class="text-teal hover:underline mb-2"
				>
					← Back to Services
				</button>
				<h1 class="text-3xl font-bold text-navy">
					{service.service_type?.name || 'Service'}
				</h1>
				<p class="text-secondary">{formatDate(service.service_date)} {service.service_time || ''}</p>
			</div>
			<div class="flex gap-2">
				<select
					bind:value={service.status}
					class="px-4 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md"
				>
					<option value="planning">Planning</option>
					<option value="confirmed">Confirmed</option>
					<option value="completed">Completed</option>
				</select>
			</div>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Order of Service -->
			<div class="lg:col-span-2 bg-surface rounded-lg shadow p-6">
				<div class="flex justify-between items-center mb-4">
					<h2 class="text-xl font-semibold text-navy">Order of Service</h2>
					<button
						on:click={() => (showAddItemModal = true)}
						class="px-4 py-2 bg-teal text-white rounded-md hover:bg-opacity-90 text-sm"
					>
						+ Add Item
					</button>
				</div>

				{#if items.length === 0}
					<p class="text-secondary text-center py-8">No items yet. Add your first item to get started.</p>
				{:else}
					<div class="space-y-2">
						{#each items as item, index}
							<div class="border rounded-lg p-4 hover:bg-[var(--surface-hover)]">
								<div class="flex items-start gap-3">
									<span class="text-2xl">{getItemTypeIcon(item.item_type)}</span>
									<div class="flex-1">
										<div class="font-medium text-navy">{item.title}</div>
										{#if item.song && item.song_key}
											<div class="text-sm text-secondary">Key: {item.song_key}</div>
										{/if}
										{#if item.assigned_to}
											<div class="text-sm text-secondary">Assigned: {item.assigned_to}</div>
										{/if}
										{#if item.duration_minutes}
											<div class="text-sm text-secondary">{item.duration_minutes} min</div>
										{/if}
									</div>
									<div class="flex flex-col gap-1">
										<button
											on:click={() => moveItem(item.id, 'up')}
											disabled={index === 0}
											class="px-2 py-1 text-xs border rounded disabled:opacity-30"
										>
											↑
										</button>
										<button
											on:click={() => moveItem(item.id, 'down')}
											disabled={index === items.length - 1}
											class="px-2 py-1 text-xs border rounded disabled:opacity-30"
										>
											↓
										</button>
										<button
											on:click={() => deleteItem(item.id)}
											class="px-2 py-1 text-xs border border-red-300 text-red-600 rounded hover:bg-red-50"
										>
											✕
										</button>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Team -->
			<div class="bg-surface rounded-lg shadow p-6">
				<div class="flex justify-between items-center mb-4">
					<h2 class="text-xl font-semibold text-navy">Team</h2>
					<button
						on:click={() => (showAddTeamModal = true)}
						class="px-3 py-1 bg-teal text-white rounded-md hover:bg-opacity-90 text-sm"
					>
						+ Add
					</button>
				</div>

				{#if team.length === 0}
					<p class="text-secondary text-sm">No team members assigned yet.</p>
				{:else}
					<div class="space-y-3">
						{#each team as member}
							<div class="border rounded p-3">
								<div class="font-medium text-navy text-sm">
									{member.person_first_name} {member.person_last_name}
								</div>
								<div class="text-xs text-secondary">{member.role}</div>
								<div class="mt-2 flex gap-1">
									<button
										on:click={() => updateTeamStatus(member.id, 'accepted')}
										class="px-2 py-1 text-xs border rounded {member.status === 'accepted' ? 'bg-green-100 border-green-300' : ''}"
									>
										✓
									</button>
									<button
										on:click={() => updateTeamStatus(member.id, 'pending')}
										class="px-2 py-1 text-xs border rounded {member.status === 'pending' ? 'bg-yellow-100 border-yellow-300' : ''}"
									>
										?
									</button>
									<button
										on:click={() => updateTeamStatus(member.id, 'declined')}
										class="px-2 py-1 text-xs border rounded {member.status === 'declined' ? 'bg-red-100 border-red-300' : ''}"
									>
										✕
									</button>
									<button
										on:click={() => removeTeamMember(member.id)}
										class="ml-auto px-2 py-1 text-xs text-red-600 border border-red-300 rounded hover:bg-red-50"
									>
										Remove
									</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}

<!-- Add Item Modal -->
{#if showAddItemModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
		<div class="bg-surface rounded-lg max-w-md w-full p-6 max-h-[90vh] overflow-y-auto">
			<h2 class="text-2xl font-bold text-navy mb-4">Add Item</h2>
			<form on:submit|preventDefault={addItem} class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-primary">Type *</label>
					<select
						bind:value={newItem.item_type}
						required
						class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					>
						<option value="song">Song</option>
						<option value="prayer">Prayer</option>
						<option value="reading">Scripture Reading</option>
						<option value="sermon">Sermon</option>
						<option value="announcement">Announcement</option>
						<option value="other">Other</option>
					</select>
				</div>

				{#if newItem.item_type === 'song'}
					<div>
						<label class="block text-sm font-medium text-primary">Song *</label>
						<div class="flex gap-2 mt-1">
							<input
								type="text"
								bind:value={newItem.title}
								placeholder="Select a song"
								readonly
								class="flex-1 px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md bg-[var(--surface-hover)]"
							/>
							<button
								type="button"
								on:click={() => (showSongSearchModal = true)}
								class="px-4 py-2 bg-navy text-white rounded-md hover:bg-opacity-90"
							>
								Browse
							</button>
						</div>
					</div>
					{#if newItem.song_id}
						<div>
							<label class="block text-sm font-medium text-primary">Key</label>
							<input
								type="text"
								bind:value={newItem.song_key}
								placeholder="G, C, Bb, etc."
								class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
							/>
						</div>
					{/if}
				{:else}
					<div>
						<label class="block text-sm font-medium text-primary">Title *</label>
						<input
							type="text"
							bind:value={newItem.title}
							required
							class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
						/>
					</div>
				{/if}

				<div>
					<label class="block text-sm font-medium text-primary">Assigned To</label>
					<input
						type="text"
						bind:value={newItem.assigned_to}
						placeholder="Person or role"
						class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-primary">Duration (minutes)</label>
					<input
						type="number"
						bind:value={newItem.duration_minutes}
						class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-primary">Notes</label>
					<textarea
						bind:value={newItem.notes}
						rows="2"
						class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					></textarea>
				</div>

				<div class="flex gap-2 pt-4">
					<button
						type="button"
						on:click={() => (showAddItemModal = false)}
						class="flex-1 px-4 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md hover:bg-[var(--surface-hover)]"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="flex-1 px-4 py-2 bg-teal text-white rounded-md hover:bg-opacity-90"
					>
						Add
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Song Search Modal -->
{#if showSongSearchModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
		<div class="bg-surface rounded-lg max-w-2xl w-full p-6 max-h-[90vh] overflow-y-auto">
			<h2 class="text-2xl font-bold text-navy mb-4">Select a Song</h2>
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search songs..."
				class="w-full px-4 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md mb-4"
			/>
			<div class="space-y-2 max-h-96 overflow-y-auto">
				{#each filteredSongs as song}
					<div
						on:click={() => selectSong(song)}
						class="p-3 border rounded hover:bg-[var(--surface-hover)] cursor-pointer"
					>
						<div class="font-medium">{song.title}</div>
						{#if song.artist}
							<div class="text-sm text-secondary">{song.artist}</div>
						{/if}
						{#if song.default_key}
							<div class="text-xs text-secondary">Key: {song.default_key}</div>
						{/if}
					</div>
				{/each}
			</div>
			<button
				on:click={() => { showSongSearchModal = false; searchQuery = ''; }}
				class="mt-4 w-full px-4 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md hover:bg-[var(--surface-hover)]"
			>
				Close
			</button>
		</div>
	</div>
{/if}

<!-- Add Team Member Modal -->
{#if showAddTeamModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
		<div class="bg-surface rounded-lg max-w-md w-full p-6">
			<h2 class="text-2xl font-bold text-navy mb-4">Add Team Member</h2>
			<form on:submit|preventDefault={addTeamMember} class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-primary">Person *</label>
					<select
						bind:value={newTeamMember.person_id}
						required
						class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					>
						<option value="">Select a person</option>
						{#each people as person}
							<option value={person.id}>{person.first_name} {person.last_name}</option>
						{/each}
					</select>
				</div>
				<div>
					<label class="block text-sm font-medium text-primary">Role *</label>
					<input
						type="text"
						bind:value={newTeamMember.role}
						required
						placeholder="Worship Leader, Keys, Sound, etc."
						class="mt-1 block w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					/>
				</div>
				<div class="flex gap-2 pt-4">
					<button
						type="button"
						on:click={() => (showAddTeamModal = false)}
						class="flex-1 px-4 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md hover:bg-[var(--surface-hover)]"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="flex-1 px-4 py-2 bg-teal text-white rounded-md hover:bg-opacity-90"
					>
						Add
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
	:global(.bg-navy) {
		background-color: #1b3a4b;
	}
	:global(.text-navy) {
		color: #1b3a4b;
	}
	:global(.bg-teal) {
		background-color: #4a8b8c;
	}
	:global(.text-teal) {
		color: #4a8b8c;
	}
	:global(.ring-teal) {
		--tw-ring-color: #4a8b8c;
	}
</style>
