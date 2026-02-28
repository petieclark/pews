<script>
	import { onMount } from 'svelte';
	import { Mail } from 'lucide-svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let campaign = null;
	let recipients = [];
	let loading = true;
	let error = '';
	let sending = false;
	let showSchedule = false;
	let scheduledDate = '';
	let scheduledTime = '';
	let showPreview = false;

	$: campaignId = $page.params.id;

	onMount(async () => {
		await loadCampaign();
	});

	async function loadCampaign() {
		try {
			loading = true;
			const [c, r] = await Promise.all([
				api(`/api/communication/campaigns/${campaignId}`),
				api(`/api/communication/campaigns/${campaignId}/recipients`).catch(() => [])
			]);
			campaign = c;
			recipients = r || [];
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	async function sendNow() {
		if (!confirm('Send this campaign now? This cannot be undone.')) return;
		sending = true;
		try {
			await api(`/api/communication/campaigns/${campaignId}/send`, {
				method: 'POST',
				body: JSON.stringify({})
			});
			await loadCampaign();
		} catch (err) {
			error = err.message;
		} finally {
			sending = false;
		}
	}

	async function schedule() {
		if (!scheduledDate || !scheduledTime) {
			error = 'Please select a date and time';
			return;
		}
		sending = true;
		try {
			const scheduledAt = new Date(`${scheduledDate}T${scheduledTime}`).toISOString();
			await api(`/api/communication/campaigns/${campaignId}/send`, {
				method: 'POST',
				body: JSON.stringify({ scheduled_at: scheduledAt })
			});
			showSchedule = false;
			await loadCampaign();
		} catch (err) {
			error = err.message;
		} finally {
			sending = false;
		}
	}

	function getStatusBadge(status) {
		const badges = {
			draft: { bg: '#e2e8f0', color: '#475569', text: 'Draft' },
			scheduled: { bg: '#dbeafe', color: '#1e40af', text: 'Scheduled' },
			sending: { bg: '#fef3c7', color: '#92400e', text: 'Sending' },
			sent: { bg: '#d1fae5', color: '#065f46', text: 'Sent' },
			failed: { bg: '#fee2e2', color: '#991b1b', text: 'Failed' }
		};
		return badges[status] || badges.draft;
	}

	function formatDate(d) {
		if (!d) return '—';
		return new Date(d).toLocaleString();
	}
</script>

<div class="max-w-4xl mx-auto">
	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2" style="border-color: var(--teal)"></div>
		</div>
	{:else if error && !campaign}
		<div class="px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">{error}</div>
	{:else if campaign}
		<!-- Header -->
		<div class="mb-6">
			<a href="/dashboard/communication/campaigns" class="text-sm font-medium" style="color: var(--teal)">← Campaigns</a>
			<div class="flex items-center justify-between mt-2">
				<div class="flex items-center gap-3">
					<h1 class="text-3xl font-bold" style="color: var(--text-primary)">{campaign.name}</h1>
					{#if true}{@const badge = getStatusBadge(campaign.status)}<span class="px-3 py-1 rounded-full text-sm font-medium" style="background: {badge.bg}; color: {badge.color}">{badge.text}</span>{/if}
				</div>
				{#if campaign.status === 'draft'}
					<div class="flex gap-3">
						<button
							on:click={() => goto(`/dashboard/communication/campaigns/new?edit=${campaignId}`)}
							class="px-4 py-2 rounded-lg font-medium border"
							style="background: var(--surface); border-color: var(--border); color: var(--text-primary)"
						>
							Edit
						</button>
						<button
							on:click={() => showSchedule = !showSchedule}
							class="px-4 py-2 rounded-lg font-medium border"
							style="background: var(--surface); border-color: var(--border); color: var(--text-primary)"
						>
							Schedule
						</button>
						<button
							on:click={sendNow}
							disabled={sending}
							class="px-4 py-2 rounded-lg font-medium"
							style="background: var(--teal); color: white; opacity: {sending ? 0.5 : 1}"
						>
							{sending ? 'Sending...' : 'Send Now'}
						</button>
					</div>
				{/if}
			</div>
		</div>

		{#if error}
			<div class="mb-4 px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">{error}</div>
		{/if}

		<!-- Schedule Modal -->
		{#if showSchedule}
			<div class="mb-6 rounded-lg shadow border p-6" style="background: var(--surface); border-color: var(--border)">
				<h3 class="text-lg font-semibold mb-4" style="color: var(--text-primary)">Schedule Campaign</h3>
				<div class="grid grid-cols-2 gap-4 mb-4">
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Date</label>
						<input type="date" bind:value={scheduledDate} class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Time</label>
						<input type="time" bind:value={scheduledTime} class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
					</div>
				</div>
				<div class="flex gap-3">
					<button on:click={schedule} disabled={sending} class="px-4 py-2 rounded-lg font-medium" style="background: var(--teal); color: white; opacity: {sending ? 0.5 : 1}">
						{sending ? 'Scheduling...' : 'Schedule'}
					</button>
					<button on:click={() => showSchedule = false} class="px-4 py-2 rounded-lg font-medium border" style="background: var(--surface); border-color: var(--border); color: var(--text-primary)">Cancel</button>
				</div>
			</div>
		{/if}

		<!-- Stats (for sent campaigns) -->
		{#if campaign.status === 'sent' || campaign.status === 'scheduled'}
			<div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
				<div class="rounded-lg shadow border p-5" style="background: var(--surface); border-color: var(--border)">
					<div class="text-sm font-medium" style="color: var(--text-secondary)">Recipients</div>
					<div class="text-2xl font-bold mt-1" style="color: var(--text-primary)">{campaign.recipient_count}</div>
				</div>
				<div class="rounded-lg shadow border p-5" style="background: var(--surface); border-color: var(--border)">
					<div class="text-sm font-medium" style="color: var(--text-secondary)">Opened</div>
					<div class="text-2xl font-bold mt-1" style="color: var(--teal)">{campaign.opened_count}</div>
					{#if campaign.recipient_count > 0}
						<div class="text-xs mt-1" style="color: var(--text-secondary)">{((campaign.opened_count / campaign.recipient_count) * 100).toFixed(1)}%</div>
					{/if}
				</div>
				<div class="rounded-lg shadow border p-5" style="background: var(--surface); border-color: var(--border)">
					<div class="text-sm font-medium" style="color: var(--text-secondary)">Clicked</div>
					<div class="text-2xl font-bold mt-1" style="color: var(--text-primary)">{campaign.clicked_count}</div>
					{#if campaign.recipient_count > 0}
						<div class="text-xs mt-1" style="color: var(--text-secondary)">{((campaign.clicked_count / campaign.recipient_count) * 100).toFixed(1)}%</div>
					{/if}
				</div>
				<div class="rounded-lg shadow border p-5" style="background: var(--surface); border-color: var(--border)">
					<div class="text-sm font-medium" style="color: var(--text-secondary)">{campaign.status === 'scheduled' ? 'Scheduled For' : 'Sent At'}</div>
					<div class="text-sm font-bold mt-1" style="color: var(--text-primary)">
						{campaign.status === 'scheduled' ? formatDate(campaign.scheduled_at) : formatDate(campaign.sent_at)}
					</div>
				</div>
			</div>
		{/if}

		<!-- Campaign Details -->
		<div class="rounded-lg shadow border p-6 mb-6" style="background: var(--surface); border-color: var(--border)">
			<h2 class="text-lg font-semibold mb-4" style="color: var(--text-primary)">Campaign Details</h2>
			<div class="space-y-3">
				<div class="grid grid-cols-3 gap-4 text-sm">
					<div>
						<span class="font-medium" style="color: var(--text-secondary)">Channel</span>
						<div style="color: var(--text-primary)">{#if campaign.channel === 'email'}<Mail size={14} class="inline" /> Email{:else}💬 SMS{/if}</div>
					</div>
					<div>
						<span class="font-medium" style="color: var(--text-secondary)">Target</span>
						<div style="color: var(--text-primary)" class="capitalize">{campaign.target_type}{campaign.target_id ? `: ${campaign.target_id}` : ''}</div>
					</div>
					<div>
						<span class="font-medium" style="color: var(--text-secondary)">Created</span>
						<div style="color: var(--text-primary)">{formatDate(campaign.created_at)}</div>
					</div>
				</div>
				{#if campaign.subject}
					<div>
						<span class="text-sm font-medium" style="color: var(--text-secondary)">Subject</span>
						<div class="font-medium" style="color: var(--text-primary)">{campaign.subject}</div>
					</div>
				{/if}
			</div>
		</div>

		<!-- Message Preview -->
		<div class="rounded-lg shadow border p-6 mb-6" style="background: var(--surface); border-color: var(--border)">
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-lg font-semibold" style="color: var(--text-primary)">Message</h2>
				<button on:click={() => showPreview = !showPreview} class="text-sm font-medium" style="color: var(--teal)">
					{showPreview ? 'Show Source' : 'Preview'}
				</button>
			</div>
			{#if showPreview}
				<div class="rounded-lg p-6 border" style="background: white; color: #333; border-color: var(--border)">
					{@html campaign.body.replace(/\n/g, '<br>')}
				</div>
			{:else}
				<pre class="whitespace-pre-wrap text-sm p-4 rounded-lg" style="background: var(--bg); color: var(--text-primary)">{campaign.body}</pre>
			{/if}
		</div>

		<!-- Recipients -->
		{#if recipients.length > 0}
			<div class="rounded-lg shadow border overflow-hidden" style="background: var(--surface); border-color: var(--border)">
				<div class="px-6 py-4 border-b" style="border-color: var(--border)">
					<h2 class="text-lg font-semibold" style="color: var(--text-primary)">Recipients ({recipients.length})</h2>
				</div>
				<table class="w-full">
					<thead style="background: var(--surface-hover)">
						<tr>
							<th class="text-left px-6 py-3 text-sm font-medium" style="color: var(--text-secondary)">Person</th>
							<th class="text-left px-6 py-3 text-sm font-medium" style="color: var(--text-secondary)">Status</th>
							<th class="text-left px-6 py-3 text-sm font-medium" style="color: var(--text-secondary)">Sent</th>
							<th class="text-left px-6 py-3 text-sm font-medium" style="color: var(--text-secondary)">Opened</th>
						</tr>
					</thead>
					<tbody class="divide-y" style="border-color: var(--border)">
						{#each recipients as r}
							<tr>
								<td class="px-6 py-3 text-sm" style="color: var(--text-primary)">{r.person_id}</td>
								<td class="px-6 py-3">
									<span class="px-2 py-0.5 rounded-full text-xs font-medium capitalize" style="background: var(--bg); color: var(--text-secondary)">{r.status}</span>
								</td>
								<td class="px-6 py-3 text-sm" style="color: var(--text-secondary)">{formatDate(r.sent_at)}</td>
								<td class="px-6 py-3 text-sm" style="color: var(--text-secondary)">{formatDate(r.opened_at)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	{/if}
</div>
