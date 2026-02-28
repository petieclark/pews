<script>
	import { onMount } from 'svelte';
	import { Mail, MessageSquare } from 'lucide-svelte';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let campaigns = [];
	let filteredCampaigns = [];
	let loading = true;
	let error = '';
	let statusFilter = 'all';

	onMount(async () => {
		await loadCampaigns();
	});

	async function loadCampaigns() {
		try {
			loading = true;
			campaigns = await api('/api/communication/campaigns');
			filterCampaigns();
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	function filterCampaigns() {
		if (statusFilter === 'all') {
			filteredCampaigns = campaigns;
		} else {
			filteredCampaigns = campaigns.filter(c => c.status === statusFilter);
		}
	}

	$: {
		statusFilter;
		filterCampaigns();
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
</script>

<div>
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-3xl font-bold" style="color: var(--text-primary)">Campaigns</h1>
		<button
			on:click={() => goto('/dashboard/communication/campaigns/new')}
			class="px-4 py-2 rounded-lg font-medium"
			style="background: var(--teal); color: white"
		>
			New Campaign
		</button>
	</div>

	<!-- Status Filter -->
	<div class="mb-6 flex gap-2">
		{#each ['all', 'draft', 'scheduled', 'sent'] as status}
			<button
				on:click={() => statusFilter = status}
				class="px-4 py-2 rounded-lg font-medium transition capitalize"
				style="background: {statusFilter === status ? 'var(--teal)' : 'var(--surface)'}; color: {statusFilter === status ? 'white' : 'var(--text-primary)'}; border: 1px solid {statusFilter === status ? 'var(--teal)' : 'var(--border)'}"
			>
				{status}
			</button>
		{/each}
	</div>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2" style="border-color: var(--teal)"></div>
		</div>
	{:else if error}
		<div class="px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">
			{error}
		</div>
	{:else if filteredCampaigns.length === 0}
		<div class="rounded-lg shadow border p-12 text-center" style="background: var(--surface); border-color: var(--border)">
			<div class="mb-4"><MessageSquare size={48} /></div>
			<h2 class="text-xl font-semibold mb-2" style="color: var(--text-primary)">No campaigns yet</h2>
			<p class="mb-6" style="color: var(--text-secondary)">Create your first campaign to reach your members</p>
			<button
				on:click={() => goto('/dashboard/communication/campaigns/new')}
				class="px-6 py-2 rounded-lg font-medium"
				style="background: var(--teal); color: white"
			>
				Create Campaign
			</button>
		</div>
	{:else}
		<div class="rounded-lg shadow border overflow-hidden" style="background: var(--surface); border-color: var(--border)">
			<table class="w-full">
				<thead style="background: var(--surface-hover)">
					<tr>
						<th class="text-left px-6 py-3 text-sm font-medium" style="color: var(--text-secondary)">Name</th>
						<th class="text-left px-6 py-3 text-sm font-medium" style="color: var(--text-secondary)">Channel</th>
						<th class="text-left px-6 py-3 text-sm font-medium" style="color: var(--text-secondary)">Status</th>
						<th class="text-left px-6 py-3 text-sm font-medium" style="color: var(--text-secondary)">Recipients</th>
						<th class="text-left px-6 py-3 text-sm font-medium" style="color: var(--text-secondary)">Open Rate</th>
						<th class="text-left px-6 py-3 text-sm font-medium" style="color: var(--text-secondary)">Sent</th>
					</tr>
				</thead>
				<tbody class="divide-y" style="border-color: var(--border)">
					{#each filteredCampaigns as campaign}
						<tr
							on:click={() => goto(`/dashboard/communication/campaigns/${campaign.id}`)}
							class="cursor-pointer transition"
							style="background: var(--surface); hover:background: var(--surface-hover)"
						>
							<td class="px-6 py-4">
								<div class="font-medium" style="color: var(--text-primary)">{campaign.name}</div>
							</td>
							<td class="px-6 py-4">
								{#if campaign.channel === 'email'}<Mail size={20} />{:else}<span class="text-lg">💬</span>{/if}
							</td>
							<td class="px-6 py-4">
								{#if campaign.status}
									{@const badge = getStatusBadge(campaign.status)}
									<span class="px-2 py-1 rounded-full text-xs font-medium" style="background: {badge.bg}; color: {badge.color}">
										{badge.text}
									</span>
								{/if}
							</td>
							<td class="px-6 py-4" style="color: var(--text-primary)">
								{campaign.recipient_count || 0}
							</td>
							<td class="px-6 py-4" style="color: var(--text-primary)">
								{#if campaign.channel === 'email' && campaign.recipient_count > 0}
									{((campaign.opened_count / campaign.recipient_count) * 100).toFixed(1)}%
								{:else}
									—
								{/if}
							</td>
							<td class="px-6 py-4 text-sm" style="color: var(--text-secondary)">
								{#if campaign.sent_at}
									{new Date(campaign.sent_at).toLocaleDateString()}
								{:else if campaign.scheduled_at}
									Scheduled: {new Date(campaign.scheduled_at).toLocaleDateString()}
								{:else}
									—
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
