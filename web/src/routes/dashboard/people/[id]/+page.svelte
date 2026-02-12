<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let person = null;
	let loading = true;
	let editing = false;
	let editForm = {};
	let availableTags = [];
	let showAddTag = false;
	let selectedTagId = '';
	let engagementScore = null;
	let groups = [];
	let activityLog = [];
	let givingHistory = [];
	let activeTab = 'info';

	$: personId = $page.params.id;

	onMount(() => {
		loadPerson();
		loadAvailableTags();
		loadEngagementScore();
		loadGroups();
		loadActivity();
		loadGiving();
	});

	async function loadPerson() {
		loading = true;
		try {
			person = await api(`/api/people/${personId}`);
			editForm = { ...person };
		} catch (error) {
			console.error('Failed to load person:', error);
			goto('/dashboard/people');
		} finally {
			loading = false;
		}
	}

	async function loadAvailableTags() {
		try { availableTags = await api('/api/tags', { silent: true }); } catch (e) { }
	}

	async function loadEngagementScore() {
		try { engagementScore = await api(`/api/engagement/scores/${personId}`, { silent: true }); } catch (e) { engagementScore = null; }
	}

	async function loadGroups() {
		try { groups = await api(`/api/groups/person/${personId}`, { silent: true }); } catch (e) { groups = []; }
	}

	async function loadActivity() {
		try {
			const res = await api(`/api/activity?entity_type=people&limit=10`, { silent: true });
			activityLog = (res.data || []).filter(a => a.entity_id === personId);
		} catch (e) { activityLog = []; }
	}

	async function loadGiving() {
		try { givingHistory = await api(`/api/giving/person/${personId}`, { silent: true }); } catch (e) { givingHistory = []; }
	}

	async function recalculateEngagement() {
		try { engagementScore = await api(`/api/engagement/scores/${personId}/calculate`, { method: 'POST' }); } catch (e) { }
	}

	async function savePerson() {
		try {
			await api(`/api/people/${personId}`, { method: 'PUT', body: JSON.stringify(editForm) });
			editing = false;
			loadPerson();
		} catch (error) { alert('Failed to save: ' + error.message); }
	}

	async function deletePerson() {
		if (!confirm('Are you sure you want to delete this person?')) return;
		try {
			await api(`/api/people/${personId}`, { method: 'DELETE' });
			goto('/dashboard/people');
		} catch (error) { alert('Failed to delete: ' + error.message); }
	}

	async function addTag() {
		if (!selectedTagId) return;
		try {
			await api(`/api/people/${personId}/tags`, { method: 'POST', body: JSON.stringify({ tag_id: selectedTagId }) });
			showAddTag = false; selectedTagId = '';
			loadPerson();
		} catch (error) { alert('Failed to add tag: ' + error.message); }
	}

	async function removeTag(tagId) {
		try {
			await api(`/api/people/${personId}/tags/${tagId}`, { method: 'DELETE' });
			loadPerson();
		} catch (error) { alert('Failed to remove tag: ' + error.message); }
	}

	function getInitials(first, last) {
		return ((first?.[0] || '') + (last?.[0] || '')).toUpperCase();
	}

	function getAvatarColor(name) {
		const colors = ['#4A8B8C', '#1B3A4B', '#8FBCB0', '#6366f1', '#ec4899', '#f59e0b', '#10b981', '#8b5cf6', '#ef4444', '#06b6d4'];
		let hash = 0;
		for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash);
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
		if (cleaned.length === 10) return `(${cleaned.slice(0, 3)}) ${cleaned.slice(3, 6)}-${cleaned.slice(6)}`;
		return phone;
	}

	function formatDate(d) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
	}

	function formatCurrency(cents) {
		return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(cents / 100);
	}

	function getCustomFields(person) {
		if (!person.custom_fields) return [];
		try {
			const cf = typeof person.custom_fields === 'string' ? JSON.parse(person.custom_fields) : person.custom_fields;
			return Object.entries(cf).filter(([k, v]) => v != null && v !== '');
		} catch { return []; }
	}

	function humanizeKey(key) {
		return key.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
	}

	function getEngagementColor(score) {
		if (score >= 75) return '#4A8B8C';
		if (score >= 50) return '#8FBCB0';
		if (score >= 25) return '#f59e0b';
		return '#ef4444';
	}

	function timeAgo(date) {
		const seconds = Math.floor((new Date() - new Date(date)) / 1000);
		if (seconds < 60) return 'just now';
		if (seconds < 3600) return Math.floor(seconds / 60) + 'm ago';
		if (seconds < 86400) return Math.floor(seconds / 3600) + 'h ago';
		return Math.floor(seconds / 86400) + 'd ago';
	}

	const tabs = [
		{ id: 'info', label: 'Info' },
		{ id: 'groups', label: 'Groups & Teams' },
		{ id: 'giving', label: 'Giving' },
		{ id: 'activity', label: 'Activity' },
	];
