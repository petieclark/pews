<script>
	import { goto } from '$app/navigation';
	import { api, setToken } from '$lib/api';
	import ThemeToggle from '$lib/ThemeToggle.svelte';

	let tenantSlug = '';
	let email = '';
	let password = '';
	let error = '';
	let loading = false;
	let showHelp = false;

	async function handleLogin() {
		error = '';
		loading = true;

		try {
			const data = await api('/api/auth/login', {
				method: 'POST',
				body: JSON.stringify({ tenant_slug: tenantSlug, email, password }),
				silent: true
			});

			setToken(data.token);
			localStorage.setItem('tenant_id', data.tenant_id);
			localStorage.setItem('tenant_slug', tenantSlug);
			localStorage.setItem('email', data.email);
			localStorage.setItem('role', data.role);

			goto('/dashboard');
		} catch (err) {
			error = err.message || 'Login failed. Please check your credentials.';
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen flex items-center justify-center relative" style="background: linear-gradient(to bottom right, var(--navy), var(--teal));">
	<div class="absolute top-4 right-4">
		<ThemeToggle />
	</div>
	
	<div class="bg-surface rounded-lg shadow-xl p-8 w-full max-w-md border border-custom">
		<h1 class="text-3xl font-bold text-primary mb-2">Pews</h1>
		<p class="text-secondary mb-6">Church Management Platform</p>

		<form on:submit|preventDefault={handleLogin} class="space-y-4">
			<div>
				<div class="flex items-center justify-between mb-1">
					<label for="tenantSlug" class="block text-sm font-medium text-primary">
						Church Slug <span class="text-red-500">*</span>
					</label>
					<button
						type="button"
						on:click={() => showHelp = !showHelp}
						class="text-xs text-[var(--teal)] hover:underline focus:outline-none"
					>
						{showHelp ? 'Hide help' : 'What\'s this?'}
					</button>
				</div>
				
				{#if showHelp}
					<div class="mb-2 p-3 bg-blue-50 dark:bg-blue-900 border border-blue-200 dark:border-blue-700 rounded-lg">
						<p class="text-xs text-secondary mb-2">
							Your church slug is a unique identifier created from your church name.
						</p>
						<p class="text-xs text-secondary font-medium">
							Examples: <span class="font-mono text-[var(--teal)]">grace-community</span>, 
							<span class="font-mono text-[var(--teal)]">first-baptist</span>
						</p>
					</div>
				{/if}

				<input
					id="tenantSlug"
					type="text"
					bind:value={tenantSlug}
					placeholder="your-church"
					required
					class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary font-mono"
				/>
				<p class="text-xs text-secondary mt-1">
					This was shown to you when you registered
				</p>
			</div>

			<div>
				<label for="email" class="block text-sm font-medium text-primary mb-1">
					Email <span class="text-red-500">*</span>
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
					Password <span class="text-red-500">*</span>
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
				<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg text-sm">
					{error}
				</div>
			{/if}

			<button
				type="submit"
				disabled={loading}
				class="w-full bg-[var(--teal)] text-white py-2 px-4 rounded-lg font-medium hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed transition-opacity"
			>
				{loading ? 'Signing in...' : 'Sign In'}
			</button>
		</form>

		<p class="mt-6 text-center text-sm text-secondary">
			Don't have an account?
			<a href="/register" class="text-[var(--teal)] font-medium hover:underline">Register</a>
		</p>
	</div>
</div>
