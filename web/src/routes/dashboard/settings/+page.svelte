<script>
	const settingsPages = [
		{ href: '/dashboard/settings/profile', label: 'Church Profile', desc: 'Name, address, logo, timezone', icon: '🏛️' },
		{ href: '/dashboard/settings/modules', label: 'Modules', desc: 'Enable or disable features', icon: '🧩' },
		{ href: '/dashboard/settings/users', label: 'Users & Roles', desc: 'Manage team access and permissions', icon: '👥' },
		{ href: '/dashboard/settings/billing', label: 'Billing', desc: 'Plan, usage, and payments', icon: '💳' },
		{ href: '/dashboard/settings/import', label: 'Import', desc: 'Import from PCO or CSV', icon: '📥' },
		{ href: '/dashboard/settings/qr', label: 'QR Codes', desc: 'Generate printable QR codes', icon: '📱' }
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
				<div class="text-2xl mb-2">{page.icon}</div>
				<h2 class="text-lg font-semibold text-primary group-hover:text-[var(--teal)] transition-colors">{page.label}</h2>
				<p class="text-sm text-secondary mt-1">{page.desc}</p>
			</a>
		{/each}
	</div>
</div>
