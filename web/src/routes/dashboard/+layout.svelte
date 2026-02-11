<script>
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { getToken, clearToken, api } from '$lib/api';
	import ThemeToggle from '$lib/ThemeToggle.svelte';
	import GlobalSearch from '$lib/GlobalSearch.svelte';
	import NotificationBell from '$lib/NotificationBell.svelte';

	let email = '';
	let churchName = 'Pews';
	let churchLogo = '';
	let showUserDropdown = false;
	let activeNavSection = '';

	// Navigation structure
	const navSections = [
		{
			name: 'Core',
			items: [
				{ href: '/dashboard', label: 'Dashboard' },
				{ href: '/dashboard/people', label: 'People' },
				{ href: '/dashboard/groups', label: 'Groups' }
			]
		},
		{
			name: 'Ministry',
			items: [
				{ href: '/dashboard/services', label: 'Services' },
				{ href: '/dashboard/calendar', label: 'Calendar' },
				{ href: '/dashboard/checkins', label: 'Check-Ins' },
				{ href: '/dashboard/care', label: 'Care' },
				{ href: '/dashboard/communication', label: 'Communication' }
			]
		},
		{
			name: 'Media',
			items: [
				{ href: '/dashboard/streaming', label: 'Streaming' },
				{ href: '/dashboard/media', label: 'Media Library' }
			]
		},
		{
			name: 'Finance',
			items: [
				{ href: '/dashboard/giving', label: 'Giving' }
			]
		},
		{
			name: 'Admin',
			items: [
				{ href: '/dashboard/reports', label: 'Reports' },
				{ href: '/dashboard/settings', label: 'Settings' }
			]
		}
	];

	onMount(async () => {
		if (!getToken()) {
			goto('/login');
			return;
		}
		email = localStorage.getItem('email') || '';
		
		// Fetch tenant profile for logo and name
		try {
			const tenant = await api('/api/tenant/profile');
			churchName = tenant.name || 'Pews';
			churchLogo = tenant.logo || '';
		} catch (err) {
			console.error('Failed to load tenant profile:', err);
		}
	});

	function logout() {
		clearToken();
		localStorage.clear();
		goto('/login');
	}

	function isActive(href) {
		if (!$page) return false;
		if (href === '/dashboard') {
			return $page.url.pathname === '/dashboard';
		}
		return $page.url.pathname.startsWith(href);
	}

	function closeUserDropdown() {
		showUserDropdown = false;
	}
</script>

<div class="min-h-screen bg-[var(--bg)]">
	<nav class="bg-surface shadow-sm border-b border-custom sticky top-0 z-40">
		<!-- Top row: Logo, Search, Notifications, Theme, User -->
		<div class="border-b border-custom">
			<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div class="flex justify-between h-16 items-center">
					<a href="/dashboard" class="flex items-center space-x-3">
						{#if churchLogo}
							<img src={churchLogo} alt="{churchName} logo" class="h-10 w-10 object-contain" />
						{/if}
						<span class="text-2xl font-bold text-[var(--text-primary)]">{churchName}</span>
					</a>
					
					<div class="flex items-center space-x-4">
						<GlobalSearch />
						<NotificationBell />
						<ThemeToggle />
						
						<!-- User dropdown -->
						<div class="relative">
							<button
								on:click={() => showUserDropdown = !showUserDropdown}
								class="flex items-center space-x-2 text-sm text-secondary hover:text-primary"
							>
								<span>{email}</span>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
								</svg>
							</button>
							
							{#if showUserDropdown}
								<div class="absolute right-0 mt-2 w-48 bg-surface rounded-lg shadow-lg border border-custom z-50">
									<button
										on:click={logout}
										class="w-full text-left px-4 py-2 text-sm text-secondary hover:bg-[var(--surface-hover)] hover:text-primary rounded-lg"
									>
										Logout
									</button>
								</div>
							{/if}
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Second row: Navigation sections -->
		<div class="bg-surface">
			<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div class="flex space-x-6 h-12 items-center overflow-x-auto">
					{#each navSections as section}
						<div class="relative group">
							<button class="text-sm font-medium text-secondary hover:text-primary whitespace-nowrap py-3 px-2">
								{section.name}
								<svg class="w-3 h-3 inline ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
								</svg>
							</button>
							
							<!-- Dropdown menu -->
							<div class="absolute left-0 mt-1 hidden group-hover:block z-50">
								<div class="bg-surface rounded-lg shadow-lg border border-custom py-2 min-w-[180px]">
									{#each section.items as item}
										<a
											href={item.href}
											class="block px-4 py-2 text-sm hover:bg-[var(--surface-hover)] transition-colors {isActive(item.href) ? 'text-[var(--teal)] font-semibold bg-[var(--surface-hover)]' : 'text-secondary hover:text-primary'}"
										>
											{item.label}
										</a>
									{/each}
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</nav>

	<main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
		<slot />
	</main>
</div>

<!-- Close user dropdown when clicking outside -->
{#if showUserDropdown}
	<button
		class="fixed inset-0 z-30"
		on:click={closeUserDropdown}
		aria-label="Close dropdown"
	></button>
{/if}
