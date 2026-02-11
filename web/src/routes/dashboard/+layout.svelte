<script>
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { getToken, clearToken } from '$lib/api';

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

<div class="min-h-screen bg-bg">
	<nav class="bg-white shadow-sm">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex justify-between h-16">
				<div class="flex items-center space-x-8">
					<a href="/dashboard" class="text-2xl font-bold text-navy">Pews</a>
					<a href="/dashboard" class="text-gray-600 hover:text-navy">Dashboard</a>
					<a href="/dashboard/people" class="text-gray-600 hover:text-navy">People</a>
					<a href="/dashboard/settings" class="text-gray-600 hover:text-navy">Settings</a>
				</div>
				<div class="flex items-center space-x-4">
					<span class="text-sm text-gray-600">{email}</span>
					<button
						on:click={logout}
						class="text-sm text-gray-600 hover:text-navy"
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
