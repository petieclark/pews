<script>
	import { goto } from '$app/navigation';
	import { api, setToken } from '$lib/api';
	import ThemeToggle from '$lib/ThemeToggle.svelte';

	let tenantName = '';
	let email = '';
	let password = '';
	let error = '';
	let loading = false;

	async function handleRegister() {
		error = '';
		loading = true;

		try {
			const data = await api('/api/auth/register', {
				method: 'POST',
				body: JSON.stringify({ tenant_name: tenantName, email, password })
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
	<div class="absolute top-4 right-4">
		<ThemeToggle />
	</div>
	
	<div class="bg-surface rounded-lg shadow-xl p-8 w-full max-w-md border border-custom">
		<h1 class="text-3xl font-bold text-primary mb-2">Pews</h1>
		<p class="text-secondary mb-6">Create your church account</p>

		<form on:submit|preventDefault={handleRegister} class="space-y-4">
			<div>
				<label for="tenantName" class="block text-sm font-medium text-primary mb-1">
					Church Name
				</label>
				<input
					id="tenantName"
					type="text"
					bind:value={tenantName}
					placeholder="Grace Community Church"
					required
					class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
				/>
			</div>

			<div>
				<label for="email" class="block text-sm font-medium text-primary mb-1">
					Email
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
					Password
				</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					placeholder="••••••••"
					required
					minlength="8"
					class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
				/>
				<p class="text-xs text-secondary mt-1">At least 8 characters</p>
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
				{loading ? 'Creating account...' : 'Create Account'}
			</button>
		</form>

		<p class="mt-6 text-center text-sm text-secondary">
			Already have an account?
			<a href="/login" class="text-[var(--teal)] font-medium hover:underline">Sign in</a>
		</p>
	</div>
</div>
