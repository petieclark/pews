<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { CheckCircle, CreditCard, Building2, FileText, Sparkles, PartyPopper } from 'lucide-svelte';

	let step = 1;
	const totalSteps = 5;
	let loading = true;
	let saving = false;

	// Step 1: Church profile
	let church = {
		name: '',
		address_line1: '',
		address_line2: '',
		city: '',
		state: '',
		zip: '',
		phone: '',
		website: '',
		email: ''
	};
	
	// Step 2: Modules
	let modules = [];
	
	// Step 3: Stripe Connect (placeholder)
	let stripeConnected = false;
	
	// Step 4: Import choice
	let importChoice = 'skip'; // 'pco', 'csv', 'skip'
	
	let error = '';

	const usStates = [
		'AL','AK','AZ','AR','CA','CO','CT','DE','FL','GA','HI','ID','IL','IN','IA','KS','KY',
		'LA','ME','MD','MA','MI','MN','MS','MO','MT','NE','NV','NH','NJ','NM','NY','NC','ND',
		'OH','OK','OR','PA','RI','SC','SD','TN','TX','UT','VT','VA','WA','WV','WI','WY','DC'
	];

	onMount(async () => {
		try {
			// Load existing progress
			const [tenant, mods] = await Promise.all([
				api('/api/tenant/profile'),
				api('/api/tenant/modules').catch(() => [])
			]);
			if (tenant.name) church.name = tenant.name;
			if (tenant.address_line1) church.address_line1 = tenant.address_line1;
			if (tenant.address_line2) church.address_line2 = tenant.address_line2;
			if (tenant.city) church.city = tenant.city;
			if (tenant.state) church.state = tenant.state;
			if (tenant.zip) church.zip = tenant.zip;
			if (tenant.phone) church.phone = tenant.phone;
			if (tenant.website) church.website = tenant.website;
			if (tenant.email) church.email = tenant.email;
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
				if (!church.name.trim()) {
					error = 'Church name is required';
					saving = false;
					return;
				}
				await api('/api/tenant/profile', {
					method: 'PUT',
					body: JSON.stringify(church)
				});
			} else if (step === 2) {
				for (const mod of modules) {
					const action = mod.enabled ? 'enable' : 'disable';
					await api(`/api/tenant/modules/${mod.name}/${action}`, { method: 'POST' }).catch(() => {});
				}
			}
			// Steps 3 (Stripe) and 4 (Import) are optional / skippable

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
		} catch (e) {
			console.error('Failed to mark onboarding complete:', e);
		}
		localStorage.removeItem('onboarding_step');
		goto('/dashboard');
	}

	function toggleModule(name) {
		modules = modules.map(m => m.name === name ? { ...m, enabled: !m.enabled } : m);
	}

	const stepLabels = ['Church Info', 'Modules', 'Stripe', 'Import', 'Done'];
</script>

