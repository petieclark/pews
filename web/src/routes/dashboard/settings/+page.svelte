<script>
	import { Building2, Puzzle, Users, CreditCard, Download, Smartphone } from 'lucide-svelte';

	const settingsPages = [
		{ href: '/dashboard/settings/profile', label: 'Church Profile', desc: 'Name, address, logo, timezone', icon: Building2 },
		{ href: '/dashboard/settings/modules', label: 'Modules', desc: 'Enable or disable features', icon: Puzzle },
		{ href: '/dashboard/settings/users', label: 'Users & Roles', desc: 'Manage team access and permissions', icon: Users },
		{ href: '/dashboard/settings/billing', label: 'Billing', desc: 'Plan, usage, and payments', icon: CreditCard },
		{ href: '/dashboard/settings/import', label: 'Import', desc: 'Import from PCO or CSV', icon: Download },
		{ href: '/dashboard/settings/qr', label: 'QR Codes', desc: 'Generate printable QR codes', icon: Smartphone }
	];
</script>

<div class="max-w-4xl">
	<h1 class="text-2xl sm:text-3xl font-bold text-primary mb-6">Settings</h1>

	<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
		{#each settingsPages as page}
			<a
				href={page.href}
				class="bg-surface rounded-lg border border-custom p-5 hover:bg-[var(--surface-hover)] transition-colors group"
			>
				<div class="mb-2"><svelte:component this={page.icon} size={24} /></div>
				<h2 class="text-lg font-semibold text-primary group-hover:text-[var(--teal)] transition-colors">{page.label}</h2>
				<p class="text-sm text-secondary mt-1">{page.desc}</p>
			</a>
		{/each}
	</div>
</div>
