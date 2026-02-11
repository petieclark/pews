<script lang="ts">
	import { page } from '$app/stores';
	
	const tabs = [
		{ href: '/dashboard/giving', label: 'Dashboard', icon: '📊' },
		{ href: '/dashboard/giving/donations', label: 'Donations', icon: '💵' },
		{ href: '/dashboard/giving/funds', label: 'Funds', icon: '🗂️' },
		{ href: '/dashboard/giving/statements', label: 'Statements', icon: '📄' },
		{ href: '/dashboard/giving/settings', label: 'Settings', icon: '⚙️' }
	];
	
	$: currentPath = $page.url.pathname;
	
	function isActive(tabHref: string): boolean {
		if (tabHref === '/dashboard/giving') {
			return currentPath === '/dashboard/giving';
		}
		return currentPath.startsWith(tabHref);
	}
</script>

<nav class="bg-surface border-b border-custom mb-6">
	<div class="flex overflow-x-auto">
		{#each tabs as tab}
			<a
				href={tab.href}
				class="px-6 py-3 text-sm font-medium whitespace-nowrap transition border-b-2 {
					isActive(tab.href)
						? 'border-[var(--teal)] text-[var(--teal)]'
						: 'border-transparent text-secondary hover:text-primary hover:border-gray-300'
				}"
			>
				<span class="mr-2">{tab.icon}</span>
				{tab.label}
			</a>
		{/each}
	</div>
</nav>
