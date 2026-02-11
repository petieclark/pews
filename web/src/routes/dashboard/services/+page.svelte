<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let services = [];
	let upcomingServices = [];
	let serviceTypes = [];
	let loading = false;
	let showCreateModal = false;
	let newService = {
		service_type_id: '',
		service_date: '',
		service_time: '',
		status: 'planning'
	};

	onMount(() => {
		loadUpcomingServices();
		loadServiceTypes();
		loadAllServices();
	});

	async function loadUpcomingServices() {
		try {
			upcomingServices = await api('/api/services/upcoming?limit=4');
		} catch (error) {
			console.error('Failed to load upcoming services:', error);
		}
	}

	async function loadAllServices() {
		loading = true;
		try {
			const response = await api('/api/services?limit=20');
			services = response.services || [];
		} catch (error) {
			console.error('Failed to load services:', error);
		} finally {
			loading = false;
		}
	}

	async function loadServiceTypes() {
		try {
			serviceTypes = await api('/api/services/types');
		} catch (error) {
			console.error('Failed to load service types:', error);
		}
	}

	function viewService(id) {
		goto(`/dashboard/services/${id}`);
	}

	async function createService() {
		try {
			const created = await api('/api/services', {
				method: 'POST',
				body: JSON.stringify(newService)
			});
			showCreateModal = false;
			newService = {
				service_type_id: '',
				service_date: '',
				service_time: '',
				status: 'planning'
			};
			goto(`/dashboard/services/${created.id}`);
		} catch (error) {
			alert('Failed to create service: ' + error.message);
		}
	}

	function getStatusColor(status) {
		const colors = {
			planning: 'bg-blue-100 text-blue-800',
			confirmed: 'bg-green-100 text-green-800',
			completed: 'bg-gray-100 text-gray-800'
		};
		return colors[status] || 'bg-gray-100 text-gray-800';
	}

	function formatDate(dateStr) {
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' });
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-navy">Services</h1>
		<div class="flex gap-2">
			<button
				on:click={() => goto('/dashboard/services/songs')}
				class="px-4 py-2 bg-white border border-navy text-navy rounded-md hover:bg-gray-50"
			>
				Song Library
			</button>
			<button
				on:click={() => (showCreateModal = true)}
				class="px-4 py-2 bg-teal text-white rounded-md hover:bg-opacity-90"
			>
				Plan a Service
			</button>
		</div>
	</div>

	<!-- Upcoming Services -->
	{#if upcomingServices.length > 0}
		<div class="bg-white rounded-lg shadow p-6">
			<h2 class="text-xl font-semibold text-navy mb-4">Upcoming Services</h2>
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
				{#each upcomingServices as service}
					<div
						on:click={() => viewService(service.id)}
						class="border rounded-lg p-4 cursor-pointer hover:shadow-md transition-shadow"
						style="border-left: 4px solid {service.service_type?.color || '#4A8B8C'}"
					>
						<div class="text-sm text-gray-500">{formatDate(service.service_date)}</div>
						<div class="font-semibold text-navy mt-1">
							{service.service_type?.name || 'Service'}
						</div>
						{#if service.service_time}
							<div class="text-sm text-gray-600">{service.service_time}</div>
						{/if}
						<span
							class="inline-block mt-2 px-2 py-1 text-xs rounded {getStatusColor(service.status)}"
						>
							{service.status}
						</span>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- All Services List -->
	<div class="bg-white rounded-lg shadow overflow-hidden">
		<div class="p-4 border-b">
			<h2 class="text-lg font-semibold text-navy">All Services</h2>
		</div>
		{#if loading}
			<div class="p-8 text-center text-gray-500">Loading...</div>
		{:else if services.length === 0}
			<div class="p-8 text-center text-gray-500">
				No services found. Plan your first service to get started.
			</div>
		{:else}
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Time</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each services as service}
						<tr
							on:click={() => viewService(service.id)}
							class="hover:bg-gray-50 cursor-pointer"
						>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm font-medium text-gray-900">{formatDate(service.service_date)}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="flex items-center">
									<div
										class="w-2 h-8 rounded mr-2"
										style="background-color: {service.service_type?.color || '#4A8B8C'}"
									></div>
									<span class="text-sm text-gray-900">{service.service_type?.name || 'Service'}</span>
								</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-gray-500">{service.service_time || '—'}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full {getStatusColor(service.status)}">
									{service.status}
								</span>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>

<!-- Create service modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
		<div class="bg-white rounded-lg max-w-md w-full p-6">
			<h2 class="text-2xl font-bold text-navy mb-4">Plan a Service</h2>
			<form on:submit|preventDefault={createService} class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700">Service Type *</label>
					<select
						bind:value={newService.service_type_id}
						required
						class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					>
						<option value="">Select a type</option>
						{#each serviceTypes as type}
							<option value={type.id}>{type.name}</option>
						{/each}
					</select>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700">Date *</label>
					<input
						type="date"
						bind:value={newService.service_date}
						required
						class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700">Time</label>
					<input
						type="text"
						bind:value={newService.service_time}
						placeholder="10:30 AM"
						class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700">Status</label>
					<select
						bind:value={newService.status}
						class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
					>
						<option value="planning">Planning</option>
						<option value="confirmed">Confirmed</option>
						<option value="completed">Completed</option>
					</select>
				</div>
				<div class="flex gap-2 pt-4">
					<button
						type="button"
						on:click={() => (showCreateModal = false)}
						class="flex-1 px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="flex-1 px-4 py-2 bg-teal text-white rounded-md hover:bg-opacity-90"
					>
						Create
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
	:global(.border-navy) {
		border-color: #1b3a4b;
	}
</style>
