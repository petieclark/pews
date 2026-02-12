<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { refreshModules } from '$lib/stores/modules';

	let modules = [];
	let loading = true;
	let error = '';
	let saving = {};

	onMount(async () => {
		try {
			modules = await api('/api/tenant/modules');
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	});

	async function toggleModule(mod) {
		const newState = !mod.enabled;
		const action = newState ? 'enable' : 'disable';
		saving[mod.name] = true;
		saving = saving;
		try {
			await api(`/api/tenant/modules/${mod.name}/${action}`, { method: 'POST' });
			modules = modules.map(m => m.name === mod.name ? { ...m, enabled: newState } : m);
			await refreshModules();
		} catch (err) {
			error = err.message;
		} finally {
			saving[mod.name] = false;
			saving = saving;
		}
	}
</script>

<div class="max-w-2xl">
	<a href="/dashboard/settings" class="text-sm text-secondary hover:text-[var(--teal)] mb-4 inline-block">← Back to Settings</a>
	<h1 class="text-2xl sm:text-3xl font-bold text-primary mb-2">Modules</h1>
	<p class="text-secondary mb-6">Enable or disable features for your church. Disabled modules are hidden from navigation.</p>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else if error}
		<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg">{error}</div>
	{:else}
		<div class="space-y-3">
			{#each modules as mod}
				<div class="bg-surface rounded-lg border border-custom p-4 flex items-center justify-between gap-4">
					<div class="flex-1 min-w-0">
						<h3 class="font-medium text-primary">{mod.display_name}</h3>
						<p class="text-sm text-secondary truncate">{mod.description}</p>
					</div>
					<button
						on:click={() => toggleModule(mod)}
						disabled={saving[mod.name]}
						class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-[var(--teal)] focus:ring-offset-2
							{mod.enabled ? 'bg-[var(--teal)]' : 'bg-[var(--surface-hover)]'}"
						role="switch"
						aria-checked={mod.enabled}
						aria-label="Toggle {mod.display_name}"
					>
						<span class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out
							{mod.enabled ? 'translate-x-5' : 'translate-x-0'}"></span>
					</button>
				</div>
			{/each}
		</div>
	{/if}
</div>
