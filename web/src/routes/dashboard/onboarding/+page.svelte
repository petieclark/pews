<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let step = 1;
	const totalSteps = 5;
	let loading = true;
	let saving = false;

	// Step 1: Church profile
	let church = { name: '', address: '', timezone: 'America/New_York' };
	
	// Step 2: Modules
	let modules = [];
	
	// Step 3: Import choice
	let importChoice = 'skip'; // 'pco', 'csv', 'skip'
	
	// Step 4: Invite team
	let invites = [{ email: '', role: 'Staff' }];
	
	let error = '';

	const timezones = [
		'America/New_York', 'America/Chicago', 'America/Denver', 'America/Los_Angeles',
		'America/Phoenix', 'America/Anchorage', 'Pacific/Honolulu'
	];

	const roles = ['Admin', 'Staff', 'Volunteer', 'Member'];

	onMount(async () => {
		try {
			// Load existing progress
			const [tenant, mods] = await Promise.all([
				api('/api/tenant/profile'),
				api('/api/tenant/modules').catch(() => [])
			]);
			if (tenant.name) church.name = tenant.name;
			if (tenant.address) church.address = tenant.address;
			if (tenant.timezone) church.timezone = tenant.timezone;
			modules = (mods || []).map(m => ({ ...m, enabled: m.enabled !== false }));

			// Resume from saved step
			const savedStep = localStorage.getItem('onboarding_step');
			if (savedStep) step = parseInt(savedStep);
		} catch (err) {
			console.error('Failed to load onboarding data:', err);
		} finally {
			loading = false;
		}
	});

	function saveProgress() {
		localStorage.setItem('onboarding_step', step.toString());
	}

	async function nextStep() {
		error = '';
		saving = true;
		try {
			if (step === 1) {
				await api('/api/tenant/profile', {
					method: 'PUT',
					body: JSON.stringify(church)
				});
			} else if (step === 2) {
				for (const mod of modules) {
					const action = mod.enabled ? 'enable' : 'disable';
					await api(`/api/tenant/modules/${mod.name}/${action}`, { method: 'POST' }).catch(() => {});
				}
			} else if (step === 4) {
				for (const inv of invites.filter(i => i.email.trim())) {
					await api('/api/tenant/users/invite', {
						method: 'POST',
						body: JSON.stringify(inv)
					}).catch(() => {});
				}
			}

			if (step < totalSteps) {
				step++;
				saveProgress();
			} else {
				await completeOnboarding();
			}
		} catch (err) {
			error = err.message;
		} finally {
			saving = false;
		}
	}

	function skipStep() {
		if (step < totalSteps) {
			step++;
			saveProgress();
		} else {
			completeOnboarding();
		}
	}

	async function completeOnboarding() {
		try {
			await api('/api/tenant', {
				method: 'PUT',
				body: JSON.stringify({ onboarding_completed: true })
			});
		} catch (e) {}
		localStorage.removeItem('onboarding_step');
		goto('/dashboard');
	}

	function addInvite() {
		invites = [...invites, { email: '', role: 'Staff' }];
	}

	function removeInvite(i) {
		invites = invites.filter((_, idx) => idx !== i);
	}

	function toggleModule(name) {
		modules = modules.map(m => m.name === name ? { ...m, enabled: !m.enabled } : m);
	}
</script>

