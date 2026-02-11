<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';

	// Kiosk state
	let step = 1; // 1=welcome, 2=amount, 3=fund, 4=info, 5=processing, 6=thank-you
	let selectedAmount = 0;
	let customAmount = '';
	let selectedFund = '';
	let donorName = '';
	let donorEmail = '';
	let loading = true;
	let error = '';
	let processing = false;

	// Config from API
	let config: any = null;
	let funds: any[] = [];
	let tenantId = '';

	// Inactivity timer
	let inactivityTimer: any = null;
	const INACTIVITY_TIMEOUT = 60000; // 60 seconds

	onMount(async () => {
		// Get tenant ID from subdomain or query param
		const urlParams = new URLSearchParams(window.location.search);
		tenantId = urlParams.get('tenant') || ''; // For dev, pass tenant ID
		
		if (!tenantId) {
			// Try to extract from subdomain
			const hostname = window.location.hostname;
			const parts = hostname.split('.');
			if (parts.length >= 2) {
				// Assume subdomain.pews.app format
				tenantId = parts[0];
			}
		}

		await loadConfig();
		startInactivityTimer();
	});

	onDestroy(() => {
		if (inactivityTimer) clearTimeout(inactivityTimer);
	});

	async function loadConfig() {
		try {
			const response = await fetch(`/api/giving/kiosk/config?tenant_id=${tenantId}`);
			if (response.ok) {
				const data = await response.json();
				config = data.config;
				funds = data.funds;

				if (!config.enabled) {
					error = 'Kiosk is not currently available';
					loading = false;
					return;
				}

				// Pre-select default fund if configured
				if (config.default_fund_id) {
					selectedFund = config.default_fund_id;
				}

				loading = false;
			} else {
				error = 'Failed to load kiosk configuration';
				loading = false;
			}
		} catch (err) {
			console.error('Failed to load config:', err);
			error = 'An error occurred. Please try again.';
			loading = false;
		}
	}

	function startInactivityTimer() {
		if (inactivityTimer) clearTimeout(inactivityTimer);
		
		inactivityTimer = setTimeout(() => {
			reset();
		}, INACTIVITY_TIMEOUT);
	}

	function resetInactivityTimer() {
		startInactivityTimer();
	}

	function reset() {
		step = 1;
		selectedAmount = 0;
		customAmount = '';
		selectedFund = config?.default_fund_id || '';
		donorName = '';
		donorEmail = '';
		processing = false;
		resetInactivityTimer();
	}

	function selectAmount(amount: number) {
		selectedAmount = amount;
		customAmount = '';
		nextStep();
	}

	function selectCustomAmount() {
		const cents = Math.round(parseFloat(customAmount) * 100);
		if (cents > 0) {
			selectedAmount = cents;
			nextStep();
		}
	}

	function selectFund(fundId: string) {
		selectedFund = fundId;
		nextStep();
	}

	function nextStep() {
		resetInactivityTimer();
		
		if (step === 1) {
			step = 2;
		} else if (step === 2 && selectedAmount > 0) {
			step = 3;
		} else if (step === 3 && selectedFund) {
			step = 4;
		}
	}

	function previousStep() {
		resetInactivityTimer();
		
		if (step > 1) {
			step--;
		}
	}

	function skipInfo() {
		submitDonation();
	}

	async function submitDonation() {
		processing = true;
		step = 5;

		try {
			const response = await fetch('/api/giving/public/checkout', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					tenant_id: tenantId,
					fund_id: selectedFund,
					amount_cents: selectedAmount,
					name: donorName || null,
					email: donorEmail || null
				})
			});

			if (response.ok) {
				const data = await response.json();
				// Redirect to Stripe Checkout
				window.location.href = data.url;
			} else {
				error = 'Failed to create checkout session';
				processing = false;
				step = 4;
			}
		} catch (err) {
			console.error('Failed to submit:', err);
			error = 'An error occurred. Please try again.';
			processing = false;
			step = 4;
		}
	}

	function formatCurrency(cents: number): string {
		return `$${(cents / 100).toFixed(2)}`;
	}
</script>

<svelte:window on:click={resetInactivityTimer} on:touchstart={resetInactivityTimer} />

