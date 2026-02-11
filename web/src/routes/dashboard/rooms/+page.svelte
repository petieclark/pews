<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	let rooms = [];
	let loading = true;
	let showCreateModal = false;
	let newRoom = {
		name: '',
		capacity: null,
		description: '',
		color: '#3B82F6',
		amenities: []
	};
	let amenityInput = '';

	onMount(async () => {
		await loadRooms();
		loading = false;
	});

	async function loadRooms() {
		try {
			const response = await fetch('/api/rooms?active=true', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});
			if (response.ok) {
				rooms = await response.json();
			}
		} catch (error) {
			console.error('Failed to load rooms:', error);
		}
	}

	async function createRoom() {
		try {
			const response = await fetch('/api/rooms', {
				method: 'POST',
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`,
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(newRoom)
			});

			if (response.ok) {
				showCreateModal = false;
				resetForm();
				await loadRooms();
			} else {
				alert('Failed to create room');
			}
		} catch (error) {
			console.error('Failed to create room:', error);
			alert('Failed to create room');
		}
	}

	function resetForm() {
		newRoom = {
			name: '',
			capacity: null,
			description: '',
			color: '#3B82F6',
			amenities: []
		};
		amenityInput = '';
	}

	function addAmenity() {
		if (amenityInput.trim()) {
			newRoom.amenities = [...newRoom.amenities, amenityInput.trim()];
			amenityInput = '';
		}
	}

	function removeAmenity(index: number) {
		newRoom.amenities = newRoom.amenities.filter((_, i) => i !== index);
	}

	function viewRoom(roomId: string) {
		goto(`/dashboard/rooms/${roomId}`);
	}
</script>

<div class="p-6">
	<div class="flex justify-between items-center mb-6">
		<div>
			<h1 class="text-3xl font-bold">Rooms & Facilities</h1>
			<p class="text-gray-600 mt-1">Manage rooms and facility bookings</p>
		</div>
		<button
			on:click={() => showCreateModal = true}
			class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg"
		>
			+ New Room
		</button>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
		</div>
	{:else if rooms.length === 0}
		<div class="bg-white rounded-lg shadow p-12 text-center">
			<p class="text-gray-600 mb-4">No rooms created yet.</p>
			<button
				on:click={() => showCreateModal = true}
				class="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg"
			>
				Create Your First Room
			</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each rooms as room}
				<div
					class="bg-white rounded-lg shadow hover:shadow-lg transition-shadow cursor-pointer border-l-4"
					style="border-left-color: {room.color}"
					on:click={() => viewRoom(room.id)}
				>
					<div class="p-6">
						<div class="flex justify-between items-start mb-2">
							<h3 class="text-xl font-semibold">{room.name}</h3>
							{#if room.capacity}
								<span class="bg-gray-100 text-gray-700 px-2 py-1 rounded text-sm">
									{room.capacity} people
								</span>
							{/if}
						</div>
						
						{#if room.description}
							<p class="text-gray-600 text-sm mb-4">{room.description}</p>
						{/if}

						{#if room.amenities && room.amenities.length > 0}
							<div class="flex flex-wrap gap-2 mt-3">
								{#each room.amenities.slice(0, 3) as amenity}
									<span class="bg-blue-50 text-blue-700 px-2 py-1 rounded-full text-xs">
										{amenity}
									</span>
								{/each}
								{#if room.amenities.length > 3}
									<span class="text-gray-500 text-xs py-1">
										+{room.amenities.length - 3} more
									</span>
								{/if}
							</div>
						{/if}

						<div class="mt-4 pt-4 border-t">
							<button class="text-blue-600 hover:text-blue-800 text-sm font-medium">
								View Calendar →
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Create Room Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<div class="bg-white rounded-lg p-6 max-w-md w-full mx-4 max-h-[90vh] overflow-y-auto">
			<h2 class="text-2xl font-bold mb-4">Create New Room</h2>
			
			<form on:submit|preventDefault={createRoom}>
				<div class="space-y-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">
							Room Name *
						</label>
						<input
							type="text"
							bind:value={newRoom.name}
							required
							class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
							placeholder="e.g., Sanctuary, Fellowship Hall"
						/>
					</div>

					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">
							Capacity
						</label>
						<input
							type="number"
							bind:value={newRoom.capacity}
							class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
							placeholder="Number of people"
						/>
					</div>

					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">
							Description
						</label>
						<textarea
							bind:value={newRoom.description}
							rows="3"
							class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
							placeholder="Room description..."
						></textarea>
					</div>

					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">
							Color
						</label>
						<input
							type="color"
							bind:value={newRoom.color}
							class="w-full h-10 px-3 py-1 border rounded-lg cursor-pointer"
						/>
					</div>

					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">
							Amenities
						</label>
						<div class="flex gap-2 mb-2">
							<input
								type="text"
								bind:value={amenityInput}
								on:keypress={(e) => e.key === 'Enter' && (e.preventDefault(), addAmenity())}
								class="flex-1 px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
								placeholder="e.g., Projector, WiFi"
							/>
							<button
								type="button"
								on:click={addAmenity}
								class="bg-gray-200 hover:bg-gray-300 px-3 py-2 rounded-lg"
							>
								Add
							</button>
						</div>
						<div class="flex flex-wrap gap-2">
							{#each newRoom.amenities as amenity, i}
								<span class="bg-blue-100 text-blue-800 px-3 py-1 rounded-full text-sm flex items-center gap-2">
									{amenity}
									<button
										type="button"
										on:click={() => removeAmenity(i)}
										class="text-blue-600 hover:text-blue-800"
									>
										×
									</button>
								</span>
							{/each}
						</div>
					</div>
				</div>

				<div class="flex gap-3 mt-6">
					<button
						type="button"
						on:click={() => { showCreateModal = false; resetForm(); }}
						class="flex-1 px-4 py-2 border rounded-lg hover:bg-gray-50"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="flex-1 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg"
					>
						Create Room
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
