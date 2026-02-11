<script>
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { getToken, clearToken } from '$lib/api';
	import ThemeToggle from '$lib/ThemeToggle.svelte';

	let email = '';

	onMount(() => {
		if (!getToken()) {
			goto('/login');
			return;
		}
		email = localStorage.getItem('email') || '';
	});

	function logout() {
		clearToken();
		localStorage.clear();
		goto('/login');
	}
</script>

<div class="min-h-screen bg-[var(--bg)]">
	<!-- Skip Navigation -->
	<a href="#main-content" class="skip-link">Skip to main content</a>
	
	<nav class="bg-surface shadow-sm border-b border-custom" aria-label="Main navigation">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex justify-between h-16">
				<div class="flex items-center space-x-8">
					<a href="/dashboard" class="text-2xl font-bold text-[var(--text-primary)]" aria-label="Pews home">Pews</a>
					<a href="/dashboard" class="text-secondary hover:text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)] focus:ring-offset-2 rounded px-2 py-1">Dashboard</a>
					<a href="/dashboard/people" class="text-secondary hover:text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)] focus:ring-offset-2 rounded px-2 py-1">People</a>
					<a href="/dashboard/groups" class="text-secondary hover:text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)] focus:ring-offset-2 rounded px-2 py-1">Groups</a>
					<a href="/dashboard/services" class="text-secondary hover:text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)] focus:ring-offset-2 rounded px-2 py-1">Services</a>
					<a href="/dashboard/checkins" class="text-secondary hover:text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)] focus:ring-offset-2 rounded px-2 py-1">Check-Ins</a>
					<a href="/dashboard/streaming" class="text-secondary hover:text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)] focus:ring-offset-2 rounded px-2 py-1">Streaming</a>
					<a href="/dashboard/giving" class="text-secondary hover:text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)] focus:ring-offset-2 rounded px-2 py-1">Giving</a>
					<a href="/dashboard/communication" class="text-secondary hover:text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)] focus:ring-offset-2 rounded px-2 py-1">Communication</a>
					<a href="/dashboard/settings" class="text-secondary hover:text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)] focus:ring-offset-2 rounded px-2 py-1">Settings</a>
				</div>
				<div class="flex items-center space-x-4">
					<ThemeToggle />
					<span class="text-sm text-secondary" aria-label="Logged in as {email}">{email}</span>
					<button
						on:click={logout}
						class="text-sm text-secondary hover:text-primary focus:outline-none focus:ring-2 focus:ring-[var(--teal)] focus:ring-offset-2 rounded px-2 py-1"
						aria-label="Log out of Pews"
					>
						Logout
					</button>
				</div>
			</div>
		</div>
	</nav>

	<main id="main-content" class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
		<slot />
	</main>
</div>

<style>
	.skip-link {
		position: absolute;
		top: -40px;
		left: 0;
		background: var(--teal);
		color: white;
		padding: 8px 16px;
		text-decoration: none;
		border-radius: 0 0 4px 0;
		z-index: 100;
	}
	
	.skip-link:focus {
		top: 0;
	}
</style>
