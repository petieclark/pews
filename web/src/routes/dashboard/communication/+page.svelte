<script>
	import { onMount } from 'svelte';
	import { Mail, FileEdit, CreditCard } from 'lucide-svelte';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let stats = {
		emails_sent_this_month: 0,
		sms_sent_this_month: 0,
		open_rate: 0,
		active_journeys: 0,
		unprocessed_cards: 0
	};
	let campaigns = [];
	let loading = true;
	let error = '';

	onMount(async () => {
		try {
			const [statsData, campaignsData] = await Promise.all([
				api('/api/communication/stats'),
				api('/api/communication/campaigns?status=sent')
			]);
			stats = statsData;
			campaigns = campaignsData.slice(0, 5); // Recent 5
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	});
</script>

<div>
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-3xl font-bold" style="color: var(--text-primary)">Communication</h1>
		<div class="flex gap-3">
			<button
				on:click={() => goto('/dashboard/communication/campaigns/new')}
				class="px-4 py-2 rounded-lg font-medium"
				style="background: var(--teal); color: white"
			>
				New Campaign
			</button>
			<button
				on:click={() => goto('/dashboard/communication/journeys')}
				class="px-4 py-2 rounded-lg font-medium border"
				style="background: var(--surface); border-color: var(--border); color: var(--text-primary)"
			>
				Manage Journeys
			</button>
		</div>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2" style="border-color: var(--teal)"></div>
		</div>
	{:else if error}
		<div class="px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">
			{error}
		</div>
	{:else}
		<!-- Stats Cards -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
			<div class="rounded-lg shadow p-6 border" style="background: var(--surface); border-color: var(--border)">
				<div class="text-sm font-medium" style="color: var(--text-secondary)">Emails Sent This Month</div>
				<div class="text-3xl font-bold mt-2" style="color: var(--text-primary)">{stats.emails_sent_this_month}</div>
			</div>
			<div class="rounded-lg shadow p-6 border" style="background: var(--surface); border-color: var(--border)">
				<div class="text-sm font-medium" style="color: var(--text-secondary)">SMS Sent This Month</div>
				<div class="text-3xl font-bold mt-2" style="color: var(--text-primary)">{stats.sms_sent_this_month}</div>
			</div>
			<div class="rounded-lg shadow p-6 border" style="background: var(--surface); border-color: var(--border)">
				<div class="text-sm font-medium" style="color: var(--text-secondary)">Open Rate</div>
				<div class="text-3xl font-bold mt-2" style="color: var(--teal)">{stats.open_rate.toFixed(1)}%</div>
			</div>
			<div class="rounded-lg shadow p-6 border" style="background: var(--surface); border-color: var(--border)">
				<div class="text-sm font-medium" style="color: var(--text-secondary)">Active Journeys</div>
				<div class="text-3xl font-bold mt-2" style="color: var(--text-primary)">{stats.active_journeys}</div>
			</div>
		</div>

		{#if stats.unprocessed_cards > 0}
			<div class="mb-6 p-4 rounded-lg border flex items-center justify-between" style="background: #fff3cd; border-color: #ffc107; color: #856404">
				<div>
					<span class="font-semibold">{stats.unprocessed_cards} new connection card{stats.unprocessed_cards !== 1 ? 's' : ''}</span> waiting to be processed
				</div>
				<button
					on:click={() => goto('/dashboard/communication/cards')}
					class="px-4 py-2 rounded-lg font-medium"
					style="background: #ffc107; color: #000"
				>
					View Cards
				</button>
			</div>
		{/if}

		<!-- Recent Campaigns -->
		<div class="rounded-lg shadow border" style="background: var(--surface); border-color: var(--border)">
			<div class="px-6 py-4 border-b flex items-center justify-between" style="border-color: var(--border)">
				<h2 class="text-lg font-semibold" style="color: var(--text-primary)">Recent Campaigns</h2>
				<a href="/dashboard/communication/campaigns" class="text-sm font-medium" style="color: var(--teal)">View All</a>
			</div>
			{#if campaigns.length === 0}
				<div class="px-6 py-12 text-center" style="color: var(--text-secondary)">
					<p>No campaigns yet</p>
					<button
						on:click={() => goto('/dashboard/communication/campaigns/new')}
						class="mt-4 px-4 py-2 rounded-lg font-medium"
						style="background: var(--teal); color: white"
					>
						Create Your First Campaign
					</button>
				</div>
			{:else}
				<div class="divide-y" style="color: var(--border)">
					{#each campaigns as campaign}
						<div class="px-6 py-4 flex items-center justify-between hover:bg-opacity-50" style="background: var(--surface-hover)">
							<div>
								<div class="font-medium" style="color: var(--text-primary)">{campaign.name}</div>
								<div class="text-sm" style="color: var(--text-secondary)">
									{#if campaign.channel === 'email'}<Mail size={14} class="inline" />{:else}💬{/if} {campaign.recipient_count} recipients
									{#if campaign.channel === 'email'}
										• {campaign.recipient_count > 0 ? ((campaign.opened_count / campaign.recipient_count) * 100).toFixed(0) : 0}% opened
									{/if}
								</div>
							</div>
							<div class="text-sm" style="color: var(--text-secondary)">
								{new Date(campaign.sent_at).toLocaleDateString()}
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Quick Actions -->
		<div class="mt-8 grid grid-cols-1 md:grid-cols-3 gap-4">
			<button
				on:click={() => goto('/dashboard/communication/templates')}
				class="p-6 rounded-lg border text-left hover:shadow-md transition"
				style="background: var(--surface); border-color: var(--border)"
			>
				<div class="mb-2"><FileEdit size={24} /></div>
				<div class="font-semibold mb-1" style="color: var(--text-primary)">Message Templates</div>
				<div class="text-sm" style="color: var(--text-secondary)">Create reusable message templates</div>
			</button>
			<button
				on:click={() => goto('/dashboard/communication/journeys')}
				class="p-6 rounded-lg border text-left hover:shadow-md transition"
				style="background: var(--surface); border-color: var(--border)"
			>
				<div class="text-2xl mb-2">🗺️</div>
				<div class="font-semibold mb-1" style="color: var(--text-primary)">Automated Journeys</div>
				<div class="text-sm" style="color: var(--text-secondary)">Set up automated message sequences</div>
			</button>
			<button
				on:click={() => goto('/dashboard/communication/cards')}
				class="p-6 rounded-lg border text-left hover:shadow-md transition"
				style="background: var(--surface); border-color: var(--border)"
			>
				<div class="mb-2"><CreditCard size={24} /></div>
				<div class="font-semibold mb-1" style="color: var(--text-primary)">Connection Cards</div>
				<div class="text-sm" style="color: var(--text-secondary)">View and process visitor submissions</div>
			</button>
		</div>
	{/if}
</div>