<div class="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 flex items-center justify-center p-8">
	{#if loading}
		<div class="text-center">
			<div class="animate-spin rounded-full h-24 w-24 border-b-4 border-blue-600 mx-auto mb-6"></div>
			<p class="text-2xl text-gray-700">Loading...</p>
		</div>
	{:else if error}
		<div class="bg-white rounded-3xl shadow-2xl p-12 max-w-lg text-center">
			<div class="text-6xl mb-6">⚠️</div>
			<h1 class="text-3xl font-bold text-gray-800 mb-4">Oops!</h1>
			<p class="text-xl text-gray-600">{error}</p>
		</div>
	{:else if step === 1}
		<!-- Welcome Screen -->
		<div class="bg-white rounded-3xl shadow-2xl p-16 max-w-2xl text-center">
			<div class="text-8xl mb-8">❤️</div>
			<h1 class="text-5xl font-bold text-gray-800 mb-6">Welcome</h1>
			<p class="text-2xl text-gray-600 mb-12">
				Thank you for your generosity!
			</p>
			<button
				on:click={() => nextStep()}
				class="w-full py-8 bg-gradient-to-r from-blue-600 to-purple-600 text-white text-3xl font-bold rounded-2xl hover:from-blue-700 hover:to-purple-700 transition-all transform hover:scale-105 shadow-xl"
			>
				Give Now →
			</button>
		</div>
	{:else if step === 2}
		<!-- Amount Selection -->
		<div class="bg-white rounded-3xl shadow-2xl p-12 max-w-4xl">
			<h2 class="text-4xl font-bold text-gray-800 mb-8 text-center">Select Amount</h2>
			
			<div class="grid grid-cols-2 gap-6 mb-8">
				{#each config.quick_amounts as amount}
					<button
						on:click={() => selectAmount(amount)}
						class="py-10 bg-gradient-to-br from-blue-500 to-blue-600 text-white text-4xl font-bold rounded-2xl hover:from-blue-600 hover:to-blue-700 transition-all transform hover:scale-105 shadow-lg"
					>
						{formatCurrency(amount)}
					</button>
				{/each}
			</div>

			<div class="mb-8">
				<label class="block text-2xl font-semibold text-gray-700 mb-4">Custom Amount</label>
				<div class="flex gap-4">
					<input
						type="number"
						bind:value={customAmount}
						placeholder="Enter amount"
						class="flex-1 px-6 py-6 text-3xl border-4 border-gray-300 rounded-xl focus:border-blue-500 focus:outline-none"
					/>
					<button
						on:click={selectCustomAmount}
						disabled={!customAmount || parseFloat(customAmount) <= 0}
						class="px-12 py-6 bg-green-600 text-white text-2xl font-bold rounded-xl hover:bg-green-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Continue
					</button>
				</div>
			</div>

			<button
				on:click={previousStep}
				class="w-full py-6 bg-gray-200 text-gray-700 text-2xl font-semibold rounded-xl hover:bg-gray-300 transition"
			>
				← Back
			</button>
		</div>
	{:else if step === 3}
		<!-- Fund Selection -->
		<div class="bg-white rounded-3xl shadow-2xl p-12 max-w-4xl">
			<h2 class="text-4xl font-bold text-gray-800 mb-8 text-center">Select Fund</h2>
			
			<div class="space-y-4 mb-8">
				{#each funds as fund}
					<button
						on:click={() => selectFund(fund.id)}
						class="w-full p-8 bg-gradient-to-r from-purple-500 to-purple-600 text-white text-left rounded-2xl hover:from-purple-600 hover:to-purple-700 transition-all transform hover:scale-105 shadow-lg"
					>
						<div class="text-3xl font-bold mb-2">{fund.name}</div>
						{#if fund.description}
							<div class="text-xl opacity-90">{fund.description}</div>
						{/if}
					</button>
				{/each}
			</div>

			<button
				on:click={previousStep}
				class="w-full py-6 bg-gray-200 text-gray-700 text-2xl font-semibold rounded-xl hover:bg-gray-300 transition"
			>
				← Back
			</button>
		</div>
	{:else if step === 4}
		<!-- Optional Info -->
		<div class="bg-white rounded-3xl shadow-2xl p-12 max-w-3xl">
			<h2 class="text-4xl font-bold text-gray-800 mb-4 text-center">Almost Done!</h2>
			<p class="text-xl text-gray-600 mb-8 text-center">
				Would you like a receipt? (Optional)
			</p>
			
			<div class="space-y-6 mb-8">
				<div>
					<label class="block text-2xl font-semibold text-gray-700 mb-3">Name</label>
					<input
						type="text"
						bind:value={donorName}
						placeholder="Your name"
						class="w-full px-6 py-6 text-2xl border-4 border-gray-300 rounded-xl focus:border-blue-500 focus:outline-none"
					/>
				</div>
				<div>
					<label class="block text-2xl font-semibold text-gray-700 mb-3">Email</label>
					<input
						type="email"
						bind:value={donorEmail}
						placeholder="your@email.com"
						class="w-full px-6 py-6 text-2xl border-4 border-gray-300 rounded-xl focus:border-blue-500 focus:outline-none"
					/>
				</div>
			</div>

			<div class="space-y-4">
				<button
					on:click={submitDonation}
					disabled={processing}
					class="w-full py-8 bg-gradient-to-r from-green-600 to-green-700 text-white text-3xl font-bold rounded-2xl hover:from-green-700 hover:to-green-800 transition-all transform hover:scale-105 shadow-xl disabled:opacity-50"
				>
					{processing ? 'Processing...' : 'Continue to Payment →'}
				</button>
				
				<button
					on:click={skipInfo}
					disabled={processing}
					class="w-full py-6 bg-blue-100 text-blue-700 text-2xl font-semibold rounded-xl hover:bg-blue-200 transition disabled:opacity-50"
				>
					Skip (No Receipt)
				</button>

				<button
					on:click={previousStep}
					disabled={processing}
					class="w-full py-6 bg-gray-200 text-gray-700 text-2xl font-semibold rounded-xl hover:bg-gray-300 transition disabled:opacity-50"
				>
					← Back
				</button>
			</div>
		</div>
	{:else if step === 5}
		<!-- Processing -->
		<div class="bg-white rounded-3xl shadow-2xl p-16 max-w-2xl text-center">
			<div class="animate-spin rounded-full h-32 w-32 border-b-4 border-blue-600 mx-auto mb-8"></div>
			<h2 class="text-4xl font-bold text-gray-800 mb-4">Processing...</h2>
			<p class="text-2xl text-gray-600">Please wait while we prepare your checkout</p>
		</div>
	{/if}
</div>

<style>
	:global(body) {
		overflow: hidden;
		user-select: none;
		-webkit-user-select: none;
	}
</style>