{#if loading}
	<div class="flex items-center justify-center min-h-[60vh]">
		<div class="inline-block animate-spin rounded-full h-10 w-10 border-b-2 border-[var(--teal)]"></div>
	</div>
{:else}
	<div class="max-w-lg mx-auto py-8">
		<!-- Progress bar -->
		<div class="mb-8">
			<div class="flex justify-between text-xs text-secondary mb-2">
				<span>Step {step} of {totalSteps}</span>
				<span>{Math.round((step / totalSteps) * 100)}%</span>
			</div>
			<div class="h-2 bg-[var(--surface-hover)] rounded-full overflow-hidden">
				<div class="h-full bg-[var(--teal)] rounded-full transition-all duration-300" style="width: {(step / totalSteps) * 100}%"></div>
			</div>
		</div>

		<!-- Step 1: Church Profile -->
		{#if step === 1}
			<div class="text-center mb-8">
				<h1 class="text-3xl font-bold text-primary mb-2">Welcome to Pews! 🎉</h1>
				<p class="text-secondary">Let's set up your church profile</p>
			</div>

			<div class="bg-surface rounded-lg border border-custom p-6 space-y-4">
				<div>
					<label for="ob-name" class="block text-sm font-medium text-primary mb-1">Church Name *</label>
					<input id="ob-name" type="text" bind:value={church.name} required
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary" />
				</div>
				<div>
					<label for="ob-addr" class="block text-sm font-medium text-primary mb-1">Address</label>
					<textarea id="ob-addr" bind:value={church.address} rows="2"
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary"></textarea>
				</div>
				<div>
					<label for="ob-tz" class="block text-sm font-medium text-primary mb-1">Timezone</label>
					<select id="ob-tz" bind:value={church.timezone}
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary">
						{#each timezones as tz}
							<option value={tz}>{tz.replace(/_/g, ' ')}</option>
						{/each}
					</select>
				</div>
			</div>

		<!-- Step 2: Modules -->
		{:else if step === 2}
			<div class="text-center mb-8">
				<h1 class="text-3xl font-bold text-primary mb-2">Choose Your Modules</h1>
				<p class="text-secondary">Select which features you want to use</p>
			</div>

			<div class="space-y-3">
				{#each modules as mod}
					<button
						on:click={() => toggleModule(mod.name)}
						class="w-full text-left bg-surface rounded-lg border p-4 transition-colors
							{mod.enabled ? 'border-[var(--teal)] bg-[var(--teal)]/5' : 'border-custom hover:bg-[var(--surface-hover)]'}"
					>
						<div class="flex items-center justify-between">
							<div>
								<h3 class="font-medium text-primary">{mod.display_name}</h3>
								<p class="text-sm text-secondary">{mod.description}</p>
							</div>
							<div class="w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0
								{mod.enabled ? 'border-[var(--teal)] bg-[var(--teal)]' : 'border-[var(--border)]'}">
								{#if mod.enabled}
									<svg class="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
									</svg>
								{/if}
							</div>
						</div>
					</button>
				{/each}
			</div>

		<!-- Step 3: Import -->
		{:else if step === 3}
			<div class="text-center mb-8">
				<h1 class="text-3xl font-bold text-primary mb-2">Import Your Data</h1>
				<p class="text-secondary">Bring your existing data or start fresh</p>
			</div>

			<div class="space-y-3">
				<button on:click={() => { importChoice = 'pco'; }}
					class="w-full text-left bg-surface rounded-lg border p-5 transition-colors
						{importChoice === 'pco' ? 'border-[var(--teal)]' : 'border-custom hover:bg-[var(--surface-hover)]'}">
					<h3 class="font-semibold text-primary">Import from Planning Center</h3>
					<p class="text-sm text-secondary mt-1">Automatically sync people, groups, and giving data</p>
				</button>
				<button on:click={() => { importChoice = 'csv'; }}
					class="w-full text-left bg-surface rounded-lg border p-5 transition-colors
						{importChoice === 'csv' ? 'border-[var(--teal)]' : 'border-custom hover:bg-[var(--surface-hover)]'}">
					<h3 class="font-semibold text-primary">Upload CSV</h3>
					<p class="text-sm text-secondary mt-1">Import from a spreadsheet</p>
				</button>
				<button on:click={() => { importChoice = 'skip'; }}
					class="w-full text-left bg-surface rounded-lg border p-5 transition-colors
						{importChoice === 'skip' ? 'border-[var(--teal)]' : 'border-custom hover:bg-[var(--surface-hover)]'}">
					<h3 class="font-semibold text-primary">Start Fresh</h3>
					<p class="text-sm text-secondary mt-1">Enter data manually later</p>
				</button>
			</div>

		<!-- Step 4: Invite Team -->
		{:else if step === 4}
			<div class="text-center mb-8">
				<h1 class="text-3xl font-bold text-primary mb-2">Invite Your Team</h1>
				<p class="text-secondary">Add team members to collaborate</p>
			</div>

			<div class="bg-surface rounded-lg border border-custom p-6 space-y-4">
				{#each invites as inv, i}
					<div class="flex flex-col sm:flex-row gap-2">
						<input type="email" bind:value={inv.email} placeholder="email@example.com"
							class="flex-1 px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary" />
						<select bind:value={inv.role}
							class="px-3 py-2 border input-border rounded-lg bg-[var(--input-bg)] text-primary">
							{#each roles as role}
								<option value={role}>{role}</option>
							{/each}
						</select>
						{#if invites.length > 1}
							<button on:click={() => removeInvite(i)} class="text-secondary hover:text-red-500 px-2">✕</button>
						{/if}
					</div>
				{/each}
				<button on:click={addInvite} class="text-sm text-[var(--teal)] hover:underline">+ Add another</button>
			</div>

		<!-- Step 5: Done -->
		{:else if step === 5}
			<div class="text-center py-12">
				<div class="text-6xl mb-4">🎉</div>
				<h1 class="text-3xl font-bold text-primary mb-2">You're All Set!</h1>
				<p class="text-secondary mb-8">Your church is ready to go. Let's dive into the dashboard.</p>
			</div>
		{/if}

		{#if error}
			<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg mt-4">{error}</div>
		{/if}

		<!-- Navigation buttons -->
		<div class="flex justify-between mt-8 gap-3">
			{#if step > 1 && step < 5}
				<button on:click={() => { step--; saveProgress(); }}
					class="px-6 py-2 border border-custom rounded-lg text-primary hover:bg-[var(--surface-hover)]">
					Back
				</button>
			{:else}
				<div></div>
			{/if}

			<div class="flex gap-3">
				{#if step < 5}
					<button on:click={skipStep}
						class="px-6 py-2 text-secondary hover:text-primary">
						Skip
					</button>
				{/if}
				<button on:click={nextStep} disabled={saving}
					class="px-6 py-2 bg-[var(--teal)] text-white rounded-lg font-medium hover:opacity-90 disabled:opacity-50">
					{#if saving}
						Saving...
					{:else if step === 5}
						Go to Dashboard
					{:else if step === 3 && importChoice === 'pco'}
						Connect PCO →
					{:else}
						Continue →
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}
