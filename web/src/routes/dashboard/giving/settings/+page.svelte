<script lang="ts">
	import { onMount } from 'svelte';

	let connectStatus = {
		connected: false,
		onboarding_completed: false,
		charges_enabled: false,
		payouts_enabled: false,
		account_id: ''
	};

	let loading = true;
	let creatingLink = false;
	let tenantName = '';
	let tenantEmail = '';

	onMount(async () => {
		await loadConnectStatus();
		await loadTenantInfo();
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

	async function loadTenantInfo() {
		try {
			const response = await fetch('/api/tenant', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});
			if (response.ok) {
				const tenant = await response.json();
				tenantName = tenant.name;
			}
		} catch (error) {
			console.error('Failed to load tenant:', error);
		}
	}

	async function startOnboarding() {
		if (!tenantEmail) {
			alert('Please enter an email address');
			return;
		}

		creatingLink = true;

		try {
			const response = await fetch('/api/giving/connect/onboard', {
				method: 'POST',
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`,
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					tenant_name: tenantName,
					tenant_email: tenantEmail
				})
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

	async function refreshStatus() {
		loading = true;
		await loadConnectStatus();
		loading = false;
	}
</script>

<div class="p-6 max-w-4xl mx-auto">
	<div class="mb-6">
		<h1 class="text-3xl font-bold text-[#1B3A4B]">Giving Settings</h1>
		<p class="text-gray-600 mt-1">Configure online giving and payment processing</p>
	</div>

	{#if loading}
		<div class="flex justify-center items-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[#4A8B8C]"></div>
		</div>
	{:else}
		<!-- Stripe Connect Status -->
		<div class="bg-white rounded-lg shadow p-6 mb-6">
			<div class="flex justify-between items-start mb-4">
				<div>
					<h2 class="text-xl font-semibold text-[#1B3A4B]">Stripe Connect</h2>
					<p class="text-sm text-gray-600 mt-1">Accept online donations via credit card and ACH</p>
				</div>
				{#if connectStatus.connected}
					<button
						on:click={refreshStatus}
						class="text-sm text-[#4A8B8C] hover:underline"
					>
						Refresh Status
					</button>
				{/if}
			</div>

			{#if !connectStatus.connected}
				<!-- Not Connected -->
				<div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
					<p class="text-sm text-yellow-800 mb-4">
						⚠️ Online giving is not enabled yet. Connect your Stripe account to start accepting online donations.
					</p>
				</div>

				<div class="space-y-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">
							Email Address <span class="text-red-500">*</span>
						</label>
						<input
							type="email"
							bind:value={tenantEmail}
							required
							placeholder="your-email@church.org"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent"
						/>
						<p class="text-sm text-gray-500 mt-1">
							This email will be used for your Stripe account
						</p>
					</div>

					<button
						on:click={startOnboarding}
						disabled={creatingLink}
						class="w-full px-4 py-3 bg-[#635BFF] text-white rounded-lg hover:bg-[#554edd] transition disabled:opacity-50 disabled:cursor-not-allowed font-medium"
					>
						{creatingLink ? 'Creating...' : 'Connect with Stripe'}
					</button>
				</div>
			{:else}
				<!-- Connected -->
				<div class="space-y-4">
					<div class="grid grid-cols-2 gap-4">
						<div class="bg-gray-50 rounded-lg p-4">
							<p class="text-sm text-gray-600 mb-1">Account Status</p>
							<p class="text-lg font-semibold text-[#1B3A4B]">
								{connectStatus.onboarding_completed ? '✅ Active' : '⏳ Pending'}
							</p>
						</div>

						<div class="bg-gray-50 rounded-lg p-4">
							<p class="text-sm text-gray-600 mb-1">Stripe Account</p>
							<p class="text-sm font-mono text-gray-700">{connectStatus.account_id}</p>
						</div>
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div class="flex items-center gap-2">
							{#if connectStatus.charges_enabled}
								<span class="text-green-600">✅</span>
							{:else}
								<span class="text-red-600">❌</span>
							{/if}
							<span class="text-sm text-gray-700">Charges Enabled</span>
						</div>

						<div class="flex items-center gap-2">
							{#if connectStatus.payouts_enabled}
								<span class="text-green-600">✅</span>
							{:else}
								<span class="text-red-600">❌</span>
							{/if}
							<span class="text-sm text-gray-700">Payouts Enabled</span>
						</div>
					</div>

					{#if !connectStatus.onboarding_completed}
						<div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
							<p class="text-sm text-yellow-800 mb-2">
								⚠️ Complete your Stripe onboarding to start accepting donations
							</p>
							<button
								on:click={startOnboarding}
								class="text-sm text-yellow-900 underline hover:no-underline"
							>
								Continue Onboarding
							</button>
						</div>
					{:else}
						<div class="bg-green-50 border border-green-200 rounded-lg p-4">
							<p class="text-sm text-green-800">
								✅ Online giving is active! Donations will be processed through your connected Stripe account.
							</p>
						</div>
					{/if}
				</div>
			{/if}
		</div>

		<!-- Platform Fee Info -->
		<div class="bg-blue-50 border border-blue-200 rounded-lg p-6">
			<h3 class="font-semibold text-blue-900 mb-2">💰 Platform Fee</h3>
			<p class="text-sm text-blue-800 mb-2">
				Pews charges a 1% platform fee on all online donations (minimum $0.30 per transaction).
				This is in addition to Stripe's standard payment processing fees.
			</p>
			<p class="text-sm text-blue-800">
				Example: A $100 donation will have a $1 platform fee + Stripe fees (~$3.20) = $95.80 net to your church.
			</p>
		</div>
	{/if}
</div>
