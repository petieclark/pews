<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let modules = [];
	let tenant = null;
	let loading = true;
	let error = '';
	let isFirstTime = false;
	let showWelcome = false;

	onMount(async () => {
		try {
			// Check if this is first login
			const hasSeenWelcome = localStorage.getItem('has_seen_welcome');
			if (!hasSeenWelcome) {
				showWelcome = true;
				localStorage.setItem('has_seen_welcome', 'true');
			}

			// Fetch modules and tenant info
			const [modulesData, tenantData] = await Promise.all([
				api('/api/tenant/modules'),
				api('/api/tenant')
			]);
			
			modules = modulesData;
			tenant = tenantData;
			
			// Check if user hasn't enabled any modules yet
			isFirstTime = !modules.some(m => m.enabled);
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	});

	function dismissWelcome() {
		showWelcome = false;
	}
</script>

<div>
	{#if showWelcome && !loading}
		<div class="mb-8 bg-gradient-to-r from-[var(--teal)] to-[var(--sage)] rounded-lg p-6 text-white shadow-lg">
			<div class="flex items-start justify-between">
				<div class="flex-1">
					<h2 class="text-2xl font-bold mb-2">🎉 Welcome to Pews!</h2>
					<p class="mb-4 opacity-90">
						You're all set up and ready to go. Let's get started by enabling some modules and setting up your church profile.
					</p>
					<div class="flex flex-wrap gap-3">
						<a
							href="/dashboard/settings"
							class="inline-block bg-white text-[var(--navy)] px-4 py-2 rounded-lg font-medium hover:bg-opacity-90 transition-opacity"
						>
							⚙️ Set Up Church Profile
						</a>
						<button
							on:click={dismissWelcome}
							class="inline-block bg-white bg-opacity-20 text-white px-4 py-2 rounded-lg font-medium hover:bg-opacity-30 transition-opacity"
						>
							I'll do this later
						</button>
					</div>
				</div>
				<button
					on:click={dismissWelcome}
					class="ml-4 text-white hover:text-gray-200 focus:outline-none"
					aria-label="Dismiss welcome message"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
					</svg>
				</button>
			</div>
		</div>
	{/if}

	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-3xl font-bold text-primary">Dashboard</h1>
			{#if tenant}
				<p class="text-secondary mt-1">Welcome to {tenant.name}</p>
			{/if}
		</div>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
			<p class="text-secondary mt-3">Loading your workspace...</p>
		</div>
	{:else if error}
		<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg">
			{error}
		</div>
	{:else}
		{#if isFirstTime}
			<!-- First-time user onboarding -->
			<div class="mb-8 bg-blue-50 dark:bg-blue-900 border-2 border-blue-200 dark:border-blue-700 rounded-lg p-6">
				<h2 class="text-xl font-semibold text-primary mb-3">🚀 Quick Start Guide</h2>
				<p class="text-secondary mb-4">
					Get the most out of Pews by completing these steps:
				</p>
				<div class="space-y-3">
					<div class="flex items-start gap-3">
						<div class="flex-shrink-0 w-6 h-6 rounded-full bg-blue-500 text-white flex items-center justify-center text-sm font-bold">
							1
						</div>
						<div class="flex-1">
							<h3 class="font-medium text-primary">Enable Your First Module</h3>
							<p class="text-sm text-secondary">
								Choose from People Management, Giving, Services, Groups, or Check-ins below
							</p>
						</div>
					</div>
					<div class="flex items-start gap-3">
						<div class="flex-shrink-0 w-6 h-6 rounded-full bg-blue-500 text-white flex items-center justify-center text-sm font-bold">
							2
						</div>
						<div class="flex-1">
							<h3 class="font-medium text-primary">Complete Church Profile</h3>
							<p class="text-sm text-secondary">
								Add your church's information in <a href="/dashboard/settings" class="text-[var(--teal)] hover:underline">Settings</a>
							</p>
						</div>
					</div>
					<div class="flex items-start gap-3">
						<div class="flex-shrink-0 w-6 h-6 rounded-full bg-blue-500 text-white flex items-center justify-center text-sm font-bold">
							3
						</div>
						<div class="flex-1">
							<h3 class="font-medium text-primary">Start Using Features</h3>
							<p class="text-sm text-secondary">
								Add members, schedule services, or set up giving
							</p>
						</div>
					</div>
				</div>
			</div>
		{/if}

		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each modules as module}
				<div class="bg-surface rounded-lg shadow-md p-6 border transition-all hover:shadow-lg {module.enabled ? 'border-[var(--teal)]' : 'border-custom'}">
					<div class="flex items-start justify-between mb-3">
						<h3 class="text-xl font-semibold text-primary">{module.display_name}</h3>
						{#if module.enabled}
							<span class="bg-[var(--teal)] text-white text-xs px-2 py-1 rounded-full font-medium">Active</span>
						{:else}
							<span class="bg-[var(--surface-hover)] text-secondary text-xs px-2 py-1 rounded-full">Disabled</span>
						{/if}
					</div>
					<p class="text-secondary text-sm mb-4">{module.description}</p>
					{#if module.enabled}
						<a
							href="/dashboard/{module.module_name}"
							class="block w-full bg-[var(--teal)] text-white py-2 px-4 rounded-lg font-medium hover:opacity-90 transition-opacity text-center"
						>
							Open {module.display_name}
						</a>
					{:else}
						<a
							href="/dashboard/settings"
							class="block w-full bg-[var(--surface-hover)] text-primary py-2 px-4 rounded-lg font-medium hover:bg-[var(--teal)] hover:text-white transition-colors text-center"
						>
							Enable Module
						</a>
					{/if}
				</div>
			{/each}
		</div>

		{#if modules.filter(m => !m.enabled).length > 0}
			<div class="mt-8 bg-surface border border-[var(--sage)] rounded-lg p-6">
				<h2 class="text-lg font-semibold text-primary mb-2">💡 Want more features?</h2>
				<p class="text-secondary mb-4">
					Enable additional modules in your settings to unlock more functionality for your church management.
				</p>
				<a
					href="/dashboard/settings"
					class="inline-block bg-[var(--teal)] text-white py-2 px-6 rounded-lg font-medium hover:opacity-90 transition-opacity"
				>
					⚙️ Manage Modules
				</a>
			</div>
		{/if}
	{/if}
</div>
