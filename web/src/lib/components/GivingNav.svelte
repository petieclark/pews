<script lang="ts">
	import { page } from '$app/stores';
	import { BarChart3, Banknote, FolderOpen, FileText, Settings } from 'lucide-svelte';
	
	const tabs = [
		{ href: '/dashboard/giving', label: 'Dashboard', icon: BarChart3 },
		{ href: '/dashboard/giving/donations', label: 'Donations', icon: Banknote },
		{ href: '/dashboard/giving/funds', label: 'Funds', icon: FolderOpen },
		{ href: '/dashboard/giving/statements', label: 'Statements', icon: FileText },
		{ href: '/dashboard/giving/settings', label: 'Settings', icon: Settings }
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
				<span class="mr-2 inline-flex"><svelte:component this={tab.icon} size={16} /></span>
				{tab.label}
			</a>
		{/each}
	</div>
</nav>
