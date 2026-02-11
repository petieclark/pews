<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { t, localeNames } from '$lib/i18n.js';

	let tenant = {};
	let modules = [];
	let subscription = {};
	let loading = true;
	let saving = false;
	let error = '';
	let success = '';
	let translate;
	let supportedLocales = ['en', 'es', 'pt', 'ko'];

	// Subscribe to translation updates
	const unsubscribe = t.subscribe(value => {
		translate = value;
	});

	onMount(async () => {
		try {
			const [tenantData, modulesData, subData] = await Promise.all([
				api('/api/tenant'),
				api('/api/tenant/modules'),
				api('/api/billing/subscription')
			]);
			tenant = tenantData;
			modules = modulesData;
			subscription = subData;
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	});

	async function updateTenant() {
		saving = true;
		error = '';
		success = '';

		try {
			await api('/api/tenant', {
				method: 'PUT',
				body: JSON.stringify({ 
					name: tenant.name, 
					domain: tenant.domain || '',
					default_locale: tenant.default_locale || 'en'
				})
			});
			success = translate ? translate('common.success') : 'Settings updated successfully';
		} catch (err) {
			error = err.message;
		} finally {
			saving = false;
		}
	}

	async function toggleModule(moduleName, enabled) {
		const action = enabled ? 'enable' : 'disable';
		try {
			await api(`/api/tenant/modules/${moduleName}/${action}`, { method: 'POST' });
			modules = modules.map(m => 
				m.name === moduleName ? { ...m, enabled } : m
			);
		} catch (err) {
			error = err.message;
		}
	}

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

<div class="max-w-4xl">
	<h1 class="text-3xl font-bold text-primary mb-6">{translate ? translate('settings.title') : 'Settings'}</h1>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else}
		<!-- Tenant Settings -->
		<div class="bg-surface rounded-lg shadow-md p-6 mb-6 border border-custom">
			<h2 class="text-xl font-semibold text-primary mb-4">Church Information</h2>
			
			<form on:submit|preventDefault={updateTenant} class="space-y-4">
				<div>
					<label for="name" class="block text-sm font-medium text-primary mb-1">
						Church Name
					</label>
					<input
						id="name"
						type="text"
						bind:value={tenant.name}
						required
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
					/>
				</div>

				<div>
					<label for="slug" class="block text-sm font-medium text-primary mb-1">
						Slug (read-only)
					</label>
					<input
						id="slug"
						type="text"
						value={tenant.slug}
						disabled
						class="w-full px-4 py-2 border border-custom rounded-lg bg-[var(--surface-hover)] text-secondary"
					/>
				</div>

				<div>
					<label for="domain" class="block text-sm font-medium text-primary mb-1">
						Custom Domain (optional)
					</label>
					<input
						id="domain"
						type="text"
						bind:value={tenant.domain}
						placeholder="church.example.com"
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
					/>
				</div>

				<div>
					<label for="locale" class="block text-sm font-medium text-primary mb-1">
						{translate ? translate('settings.language') : 'Language'}
					</label>
					<select
						id="locale"
						bind:value={tenant.default_locale}
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
					>
						{#each supportedLocales as loc}
							<option value={loc}>{localeNames[loc] || loc}</option>
						{/each}
					</select>
					<p class="text-xs text-secondary mt-1">Default language for your church</p>
				</div>

				{#if error}
					<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg">
						{error}
					</div>
				{/if}

				{#if success}
					<div class="success-box">
						{success}
					</div>
				{/if}

				<button
					type="submit"
					disabled={saving}
					class="bg-[var(--teal)] text-white py-2 px-6 rounded-lg font-medium hover:opacity-90 disabled:opacity-50"
				>
					{saving ? (translate ? translate('common.loading') : 'Saving...') : (translate ? translate('common.save') : 'Save Changes')}
				</button>
			</form>
		</div>

		<!-- Module Management -->
		<div class="bg-surface rounded-lg shadow-md p-6 mb-6 border border-custom">
			<h2 class="text-xl font-semibold text-primary mb-4">Modules</h2>
			
			<div class="space-y-4">
				{#each modules as module}
					<div class="flex items-center justify-between p-4 border border-custom rounded-lg hover:bg-[var(--surface-hover)]">
						<div class="flex-1">
							<h3 class="font-medium text-primary">{module.display_name}</h3>
							<p class="text-sm text-secondary">{module.description}</p>
						</div>
						<label class="relative inline-flex items-center cursor-pointer">
							<input
								type="checkbox"
								checked={module.enabled}
								on:change={(e) => toggleModule(module.name, e.target.checked)}
								class="sr-only peer"
							/>
							<div class="w-11 h-6 bg-[var(--surface-hover)] peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-[var(--teal)]/30 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-[var(--border)] after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-[var(--teal)]"></div>
						</label>
					</div>
				{/each}
			</div>
		</div>

		<!-- Billing -->
		<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
			<h2 class="text-xl font-semibold text-primary mb-4">Subscription</h2>
			
			<div class="flex items-center justify-between mb-4">
				<div>
					<p class="text-sm text-secondary">Current Plan</p>
					<p class="text-2xl font-bold text-primary capitalize">{subscription.plan}</p>
					<p class="text-sm text-secondary capitalize">{subscription.status}</p>
				</div>
				
				{#if subscription.plan === 'free'}
					<button
						on:click={upgradeToPro}
						class="bg-[var(--teal)] text-white py-2 px-6 rounded-lg font-medium hover:opacity-90"
					>
						Upgrade to Pro - $100/mo
					</button>
				{:else}
					<button
						on:click={manageBilling}
						class="bg-[var(--text-secondary)] text-white py-2 px-6 rounded-lg font-medium hover:opacity-90"
					>
						Manage Billing
					</button>
				{/if}
			</div>

			{#if subscription.plan === 'free'}
				<div class="pro-features-box">
					<h3 class="font-semibold text-primary mb-2">Pro Features:</h3>
					<ul class="list-disc list-inside text-sm text-secondary space-y-1">
						<li>Unlimited members</li>
						<li>Advanced reporting</li>
						<li>Custom branding</li>
						<li>Priority support</li>
					</ul>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.success-box {
		background-color: #D1FAE5;
		border: 1px solid #6EE7B7;
		color: #065F46;
		padding: 1rem;
		border-radius: 0.5rem;
	}
	:global(.dark) .success-box {
		background-color: #064E3B;
		border-color: #059669;
		color: #6EE7B7;
	}
	
	.pro-features-box {
		background-color: #CCFBF1;
		border: 1px solid #5EEAD4;
		padding: 1rem;
		border-radius: 0.5rem;
	}
	:global(.dark) .pro-features-box {
		background-color: #134E4A;
		border-color: #14B8A6;
	}
</style>
