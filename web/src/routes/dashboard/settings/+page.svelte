<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let tenant = {};
	let modules = [];
	let subscription = {};
	let loading = true;
	let saving = false;
	let error = '';
	let success = '';

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
				body: JSON.stringify({ name: tenant.name, domain: tenant.domain || '' })
			});
			success = 'Settings updated successfully';
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
	<h1 class="text-3xl font-bold text-navy mb-6">Settings</h1>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-teal"></div>
		</div>
	{:else}
		<!-- Tenant Settings -->
		<div class="bg-white rounded-lg shadow-md p-6 mb-6">
			<h2 class="text-xl font-semibold text-navy mb-4">Church Information</h2>
			
			<form on:submit|preventDefault={updateTenant} class="space-y-4">
				<div>
					<label for="name" class="block text-sm font-medium text-gray-700 mb-1">
						Church Name
					</label>
					<input
						id="name"
						type="text"
						bind:value={tenant.name}
						required
						class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-teal focus:border-transparent"
					/>
				</div>

				<div>
					<label for="slug" class="block text-sm font-medium text-gray-700 mb-1">
						Slug (read-only)
					</label>
					<input
						id="slug"
						type="text"
						value={tenant.slug}
						disabled
						class="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500"
					/>
				</div>

				<div>
					<label for="domain" class="block text-sm font-medium text-gray-700 mb-1">
						Custom Domain (optional)
					</label>
					<input
						id="domain"
						type="text"
						bind:value={tenant.domain}
						placeholder="church.example.com"
						class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-teal focus:border-transparent"
					/>
				</div>

				{#if error}
					<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
						{error}
					</div>
				{/if}

				{#if success}
					<div class="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg">
						{success}
					</div>
				{/if}

				<button
					type="submit"
					disabled={saving}
					class="bg-teal text-white py-2 px-6 rounded-lg font-medium hover:bg-teal/90 disabled:opacity-50"
				>
					{saving ? 'Saving...' : 'Save Changes'}
				</button>
			</form>
		</div>

		<!-- Module Management -->
		<div class="bg-white rounded-lg shadow-md p-6 mb-6">
			<h2 class="text-xl font-semibold text-navy mb-4">Modules</h2>
			
			<div class="space-y-4">
				{#each modules as module}
					<div class="flex items-center justify-between p-4 border border-gray-200 rounded-lg">
						<div class="flex-1">
							<h3 class="font-medium text-navy">{module.display_name}</h3>
							<p class="text-sm text-gray-600">{module.description}</p>
						</div>
						<label class="relative inline-flex items-center cursor-pointer">
							<input
								type="checkbox"
								checked={module.enabled}
								on:change={(e) => toggleModule(module.name, e.target.checked)}
								class="sr-only peer"
							/>
							<div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-teal/30 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-teal"></div>
						</label>
					</div>
				{/each}
			</div>
		</div>

		<!-- Billing -->
		<div class="bg-white rounded-lg shadow-md p-6">
			<h2 class="text-xl font-semibold text-navy mb-4">Subscription</h2>
			
			<div class="flex items-center justify-between mb-4">
				<div>
					<p class="text-sm text-gray-600">Current Plan</p>
					<p class="text-2xl font-bold text-navy capitalize">{subscription.plan}</p>
					<p class="text-sm text-gray-600 capitalize">{subscription.status}</p>
				</div>
				
				{#if subscription.plan === 'free'}
					<button
						on:click={upgradeToPro}
						class="bg-teal text-white py-2 px-6 rounded-lg font-medium hover:bg-teal/90"
					>
						Upgrade to Pro - $29/mo
					</button>
				{:else}
					<button
						on:click={manageBilling}
						class="bg-gray-600 text-white py-2 px-6 rounded-lg font-medium hover:bg-gray-700"
					>
						Manage Billing
					</button>
				{/if}
			</div>

			{#if subscription.plan === 'free'}
				<div class="bg-sage/10 border border-sage rounded-lg p-4">
					<h3 class="font-semibold text-navy mb-2">Pro Features:</h3>
					<ul class="list-disc list-inside text-sm text-gray-700 space-y-1">
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
