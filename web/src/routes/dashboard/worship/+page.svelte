<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let plans = [];
	let services = [];
	let loading = false;
	let showCreateModal = false;
	let newPlan = {
		service_id: '',
		notes: ''
	};

	onMount(() => {
		loadPlans();
		loadServices();
	});

	async function loadPlans() {
		loading = true;
		try {
			plans = await api('/api/worship/plans');
		} catch (error) {
			console.error('Failed to load worship plans:', error);
		} finally {
			loading = false;
		}
	}

	async function loadServices() {
		try {
			const response = await api('/api/services?limit=50');
			services = response.services || [];
		} catch (error) {
			console.error('Failed to load services:', error);
		}
	}

	async function createPlan() {
		if (!newPlan.service_id) {
			alert('Please select a service');
			return;
		}

		try {
			const created = await api('/api/worship/plans', {
				method: 'POST',
				body: JSON.stringify(newPlan)
			});
			showCreateModal = false;
			newPlan = {
				service_id: '',
				notes: ''
			};
			goto(`/dashboard/worship/${created.id}`);
		} catch (error) {
			alert('Failed to create plan: ' + error.message);
		}
	}

	function viewPlan(id) {
		goto(`/dashboard/worship/${id}`);
	}

	function getStatusColor(status) {
		const colors = {
			draft: 'bg-yellow-100 text-yellow-800',
			published: 'bg-green-100 text-green-800'
		};
		return colors[status] || 'bg-gray-100 text-gray-800';
	}

	function getServiceName(serviceId) {
		const service = services.find(s => s.id === serviceId);
		return service ? service.name || service.service_type?.name : 'Unknown Service';
	}

	function formatDate(dateStr) {
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' });
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-navy">Worship Planning</h1>
		<button
			on:click={() => (showCreateModal = true)}
			class="px-4 py-2 bg-navy text-white rounded-md hover:bg-navy-dark"
		>
			New Service Plan
		</button>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-navy"></div>
		</div>
	{:else if plans.length === 0}
		<div class="bg-white rounded-lg p-12 text-center">
			<p class="text-gray-500 mb-4">No service plans yet.</p>
			<button
				on:click={() => (showCreateModal = true)}
				class="px-4 py-2 bg-navy text-white rounded-md hover:bg-navy-dark"
			>
				Create Your First Plan
			</button>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm overflow-hidden">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Service
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Status
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Created
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							Items
						</th>
						<th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
							Actions
						</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each plans as plan}
						<tr class="hover:bg-gray-50 cursor-pointer" on:click={() => viewPlan(plan.id)}>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm font-medium text-gray-900">
									{getServiceName(plan.service_id)}
								</div>
								{#if plan.notes}
									<div class="text-sm text-gray-500">{plan.notes}</div>
								{/if}
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full {getStatusColor(plan.status)}">
									{plan.status}
								</span>
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
								{formatDate(plan.created_at)}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
								{plan.items?.length || 0} items
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
								<button
									on:click|stopPropagation={() => viewPlan(plan.id)}
									class="text-navy hover:text-navy-dark"
								>
									Edit
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<!-- Create Plan Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<div class="bg-white rounded-lg p-6 max-w-md w-full">
			<h2 class="text-xl font-bold mb-4">Create Service Plan</h2>

			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Service</label>
					<select
						bind:value={newPlan.service_id}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-navy focus:border-navy"
					>
						<option value="">Select a service</option>
						{#each services as service}
							<option value={service.id}>
								{service.name || service.service_type?.name} - {formatDate(service.service_date)}
							</option>
						{/each}
					</select>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Notes (Optional)</label>
					<textarea
						bind:value={newPlan.notes}
						rows="3"
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-navy focus:border-navy"
						placeholder="Add any planning notes..."
					></textarea>
				</div>
			</div>

			<div class="flex justify-end gap-2 mt-6">
				<button
					on:click={() => (showCreateModal = false)}
					class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
				>
					Cancel
				</button>
				<button
					on:click={createPlan}
					class="px-4 py-2 bg-navy text-white rounded-md hover:bg-navy-dark"
				>
					Create Plan
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
	.hover\:text-navy-dark:hover {
		color: #1e40af;
	}
</style>
