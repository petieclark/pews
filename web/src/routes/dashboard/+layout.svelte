<script>
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { getToken, clearToken, api } from '$lib/api';
	import ThemeToggle from '$lib/ThemeToggle.svelte';
	import GlobalSearch from '$lib/GlobalSearch.svelte';
	import NotificationBell from '$lib/NotificationBell.svelte';
	import { enabledModules as enabledModulesStore, refreshModules } from '$lib/stores/modules';

	let email = '';
	let churchName = 'Pews';
	let churchLogo = '';
	let showUserDropdown = false;
	let openNavSection = '';
	let mobileMenuOpen = false;
	let enabledModules = [];

	const navSections = [
		{
			name: 'Core',
			items: [
				{ href: '/dashboard', label: 'Dashboard', icon: 'home', module: null },
				{ href: '/dashboard/people', label: 'People', icon: 'people', module: 'people' },
				{ href: '/dashboard/groups', label: 'Groups', icon: 'groups', module: 'groups' }
			]
		},
		{
			name: 'Ministry',
			items: [
				{ href: '/dashboard/services', label: 'Services', icon: 'services', module: 'services' },
				{ href: '/dashboard/services/teams', label: 'Teams', icon: 'groups', module: 'services' },
				{ href: '/dashboard/schedule', label: 'Schedule', icon: 'calendar', module: 'services' },
				{ href: '/dashboard/calendar', label: 'Calendar', icon: 'calendar', module: 'calendar' },
				{ href: '/dashboard/checkins', label: 'Check-Ins', icon: 'checkins', module: 'checkins' },
				{ href: '/dashboard/care', label: 'Care', icon: 'care', module: 'care' },
				{ href: '/dashboard/prayer', label: 'Prayer', icon: 'prayer', module: 'prayer' },
				{ href: '/dashboard/sermons', label: 'Sermons', icon: 'services', module: 'sermons' },
				{ href: '/dashboard/communication', label: 'Communication', icon: 'comm', module: 'communication' }
			]
		},
		{
			name: 'Media',
			items: [
				{ href: '/dashboard/streaming', label: 'Streaming', icon: 'stream', module: 'streaming' },
				{ href: '/dashboard/media', label: 'Media Library', icon: 'media', module: 'media' }
			]
		},
		{
			name: 'Finance',
			items: [
				{ href: '/dashboard/giving', label: 'Giving', icon: 'giving', module: 'giving' },
				{ href: '/dashboard/reports', label: 'Reports', icon: 'reports', module: null }
			]
		},
		{
			name: 'Admin',
			items: [
				{ href: '/dashboard/settings', label: 'Settings', icon: 'settings', module: null },
				{ href: '/dashboard/settings/profile', label: 'Church Profile', icon: 'profile', module: null },
				{ href: '/dashboard/settings/modules', label: 'Modules', icon: 'modules', module: null },
				{ href: '/dashboard/settings/users', label: 'Users & Roles', icon: 'users', module: null },
				{ href: '/dashboard/settings/billing', label: 'Billing', icon: 'billing', module: null },
				{ href: '/dashboard/settings/import', label: 'Import', icon: 'import', module: null },
				{ href: '/dashboard/settings/qr', label: 'QR Codes', icon: 'qr', module: null },
				{ href: '/dashboard/settings/ccli', label: 'CCLI', icon: 'ccli', module: null }
			]
		}
	];

	// Subscribe to module store
	$: enabledModules = $enabledModulesStore;

	// Filter nav based on enabled modules
	$: filteredNavSections = navSections.map(section => ({
		...section,
		items: section.items.filter(item => 
			!item.module || enabledModules.length === 0 || enabledModules.includes(item.module)
		)
	})).filter(section => section.items.length > 0);

	onMount(async () => {
		if (!getToken()) {
			goto('/login');
			return;
		}
		email = localStorage.getItem('email') || '';
		
		try {
			const [tenant] = await Promise.all([
				api('/api/tenant/profile'),
				refreshModules()
			]);
			churchName = tenant.name || 'Pews';
			churchLogo = tenant.logo || '';
		} catch (err) {
			console.error('Failed to load tenant profile:', err);
		}

		// Check onboarding
		try {
			const tenant = await api('/api/tenant');
			if (tenant && !tenant.onboarding_completed && $page.url.pathname !== '/dashboard/onboarding') {
				goto('/dashboard/onboarding');
			}
		} catch (e) {}
	});

	function logout() {
		clearToken();
		localStorage.clear();
		goto('/login');
	}

	function isActive(href) {
		if (!$page) return false;
		if (href === '/dashboard') return $page.url.pathname === '/dashboard';
		return $page.url.pathname.startsWith(href);
	}

	function toggleNav(name) {
		openNavSection = openNavSection === name ? '' : name;
	}

	function closeNav() {
		openNavSection = '';
	}

	function closeMobile() {
		mobileMenuOpen = false;
	}

	function navigateMobile(href) {
		closeMobile();
		closeNav();
		goto(href);
	}
