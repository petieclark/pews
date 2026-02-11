<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let subscription = { plan: 'free', status: 'active' };
	let loading = true;
	let error = '';

	onMount(async () => {
		try {
			subscription = await api('/api/billing/subscription');
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	});

	async function upgradeToPro() {
		try {
			const { url } = await api('/api/billing/checkout', { method: 'POST' });
			window.location.href = url;
		} catch (err) {
			error = err.message;
		}
	}

	async function manageBilling() {
		try {
			const { url } = await api('/api/billing/portal', { method: 'POST' });
			window.location.href = url;
		} catch (err) {
			error = err.message;
		}
	}
</script>

<div class="max-w-2xl">
	<a href="/dashboard/settings" class="text-sm text-secondary hover:text-[var(--teal)] mb-4 inline-block">← Back to Settings</a>
	<h1 class="text-2xl sm:text-3xl font-bold text-primary mb-6">Billing</h1>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else}
		{#if error}
			<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg mb-4">{error}</div>
		{/if}

		<div class="bg-surface rounded-lg border border-custom p-6 mb-6">
			<div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
				<div>
					<p class="text-sm text-secondary">Current Plan</p>
					<p class="text-3xl font-bold text-primary capitalize">{subscription.plan}</p>
					<p class="text-sm text-secondary capitalize mt-1">Status: {subscription.status}</p>
					{#if subscription.current_period_end}
						<p class="text-xs text-secondary mt-1">Renews: {new Date(subscription.current_period_end).toLocaleDateString()}</p>
					{/if}
				</div>
				
				{#if subscription.plan === 'free'}
					<button on:click={upgradeToPro}
						class="w-full sm:w-auto bg-[var(--teal)] text-white py-3 px-6 rounded-lg font-medium hover:opacity-90">
						Upgrade to Pro — $100/mo
					</button>
				{:else}
					<button on:click={manageBilling}
						class="w-full sm:w-auto bg-[var(--surface-hover)] text-primary py-3 px-6 rounded-lg font-medium hover:opacity-80 border border-custom">
						Manage Billing
					</button>
				{/if}
			</div>
		</div>

		{#if subscription.plan === 'free'}
			<div class="bg-surface rounded-lg border border-custom p-6">
				<h2 class="text-lg font-semibold text-primary mb-4">Why Upgrade?</h2>
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
					{#each ['Unlimited members', 'Advanced reporting', 'Custom branding', 'Priority support', 'SMS messaging', 'Streaming integration'] as feature}
						<div class="flex items-center gap-2">
							<svg class="w-5 h-5 text-[var(--teal)] flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
							</svg>
							<span class="text-sm text-primary">{feature}</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		{#if subscription.usage}
			<div class="bg-surface rounded-lg border border-custom p-6 mt-6">
				<h2 class="text-lg font-semibold text-primary mb-4">Usage</h2>
				<div class="space-y-3">
					{#if subscription.usage.members !== undefined}
						<div>
							<div class="flex justify-between text-sm mb-1">
								<span class="text-secondary">Members</span>
								<span class="text-primary">{subscription.usage.members}{subscription.usage.members_limit ? ` / ${subscription.usage.members_limit}` : ''}</span>
							</div>
							{#if subscription.usage.members_limit}
								<div class="h-2 bg-[var(--surface-hover)] rounded-full overflow-hidden">
									<div class="h-full bg-[var(--teal)] rounded-full" style="width: {Math.min(100, (subscription.usage.members / subscription.usage.members_limit) * 100)}%"></div>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>
		{/if}
	{/if}
</div>
