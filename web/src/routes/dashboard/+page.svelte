<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let modules = [];
	let loading = true;
	let error = '';

	onMount(async () => {
		try {
			modules = await api('/api/tenant/modules');
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	});
</script>

<div>
	<h1 class="text-3xl font-bold text-navy mb-6">Dashboard</h1>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-teal"></div>
		</div>
	{:else if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
			{error}
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each modules as module}
				<div class="bg-white rounded-lg shadow-md p-6 {module.enabled ? 'border-2 border-teal' : ''}">
					<div class="flex items-start justify-between mb-3">
						<h3 class="text-xl font-semibold text-navy">{module.display_name}</h3>
						{#if module.enabled}
							<span class="bg-teal text-white text-xs px-2 py-1 rounded-full">Active</span>
						{:else}
							<span class="bg-gray-200 text-gray-600 text-xs px-2 py-1 rounded-full">Disabled</span>
						{/if}
					</div>
					<p class="text-gray-600 text-sm">{module.description}</p>
					{#if module.enabled}
						<button class="mt-4 w-full bg-teal text-white py-2 px-4 rounded-lg font-medium hover:bg-teal/90">
							Open Module
						</button>
					{/if}
				</div>
			{/each}
		</div>

		{#if modules.filter(m => !m.enabled).length > 0}
			<div class="mt-8 bg-sage/10 border border-sage rounded-lg p-6">
				<h2 class="text-lg font-semibold text-navy mb-2">Want more features?</h2>
				<p class="text-gray-600 mb-4">
					Enable additional modules in your settings to unlock more functionality.
				</p>
				<a
					href="/dashboard/settings"
					class="inline-block bg-teal text-white py-2 px-6 rounded-lg font-medium hover:bg-teal/90"
				>
					Manage Modules
				</a>
			</div>
		{/if}
	{/if}
</div>
