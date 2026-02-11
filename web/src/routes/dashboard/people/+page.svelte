<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import Modal from '$lib/Modal.svelte';

	let people = [];
	let total = 0;
	let page = 1;
	let limit = 25;
	let searchQuery = '';
	let statusFilter = 'all';
	let sortBy = 'name';
	let loading = false;
	let showCreateModal = false;
	let selectedIds = new Set();
	let selectAll = false;
	let viewMode = 'cards'; // 'cards' or 'table'
	let searchTimeout;

	let newPerson = {
		first_name: '',
		last_name: '',
		email: '',
		phone: '',
		address_line1: '',
		address_line2: '',
		city: '',
		state: '',
		zip: '',
		birthdate: '',
		gender: '',
		membership_status: 'visitor',
		notes: ''
	};

	const statuses = ['all', 'active', 'member', 'visitor', 'inactive'];

	onMount(() => {
		loadPeople();
	});

	async function loadPeople() {
		loading = true;
		try {
			const params = new URLSearchParams({
				page: page.toString(),
				limit: limit.toString(),
				sort: sortBy
			});
			if (searchQuery) params.append('q', searchQuery);
			if (statusFilter !== 'all') params.append('status', statusFilter);

			const response = await api(`/api/people?${params}`);
			people = response.people || [];
			total = response.total || 0;
		} catch (error) {
			console.error('Failed to load people:', error);
		} finally {
			loading = false;
		}
	}

	function handleSearchInput() {
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => {
			page = 1;
			loadPeople();
		}, 300);
	}

	function handleFilterChange() {
		page = 1;
		selectedIds = new Set();
		selectAll = false;
		loadPeople();
	}

	function viewPerson(id) {
		goto(`/dashboard/people/${id}`);
	}

	function toggleSelect(id) {
		if (selectedIds.has(id)) {
			selectedIds.delete(id);
		} else {
			selectedIds.add(id);
		}
		selectedIds = selectedIds; // trigger reactivity
		selectAll = selectedIds.size === people.length;
	}

	function toggleSelectAll() {
		if (selectAll) {
			selectedIds = new Set();
		} else {
			selectedIds = new Set(people.map(p => p.id));
		}
		selectAll = !selectAll;
		selectedIds = selectedIds;
	}

	async function bulkChangeStatus(status) {
		if (selectedIds.size === 0) return;
		try {
			await api('/api/people/bulk/status', {
				method: 'POST',
				body: JSON.stringify({ person_ids: [...selectedIds], status })
			});
			selectedIds = new Set();
			selectAll = false;
			loadPeople();
		} catch (error) {
			alert('Failed to update: ' + error.message);
		}
	}

	async function exportCSV() {
		const ids = selectedIds.size > 0 ? [...selectedIds].join(',') : '';
		const params = new URLSearchParams();
		if (searchQuery) params.append('q', searchQuery);
		if (statusFilter !== 'all') params.append('status', statusFilter);
		if (ids) params.append('ids', ids);

		const token = localStorage.getItem('token');
		const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8190';
		const response = await fetch(`${API_URL}/api/people/export?${params}`, {
			headers: { Authorization: `Bearer ${token}` }
		});
		const blob = await response.blob();
		const url = window.URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = 'people.csv';
		a.click();
		window.URL.revokeObjectURL(url);
	}

	async function createPerson() {
		try {
			const body = { ...newPerson };
			if (!body.birthdate) delete body.birthdate;
			await api('/api/people', {
				method: 'POST',
				body: JSON.stringify(body)
			});
			showCreateModal = false;
			newPerson = {
				first_name: '', last_name: '', email: '', phone: '',
				address_line1: '', address_line2: '', city: '', state: '', zip: '',
				birthdate: '', gender: '', membership_status: 'visitor', notes: ''
			};
			loadPeople();
		} catch (error) {
			alert('Failed to create person: ' + error.message);
		}
	}

	function getInitials(first, last) {
		return ((first?.[0] || '') + (last?.[0] || '')).toUpperCase();
	}

	function getAvatarColor(name) {
		const colors = [
			'#4A8B8C', '#1B3A4B', '#8FBCB0', '#6366f1', '#ec4899',
			'#f59e0b', '#10b981', '#8b5cf6', '#ef4444', '#06b6d4'
		];
		let hash = 0;
		for (let i = 0; i < name.length; i++) {
			hash = name.charCodeAt(i) + ((hash << 5) - hash);
		}
		return colors[Math.abs(hash) % colors.length];
	}

	function getStatusBadge(status) {
		const map = {
			active: { bg: 'rgba(16, 185, 129, 0.15)', color: '#10b981', label: 'Active' },
			member: { bg: 'rgba(74, 139, 140, 0.15)', color: '#4A8B8C', label: 'Member' },
			visitor: { bg: 'rgba(59, 130, 246, 0.15)', color: '#3b82f6', label: 'Visitor' },
			inactive: { bg: 'rgba(107, 114, 128, 0.15)', color: '#6b7280', label: 'Inactive' }
		};
		return map[status] || map.inactive;
	}

	function formatPhone(phone) {
		if (!phone) return '';
		const cleaned = phone.replace(/\D/g, '');
		if (cleaned.length === 10) {
			return `(${cleaned.slice(0, 3)}) ${cleaned.slice(3, 6)}-${cleaned.slice(6)}`;
		}
		return phone;
	}
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
		<div>
			<h1 class="text-3xl font-bold text-primary">People</h1>
			<p class="text-secondary text-sm mt-1">{total} {total === 1 ? 'person' : 'people'} total</p>
		</div>
		<button
			on:click={() => (showCreateModal = true)}
			class="px-5 py-2.5 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 transition-opacity flex items-center gap-2 font-medium"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
			</svg>
			Add Person
		</button>
	</div>

	<!-- Search, Filter, Sort Bar -->
	<div class="bg-surface rounded-xl shadow-sm p-4 border border-custom">
		<div class="flex flex-col md:flex-row gap-3">
			<div class="flex-1 relative">
				<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
				<input
					type="search"
					bind:value={searchQuery}
					on:input={handleSearchInput}
					placeholder="Search by name, email, or phone..."
					class="w-full pl-10 pr-4 py-2.5 border input-border rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary text-sm"
				/>
			</div>
			<select
				bind:value={statusFilter}
				on:change={handleFilterChange}
				class="px-4 py-2.5 border input-border rounded-lg bg-[var(--input-bg)] text-primary text-sm focus:outline-none focus:ring-2 focus:ring-[var(--teal)]"
			>
				{#each statuses as s}
					<option value={s}>{s === 'all' ? 'All Statuses' : s.charAt(0).toUpperCase() + s.slice(1)}</option>
				{/each}
			</select>
			<select
				bind:value={sortBy}
				on:change={handleFilterChange}
				class="px-4 py-2.5 border input-border rounded-lg bg-[var(--input-bg)] text-primary text-sm focus:outline-none focus:ring-2 focus:ring-[var(--teal)]"
			>
				<option value="name">Name A→Z</option>
				<option value="name_desc">Name Z→A</option>
				<option value="newest">Newest First</option>
				<option value="oldest">Oldest First</option>
			</select>
			<div class="flex gap-1 border input-border rounded-lg overflow-hidden">
				<button
					on:click={() => viewMode = 'cards'}
					class="px-3 py-2 text-sm transition-colors {viewMode === 'cards' ? 'bg-[var(--teal)] text-white' : 'bg-[var(--input-bg)] text-secondary hover:text-primary'}"
					title="Card view"
				>
					<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M4 5h6v6H4V5zm0 8h6v6H4v-6zm8-8h6v6h-6V5zm0 8h6v6h-6v-6z"/></svg>
				</button>
				<button
					on:click={() => viewMode = 'table'}
					class="px-3 py-2 text-sm transition-colors {viewMode === 'table' ? 'bg-[var(--teal)] text-white' : 'bg-[var(--input-bg)] text-secondary hover:text-primary'}"
					title="Table view"
				>
					<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M3 4h18v2H3V4zm0 7h18v2H3v-2zm0 7h18v2H3v-2z"/></svg>
				</button>
			</div>
		</div>
	</div>

	<!-- Bulk Actions Bar -->
	{#if selectedIds.size > 0}
		<div class="bg-[var(--teal)] bg-opacity-10 border border-[var(--teal)] rounded-xl p-3 flex flex-wrap items-center gap-3">
			<span class="text-sm font-medium text-primary">{selectedIds.size} selected</span>
			<div class="flex gap-2">
				<button on:click={exportCSV} class="px-3 py-1.5 text-xs bg-surface border border-custom rounded-lg hover:bg-[var(--surface-hover)] text-primary">
					Export CSV
				</button>
				<select
					on:change={(e) => { if (e.target.value) { bulkChangeStatus(e.target.value); e.target.value = ''; }}}
					class="px-3 py-1.5 text-xs bg-surface border border-custom rounded-lg text-primary"
				>
					<option value="">Change Status...</option>
					<option value="active">Active</option>
					<option value="member">Member</option>
					<option value="visitor">Visitor</option>
					<option value="inactive">Inactive</option>
				</select>
			</div>
			<button on:click={() => { selectedIds = new Set(); selectAll = false; }} class="ml-auto text-xs text-secondary hover:text-primary">
				Clear selection
			</button>
		</div>
	{/if}

	<!-- People Grid / Table -->
	{#if loading}
		<div class="flex justify-center py-16">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else if people.length === 0}
		<div class="bg-surface rounded-xl shadow-sm border border-custom p-16 text-center">
			<svg class="w-16 h-16 mx-auto mb-4 text-secondary opacity-40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
			</svg>
			<p class="text-secondary text-lg">
				{#if searchQuery || statusFilter !== 'all'}No people match your filters.{:else}No people yet. Add your first person to get started.{/if}
			</p>
		</div>
	{:else if viewMode === 'cards'}
		<!-- Card View -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each people as person}
				{@const badge = getStatusBadge(person.membership_status)}
				<div
					class="bg-surface rounded-xl shadow-sm border border-custom p-5 hover:shadow-md hover:border-[var(--teal)] transition-all cursor-pointer relative group"
					on:click={() => viewPerson(person.id)}
					on:keydown={(e) => (e.key === 'Enter') && viewPerson(person.id)}
					tabindex="0"
					role="button"
				>
					<!-- Checkbox -->
					<div class="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity">
						<input
							type="checkbox"
							checked={selectedIds.has(person.id)}
							on:click|stopPropagation={() => toggleSelect(person.id)}
							on:change|stopPropagation
							class="w-4 h-4 rounded border-gray-300 text-[var(--teal)] focus:ring-[var(--teal)]"
						/>
					</div>

					<div class="flex items-start gap-4">
						<!-- Avatar -->
						<div
							class="w-12 h-12 rounded-full flex items-center justify-center text-white font-semibold text-sm flex-shrink-0"
							style="background-color: {getAvatarColor(person.first_name + person.last_name)}"
						>
							{getInitials(person.first_name, person.last_name)}
						</div>
						<div class="min-w-0 flex-1">
							<h3 class="font-semibold text-primary truncate">{person.first_name} {person.last_name}</h3>
							<span class="inline-block mt-1 px-2 py-0.5 text-xs font-medium rounded-full" style="background-color: {badge.bg}; color: {badge.color}">
								{badge.label}
							</span>
						</div>
					</div>

					<div class="mt-4 space-y-1.5">
						{#if person.email}
							<div class="flex items-center gap-2 text-sm text-secondary truncate">
								<svg class="w-3.5 h-3.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
								</svg>
								<span class="truncate">{person.email}</span>
							</div>
						{/if}
						{#if person.phone}
							<div class="flex items-center gap-2 text-sm text-secondary">
								<svg class="w-3.5 h-3.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
								</svg>
								<span>{formatPhone(person.phone)}</span>
							</div>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{:else}
		<!-- Table View -->
		<div class="bg-surface rounded-xl shadow-sm border border-custom overflow-hidden">
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-[var(--border)]">
					<thead class="bg-[var(--surface-hover)]">
						<tr>
							<th class="px-4 py-3 w-10">
								<input type="checkbox" checked={selectAll} on:change={toggleSelectAll} class="w-4 h-4 rounded" />
							</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">Name</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">Email</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">Phone</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">Status</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-[var(--border)]">
						{#each people as person}
							{@const badge = getStatusBadge(person.membership_status)}
							<tr
								class="hover:bg-[var(--surface-hover)] cursor-pointer transition-colors"
								on:click={() => viewPerson(person.id)}
							>
								<td class="px-4 py-3" on:click|stopPropagation>
									<input
										type="checkbox"
										checked={selectedIds.has(person.id)}
										on:change={() => toggleSelect(person.id)}
										class="w-4 h-4 rounded"
									/>
								</td>
								<td class="px-4 py-3 whitespace-nowrap">
									<div class="flex items-center gap-3">
										<div
											class="w-8 h-8 rounded-full flex items-center justify-center text-white font-medium text-xs flex-shrink-0"
											style="background-color: {getAvatarColor(person.first_name + person.last_name)}"
										>
											{getInitials(person.first_name, person.last_name)}
										</div>
										<span class="font-medium text-primary">{person.first_name} {person.last_name}</span>
									</div>
								</td>
								<td class="px-4 py-3 whitespace-nowrap text-sm text-secondary">{person.email || '—'}</td>
								<td class="px-4 py-3 whitespace-nowrap text-sm text-secondary">{formatPhone(person.phone) || '—'}</td>
								<td class="px-4 py-3 whitespace-nowrap">
									<span class="px-2 py-0.5 text-xs font-medium rounded-full" style="background-color: {badge.bg}; color: {badge.color}">
										{badge.label}
									</span>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}

	<!-- Pagination -->
	{#if total > limit}
		<div class="flex justify-center items-center gap-3">
			<button
				on:click={() => { page--; loadPeople(); }}
				disabled={page === 1}
				class="px-4 py-2 bg-surface border border-custom rounded-lg disabled:opacity-40 text-primary text-sm hover:bg-[var(--surface-hover)] transition-colors"
			>
				← Previous
			</button>
			<span class="text-sm text-secondary">
				Page {page} of {Math.ceil(total / limit)}
			</span>
			<button
				on:click={() => { page++; loadPeople(); }}
				disabled={page >= Math.ceil(total / limit)}
				class="px-4 py-2 bg-surface border border-custom rounded-lg disabled:opacity-40 text-primary text-sm hover:bg-[var(--surface-hover)] transition-colors"
			>
				Next →
			</button>
		</div>
	{/if}
</div>

<!-- Create Person Modal -->
<Modal show={showCreateModal} title="Add Person" onClose={() => (showCreateModal = false)}>
	<form on:submit|preventDefault={createPerson} class="space-y-4">
		<div class="grid grid-cols-2 gap-3">
			<div>
				<label for="firstName" class="block text-sm font-medium text-primary mb-1">First Name *</label>
				<input id="firstName" type="text" bind:value={newPerson.first_name} required
					class="w-full px-3 py-2 border input-border rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary text-sm" />
			</div>
			<div>
				<label for="lastName" class="block text-sm font-medium text-primary mb-1">Last Name *</label>
				<input id="lastName" type="text" bind:value={newPerson.last_name} required
					class="w-full px-3 py-2 border input-border rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary text-sm" />
			</div>
		</div>
		<div>
			<label for="email" class="block text-sm font-medium text-primary mb-1">Email</label>
			<input id="email" type="email" bind:value={newPerson.email}
				class="w-full px-3 py-2 border input-border rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary text-sm" />
		</div>
		<div class="grid grid-cols-2 gap-3">
			<div>
				<label for="phone" class="block text-sm font-medium text-primary mb-1">Phone</label>
				<input id="phone" type="tel" bind:value={newPerson.phone} placeholder="(555) 555-5555"
					class="w-full px-3 py-2 border input-border rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary text-sm" />
			</div>
			<div>
				<label for="gender" class="block text-sm font-medium text-primary mb-1">Gender</label>
				<select id="gender" bind:value={newPerson.gender}
					class="w-full px-3 py-2 border input-border rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary text-sm">
					<option value="">—</option>
					<option value="male">Male</option>
					<option value="female">Female</option>
				</select>
			</div>
		</div>
		<div>
			<label for="address" class="block text-sm font-medium text-primary mb-1">Address</label>
			<input id="address" type="text" bind:value={newPerson.address_line1} placeholder="Street address"
				class="w-full px-3 py-2 border input-border rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary text-sm" />
			<div class="grid grid-cols-3 gap-2 mt-2">
				<input type="text" bind:value={newPerson.city} placeholder="City"
					class="px-3 py-2 border input-border rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary text-sm" />
				<input type="text" bind:value={newPerson.state} placeholder="State"
					class="px-3 py-2 border input-border rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary text-sm" />
				<input type="text" bind:value={newPerson.zip} placeholder="ZIP"
					class="px-3 py-2 border input-border rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary text-sm" />
			</div>
		</div>
		<div class="grid grid-cols-2 gap-3">
			<div>
				<label for="birthdate" class="block text-sm font-medium text-primary mb-1">Birthdate</label>
				<input id="birthdate" type="date" bind:value={newPerson.birthdate}
					class="w-full px-3 py-2 border input-border rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary text-sm" />
			</div>
			<div>
				<label for="status" class="block text-sm font-medium text-primary mb-1">Status</label>
				<select id="status" bind:value={newPerson.membership_status}
					class="w-full px-3 py-2 border input-border rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary text-sm">
					<option value="visitor">Visitor</option>
					<option value="active">Active</option>
					<option value="member">Member</option>
					<option value="inactive">Inactive</option>
				</select>
			</div>
		</div>
		<div>
			<label for="notes" class="block text-sm font-medium text-primary mb-1">Notes</label>
			<textarea id="notes" bind:value={newPerson.notes} rows="2"
				class="w-full px-3 py-2 border input-border rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary text-sm" />
		</div>
		<div class="flex gap-3 pt-2">
			<button type="button" on:click={() => (showCreateModal = false)}
				class="flex-1 px-4 py-2.5 border border-custom rounded-lg hover:bg-[var(--surface-hover)] text-primary text-sm font-medium">
				Cancel
			</button>
			<button type="submit"
				class="flex-1 px-4 py-2.5 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 text-sm font-medium">
				Add Person
			</button>
		</div>
	</form>
</Modal>
