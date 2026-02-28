<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { DollarSign, CheckCircle, XCircle } from 'lucide-svelte';

	let connectStatus = {
		connected: false,
		onboarding_completed: false,
		charges_enabled: false,
		payouts_enabled: false,
		account_id: '',
		is_test_mode: false
	};
	
	let testModeDismissed = false;

	let loading = true;
	let creatingLink = false;

	onMount(async () => {
		await loadConnectStatus();
		
		// Check if returning from Stripe onboarding
		const setup = $page.url.searchParams.get('setup');
		if (setup === 'complete') {
			await handleOnboardingReturn();
		} else if (setup === 'refresh') {
			await handleOnboardingRefresh();
		}
		
		loading = false;
	});

	async function loadConnectStatus() {
		try {
			const response = await fetch('/api/giving/connect/status', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});
			if (response.ok) {
				connectStatus = await response.json();
			}
		} catch (error) {
			console.error('Failed to load connect status:', error);
		}
	}

	async function handleOnboardingReturn() {
		try {
			const response = await fetch('/api/giving/connect/return', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});
			if (response.ok) {
				connectStatus = await response.json();
			}
		} catch (error) {
			console.error('Failed to handle return:', error);
		}
	}

	async function handleOnboardingRefresh() {
		// If refresh needed, redirect to new onboarding link
		await startOnboarding();
	}

	async function startOnboarding() {
		creatingLink = true;

		try {
			const response = await fetch('/api/giving/connect/onboard', {
				method: 'POST',
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`,
					'Content-Type': 'application/json'
				}
			});

			if (response.ok) {
				const data = await response.json();
				// Redirect to Stripe onboarding
				window.location.href = data.url;
			} else {
				alert('Failed to create onboarding link');
			}
		} catch (error) {
			console.error('Failed to start onboarding:', error);
			alert('An error occurred');
		} finally {
			creatingLink = false;
		}
	}

	async function completeSetup() {
		creatingLink = true;

		try {
			const response = await fetch('/api/giving/connect/refresh', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});

			if (response.ok) {
				const data = await response.json();
				// Redirect to Stripe onboarding
				window.location.href = data.url;
			} else {
				alert('Failed to generate onboarding link');
			}
		} catch (error) {
			console.error('Failed to complete setup:', error);
			alert('An error occurred');
		} finally {
			creatingLink = false;
		}
	}

	async function refreshStatus() {
		loading = true;
		await loadConnectStatus();
		loading = false;
	}

	function openStripeDashboard() {
		window.open(`https://dashboard.stripe.com/${connectStatus.account_id}`, '_blank');
	}
</script>

<div class="p-6 max-w-4xl mx-auto">
	<div class="mb-6">
		<h1 class="text-3xl font-bold text-primary">Giving Settings</h1>
		<p class="text-secondary mt-1">Configure online giving and payment processing</p>
	</div>

	{#if loading}
		<div class="flex justify-center items-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[#4A8B8C]"></div>
		</div>
	{:else}
		<div class="bg-surface rounded-lg shadow border border-custom p-8 mb-6">
			{#if !connectStatus.connected}
				<!-- Not Connected -->
				<div class="text-center">
					<div class="mb-6">
						<div class="inline-block mb-4"><DollarSign size={64} /></div>
						<h2 class="text-2xl font-bold text-primary mb-2">Accept Online Donations</h2>
					</div>

					<div class="max-w-2xl mx-auto text-left mb-8">
						<p class="text-primary mb-4">
							Set up your church to receive online donations via credit card and bank transfer.
						</p>
						<p class="text-primary mb-4">
							We'll guide you through a quick setup with Stripe, our payment processor. You'll need:
						</p>
						<ul class="list-disc list-inside text-primary mb-4 space-y-2">
							<li>Your church's bank account details</li>
							<li>Basic organization information</li>
						</ul>
						<p class="text-sm text-secondary mb-6">
							Takes about 5 minutes.
						</p>
					</div>

					<button
						on:click={startOnboarding}
						disabled={creatingLink}
						class="px-8 py-4 bg-[var(--teal)] text-white text-lg font-semibold rounded-lg hover:opacity-90 transition disabled:opacity-50 disabled:cursor-not-allowed shadow border border-custom"
					>
						{creatingLink ? 'Creating...' : 'Enable Online Giving →'}
					</button>

					<div class="mt-8 p-4 bg-blue-50 border border-blue-200 rounded-lg text-left max-w-2xl mx-auto">
						<div class="flex gap-2">
							<span class="text-blue-600">ℹ️</span>
							<div>
								<p class="font-semibold text-blue-900 mb-1">Platform Fee</p>
								<p class="text-sm text-blue-800 mb-2">
									1% on donations (min $0.30) + Stripe fees
								</p>
								<p class="text-sm text-blue-700">
									Example: $100 donation → $95.80 to church
								</p>
							</div>
						</div>
					</div>
				</div>
			{:else if !connectStatus.onboarding_completed || !connectStatus.charges_enabled}
				<!-- Pending Verification -->
				<div class="text-center">
					<div class="mb-6">
						<div class="inline-block text-6xl mb-4">⏳</div>
						<h2 class="text-2xl font-bold text-primary mb-2">Stripe Setup In Progress</h2>
					</div>

					<p class="text-primary mb-6 max-w-lg mx-auto">
						Your Stripe account is being verified. This usually takes 1-2 business days.
					</p>

					{#if !connectStatus.onboarding_completed}
						<button
							on:click={completeSetup}
							disabled={creatingLink}
							class="px-6 py-3 bg-[var(--teal)] text-white font-semibold rounded-lg hover:opacity-90 transition disabled:opacity-50 disabled:cursor-not-allowed"
						>
							{creatingLink ? 'Loading...' : 'Complete Setup →'}
						</button>
					{/if}

					<div class="mt-6">
						<button
							on:click={refreshStatus}
							class="text-sm text-[var(--teal)] hover:underline"
						>
							Refresh Status
						</button>
					</div>
				</div>
			{:else}
				<!-- Fully Connected -->
				<div class="text-center">
					<div class="mb-6">
						<div class="inline-block mb-4"><CheckCircle size={64} /></div>
						<h2 class="text-2xl font-bold text-primary mb-2">Online Giving Active</h2>
					</div>

					<div class="max-w-md mx-auto mb-8">
						<div class="grid grid-cols-2 gap-4 text-left">
							<div class="flex items-center gap-2 p-3 bg-[var(--surface-hover)] rounded-lg">
								<span>{connectStatus.charges_enabled ? '' : ''}</span>{#if connectStatus.charges_enabled}<CheckCircle size={24} />{:else}<XCircle size={24} />{/if}
								<span class="text-sm text-primary">Charges enabled</span>
							</div>
							<div class="flex items-center gap-2 p-3 bg-[var(--surface-hover)] rounded-lg">
								<span>{connectStatus.payouts_enabled ? '' : ''}</span>{#if connectStatus.payouts_enabled}<CheckCircle size={24} />{:else}<XCircle size={24} />{/if}
								<span class="text-sm text-primary">Payouts enabled</span>
							</div>
						</div>
					</div>

					<button
						on:click={openStripeDashboard}
						class="px-6 py-3 bg-[#635BFF] text-white font-semibold rounded-lg hover:bg-[#554edd] transition"
					>
						Manage Stripe Dashboard →
					</button>

					<div class="mt-6">
						<button
							on:click={refreshStatus}
							class="text-sm text-[var(--teal)] hover:underline"
						>
							Refresh Status
						</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	/* Add any custom styles here */
</style>
