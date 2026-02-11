<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let campaigns = [];
	let loading = true;
	let error = '';

	onMount(async () => {
		await loadCampaigns();
	});

	async function loadCampaigns() {
		try {
			loading = true;
			campaigns = await api('/api/drip/campaigns');
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	async function toggleActive(campaign) {
		try {
			await api(`/api/drip/campaigns/${campaign.id}`, {
				method: 'PUT',
				body: JSON.stringify({ is_active: !campaign.is_active })
			});
			await loadCampaigns();
		} catch (err) {
			error = err.message;
		}
	}

	async function deleteCampaign(id) {
		if (!confirm('Are you sure you want to delete this campaign?')) return;
		
		try {
			await api(`/api/drip/campaigns/${id}`, { method: 'DELETE' });
			await loadCampaigns();
		} catch (err) {
			error = err.message;
		}
	}

	function getTriggerLabel(trigger) {
		const labels = {
			'new_member': 'New Member',
			'connection_card': 'Connection Card',
			'first_visit': 'First Visit'
		};
		return labels[trigger] || trigger;
	}
</script>

<div>
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-3xl font-bold" style="color: var(--text)">Drip Campaigns</h1>
			<p class="mt-1" style="color: var(--text-secondary)">Automated sequences triggered by specific events</p>
		</div>
		<button
			on:click={() => goto('/dashboard/communication/drip/new')}
			class="px-4 py-2 rounded-lg font-medium"
			style="background: var(--teal); color: white"
		>
			Create Campaign
		</button>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2" style="border-color: var(--teal)"></div>
		</div>
	{:else if error}
		<div class="px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">
			{error}
		</div>
	{:else if campaigns.length === 0}
		<div class="rounded-lg shadow border p-12 text-center" style="background: var(--surface); border-color: var(--border)">
			<div class="text-5xl mb-4">💧</div>
			<h2 class="text-xl font-semibold mb-2" style="color: var(--text)">No drip campaigns yet</h2>
			<p class="mb-6" style="color: var(--text-secondary)">Create automated communication sequences triggered by visitor and member actions</p>
			<button
				on:click={() => goto('/dashboard/communication/drip/new')}
				class="px-6 py-2 rounded-lg font-medium"
				style="background: var(--teal); color: white"
			>
				Create First Campaign
			</button>
		</div>
	{:else}
		<div class="space-y-4">
			{#each campaigns as campaign}
				<div class="rounded-lg shadow border p-6" style="background: var(--surface); border-color: var(--border)">
					<div class="flex items-start justify-between">
						<div class="flex-1">
							<div class="flex items-center gap-3 mb-2">
								<h3 class="font-semibold text-lg" style="color: var(--text)">{campaign.name}</h3>
								<button
									on:click={() => toggleActive(campaign)}
									class="px-3 py-1 rounded-full text-xs font-medium"
									style="background: {campaign.is_active ? 'var(--teal)' : 'var(--surface-hover)'}; color: {campaign.is_active ? 'white' : 'var(--text-secondary)'}"
								>
									{campaign.is_active ? 'Active' : 'Inactive'}
								</button>
							</div>

							<div class="grid grid-cols-3 gap-4 text-sm" style="color: var(--text-secondary)">
								<div>
									<span class="font-medium">Trigger:</span>
									<span class="ml-1">{getTriggerLabel(campaign.trigger_event)}</span>
								</div>
								<div>
									<span class="font-medium">Steps:</span>
									<span class="ml-1">{campaign.steps?.length || 0}</span>
								</div>
								<div>
									<span class="font-medium">Enrolled:</span>
									<span class="ml-1">{campaign.enrollment_count || 0} people</span>
								</div>
							</div>
						</div>

						<div class="flex gap-2 ml-4">
							<button
								on:click={() => goto(`/dashboard/communication/drip/${campaign.id}`)}
								class="px-4 py-2 rounded-lg font-medium border"
								style="background: var(--bg); border-color: var(--border); color: var(--text)"
							>
								Edit
							</button>
							<button
								on:click={() => deleteCampaign(campaign.id)}
								class="px-4 py-2 rounded-lg font-medium border"
								style="background: var(--bg); border-color: #fcc; color: #c33"
							>
								Delete
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
