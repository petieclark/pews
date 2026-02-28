<script>
	import { goto } from '$app/navigation';
	import { FileEdit } from 'lucide-svelte';
	import { api, setToken } from '$lib/api';
	import ThemeToggle from '$lib/ThemeToggle.svelte';

	let tenantName = '';
	let email = '';
	let password = '';
	let confirmPassword = '';
	let error = '';
	let loading = false;
	let showSlug = false;
	let generatedSlug = '';

	// Form validation states
	let passwordError = '';
	let emailError = '';
	let nameError = '';

	// Real-time slug preview
	$: generatedSlug = tenantName ? slugify(tenantName) : '';

	function slugify(text) {
		return text
			.toLowerCase()
			.trim()
			.replace(/[^\w\s-]/g, '')
			.replace(/[\s_-]+/g, '-')
			.replace(/^-+|-+$/g, '');
	}

	function validateEmail(email) {
		const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
		return re.test(email);
	}

	function validatePassword(password) {
		if (password.length < 8) {
			return 'Password must be at least 8 characters';
		}
		if (!/[A-Z]/.test(password)) {
			return 'Password must contain at least one uppercase letter';
		}
		if (!/[a-z]/.test(password)) {
			return 'Password must contain at least one lowercase letter';
		}
		if (!/[0-9]/.test(password)) {
			return 'Password must contain at least one number';
		}
		return '';
	}

	function getPasswordStrength(password) {
		let strength = 0;
		if (password.length >= 8) strength++;
		if (password.length >= 12) strength++;
		if (/[A-Z]/.test(password)) strength++;
		if (/[a-z]/.test(password)) strength++;
		if (/[0-9]/.test(password)) strength++;
		if (/[^A-Za-z0-9]/.test(password)) strength++;
		return strength;
	}

	$: passwordStrength = password ? getPasswordStrength(password) : 0;
	$: passwordStrengthLabel = 
		passwordStrength === 0 ? '' :
		passwordStrength <= 2 ? 'Weak' :
		passwordStrength <= 4 ? 'Fair' :
		passwordStrength <= 5 ? 'Good' : 'Strong';
	$: passwordStrengthColor = 
		passwordStrength <= 2 ? '#ef4444' :
		passwordStrength <= 4 ? '#f59e0b' :
		passwordStrength <= 5 ? '#10b981' : '#059669';

	async function handleRegister() {
		error = '';
		nameError = '';
		emailError = '';
		passwordError = '';

		// Validate inputs
		if (!tenantName.trim()) {
			nameError = 'Church name is required';
			return;
		}

		if (!email || !validateEmail(email)) {
			emailError = 'Please enter a valid email address';
			return;
		}

		const pwdValidation = validatePassword(password);
		if (pwdValidation) {
			passwordError = pwdValidation;
			return;
		}

		if (password !== confirmPassword) {
			passwordError = 'Passwords do not match';
			return;
		}

		loading = true;

		try {
			const data = await api('/api/auth/register', {
				method: 'POST',
				body: JSON.stringify({ tenant_name: tenantName, email, password }),
				silent: true
			});

			setToken(data.token);
			localStorage.setItem('tenant_id', data.tenant_id);
			localStorage.setItem('tenant_slug', generatedSlug);
			localStorage.setItem('email', data.email);
			localStorage.setItem('role', data.role);

			// Show success message with slug
			showSlug = true;
			
			// Redirect after showing slug
			setTimeout(() => {
				goto('/dashboard');
			}, 3000);
		} catch (err) {
			error = err.message || 'Registration failed. Please try again.';
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
		{#if !showSlug}
			<h1 class="text-3xl font-bold text-primary mb-2">Pews</h1>
			<p class="text-secondary mb-6">Create your church account</p>

			<form on:submit|preventDefault={handleRegister} class="space-y-4">
				<div>
					<label for="tenantName" class="block text-sm font-medium text-primary mb-1">
						Church Name <span class="text-red-500">*</span>
					</label>
					<input
						id="tenantName"
						type="text"
						bind:value={tenantName}
						placeholder="Grace Community Church"
						required
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
						class:border-red-500={nameError}
					/>
					{#if nameError}
						<p class="text-xs text-red-500 mt-1">{nameError}</p>
					{/if}
					{#if generatedSlug}
						<p class="text-xs text-secondary mt-1">
							Your slug will be: <span class="font-mono font-semibold text-[var(--teal)]">{generatedSlug}</span>
						</p>
					{/if}
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
						class:border-red-500={emailError}
					/>
					{#if emailError}
						<p class="text-xs text-red-500 mt-1">{emailError}</p>
					{/if}
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
						minlength="8"
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
						class:border-red-500={passwordError}
					/>
					{#if password}
						<div class="mt-2">
							<div class="flex items-center justify-between mb-1">
								<span class="text-xs text-secondary">Password strength:</span>
								<span class="text-xs font-medium" style="color: {passwordStrengthColor}">
									{passwordStrengthLabel}
								</span>
							</div>
							<div class="w-full h-1 bg-gray-200 rounded-full overflow-hidden">
								<div 
									class="h-full transition-all duration-300"
									style="width: {(passwordStrength / 6) * 100}%; background-color: {passwordStrengthColor};"
								></div>
							</div>
						</div>
					{/if}
					<p class="text-xs text-secondary mt-1">
						Must be at least 8 characters with uppercase, lowercase, and numbers
					</p>
					{#if passwordError}
						<p class="text-xs text-red-500 mt-1">{passwordError}</p>
					{/if}
				</div>

				<div>
					<label for="confirmPassword" class="block text-sm font-medium text-primary mb-1">
						Confirm Password <span class="text-red-500">*</span>
					</label>
					<input
						id="confirmPassword"
						type="password"
						bind:value={confirmPassword}
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
					{loading ? 'Creating account...' : 'Create Account'}
				</button>
			</form>

			<p class="mt-6 text-center text-sm text-secondary">
				Already have an account?
				<a href="/login" class="text-[var(--teal)] font-medium hover:underline">Sign in</a>
			</p>
		{:else}
			<!-- Success state -->
			<div class="text-center">
				<div class="mb-6">
					<svg class="w-16 h-16 mx-auto text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
					</svg>
				</div>
				<h2 class="text-2xl font-bold text-primary mb-2">Welcome to Pews!</h2>
				<p class="text-secondary mb-4">Your church account has been created successfully.</p>
				
				<div class="bg-blue-50 dark:bg-blue-900 border border-blue-200 dark:border-blue-700 rounded-lg p-4 mb-6">
					<p class="text-sm font-medium text-primary mb-2"><FileEdit size={14} class="inline" /> Important: Save your church slug</p>
					<div class="bg-white dark:bg-gray-800 rounded px-3 py-2 border border-gray-200 dark:border-gray-700">
						<code class="text-lg font-mono font-bold text-[var(--teal)]">{generatedSlug}</code>
					</div>
					<p class="text-xs text-secondary mt-2">You'll need this slug to log in</p>
				</div>

				<p class="text-sm text-secondary">Redirecting to your dashboard...</p>
			</div>
		{/if}
	</div>
</div>
