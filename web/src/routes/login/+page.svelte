<script>
	import { goto } from '$app/navigation';
	import { api, setToken } from '$lib/api';
	import ThemeToggle from '$lib/ThemeToggle.svelte';
	import LanguageSelector from '$lib/LanguageSelector.svelte';
	import { t } from '$lib/i18n.js';

	let tenantSlug = '';
	let email = '';
	let password = '';
	let error = '';
	let loading = false;
	let translate;

	// Subscribe to translation updates
	const unsubscribe = t.subscribe(value => {
		translate = value;
	});

	async function handleLogin() {
		error = '';
		loading = true;

		try {
			const data = await api('/api/auth/login', {
				method: 'POST',
				body: JSON.stringify({ tenant_slug: tenantSlug, email, password })
			});

			setToken(data.token);
			localStorage.setItem('tenant_id', data.tenant_id);
			localStorage.setItem('email', data.email);
			localStorage.setItem('role', data.role);

			goto('/dashboard');
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen flex items-center justify-center relative" style="background: linear-gradient(to bottom right, var(--navy), var(--teal));">
	<div class="absolute top-4 right-4 flex items-center gap-4">
		<LanguageSelector />
		<ThemeToggle />
	</div>
	
	<div class="bg-surface rounded-lg shadow-xl p-8 w-full max-w-md border border-custom">
		<h1 class="text-3xl font-bold text-primary mb-2">Pews</h1>
		<p class="text-secondary mb-6">Church Management Platform</p>

		<form on:submit|preventDefault={handleLogin} class="space-y-4">
			<div>
				<label for="tenantSlug" class="block text-sm font-medium text-primary mb-1">
					Church Slug
				</label>
				<input
					id="tenantSlug"
					type="text"
					bind:value={tenantSlug}
					placeholder="your-church"
					required
					class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
				/>
			</div>

			<div>
				<label for="email" class="block text-sm font-medium text-primary mb-1">
					{translate ? translate('auth.email') : 'Email'}
				</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					placeholder="you@example.com"
					required
					class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
				/>
			</div>

			<div>
				<label for="password" class="block text-sm font-medium text-primary mb-1">
					{translate ? translate('auth.password') : 'Password'}
				</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					placeholder="••••••••"
					required
					class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
				/>
			</div>

			{#if error}
				<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg">
					{error}
				</div>
			{/if}

			<button
				type="submit"
				disabled={loading}
				class="w-full bg-[var(--teal)] text-white py-2 px-4 rounded-lg font-medium hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{loading ? (translate ? translate('common.loading') : 'Loading...') : (translate ? translate('auth.signin') : 'Sign In')}
			</button>
		</form>

		<p class="mt-6 text-center text-sm text-secondary">
			Don't have an account?
			<a href="/register" class="text-[var(--teal)] font-medium hover:underline">
				{translate ? translate('auth.register') : 'Register'}
			</a>
		</p>
	</div>
</div>
