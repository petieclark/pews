<script>
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { getToken, clearToken } from '$lib/api';
	import ThemeToggle from '$lib/ThemeToggle.svelte';
	import LanguageSelector from '$lib/LanguageSelector.svelte';
	import { t } from '$lib/i18n.js';

	let email = '';
	let translate;

	// Subscribe to translation updates
	const unsubscribe = t.subscribe(value => {
		translate = value;
	});

	onMount(() => {
		if (!getToken()) {
			goto('/login');
			return;
		}
		email = localStorage.getItem('email') || '';
		
		return unsubscribe;
	});

	function logout() {
		clearToken();
		localStorage.clear();
		goto('/login');
	}
</script>

<div class="min-h-screen bg-[var(--bg)]">
	<nav class="bg-surface shadow-sm border-b border-custom">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex justify-between h-16">
				<div class="flex items-center space-x-8">
					<a href="/dashboard" class="text-2xl font-bold text-[var(--text-primary)]">Pews</a>
					{#if translate}
						<a href="/dashboard" class="text-secondary hover:text-primary">{translate('nav.dashboard')}</a>
						<a href="/dashboard/people" class="text-secondary hover:text-primary">{translate('nav.people')}</a>
						<a href="/dashboard/groups" class="text-secondary hover:text-primary">{translate('nav.groups')}</a>
						<a href="/dashboard/services" class="text-secondary hover:text-primary">{translate('nav.services')}</a>
						<a href="/dashboard/checkins" class="text-secondary hover:text-primary">{translate('nav.checkins')}</a>
						<a href="/dashboard/streaming" class="text-secondary hover:text-primary">{translate('nav.streaming')}</a>
						<a href="/dashboard/giving" class="text-secondary hover:text-primary">{translate('nav.giving')}</a>
						<a href="/dashboard/communication" class="text-secondary hover:text-primary">{translate('nav.communication')}</a>
						<a href="/dashboard/settings" class="text-secondary hover:text-primary">{translate('nav.settings')}</a>
					{/if}
				</div>
				<div class="flex items-center space-x-4">
					<LanguageSelector />
					<ThemeToggle />
					<span class="text-sm text-secondary">{email}</span>
					{#if translate}
						<button
							on:click={logout}
							class="text-sm text-secondary hover:text-primary"
						>
							{translate('nav.logout')}
						</button>
					{/if}
				</div>
			</div>
		</div>
	</nav>

	<main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
		<slot />
	</main>
</div>
