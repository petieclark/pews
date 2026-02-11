<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let services = [];
	let upcomingServices = [];
	let serviceTypes = [];
	let templates = [];
	let loading = false;
	let showCreateModal = false;
	let showCopyModal = false;
	let copySourceId = '';
	let copyDate = '';

	// Filters
	let filterTypeId = '';
	let filterStatus = '';
	let filterFrom = '';
	let filterTo = '';
	let activeTab = 'all';

	let newService = {
		service_type_id: '',
		service_date: '',
		service_time: '',
		status: 'draft',
		template_id: ''
	};

	const statuses = [
		{ value: 'draft', label: 'Draft', color: 'status-draft' },
		{ value: 'planning', label: 'Planning', color: 'status-planning' },
		{ value: 'rehearsal', label: 'Rehearsal', color: 'status-rehearsal' },
		{ value: 'ready', label: 'Ready', color: 'status-ready' },
		{ value: 'completed', label: 'Completed', color: 'status-completed' }
	];

	onMount(() => {
		loadUpcomingServices();
		loadServiceTypes();
		loadAllServices();
		loadTemplates();
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
			let url = '/api/services?limit=50';
			if (filterTypeId) url += `&type_id=${filterTypeId}`;
			if (filterStatus) url += `&status=${filterStatus}`;
			if (filterFrom) url += `&from=${filterFrom}`;
			if (filterTo) url += `&to=${filterTo}`;
			const response = await api(url);
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

	async function loadTemplates() {
		try {
			templates = await api('/api/services/templates');
		} catch (error) {
			console.error('Failed to load templates:', error);
		}
	}

	function viewService(id) {
		goto(`/dashboard/services/${id}`);
	}

	function onTypeSelect(typeId) {
		if (typeId === filterTypeId) {
			filterTypeId = '';
		} else {
			filterTypeId = typeId;
		}
		activeTab = filterTypeId || 'all';
		loadAllServices();
	}

	function onServiceTypeChange() {
		const selectedType = serviceTypes.find(t => t.id === newService.service_type_id);
		if (selectedType) {
			if (selectedType.default_time && !newService.service_time) {
				newService.service_time = selectedType.default_time;
			}
			if (selectedType.default_day && !newService.service_date) {
				// Auto-fill next occurrence of that day
				const days = ['sunday', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'];
				const targetDay = days.indexOf(selectedType.default_day.toLowerCase());
				if (targetDay >= 0) {
					const today = new Date();
					const diff = (targetDay - today.getDay() + 7) % 7 || 7;
					const nextDate = new Date(today);
					nextDate.setDate(today.getDate() + diff);
					newService.service_date = nextDate.toISOString().split('T')[0];
				}
			}
		}
	}

	async function createService() {
		try {
			const payload = {
				service_type_id: newService.service_type_id,
				service_date: newService.service_date,
				service_time: newService.service_time,
				status: newService.status
			};
			const created = await api('/api/services', {
				method: 'POST',
				body: JSON.stringify(payload)
			});

			// If template selected, apply it (add items from template)
			if (newService.template_id) {
				try {
					const template = await api(`/api/services/templates/${newService.template_id}`);
					const data = typeof template.template_data === 'string' ? JSON.parse(template.template_data) : template.template_data;
					if (data.items) {
						for (const item of data.items) {
							await api(`/api/services/${created.id}/items`, {
								method: 'POST',
								body: JSON.stringify(item)
							});
						}
					}
				} catch (e) {
					console.error('Failed to apply template:', e);
				}
			}

			showCreateModal = false;
			newService = { service_type_id: '', service_date: '', service_time: '', status: 'draft', template_id: '' };
			goto(`/dashboard/services/${created.id}`);
		} catch (error) {
			alert('Failed to create service: ' + error.message);
		}
	}

	async function copyService() {
		try {
			const copied = await api(`/api/services/${copySourceId}/copy`, {
				method: 'POST',
				body: JSON.stringify({ service_date: copyDate })
			});
			showCopyModal = false;
			copySourceId = '';
			copyDate = '';
			goto(`/dashboard/services/${copied.id}`);
		} catch (error) {
			alert('Failed to copy service: ' + error.message);
		}
	}

	function openCopyModal(e, serviceId) {
		e.stopPropagation();
		copySourceId = serviceId;
		copyDate = '';
		showCopyModal = true;
	}

	function getStatusBadge(status) {
		const s = statuses.find(s => s.value === status);
		return s || { label: status, color: 'status-draft' };
	}

	function formatDate(dateStr) {
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' });
	}

	function formatDateShort(dateStr) {
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	function getDaysUntil(dateStr) {
		const date = new Date(dateStr);
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		date.setHours(0, 0, 0, 0);
		const diff = Math.ceil((date - today) / (1000 * 60 * 60 * 24));
		if (diff === 0) return 'Today';
		if (diff === 1) return 'Tomorrow';
		return `${diff} days`;
	}

	function clearFilters() {
		filterTypeId = '';
		filterStatus = '';
		filterFrom = '';
		filterTo = '';
		activeTab = 'all';
		loadAllServices();
	}

	$: hasFilters = filterTypeId || filterStatus || filterFrom || filterTo;
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex justify-between items-center">
		<div>
			<h1 class="text-3xl font-bold text-[var(--text-primary)]">Services</h1>
			<p class="text-[var(--text-secondary)] mt-1">Plan and manage your worship services</p>
		</div>
		<div class="flex gap-2">
			<button
				on:click={() => goto('/dashboard/services/templates')}
				class="btn-secondary"
			>
				<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z"/></svg>
				Templates
			</button>
			<button
				on:click={() => goto('/dashboard/services/songs')}
				class="btn-secondary"
			>
				<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3"/></svg>
				Song Library
			</button>
			<button
				on:click={() => (showCreateModal = true)}
				class="btn-primary"
			>
				<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
				Plan a Service
			</button>
		</div>
	</div>

	<!-- Upcoming Services -->
	{#if upcomingServices.length > 0}
		<div class="card p-6">
			<h2 class="text-lg font-semibold text-[var(--text-primary)] mb-4">Upcoming Services</h2>
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
				{#each upcomingServices as service}
					<button
						on:click={() => viewService(service.id)}
						class="upcoming-card text-left"
						style="border-left: 4px solid {service.service_type?.color || '#4A8B8C'}"
					>
						<div class="flex items-center justify-between mb-2">
							<span class="text-xs font-medium px-2 py-0.5 rounded-full" style="background: {service.service_type?.color || '#4A8B8C'}20; color: {service.service_type?.color || '#4A8B8C'}">
								{getDaysUntil(service.service_date)}
							</span>
							<span class="status-badge {getStatusBadge(service.status).color}">
								{getStatusBadge(service.status).label}
							</span>
						</div>
						<div class="font-semibold text-[var(--text-primary)]">
							{service.service_type?.name || 'Service'}
						</div>
						<div class="text-sm text-[var(--text-secondary)] mt-1">
							{formatDateShort(service.service_date)}{service.service_time ? ` · ${service.service_time}` : ''}
						</div>
					</button>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Type Tabs + Filters -->
	<div class="card overflow-hidden">
		<div class="border-b border-[var(--border)]">
			<div class="flex items-center justify-between px-4">
				<div class="flex gap-1 overflow-x-auto py-2">
					<button
						on:click={() => onTypeSelect('')}
						class="tab-btn {activeTab === 'all' ? 'tab-active' : ''}"
					>
						All
					</button>
					{#each serviceTypes.filter(t => t.is_active) as type}
						<button
							on:click={() => onTypeSelect(type.id)}
							class="tab-btn {activeTab === type.id ? 'tab-active' : ''}"
						>
							<span class="w-2 h-2 rounded-full inline-block mr-1.5" style="background-color: {type.color}"></span>
							{type.name}
						</button>
					{/each}
				</div>
				<div class="flex items-center gap-2 py-2 ml-4 flex-shrink-0">
					<select
						bind:value={filterStatus}
						on:change={loadAllServices}
						class="input-field text-sm py-1.5"
					>
						<option value="">All Statuses</option>
						{#each statuses as s}
							<option value={s.value}>{s.label}</option>
						{/each}
					</select>
					<input type="date" bind:value={filterFrom} on:change={loadAllServices} class="input-field text-sm py-1.5" placeholder="From" />
					<input type="date" bind:value={filterTo} on:change={loadAllServices} class="input-field text-sm py-1.5" placeholder="To" />
					{#if hasFilters}
						<button on:click={clearFilters} class="text-xs text-[var(--teal)] hover:underline whitespace-nowrap">Clear</button>
					{/if}
				</div>
			</div>
		</div>

		<!-- Service List -->
		{#if loading}
			<div class="p-12 text-center text-[var(--text-secondary)]">
				<div class="inline-block w-6 h-6 border-2 border-[var(--teal)] border-t-transparent rounded-full animate-spin"></div>
				<p class="mt-2">Loading services...</p>
			</div>
		{:else if services.length === 0}
			<div class="p-12 text-center">
				<svg class="w-16 h-16 mx-auto text-[var(--text-secondary)] opacity-30 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/></svg>
				<p class="text-[var(--text-secondary)] text-lg">No services found</p>
				<p class="text-[var(--text-secondary)] text-sm mt-1">Plan your first service to get started</p>
				<button on:click={() => (showCreateModal = true)} class="btn-primary mt-4">Plan a Service</button>
			</div>
		{:else}
			<div class="divide-y divide-[var(--border)]">
				{#each services as service}
					<button
						on:click={() => viewService(service.id)}
						class="service-row w-full text-left"
					>
						<div class="flex items-center gap-4 px-5 py-4">
							<div class="w-1 h-10 rounded-full flex-shrink-0" style="background-color: {service.service_type?.color || '#4A8B8C'}"></div>
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2">
									<span class="font-medium text-[var(--text-primary)]">{service.service_type?.name || 'Service'}</span>
									{#if service.name}
										<span class="text-[var(--text-secondary)]">— {service.name}</span>
									{/if}
								</div>
								<div class="text-sm text-[var(--text-secondary)] mt-0.5">
									{formatDate(service.service_date)}{service.service_time ? ` · ${service.service_time}` : ''}
								</div>
							</div>
							<span class="status-badge {getStatusBadge(service.status).color}">
								{getStatusBadge(service.status).label}
							</span>
							<span
								role="button"
								tabindex="0"
								on:click|stopPropagation={(e) => openCopyModal(e, service.id)}
								on:keydown={(e) => (e.key === 'Enter') && openCopyModal(e, service.id)}
								class="p-1.5 rounded hover:bg-[var(--surface-hover)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] cursor-pointer"
								title="Duplicate service"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/></svg>
							</span>
						</div>
					</button>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- Create Service Modal -->
{#if showCreateModal}
	<div class="modal-overlay" on:click|self={() => (showCreateModal = false)}>
		<div class="modal-content max-w-lg">
			<div class="flex items-center justify-between mb-6">
				<h2 class="text-xl font-bold text-[var(--text-primary)]">Plan a Service</h2>
				<button on:click={() => (showCreateModal = false)} class="p-1 rounded hover:bg-[var(--surface-hover)]">
					<svg class="w-5 h-5 text-[var(--text-secondary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
				</button>
			</div>
			<form on:submit|preventDefault={createService} class="space-y-4">
				<div>
					<label class="label">Service Type *</label>
					<select bind:value={newService.service_type_id} on:change={onServiceTypeChange} required class="input-field">
						<option value="">Select a type</option>
						{#each serviceTypes.filter(t => t.is_active) as type}
							<option value={type.id}>{type.name}</option>
						{/each}
					</select>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="label">Date *</label>
						<input type="date" bind:value={newService.service_date} required class="input-field" />
					</div>
					<div>
						<label class="label">Time</label>
						<input type="text" bind:value={newService.service_time} placeholder="10:30 AM" class="input-field" />
					</div>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="label">Status</label>
						<select bind:value={newService.status} class="input-field">
							{#each statuses as s}
								<option value={s.value}>{s.label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="label">Template</label>
						<select bind:value={newService.template_id} class="input-field">
							<option value="">None</option>
							{#each templates as t}
								<option value={t.id}>{t.name}</option>
							{/each}
						</select>
					</div>
				</div>
				<div class="flex gap-3 pt-4">
					<button type="button" on:click={() => (showCreateModal = false)} class="btn-secondary flex-1">Cancel</button>
					<button type="submit" class="btn-primary flex-1">Create Service</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Copy Service Modal -->
{#if showCopyModal}
	<div class="modal-overlay" on:click|self={() => (showCopyModal = false)}>
		<div class="modal-content max-w-sm">
			<h2 class="text-xl font-bold text-[var(--text-primary)] mb-4">Duplicate Service</h2>
			<form on:submit|preventDefault={copyService} class="space-y-4">
				<div>
					<label class="label">New Date *</label>
					<input type="date" bind:value={copyDate} required class="input-field" />
				</div>
				<div class="flex gap-3 pt-2">
					<button type="button" on:click={() => (showCopyModal = false)} class="btn-secondary flex-1">Cancel</button>
					<button type="submit" class="btn-primary flex-1">Duplicate</button>
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
		transition: background 0.15s;
	}
	.btn-secondary:hover { background: var(--surface-hover); }

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

	.tab-btn {
		padding: 0.5rem 0.75rem;
		font-size: 0.8125rem;
		font-weight: 500;
		color: var(--text-secondary);
		border-radius: 0.375rem;
		white-space: nowrap;
		display: inline-flex;
		align-items: center;
		transition: all 0.15s;
	}
	.tab-btn:hover { color: var(--text-primary); background: var(--surface-hover); }
	.tab-active {
		color: var(--teal) !important;
		background: rgba(74, 139, 140, 0.1);
	}

	.upcoming-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		padding: 1rem;
		transition: all 0.15s;
	}
	.upcoming-card:hover {
		border-color: var(--teal);
		box-shadow: 0 2px 8px rgba(0,0,0,0.1);
	}

	.service-row {
		transition: background 0.1s;
	}
	.service-row:hover {
		background: var(--surface-hover);
	}

	.status-badge {
		display: inline-flex;
		align-items: center;
		padding: 0.125rem 0.625rem;
		font-size: 0.75rem;
		font-weight: 600;
		border-radius: 9999px;
		text-transform: capitalize;
	}
	:global(.status-draft) { background: rgba(107, 114, 128, 0.15); color: #9CA3AF; }
	:global(.status-planning) { background: rgba(59, 130, 246, 0.15); color: #60A5FA; }
	:global(.status-rehearsal) { background: rgba(168, 85, 247, 0.15); color: #C084FC; }
	:global(.status-ready) { background: rgba(34, 197, 94, 0.15); color: #4ADE80; }
	:global(.status-completed) { background: rgba(107, 114, 128, 0.1); color: #6B7280; }

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
