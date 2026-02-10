<script>
	import { goto } from '$app/navigation';
	import { api, setToken } from '$lib/api';

	let tenantSlug = '';
	let email = '';
	let password = '';
	let error = '';
	let loading = false;

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

<div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-navy to-teal">
	<div class="bg-white rounded-lg shadow-xl p-8 w-full max-w-md">
		<h1 class="text-3xl font-bold text-navy mb-2">Pews</h1>
		<p class="text-gray-600 mb-6">Church Management Platform</p>

		<form on:submit|preventDefault={handleLogin} class="space-y-4">
			<div>
				<label for="tenantSlug" class="block text-sm font-medium text-gray-700 mb-1">
					Church Slug
				</label>
				<input
					id="tenantSlug"
					type="text"
					bind:value={tenantSlug}
					placeholder="your-church"
					required
					class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-teal focus:border-transparent"
				/>
			</div>

			<div>
				<label for="email" class="block text-sm font-medium text-gray-700 mb-1">
					Email
				</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					placeholder="you@example.com"
					required
					class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-teal focus:border-transparent"
				/>
			</div>

			<div>
				<label for="password" class="block text-sm font-medium text-gray-700 mb-1">
					Password
				</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					placeholder="••••••••"
					required
					class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-teal focus:border-transparent"
				/>
			</div>

			{#if error}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
					{error}
				</div>
			{/if}

			<button
				type="submit"
				disabled={loading}
				class="w-full bg-teal text-white py-2 px-4 rounded-lg font-medium hover:bg-teal/90 disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{loading ? 'Signing in...' : 'Sign In'}
			</button>
		</form>

		<p class="mt-6 text-center text-sm text-gray-600">
			Don't have an account?
			<a href="/register" class="text-teal font-medium hover:underline">Register</a>
		</p>
	</div>
</div>