</script>

<div class="min-h-screen bg-[var(--bg)]">
	<nav class="bg-surface shadow-sm border-b border-custom sticky top-0 z-40">
		<!-- Top row -->
		<div class="border-b border-custom">
			<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div class="flex justify-between h-16 items-center">
					<!-- Hamburger (mobile only) -->
					<button
						class="lg:hidden p-2 rounded-md text-secondary hover:text-primary hover:bg-[var(--surface-hover)]"
						on:click={() => mobileMenuOpen = !mobileMenuOpen}
						aria-label="Toggle menu"
					>
						<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							{#if mobileMenuOpen}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
							{:else}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
							{/if}
						</svg>
					</button>

					<a href="/dashboard" class="flex items-center space-x-3">
						{#if churchLogo}
							<img src={churchLogo} alt="{churchName} logo" class="h-10 w-10 object-contain" />
						{/if}
						<span class="text-2xl font-bold text-[var(--text-primary)]">{churchName}</span>
					</a>
					
					<div class="flex items-center space-x-2 sm:space-x-4">
						<div class="hidden sm:block">
							<GlobalSearch />
						</div>
						<NotificationBell />
						<ThemeToggle />
						
						<!-- User dropdown (desktop) -->
						<div class="relative hidden sm:block">
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

		<!-- Desktop navigation -->
		<div class="hidden lg:block bg-surface">
			<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div class="flex space-x-6 h-12 items-center overflow-visible">
					{#each filteredNavSections as section}
						<div class="relative">
							<button
								on:click={() => toggleNav(section.name)}
								class="text-sm font-medium whitespace-nowrap py-3 px-2 transition-colors
									{section.items.some(i => isActive(i.href)) ? 'text-[var(--teal)] border-b-2 border-[var(--teal)]' : 'text-secondary hover:text-primary'}"
							>
								{section.name}
								<svg class="w-3 h-3 inline ml-1 transition-transform {openNavSection === section.name ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
								</svg>
							</button>
							
							{#if openNavSection === section.name}
								<div class="absolute left-0 top-full z-50 pt-1">
									<div class="bg-surface rounded-lg shadow-lg border border-custom py-2 min-w-[180px]">
										{#each section.items as item}
											<a
												href={item.href}
												on:click={closeNav}
												class="block px-4 py-2 text-sm hover:bg-[var(--surface-hover)] transition-colors {isActive(item.href) ? 'text-[var(--teal)] font-semibold bg-[var(--surface-hover)]' : 'text-secondary hover:text-primary'}"
											>
												{item.label}
											</a>
										{/each}
									</div>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			</div>
		</div>
	</nav>

	<!-- Mobile slide-out menu -->
	{#if mobileMenuOpen}
		<div class="fixed inset-0 z-50 lg:hidden">
			<!-- Backdrop -->
			<button
				class="fixed inset-0 bg-black/50"
				on:click={closeMobile}
				aria-label="Close menu"
			></button>
			
			<!-- Sidebar -->
			<div class="fixed inset-y-0 left-0 w-72 bg-surface shadow-xl overflow-y-auto">
				<div class="p-4 border-b border-custom flex items-center justify-between">
					<span class="text-lg font-bold text-primary">{churchName}</span>
					<button on:click={closeMobile} class="p-2 text-secondary hover:text-primary">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>

				<!-- Mobile search -->
				<div class="p-4 border-b border-custom sm:hidden">
					<GlobalSearch />
				</div>

				<!-- Nav sections -->
				<div class="py-2">
					{#each filteredNavSections as section}
						<div class="px-4 py-2">
							<div class="text-xs font-semibold text-secondary uppercase tracking-wider mb-1">{section.name}</div>
							{#each section.items as item}
								<button
									on:click={() => navigateMobile(item.href)}
									class="w-full text-left block px-3 py-2 rounded-md text-sm transition-colors
										{isActive(item.href) ? 'text-[var(--teal)] bg-[var(--surface-hover)] font-semibold' : 'text-primary hover:bg-[var(--surface-hover)]'}"
								>
									{item.label}
								</button>
							{/each}
						</div>
					{/each}
				</div>

				<!-- Mobile user info -->
				<div class="border-t border-custom p-4 mt-auto">
					<p class="text-sm text-secondary mb-3 truncate">{email}</p>
					<button
						on:click={logout}
						class="w-full text-left px-3 py-2 text-sm text-red-500 hover:bg-[var(--surface-hover)] rounded-md"
					>
						Logout
					</button>
				</div>
			</div>
		</div>
	{/if}

	<main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-8">
		<slot />
	</main>
</div>

<!-- Close dropdowns when clicking outside -->
{#if showUserDropdown || openNavSection}
	<button
		class="fixed inset-0 z-30"
		on:click={() => { showUserDropdown = false; closeNav(); }}
		aria-label="Close dropdown"
	></button>
{/if}