{#if loading}
	<div class="flex items-center justify-center min-h-[60vh]">
		<div class="inline-block animate-spin rounded-full h-10 w-10 border-b-2 border-[var(--teal)]"></div>
	</div>
{:else}
	<div class="max-w-lg mx-auto py-8">
		<!-- Progress bar with step labels -->
		<div class="mb-8">
			<div class="flex justify-between text-xs text-secondary mb-2">
				<span>Step {step} of {totalSteps}</span>
				<span>{stepLabels[step - 1]}</span>
			</div>
			<div class="h-2 bg-[var(--surface-hover)] rounded-full overflow-hidden">
				<div class="h-full bg-[var(--teal)] rounded-full transition-all duration-300" style="width: {(step / totalSteps) * 100}%"></div>
			</div>
			<!-- Step dots -->
			<div class="flex justify-between mt-2">
				{#each stepLabels as label, i}
					<button
						class="flex flex-col items-center gap-1 group"
						on:click={() => { if (i + 1 <= step) { step = i + 1; saveProgress(); } }}
						disabled={i + 1 > step}
					>
						<div class="w-3 h-3 rounded-full transition-colors {i + 1 <= step ? 'bg-[var(--teal)]' : 'bg-[var(--surface-hover)]'}"></div>
						<span class="text-[10px] {i + 1 === step ? 'text-[var(--teal)] font-semibold' : 'text-secondary'}">{label}</span>
					</button>
				{/each}
			</div>
		</div>

		<!-- Step 1: Church Profile -->
		{#if step === 1}
			<div class="text-center mb-8">
				<h1 class="text-3xl font-bold text-primary mb-2">Welcome to Pews!</h1>
				<p class="text-secondary">Let's set up your church profile</p>
			</div>

			<div class="bg-surface rounded-lg border border-custom p-6 space-y-4">
				<div>
					<label for="ob-name" class="block text-sm font-medium text-primary mb-1">Church Name *</label>
					<input id="ob-name" type="text" bind:value={church.name} required
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary" />
				</div>
				<div>
					<label for="ob-addr1" class="block text-sm font-medium text-primary mb-1">Address</label>
					<input id="ob-addr1" type="text" bind:value={church.address_line1} placeholder="Street address"
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary" />
				</div>
				<div class="grid grid-cols-3 gap-3">
					<div>
						<label for="ob-city" class="block text-sm font-medium text-primary mb-1">City</label>
						<input id="ob-city" type="text" bind:value={church.city}
							class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary" />
					</div>
					<div>
						<label for="ob-state" class="block text-sm font-medium text-primary mb-1">State</label>
						<select id="ob-state" bind:value={church.state}
							class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary">
							<option value="">—</option>
							{#each usStates as st}
								<option value={st}>{st}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="ob-zip" class="block text-sm font-medium text-primary mb-1">ZIP</label>
						<input id="ob-zip" type="text" bind:value={church.zip}
							class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary" />
					</div>
				</div>
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="ob-phone" class="block text-sm font-medium text-primary mb-1">Phone</label>
						<input id="ob-phone" type="tel" bind:value={church.phone} placeholder="(555) 123-4567"
							class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary" />
					</div>
					<div>
						<label for="ob-website" class="block text-sm font-medium text-primary mb-1">Website</label>
						<input id="ob-website" type="url" bind:value={church.website} placeholder="https://..."
							class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] bg-[var(--input-bg)] text-primary" />
					</div>
				</div>
			</div>

		<!-- Step 2: Modules -->
		{:else if step === 2}
			<div class="text-center mb-8">
				<h1 class="text-3xl font-bold text-primary mb-2">Choose Your Modules</h1>
				<p class="text-secondary">Select which features you want to use. You can change these later.</p>
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
				{#if modules.length === 0}
					<div class="text-center text-secondary py-8">
						<p>No modules available yet.</p>
					</div>
				{/if}
			</div>

		<!-- Step 3: Stripe Connect -->
		{:else if step === 3}
			<div class="text-center mb-8">
				<h1 class="text-3xl font-bold text-primary mb-2">Accept Donations</h1>
				<p class="text-secondary">Connect Stripe to accept online giving</p>
			</div>

			<div class="bg-surface rounded-lg border border-custom p-8 text-center space-y-6">
				{#if stripeConnected}
					<div><CheckCircle size={48} /></div>
					<div>
						<h3 class="font-semibold text-primary text-lg">Stripe Connected!</h3>
						<p class="text-sm text-secondary mt-1">Your account is ready to accept donations.</p>
					</div>
				{:else}
					<div><CreditCard size={48} /></div>
					<div>
						<h3 class="font-semibold text-primary text-lg">Connect Stripe</h3>
						<p class="text-sm text-secondary mt-1">Set up Stripe to enable online giving. You can always do this later in Settings → Billing.</p>
					</div>
					<button
						on:click={() => goto('/dashboard/giving/settings')}
						class="px-6 py-3 bg-[#635BFF] text-white rounded-lg font-medium hover:opacity-90 inline-flex items-center gap-2"
					>
						<svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor"><path d="M13.976 9.15c-2.172-.806-3.356-1.426-3.356-2.409 0-.831.683-1.305 1.901-1.305 2.227 0 4.515.858 6.09 1.631l.89-5.494C18.252.975 15.697 0 12.165 0 9.667 0 7.589.654 6.104 1.872 4.56 3.147 3.757 4.992 3.757 7.218c0 4.039 2.467 5.76 6.476 7.219 2.585.94 3.495 1.608 3.495 2.622 0 .955-.79 1.526-2.267 1.526-1.887 0-4.84-1.007-6.799-2.35L3.787 21.8C5.51 22.928 8.648 24 12.007 24c2.623 0 4.77-.64 6.307-1.803 1.693-1.263 2.558-3.163 2.558-5.486-.024-4.118-2.502-5.78-6.896-7.56z"/></svg>
						Connect Stripe Account
					</button>
				{/if}
			</div>

		<!-- Step 4: Import -->
		{:else if step === 4}
			<div class="text-center mb-8">
				<h1 class="text-3xl font-bold text-primary mb-2">Import Your Data</h1>
				<p class="text-secondary">Bring your existing data or start fresh</p>
			</div>

			<div class="space-y-3">
				<button on:click={() => { importChoice = 'pco'; }}
					class="w-full text-left bg-surface rounded-lg border p-5 transition-colors
						{importChoice === 'pco' ? 'border-[var(--teal)]' : 'border-custom hover:bg-[var(--surface-hover)]'}">
					<div class="flex items-center gap-4">
						<Building2 size={32} />
						<div>
							<h3 class="font-semibold text-primary">Import from Planning Center</h3>
							<p class="text-sm text-secondary mt-1">Automatically sync people, groups, and giving data</p>
						</div>
					</div>
				</button>
				<button on:click={() => { importChoice = 'csv'; }}
					class="w-full text-left bg-surface rounded-lg border p-5 transition-colors
						{importChoice === 'csv' ? 'border-[var(--teal)]' : 'border-custom hover:bg-[var(--surface-hover)]'}">
					<div class="flex items-center gap-4">
						<FileText size={32} />
						<div>
							<h3 class="font-semibold text-primary">Upload CSV</h3>
							<p class="text-sm text-secondary mt-1">Import from a spreadsheet</p>
						</div>
					</div>
				</button>
				<button on:click={() => { importChoice = 'skip'; }}
					class="w-full text-left bg-surface rounded-lg border p-5 transition-colors
						{importChoice === 'skip' ? 'border-[var(--teal)]' : 'border-custom hover:bg-[var(--surface-hover)]'}">
					<div class="flex items-center gap-4">
						<Sparkles size={32} />
						<div>
							<h3 class="font-semibold text-primary">Start Fresh</h3>
							<p class="text-sm text-secondary mt-1">Enter data manually later</p>
						</div>
					</div>
				</button>
			</div>

			{#if importChoice === 'pco'}
				<div class="mt-4 bg-surface rounded-lg border border-custom p-4 text-center">
					<p class="text-sm text-secondary mb-3">You'll be redirected to Planning Center to authorize the connection.</p>
					<button
						on:click={() => goto('/dashboard/settings/import')}
						class="px-4 py-2 bg-[var(--teal)] text-white rounded-lg text-sm font-medium hover:opacity-90"
					>
						Connect Planning Center →
					</button>
				</div>
			{/if}

		<!-- Step 5: Done -->
		{:else if step === 5}
			<div class="text-center py-12">
				<div class="mb-4"><PartyPopper size={64} /></div>
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
				{#if step >= 2 && step < 5}
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
						Go to Dashboard →
					{:else if step === 1}
						Save & Continue →
					{:else}
						Continue →
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}
