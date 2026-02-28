<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { PartyPopper, Heart, Handshake } from 'lucide-svelte';

	interface PrayerRequest {
		id: string;
		person_id?: string;
		name: string;
		person_name?: string;
		request_text: string;
		is_public: boolean;
		status: string;
		notes?: string;
		submitted_at: string;
		follower_count?: number;
		is_following?: boolean;
	}

	let requests: PrayerRequest[] = [];
	let loading = true;
	let statusFilter = '';
	let privacyFilter = '';
	let showCreateModal = false;
	let showDetailModal = false;
	let showAnsweredModal = false;
	let selectedRequest: PrayerRequest | null = null;
	let answeredNote = '';
	let people: any[] = [];

	let formData = {
		name: '',
		request_text: '',
		is_public: true,
		person_id: ''
	};

	let personSearch = '';
	let filteredPeople: any[] = [];
	let isAnonymous = false;

	onMount(async () => {
		await Promise.all([loadRequests(), loadPeople()]);
	});

	async function loadRequests() {
		loading = true;
		try {
			let url = '/api/prayer-requests?';
			const params: string[] = [];
			if (statusFilter) params.push(`status=${statusFilter}`);
			if (privacyFilter) params.push(`is_public=${privacyFilter}`);
			url += params.join('&');
			requests = (await api(url)) || [];
		} catch (e) { console.error(e); }
		loading = false;
	}

	async function loadPeople() {
		try {
			const r = await api('/api/people?limit=200');
			people = r.people || [];
		} catch (e) { console.error(e); }
	}

	function searchPeople(q: string) {
		if (!q) { filteredPeople = []; return; }
		filteredPeople = people.filter(p =>
			`${p.first_name} ${p.last_name}`.toLowerCase().includes(q.toLowerCase())
		).slice(0, 10);
	}

	function selectPerson(p: any) {
		formData.person_id = p.id;
		formData.name = `${p.first_name} ${p.last_name}`;
		personSearch = formData.name;
		filteredPeople = [];
	}

	async function createRequest() {
		if (isAnonymous) {
			formData.name = 'Anonymous';
			formData.is_public = false;
		}
		try {
			let url = '/api/prayer-requests/create';
			if (formData.person_id) url += `?person_id=${formData.person_id}`;
			await api(url, { method: 'POST', body: JSON.stringify(formData) });
			showCreateModal = false;
			formData = { name: '', request_text: '', is_public: true, person_id: '' };
			personSearch = '';
			isAnonymous = false;
			loadRequests();
		} catch (e) { console.error(e); }
	}

	async function updateStatus(id: string, status: string, notes?: string) {
		try {
			const body: any = { status };
			if (notes !== undefined) body.notes = notes;
			await api(`/api/prayer-requests/${id}`, { method: 'PUT', body: JSON.stringify(body) });
			loadRequests();
		} catch (e) { console.error(e); }
	}

	async function toggleFollow(req: PrayerRequest) {
		try {
			if (req.is_following) {
				await api(`/api/prayer-requests/${req.id}/follow`, { method: 'DELETE' });
			} else {
				await api(`/api/prayer-requests/${req.id}/follow`, { method: 'POST' });
			}
			loadRequests();
		} catch (e) { console.error(e); }
	}

	function openDetail(req: PrayerRequest) {
		selectedRequest = req;
		showDetailModal = true;
	}

	function openAnsweredModal(req: PrayerRequest) {
		selectedRequest = req;
		answeredNote = '';
		showAnsweredModal = true;
	}

	async function markAsAnswered() {
		if (!selectedRequest) return;
		await updateStatus(selectedRequest.id, 'answered', answeredNote || undefined);
		showAnsweredModal = false;
		showDetailModal = false;
		loadRequests();
	}

	async function createFollowUpFromPrayer(req: PrayerRequest) {
		if (!req.person_name) return;
		try {
			// Find the person ID from the request if linked
			const personId = (req as any).person_id;
			if (!personId) {
				// Redirect to care page to create manually
				window.location.href = '/dashboard/care';
				return;
			}
			await api('/api/follow-ups', {
				method: 'POST',
				body: JSON.stringify({
					person_id: personId,
					title: `Prayer follow-up: ${req.request_text.substring(0, 60)}...`,
					type: 'prayer_response',
					priority: 'medium',
					notes: `Follow-up from prayer request: "${req.request_text}"`
				})
			});
			showDetailModal = false;
			// Show success, could navigate to care
			alert('Follow-up created successfully! View it in Care & Follow-Ups.');
		} catch (e) { console.error(e); }
	}

	function formatDate(d: string) {
		return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	function getStatusStyle(s: string) {
		const styles: Record<string, string> = {
			pending: 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20',
			praying: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
			answered: 'bg-green-500/10 text-green-500 border-green-500/20',
			archived: 'bg-gray-500/10 text-gray-500 border-gray-500/20'
		};
		return styles[s] || styles.pending;
	}

	function getStatusLabel(s: string) {
		return { pending: 'Active', praying: 'Praying', answered: 'Answered', archived: 'Archived' }[s] || s;
	}

	$: activeRequests = requests.filter(r => r.status === 'pending' || r.status === 'praying');
	$: answeredRequests = requests.filter(r => r.status === 'answered');
	$: activeCount = activeRequests.length;
	$: answeredCount = answeredRequests.length;
	$: displayRequests = statusFilter === 'answered' ? answeredRequests : statusFilter ? requests.filter(r => r.status === statusFilter) : requests;
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex justify-between items-center">
		<div>
			<h1 class="text-3xl font-bold text-[var(--text-primary)] flex items-center gap-2"><Heart size={28} /> Prayer Requests</h1>
			<p class="text-sm text-[var(--text-secondary)] mt-1">{activeCount} active · {answeredCount} answered</p>
		</div>
		<button on:click={() => showCreateModal = true} class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 font-medium">
			+ New Request
		</button>
	</div>

	<!-- Filters -->
	<div class="bg-[var(--surface)] border border-[var(--border)] rounded-lg p-4 flex flex-wrap gap-3 items-center">
		<div class="flex rounded-lg overflow-hidden border border-[var(--border)]">
			{#each [{ v: '', l: 'All' }, { v: 'pending', l: 'Active' }, { v: 'praying', l: 'Praying' }, { v: 'answered', l: 'Answered' }, { v: 'archived', l: 'Archived' }] as f}
				<button class="px-3 py-1.5 text-sm transition-colors {statusFilter === f.v ? 'bg-[var(--teal)] text-white' : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)]'}" on:click={() => { statusFilter = f.v; loadRequests(); }}>
					{f.l}
				</button>
			{/each}
		</div>

		<select bind:value={privacyFilter} on:change={loadRequests} class="px-3 py-1.5 text-sm rounded-lg border border-[var(--border)] bg-[var(--surface)] text-[var(--text-primary)]">
			<option value="">All Privacy</option>
			<option value="true">Public</option>
			<option value="false">Private</option>
		</select>
	</div>

	<!-- Prayer Cards -->
	{#if loading}
		<div class="flex justify-center py-12">
			<div class="animate-spin rounded-full h-10 w-10 border-2 border-[var(--teal)] border-t-transparent"></div>
		</div>
	{:else if requests.length === 0}
		<div class="bg-[var(--surface)] border border-[var(--border)] rounded-lg p-12 text-center">
			<div class="mb-4"><Heart size={48} /></div>
			<h3 class="text-xl font-semibold text-[var(--text-primary)] mb-2">Lift each other up in prayer</h3>
			<p class="text-[var(--text-secondary)] max-w-md mx-auto mb-6">
				Prayer requests help your church community support one another through life's joys and challenges.
				Share a need, and let others join you in prayer.
			</p>
			<button on:click={() => showCreateModal = true} class="px-6 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 font-medium">
				+ Share a Prayer Request
			</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each displayRequests as req}
				<div class="rounded-lg p-5 hover:border-[var(--teal)] transition-colors flex flex-col {req.status === 'answered' ? 'bg-green-500/5 border-2 border-green-500/30' : 'bg-[var(--surface)] border border-[var(--border)]'}">
					<div class="flex items-start justify-between mb-3">
						<div class="flex-1 min-w-0">
							<h3 class="font-semibold text-[var(--text-primary)] truncate">{req.name}</h3>
							<span class="text-xs text-[var(--text-secondary)]">{formatDate(req.submitted_at)}</span>
						</div>
						<div class="flex items-center gap-2 ml-2 flex-shrink-0">
							{#if !req.is_public}
								<span class="text-xs" title="Private">🔒</span>
							{/if}
							<span class="text-xs px-2 py-0.5 rounded-full border {getStatusStyle(req.status)}">{getStatusLabel(req.status)}</span>
						</div>
					</div>

					<p class="text-sm text-[var(--text-secondary)] mb-2 flex-1 line-clamp-3">{req.request_text}</p>
					{#if req.status === 'answered' && req.notes}
						<div class="text-xs text-green-600 bg-green-500/10 rounded px-2 py-1 mb-2">
							✨ {req.notes}
						</div>
					{/if}

					<div class="flex items-center justify-between pt-3 border-t border-[var(--border)]">
						<button class="flex items-center gap-1.5 text-sm transition-colors {req.is_following ? 'text-[var(--teal)] font-medium' : 'text-[var(--text-secondary)] hover:text-[var(--teal)]'}" on:click|stopPropagation={() => toggleFollow(req)}>
							<Heart size={14} class="inline" /> <span>{req.is_following ? 'Praying' : 'I Prayed'} ({req.follower_count || 0})</span>
						</button>

						<div class="flex items-center gap-1">
							{#if req.status !== 'answered'}
								<button on:click={() => openAnsweredModal(req)} class="px-2 py-1 text-xs text-green-500 hover:bg-green-500/10 rounded transition-colors" title="Mark as answered">
									✓ Answered
								</button>
							{/if}
							<button on:click={() => openDetail(req)} class="px-2 py-1 text-xs text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--surface-hover)] rounded transition-colors">
								View
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Detail Modal -->
{#if showDetailModal && selectedRequest}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" on:click={() => showDetailModal = false}>
		<div class="bg-[var(--surface)] rounded-xl p-6 w-full max-w-lg border border-[var(--border)] shadow-xl" on:click|stopPropagation>
			<div class="flex items-start justify-between mb-4">
				<div>
					<div class="flex items-center gap-2 mb-1">
						<span class="text-xs px-2 py-0.5 rounded-full border {getStatusStyle(selectedRequest.status)}">{getStatusLabel(selectedRequest.status)}</span>
						{#if !selectedRequest.is_public}<span class="text-xs">🔒 Private</span>{/if}
					</div>
					<h2 class="text-xl font-bold text-[var(--text-primary)]">{selectedRequest.name}</h2>
					<p class="text-xs text-[var(--text-secondary)]">{formatDate(selectedRequest.submitted_at)}</p>
				</div>
				<button on:click={() => showDetailModal = false} class="text-[var(--text-secondary)] hover:text-[var(--text-primary)] text-xl">✕</button>
			</div>

			<div class="bg-[var(--bg)] rounded-lg p-4 mb-4 border border-[var(--border)]">
				<p class="text-[var(--text-primary)] whitespace-pre-wrap">{selectedRequest.request_text}</p>
			</div>

			{#if selectedRequest.notes}
				<div class="bg-green-500/5 rounded-lg p-4 mb-4 border border-green-500/20">
					<h4 class="text-sm font-semibold text-green-500 mb-1">✨ Testimony / Note</h4>
					<p class="text-sm text-[var(--text-primary)]">{selectedRequest.notes}</p>
				</div>
			{/if}

			<div class="flex items-center gap-4 mb-4 pb-4 border-b border-[var(--border)]">
				<button class="flex items-center gap-1.5 text-sm transition-colors {selectedRequest.is_following ? 'text-[var(--teal)] font-medium' : 'text-[var(--text-secondary)] hover:text-[var(--teal)]'}" on:click={() => toggleFollow(selectedRequest)}>
					<Heart size={16} class="inline" /> {selectedRequest.is_following ? "You're praying" : 'Pray for this'}
				</button>
				<span class="text-sm text-[var(--text-secondary)]">{selectedRequest.follower_count || 0} people praying</span>
			</div>

			<!-- Status Actions -->
			<div class="flex flex-wrap gap-2 mb-4">
				<select value={selectedRequest.status} on:change={(e) => { updateStatus(selectedRequest.id, e.currentTarget.value); selectedRequest.status = e.currentTarget.value; }} class="px-3 py-1.5 text-sm rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]">
					<option value="pending">Active</option>
					<option value="praying">Praying</option>
					<option value="answered">Answered</option>
					<option value="archived">Archived</option>
				</select>
				{#if selectedRequest.status !== 'answered'}
					<button on:click={() => openAnsweredModal(selectedRequest)} class="px-3 py-1.5 text-sm bg-green-500/10 text-green-500 rounded-lg hover:bg-green-500/20 transition-colors">
						✓ Mark as Answered
					</button>
				{/if}
				{#if selectedRequest.person_name}
					<button on:click={() => createFollowUpFromPrayer(selectedRequest)} class="px-3 py-1.5 text-sm bg-[var(--teal)]/10 text-[var(--teal)] rounded-lg hover:bg-[var(--teal)]/20 transition-colors">
						<Handshake size={16} class="inline" /> Follow up with this person
					</button>
				{/if}
			</div>

			<div class="flex justify-end pt-4 border-t border-[var(--border)]">
				<button on:click={() => showDetailModal = false} class="px-4 py-2 text-sm border border-[var(--border)] rounded-lg text-[var(--text-secondary)]">Close</button>
			</div>
		</div>
	</div>
{/if}

<!-- Mark as Answered Modal -->
{#if showAnsweredModal && selectedRequest}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" on:click={() => showAnsweredModal = false}>
		<div class="bg-[var(--surface)] rounded-xl p-6 w-full max-w-md border border-[var(--border)] shadow-xl" on:click|stopPropagation>
			<div class="text-center mb-4">
				<div class="mb-2"><PartyPopper size={48} /></div>
				<h2 class="text-xl font-bold text-[var(--text-primary)]">Prayer Answered!</h2>
				<p class="text-sm text-[var(--text-secondary)]">Share how God answered this prayer.</p>
			</div>

			<div class="mb-4">
				<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Testimony / Note (optional)</label>
				<textarea bind:value={answeredNote} rows="4" placeholder="Share the answer or testimony..." class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]"></textarea>
			</div>

			<div class="flex gap-2 justify-end">
				<button on:click={() => showAnsweredModal = false} class="px-4 py-2 text-sm border border-[var(--border)] rounded-lg text-[var(--text-secondary)]">Cancel</button>
				<button on:click={markAsAnswered} class="px-6 py-2 text-sm bg-green-500 text-white rounded-lg hover:opacity-90 font-medium">Mark as Answered</button>
			</div>
		</div>
	</div>
{/if}

<!-- Create Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" on:click={() => showCreateModal = false}>
		<div class="bg-[var(--surface)] rounded-xl p-6 w-full max-w-lg border border-[var(--border)] shadow-xl" on:click|stopPropagation>
			<h2 class="text-xl font-bold text-[var(--text-primary)] mb-4">New Prayer Request</h2>
			<form on:submit|preventDefault={createRequest} class="space-y-4">
				<!-- Anonymous toggle -->
				<label class="flex items-center gap-2 text-sm text-[var(--text-primary)]">
					<input type="checkbox" bind:checked={isAnonymous} class="rounded" />
					Submit anonymously
				</label>

				{#if !isAnonymous}
					<!-- Person search -->
					<div class="relative">
						<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Requester *</label>
						<input type="text" bind:value={personSearch} on:input={() => searchPeople(personSearch)} placeholder="Search for a person or type a name..." class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]" />
						{#if filteredPeople.length > 0}
							<div class="absolute z-10 w-full mt-1 bg-[var(--surface)] border border-[var(--border)] rounded-lg shadow-lg max-h-40 overflow-y-auto">
								{#each filteredPeople as p}
									<button type="button" class="w-full text-left px-3 py-2 text-sm hover:bg-[var(--surface-hover)] text-[var(--text-primary)]" on:click={() => selectPerson(p)}>
										{p.first_name} {p.last_name}
									</button>
								{/each}
							</div>
						{/if}
						{#if !formData.person_id && personSearch}
							<input type="hidden" bind:value={formData.name} />
							{(() => { formData.name = personSearch; return ''; })()}
						{/if}
					</div>
				{/if}

				<div>
					<label class="block text-sm font-medium text-[var(--text-primary)] mb-1">Prayer Request *</label>
					<textarea bind:value={formData.request_text} required rows="4" placeholder="Share what you'd like prayer for..." class="w-full px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg)] text-[var(--text-primary)]"></textarea>
				</div>

				{#if !isAnonymous}
					<label class="flex items-center gap-2 text-sm text-[var(--text-primary)]">
						<input type="checkbox" bind:checked={formData.is_public} class="rounded" />
						Make this request public (visible to the congregation)
					</label>
				{/if}

				<div class="flex gap-2 justify-end pt-4 border-t border-[var(--border)]">
					<button type="button" on:click={() => showCreateModal = false} class="px-4 py-2 text-sm border border-[var(--border)] rounded-lg text-[var(--text-secondary)]">Cancel</button>
					<button type="submit" class="px-6 py-2 text-sm bg-[var(--teal)] text-white rounded-lg hover:opacity-90 font-medium">Submit Request</button>
				</div>
			</form>
		</div>
	</div>
{/if}
