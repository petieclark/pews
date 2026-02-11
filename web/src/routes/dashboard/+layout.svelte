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
	<nav class="bg-surface shadow-sm border-b border-custom">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex justify-between h-16">
				<div class="flex items-center space-x-8">
					<a href="/dashboard" class="text-2xl font-bold text-[var(--text-primary)]">Pews</a>
					<a href="/dashboard" class="text-secondary hover:text-primary">Dashboard</a>
					<a href="/dashboard/people" class="text-secondary hover:text-primary">People</a>
					<a href="/dashboard/groups" class="text-secondary hover:text-primary">Groups</a>
					<a href="/dashboard/services" class="text-secondary hover:text-primary">Services</a>
					<a href="/dashboard/calendar" class="text-secondary hover:text-primary">Calendar</a>
					<a href="/dashboard/checkins" class="text-secondary hover:text-primary">Check-Ins</a>
					<a href="/dashboard/streaming" class="text-secondary hover:text-primary">Streaming</a>
					<a href="/dashboard/giving" class="text-secondary hover:text-primary">Giving</a>
					<a href="/dashboard/communication" class="text-secondary hover:text-primary">Communication</a>
					<a href="/dashboard/settings" class="text-secondary hover:text-primary">Settings</a>
				</div>
				<div class="flex items-center space-x-4">
					<ThemeToggle />
					<span class="text-sm text-secondary">{email}</span>
					<button
						on:click={logout}
						class="text-sm text-secondary hover:text-primary"
					>
						Logout
					</button>
				</div>
			</div>
		</div>
	</nav>

	<main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
		<slot />
	</main>
</div>