</script>

{#if loading}
	<div class="flex justify-center py-16">
		<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
	</div>
{:else if person}
	<div class="space-y-6">
		<!-- Back + Header -->
		<button on:click={() => goto('/dashboard/people')} class="text-[var(--teal)] hover:underline text-sm flex items-center gap-1">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
			Back to People
		</button>

		<!-- Profile Header Card -->
		<div class="bg-surface rounded-xl shadow-sm border border-custom p-6">
			<div class="flex flex-col sm:flex-row items-start gap-5">
				<div
					class="w-20 h-20 rounded-full flex items-center justify-center text-white font-bold text-2xl flex-shrink-0"
					style="background-color: {getAvatarColor(person.first_name + person.last_name)}"
				>
					{getInitials(person.first_name, person.last_name)}
				</div>
				<div class="flex-1 min-w-0">
					<h1 class="text-2xl font-bold text-primary">{person.first_name} {person.last_name}</h1>
					<div class="flex flex-wrap items-center gap-2 mt-2">
						{#if true}
							{@const badge = getStatusBadge(person.membership_status)}
							<span class="px-3 py-1 text-xs font-semibold rounded-full" style="background-color: {badge.bg}; color: {badge.color}">
								{badge.label}
							</span>
						{/if}
						{#if person.tags && person.tags.length > 0}
							{#each person.tags as tag}
								<span class="px-2 py-0.5 text-xs rounded-full flex items-center gap-1" style="background-color: {tag.color}20; color: {tag.color}">
									{tag.name}
									<button on:click|stopPropagation={() => removeTag(tag.id)} class="hover:opacity-70 ml-0.5">×</button>
								</span>
							{/each}
						{/if}
						<button on:click={() => (showAddTag = !showAddTag)} class="text-[var(--teal)] text-xs hover:underline">+ Tag</button>
					</div>
					{#if showAddTag}
						<div class="flex gap-2 mt-2">
							<select bind:value={selectedTagId} class="px-2 py-1 text-sm border input-border rounded-lg bg-[var(--input-bg)] text-primary">
								<option value="">Select...</option>
								{#each availableTags as tag}<option value={tag.id}>{tag.name}</option>{/each}
							</select>
							<button on:click={addTag} class="px-3 py-1 text-sm bg-[var(--teal)] text-white rounded-lg">Add</button>
						</div>
					{/if}
				</div>
				<div class="flex gap-2 flex-shrink-0">
					{#if !editing}
						<button on:click={() => { editing = true; editForm = { ...person }; }}
							class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 text-sm font-medium">
							Edit
						</button>
					{/if}
					<button on:click={deletePerson}
						class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 text-sm font-medium">
						Delete
					</button>
				</div>
			</div>

			<!-- Engagement Score Mini -->
			{#if engagementScore}
				<div class="mt-4 flex items-center gap-4 pt-4 border-t border-custom">
					<div class="flex items-center gap-2">
						<div class="w-10 h-10 rounded-full flex items-center justify-center text-white text-sm font-bold"
							style="background-color: {getEngagementColor(engagementScore.score)}">
							{engagementScore.score}
						</div>
						<div>
							<p class="text-xs text-secondary">Engagement Score</p>
							<p class="text-xs text-primary font-medium">
								{engagementScore.score >= 75 ? 'High' : engagementScore.score >= 50 ? 'Medium' : engagementScore.score >= 25 ? 'Low' : 'Inactive'}
							</p>
						</div>
					</div>
					<button on:click={recalculateEngagement} class="text-xs text-[var(--teal)] hover:underline">↻ Recalculate</button>
				</div>
			{/if}
		</div>

		<!-- Tabs -->
		<div class="border-b border-custom">
			<nav class="flex gap-6">
				{#each tabs as tab}
					<button
						on:click={() => activeTab = tab.id}
						class="pb-3 text-sm font-medium border-b-2 transition-colors {activeTab === tab.id ? 'border-[var(--teal)] text-[var(--teal)]' : 'border-transparent text-secondary hover:text-primary'}"
					>
						{tab.label}
					</button>
				{/each}
			</nav>
		</div>

		<!-- Tab Content -->
		{#if activeTab === 'info'}
			<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
				<!-- Contact Info -->
				<div class="bg-surface rounded-xl shadow-sm border border-custom p-6">
					<h2 class="text-lg font-semibold text-primary mb-4">Contact Information</h2>
					{#if editing}
						<form on:submit|preventDefault={savePerson} class="space-y-3">
							<div class="grid grid-cols-2 gap-3">
								<div>
									<label class="block text-xs font-medium text-secondary mb-1">First Name</label>
									<input type="text" bind:value={editForm.first_name} required class="w-full px-3 py-2 border input-border rounded-lg bg-[var(--input-bg)] text-primary text-sm focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
								</div>
								<div>
									<label class="block text-xs font-medium text-secondary mb-1">Last Name</label>
									<input type="text" bind:value={editForm.last_name} required class="w-full px-3 py-2 border input-border rounded-lg bg-[var(--input-bg)] text-primary text-sm focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
								</div>
							</div>
							<div>
								<label class="block text-xs font-medium text-secondary mb-1">Email</label>
								<input type="email" bind:value={editForm.email} class="w-full px-3 py-2 border input-border rounded-lg bg-[var(--input-bg)] text-primary text-sm focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
							</div>
							<div>
								<label class="block text-xs font-medium text-secondary mb-1">Phone</label>
								<input type="tel" bind:value={editForm.phone} class="w-full px-3 py-2 border input-border rounded-lg bg-[var(--input-bg)] text-primary text-sm focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
							</div>
							<div>
								<label class="block text-xs font-medium text-secondary mb-1">Address</label>
								<input type="text" bind:value={editForm.address_line1} placeholder="Street" class="w-full px-3 py-2 border input-border rounded-lg bg-[var(--input-bg)] text-primary text-sm focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
								<div class="grid grid-cols-3 gap-2 mt-2">
									<input type="text" bind:value={editForm.city} placeholder="City" class="px-3 py-2 border input-border rounded-lg bg-[var(--input-bg)] text-primary text-sm focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
									<input type="text" bind:value={editForm.state} placeholder="State" class="px-3 py-2 border input-border rounded-lg bg-[var(--input-bg)] text-primary text-sm focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
									<input type="text" bind:value={editForm.zip} placeholder="ZIP" class="px-3 py-2 border input-border rounded-lg bg-[var(--input-bg)] text-primary text-sm focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
								</div>
							</div>
							<div class="grid grid-cols-2 gap-3">
								<div>
									<label class="block text-xs font-medium text-secondary mb-1">Gender</label>
									<select bind:value={editForm.gender} class="w-full px-3 py-2 border input-border rounded-lg bg-[var(--input-bg)] text-primary text-sm">
										<option value="">—</option><option value="male">Male</option><option value="female">Female</option>
									</select>
								</div>
								<div>
									<label class="block text-xs font-medium text-secondary mb-1">Status</label>
									<select bind:value={editForm.membership_status} class="w-full px-3 py-2 border input-border rounded-lg bg-[var(--input-bg)] text-primary text-sm">
										<option value="active">Active</option><option value="member">Member</option><option value="visitor">Visitor</option><option value="inactive">Inactive</option>
									</select>
								</div>
							</div>
							<div>
								<label class="block text-xs font-medium text-secondary mb-1">Notes</label>
								<textarea bind:value={editForm.notes} rows="3" class="w-full px-3 py-2 border input-border rounded-lg bg-[var(--input-bg)] text-primary text-sm focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
							</div>
							<div class="flex gap-3 pt-2">
								<button type="button" on:click={() => editing = false} class="flex-1 px-4 py-2 border border-custom rounded-lg text-primary text-sm hover:bg-[var(--surface-hover)]">Cancel</button>
								<button type="submit" class="flex-1 px-4 py-2 bg-[var(--teal)] text-white rounded-lg text-sm hover:opacity-90">Save</button>
							</div>
						</form>
					{:else}
						<dl class="space-y-3">
							<div class="flex items-start gap-3">
								<svg class="w-4 h-4 text-secondary mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
								<div><dt class="text-xs text-secondary">Email</dt><dd class="text-sm text-primary">{person.email || '—'}</dd></div>
							</div>
							<div class="flex items-start gap-3">
								<svg class="w-4 h-4 text-secondary mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" /></svg>
								<div><dt class="text-xs text-secondary">Phone</dt><dd class="text-sm text-primary">{formatPhone(person.phone) || '—'}</dd></div>
							</div>
							{#if person.address_line1}
								<div class="flex items-start gap-3">
									<svg class="w-4 h-4 text-secondary mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
									<div>
										<dt class="text-xs text-secondary">Address</dt>
										<dd class="text-sm text-primary">
											{person.address_line1}{#if person.address_line2}, {person.address_line2}{/if}<br/>
											{#if person.city || person.state}{person.city}, {person.state} {person.zip}{/if}
										</dd>
									</div>
								</div>
							{/if}
							{#if person.birthdate}
								<div class="flex items-start gap-3">
									<svg class="w-4 h-4 text-secondary mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
									<div><dt class="text-xs text-secondary">Birthday</dt><dd class="text-sm text-primary">{formatDate(person.birthdate)}</dd></div>
								</div>
							{/if}
							{#if person.gender}
								<div class="flex items-start gap-3">
									<svg class="w-4 h-4 text-secondary mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" /></svg>
									<div><dt class="text-xs text-secondary">Gender</dt><dd class="text-sm text-primary capitalize">{person.gender}</dd></div>
								</div>
							{/if}
							{#if person.household}
								<div class="flex items-start gap-3">
									<svg class="w-4 h-4 text-secondary mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" /></svg>
									<div><dt class="text-xs text-secondary">Household</dt><dd class="text-sm text-primary">{person.household.name}</dd></div>
								</div>
							{/if}
						</dl>
					{/if}
				</div>

				<!-- Notes + Custom Fields -->
				<div class="space-y-6">
					<!-- Notes -->
					<div class="bg-surface rounded-xl shadow-sm border border-custom p-6">
						<h2 class="text-lg font-semibold text-primary mb-3">Notes</h2>
						{#if person.notes}
							<p class="text-sm text-primary whitespace-pre-wrap">{person.notes}</p>
						{:else}
							<p class="text-sm text-secondary italic">No notes yet.</p>
						{/if}
					</div>

					<!-- Custom Fields from PCO -->
					{#if getCustomFields(person).length > 0}
						<div class="bg-surface rounded-xl shadow-sm border border-custom p-6">
							<h2 class="text-lg font-semibold text-primary mb-3">Custom Fields</h2>
							<dl class="grid grid-cols-1 sm:grid-cols-2 gap-3">
								{#each getCustomFields(person) as [key, value]}
									<div class="bg-[var(--surface-hover)] rounded-lg p-3">
										<dt class="text-xs text-secondary font-medium">{humanizeKey(key)}</dt>
										<dd class="text-sm text-primary mt-0.5">{value}</dd>
									</div>
								{/each}
							</dl>
						</div>
					{/if}

					<!-- Engagement Breakdown -->
					{#if engagementScore}
						<div class="bg-surface rounded-xl shadow-sm border border-custom p-6">
							<h2 class="text-lg font-semibold text-primary mb-3">Engagement Breakdown</h2>
							<div class="space-y-3">
								{#each [
									{ label: 'Attendance', value: engagementScore.attendance_score },
									{ label: 'Giving', value: engagementScore.giving_score },
									{ label: 'Groups', value: engagementScore.group_score },
									{ label: 'Volunteering', value: engagementScore.volunteer_score },
									{ label: 'Connection', value: engagementScore.connection_score },
								] as metric}
									<div>
										<div class="flex justify-between text-sm mb-1">
											<span class="text-secondary">{metric.label}</span>
											<span class="text-primary font-medium">{metric.value}/100</span>
										</div>
										<div class="w-full bg-[var(--surface-hover)] rounded-full h-2">
											<div class="bg-[var(--teal)] h-2 rounded-full transition-all" style="width: {metric.value}%"></div>
										</div>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</div>
			</div>
		{:else if activeTab === 'groups'}
			<div class="bg-surface rounded-xl shadow-sm border border-custom p-6">
				<h2 class="text-lg font-semibold text-primary mb-4">Groups & Teams</h2>
				{#if groups.length === 0}
					<div class="text-center py-8">
						<svg class="w-12 h-12 mx-auto mb-2 text-secondary opacity-40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
						</svg>
						<p class="text-secondary">Not a member of any groups yet.</p>
					</div>
				{:else}
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
						{#each groups as group}
							<a href="/dashboard/groups/{group.id}" class="flex items-center gap-3 p-4 bg-[var(--surface-hover)] rounded-lg hover:border-[var(--teal)] border border-transparent transition-colors">
								<div class="w-10 h-10 rounded-lg bg-[var(--teal)] bg-opacity-15 flex items-center justify-center">
									<svg class="w-5 h-5 text-[var(--teal)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
									</svg>
								</div>
								<div>
									<p class="font-medium text-primary">{group.name}</p>
									{#if group.role}<p class="text-xs text-secondary capitalize">{group.role}</p>{/if}
								</div>
							</a>
						{/each}
					</div>
				{/if}
			</div>
		{:else if activeTab === 'giving'}
			<div class="bg-surface rounded-xl shadow-sm border border-custom p-6">
				<h2 class="text-lg font-semibold text-primary mb-4">Giving History</h2>
				{#if !givingHistory || givingHistory.length === 0}
					<div class="text-center py-8">
						<svg class="w-12 h-12 mx-auto mb-2 text-secondary opacity-40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<p class="text-secondary">No giving records yet.</p>
					</div>
				{:else}
					<div class="space-y-2">
						{#each givingHistory as donation}
							<div class="flex items-center justify-between p-3 bg-[var(--surface-hover)] rounded-lg">
								<div>
									<p class="text-sm font-medium text-primary">{donation.fund_name || 'General Fund'}</p>
									<p class="text-xs text-secondary">{formatDate(donation.donation_date)}</p>
								</div>
								<p class="text-sm font-semibold text-primary">{formatCurrency(donation.amount_cents)}</p>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{:else if activeTab === 'activity'}
			<div class="bg-surface rounded-xl shadow-sm border border-custom p-6">
				<h2 class="text-lg font-semibold text-primary mb-4">Activity Timeline</h2>
				{#if activityLog.length === 0}
					<div class="text-center py-8">
						<svg class="w-12 h-12 mx-auto mb-2 text-secondary opacity-40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<p class="text-secondary">No activity recorded yet.</p>
					</div>
				{:else}
					<div class="space-y-4">
						{#each activityLog as entry}
							<div class="flex items-start gap-3">
								<div class="w-2 h-2 rounded-full bg-[var(--teal)] mt-2 flex-shrink-0"></div>
								<div>
									<p class="text-sm text-primary">{entry.action}</p>
									<p class="text-xs text-secondary">{timeAgo(entry.created_at)} · {entry.user_email || 'System'}</p>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>
{/if}
