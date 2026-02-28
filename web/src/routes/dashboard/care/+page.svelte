<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { AlertTriangle, Heart, ClipboardList, Handshake, CheckCircle, User, CalendarDays } from 'lucide-svelte';

	interface FollowUp {
		id: string;
		person_id: string;
		person_name: string;
		assigned_to?: string;
		assigned_name?: string;
		title: string;
		type: string;
		priority: string;
		status: string;
		due_date?: string;
		completed_at?: string;
		created_at: string;
		notes?: Note[];
	}

	interface Note {
		id: string;
		author_name: string;
		note: string;
		created_at: string;
	}

	let followUps: FollowUp[] = [];
	let stats: any = {};
	let people: any[] = [];
	let showCreateModal = false;
	let showDetailModal = false;
	let selectedItem: FollowUp | null = null;
	let newNote = '';

	let formData = {
		person_id: '',
		assigned_to: '',
		title: '',
		type: 'general',
		priority: 'medium',
		due_date: '',
		notes: ''
	};

	let personSearch = '';
	let filteredPeople: any[] = [];

	const types = [
		{ value: 'first_time_visitor', label: 'First-Time Visitor', icon: '👋' },
		{ value: 'visitor_followup', label: 'Visitor Follow-up', icon: '🏠' },
		{ value: 'pastoral_care', label: 'Pastoral Care', icon: '❤️' },
		{ value: 'hospital_visit', label: 'Hospital Visit', icon: '🏥' },
		{ value: 'counseling', label: 'Counseling', icon: '💬' },
		{ value: 'prayer_response', label: 'Prayer Response', icon: Heart },
		{ value: 'general', label: 'General', icon: ClipboardList },
		{ value: 'membership', label: 'Membership', icon: Handshake }
	];

	const priorities = [
		{ value: 'high', label: 'High', color: '#EF4444' },
		{ value: 'medium', label: 'Medium', color: '#F59E0B' },
		{ value: 'low', label: 'Low', color: '#6B7280' }
	];

	const columns = [
		{ status: 'new', label: 'New', icon: '🆕', color: '#3B82F6' },
		{ status: 'in_progress', label: 'In Progress', icon: '🔄', color: '#F59E0B' },
		{ status: 'waiting', label: 'Waiting', icon: '⏳', color: '#8B5CF6' },
		{ status: 'completed', label: 'Completed', icon: CheckCircle, color: '#10B981' }
	];

	onMount(() => { load(); loadPeople(); });

	async function load() {
		try {
			const [fuData, statsData] = await Promise.all([
				api('/api/follow-ups'),
				api('/api/follow-ups/stats')
			]);
			followUps = fuData.follow_ups || [];
			stats = statsData;
		} catch (e) { console.error(e); }
	}

	async function loadPeople() {
		try {
			const r = await api('/api/people?limit=200');
			people = r.people || [];
		} catch (e) { console.error(e); }
	}

	function searchPeople(q: string) {
		if (!q) { filteredPeople = []; return; }
		const lower = q.toLowerCase();
		filteredPeople = people.filter(p =>
			`${p.first_name} ${p.last_name}`.toLowerCase().includes(lower)
		).slice(0, 10);
	}

	function selectPerson(p: any) {
		formData.person_id = p.id;
		personSearch = `${p.first_name} ${p.last_name}`;
		filteredPeople = [];
	}

	async function create() {
		try {
			await api('/api/follow-ups', {
				method: 'POST',
				body: JSON.stringify({
					...formData,
					assigned_to: formData.assigned_to || null
				})
			});
			showCreateModal = false;
			formData = { person_id: '', assigned_to: '', title: '', type: 'general', priority: 'medium', due_date: '', notes: '' };
			personSearch = '';
			load();
		} catch (e) { console.error(e); }
	}

	async function updateStatus(id: string, status: string) {
		try {
			await api(`/api/follow-ups/${id}`, { method: 'PUT', body: JSON.stringify({ status }) });
			load();
			if (selectedItem?.id === id) {
				selectedItem = { ...selectedItem, status };
			}
		} catch (e) { console.error(e); }
	}

	async function openDetail(item: FollowUp) {
		try {
			selectedItem = await api(`/api/follow-ups/${item.id}`);
			showDetailModal = true;
		} catch (e) { console.error(e); }
	}

	async function addNote() {
		if (!newNote.trim() || !selectedItem) return;
		try {
			await api(`/api/follow-ups/${selectedItem.id}/notes`, {
				method: 'POST',
				body: JSON.stringify({ note: newNote })
			});
			newNote = '';
			selectedItem = await api(`/api/follow-ups/${selectedItem.id}`);
			load();
		} catch (e) { console.error(e); }
	}

	async function deleteItem(id: string) {
		if (!confirm('Delete this follow-up?')) return;
		try {
			await api(`/api/follow-ups/${id}`, { method: 'DELETE' });
			showDetailModal = false;
			load();
		} catch (e) { console.error(e); }
	}

	function getTypeInfo(t: string) { return types.find(x => x.value === t) || types[3]; }
	function getPriorityInfo(p: string) { return priorities.find(x => x.value === p) || priorities[1]; }

	function isOverdue(item: FollowUp) {
		return item.due_date && item.status !== 'completed' && new Date(item.due_date) < new Date();
	}

	function formatDate(d: string) {
		return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	function formatDateTime(d: string) {
		return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' });
	}

	function daysSince(dateStr: string): number {
		return Math.floor((Date.now() - new Date(dateStr).getTime()) / 86400000);
	}

	function getColumnItems(status: string) {
		return followUps.filter(f => f.status === status);
	}
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-[var(--text-primary)] flex items-center gap-2"><Handshake size={28} /> Care & Follow-Ups</h1>
		<button on:click={() => { showCreateModal = true; personSearch = ''; }} class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 font-medium">
			+ New Follow-Up
		</button>
	</div>

	<!-- Stats Dashboard -->
	<div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-7 gap-3">
		{#each [
			{ label: 'New', value: stats.new_count || 0, color: '#3B82F6' },
			{ label: 'In Progress', value: stats.in_progress_count || 0, color: '#F59E0B' },
			{ label: 'Waiting', value: stats.waiting_count || 0, color: '#8B5CF6' },
			{ label: 'Completed', value: stats.completed_count || 0, color: '#10B981' },
			{ label: 'Overdue', value: stats.overdue_count || 0, color: '#EF4444' },
			{ label: 'Due Today', value: stats.due_today_count || 0, color: '#F97316' },
			{ label: 'This Week', value: stats.due_this_week_count || 0, color: '#4A8B8C' }
		] as stat}
			<div class="bg-[var(--surface)] border border-[var(--border)] rounded-lg p-3 text-center">
				<div class="text-2xl font-bold" style="color: {stat.color}">{stat.value}</div>
				<div class="text-xs text-[var(--text-secondary)] mt-1">{stat.label}</div>
			</div>
		{/each}
	</div>

	<!-- Kanban Board -->
	{#if followUps.length === 0 && !stats.new_count}
		<div class="bg-[var(--surface)] border border-[var(--border)] rounded-lg p-12 text-center">
			<div class="mb-4"><Handshake size={48} /></div>
			<h3 class="text-xl font-semibold text-[var(--text-primary)] mb-2">No follow-ups yet</h3>
			<p class="text-[var(--text-secondary)] max-w-md mx-auto mb-6">
				Follow-ups help you track pastoral care, visitor connections, and member needs.
				Create your first follow-up to start caring for your community.
			</p>
			<button on:click={() => showCreateModal = true} class="px-6 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 font-medium">
				+ Create First Follow-Up
			</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
			{#each columns as col}
				{@const items = getColumnItems(col.status)}
				<div class="bg-[var(--bg)] rounded-lg border border-[var(--border)] flex flex-col">
					<div class="px-4 py-3 border-b border-[var(--border)] flex items-center gap-2">
						<span><svelte:component this={col.icon} size={16} /></span>
						<span class="font-semibold text-[var(--text-primary)]">{col.label}</span>
						<span class="ml-auto text-sm px-2 py-0.5 rounded-full bg-[var(--surface)] text-[var(--text-secondary)]">{items.length}</span>
					</div>
					<div class="p-2 space-y-2 flex-1 min-h-[200px] overflow-y-auto max-h-[60vh]">
						{#each items as item}
							<button class="w-full text-left bg-[var(--surface)] rounded-lg p-3 border border-[var(--border)] hover:border-[var(--teal)] transition-colors cursor-pointer" on:click={() => openDetail(item)}>
								<div class="flex items-start justify-between gap-2 mb-1">
									<span class="font-medium text-sm text-[var(--text-primary)] truncate inline-flex items-center gap-1"><svelte:component this={getTypeInfo(item.type).icon} size={14} /> {item.title}</span>
									<span class="w-2 h-2 rounded-full flex-shrink-0 mt-1.5" style="background-color: {getPriorityInfo(item.priority).color}" title="{getPriorityInfo(item.priority).label} priority"></span>
								</div>
								<div class="text-xs text-[var(--text-secondary)]">{item.person_name}</div>
								{#if item.assigned_name}
									<div class="text-xs text-[var(--text-secondary)] mt-0.5">→ {item.assigned_name}</div>
								{/if}
								<div class="flex items-center gap-2 mt-1">
									{#if item.due_date}
										<span class="text-xs {isOverdue(item) ? 'text-red-500 font-semibold' : 'text-[var(--text-secondary)]'}">
											{#if isOverdue(item)}<AlertTriangle size={12} class="inline" /> Overdue: {:else}Due: {/if}{formatDate(item.due_date)}
										</span>
									{/if}
									<span class="text-xs text-[var(--text-secondary)] ml-auto">{daysSince(item.created_at)}d ago</span>
								</div>
							</button>
						{/each}
						{#if items.length === 0}
							<div class="text-center text-xs text-[var(--text-secondary)] py-8 opacity-50">No items</div>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Detail Modal -->
{#if showDetailModal && selectedItem}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" on:click={() => showDetailModal = false}>
		<div class="bg-[var(--surface)] rounded-xl p-6 w-full max-w-lg border border-[var(--border)] shadow-xl max-h-[90vh] overflow-y-auto" on:click|stopPropagation>
			<div class="flex items-start justify-between mb-4">
				<div>
					<div class="flex items-center gap-2 mb-1">
						<span class="text-sm"><svelte:component this={getTypeInfo(selectedItem.type).icon} size={14} /></span>
						<span class="text-xs px-2 py-0.5 rounded-full bg-[var(--surface-hover)] text-[var(--text-secondary)]">{getTypeInfo(selectedItem.type).label}</span>
						<span class="w-2 h-2 rounded-full" style="background-color: {getPriorityInfo(selectedItem.priority).color}"></span>
						<span class="text-xs text-[var(--text-secondary)]">{getPriorityInfo(selectedItem.priority).label}</span>
					</div>
					<h2 class="text-xl font-bold text-[var(--text-primary)]">{selectedItem.title}</h2>
				</div>
				<button on:click={() => showDetailModal = false} class="text-[var(--text-secondary)] hover:text-[var(--text-primary)] text-xl">✕</button>
			</div>

			<div class="space-y-3 text-sm mb-4">
				<div class="flex items-center gap-2 text-[var(--text-secondary)]">
					<User size={14} /><a href="/dashboard/people/{selectedItem.person_id}" class="text-[var(--teal)] hover:underline">{selectedItem.person_name}</a>
				</div>
				{#if selectedItem.assigned_name}
					<div class="flex items-center gap-2 text-[var(--text-secondary)]">
						<span>→</span><span>Assigned to {selectedItem.assigned_name}</span>
					</div>
				{/if}
				{#if selectedItem.due_date}
					<div class="flex items-center gap-2 {isOverdue(selectedItem) ? 'text-red-500 font-semibold' : 'text-[var(--text-secondary)]'}">
						<CalendarDays size={14} /><span>{isOverdue(selectedItem) ? 'Overdue: ' : 'Due: '}{formatDate(selectedItem.due_date)}</span>
					</div>
				{/if}
				<div class="flex items-center gap-2 text-[var(--text-secondary)]">
					<CalendarDays size={14} /><span>Created {formatDateTime(selectedItem.created_at)}</span>
				</div>
			</div>

			<!-- Status Actions -->
			<div class="flex flex-wrap gap-2 mb-4 pb-4 border-b border-[var(--border)]">
				{#each columns as col}
					<button
						class="px-3 py-1.5 text-xs rounded-lg border transition-colors {selectedItem.status === col.status ? 'text-white border-transparent' : 'border-[var(--border)] text-[var(--text-secondary)] hover:text-[var(--text-primary)]'}"
						style={selectedItem.status === col.status ? `background-color: ${col.color}` : ''}
						on:click={() => updateStatus(selectedItem.id, col.status)}
					>
						<svelte:component this={col.icon} size={12} /> {col.label}
					</button>
				{/each}
			</div>

			<!-- Notes History -->
			<div class="mb-4">
				<h3 class="font-semibold text-[var(--text-primary)] mb-3">Notes</h3>
				{#if selectedItem.notes && selectedItem.notes.length > 0}
					<div class="space-y-3 mb-4">
						{#each selectedItem.notes as note}
							<div class="bg-[var(--bg)] rounded-lg p-3 border border-[var(--border)]">
								<div class="flex items-center justify-between mb-1">
									<span class="text-xs font-medium text-[var(--teal)]">{note.author_name}</span>
									<span class="text-xs text-[var(--text-secondary)]">{formatDateTime(note.created_at)}</span>
								</div>
								<p class="text-sm text-[var(--text-primary)] whitespace-pre-wrap">{note.note}</p>
							</div>
						{/each}
					</div>
				{:else}
					<p class="text-sm text-[var(--text-secondary)] mb-4">No notes yet.</p>
				{/if}

				<!-- Add Note -->
				<div class="flex gap-2">
					<textarea bind:value={newNote} placeholder="Add a note..." rows="2" class="flex-1 px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)] text-sm"></textarea>
					<button on:click={addNote} disabled={!newNote.trim()} class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 text-sm disabled:opacity-50 self-end">Add</button>
				</div>
			</div>

			<div class="flex justify-between pt-4 border-t border-[var(--border)]">
				<button on:click={() => deleteItem(selectedItem.id)} class="px-4 py-2 text-sm text-red-500 hover:bg-red-500/10 rounded-lg">Delete</button>
				<button on:click={() => showDetailModal = false} class="px-4 py-2 text-sm border border-[var(--border)] rounded-lg text-[var(--text-secondary)]">Close</button>
			</div>
		</div>
	</div>
{/if}

<!-- Create Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" on:click={() => showCreateModal = false}>
		<div class="bg-[var(--surface)] rounded-xl p-6 w-full max-w-lg border border-[var(--border)] shadow-xl max-h-[90vh] overflow-y-auto" on:click|stopPropagation>
			<h2 class="text-xl font-bold text-[var(--text-primary)] mb-4">New Follow-Up</h2>
			<form on:submit|preventDefault={create} class="space-y-4">
				<!-- Person Search -->
				<div class="relative">
					<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Person *</label>
					<input type="text" bind:value={personSearch} on:input={() => searchPeople(personSearch)} placeholder="Search for a person..." required={!formData.person_id} class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]" />
					{#if filteredPeople.length > 0}
						<div class="absolute z-10 w-full mt-1 bg-[var(--surface)] border border-[var(--border)] rounded-lg shadow-lg max-h-40 overflow-y-auto">
							{#each filteredPeople as p}
								<button type="button" class="w-full text-left px-3 py-2 text-sm hover:bg-[var(--surface-hover)] text-[var(--text-primary)]" on:click={() => selectPerson(p)}>
									{p.first_name} {p.last_name}
								</button>
							{/each}
						</div>
					{/if}
				</div>

				<div>
					<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Title *</label>
					<input type="text" bind:value={formData.title} required placeholder="e.g., Welcome visit for new family" class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]" />
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Type</label>
						<select bind:value={formData.type} class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]">
							{#each types as t}
								<option value={t.value}>{t.label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Priority</label>
						<select bind:value={formData.priority} class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]">
							{#each priorities as p}
								<option value={p.value}>{p.label}</option>
							{/each}
						</select>
					</div>
				</div>

				<div>
					<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Assigned To</label>
					<select bind:value={formData.assigned_to} class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]">
						<option value="">Unassigned</option>
						{#each people.slice(0, 50) as p}
							<option value={p.id}>{p.first_name} {p.last_name}</option>
						{/each}
					</select>
				</div>

				<div>
					<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Due Date</label>
					<input type="date" bind:value={formData.due_date} class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]" />
				</div>

				<div>
					<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Initial Notes</label>
					<textarea bind:value={formData.notes} rows="3" placeholder="Any context or details..." class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]"></textarea>
				</div>

				<div class="flex gap-2 justify-end pt-4 border-t border-[var(--border)]">
					<button type="button" on:click={() => showCreateModal = false} class="px-4 py-2 text-sm border border-[var(--border)] rounded-lg text-[var(--text-secondary)]">Cancel</button>
					<button type="submit" class="px-6 py-2 text-sm bg-[var(--teal)] text-white rounded-lg hover:opacity-90 font-medium">Create</button>
				</div>
			</form>
		</div>
	</div>
{/if}
