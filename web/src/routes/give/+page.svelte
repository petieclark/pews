<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8190';

	// Check for success/cancel query params
	$: status = $page.url.searchParams.get('status');

	let funds = [];
	let loading = false;
	let loadingFunds = true;
	let error = '';

	let formData = {
		fund_id: '',
		amount: '',
		guest_email: '',
		guest_name: ''
	};

	const quickAmounts = [25, 50, 100, 250, 500];

	onMount(async () => {
		await loadFunds();
	});

	async function loadFunds() {
		try {
			loadingFunds = true;
			// NOTE: This endpoint currently requires auth. 
			// For public giving, we need a public endpoint: GET /api/giving/public/funds
			// For now, we'll handle the error gracefully
			const response = await fetch(`${API_URL}/api/giving/funds`);
			
			if (response.ok) {
				const allFunds = await response.json();
				funds = allFunds.filter(f => f.is_active);
				
				// Set default fund
				if (funds.length > 0 && !formData.fund_id) {
					const defaultFund = funds.find(f => f.is_default) || funds[0];
					formData.fund_id = defaultFund.id;
				}
			} else {
				// Fallback: create a default fund option
				funds = [{ id: 'general', name: 'General Fund', description: 'General church support', is_active: true, is_default: true }];
				formData.fund_id = 'general';
			}
		} catch (err) {
			console.error('Failed to load funds:', err);
			// Fallback fund
			funds = [{ id: 'general', name: 'General Fund', description: 'General church support', is_active: true, is_default: true }];
			formData.fund_id = 'general';
		} finally {
			loadingFunds = false;
		}
	}

	function selectAmount(amount) {
		formData.amount = amount.toString();
	}

	async function submitDonation() {
		if (!formData.fund_id) {
			error = 'Please select a fund';
			return;
		}

		const amount = parseFloat(formData.amount);
		if (isNaN(amount) || amount <= 0) {
			error = 'Please enter a valid amount';
			return;
		}

		if (!formData.guest_email || !formData.guest_name) {
			error = 'Please provide your name and email';
			return;
		}

		loading = true;
		error = '';

		try {
			const amountCents = Math.round(amount * 100);

			// TODO: Create public checkout endpoint
			// POST /api/giving/public/checkout
			// Body: { fund_id, amount_cents, guest_email, guest_name }
			// Returns: { url: "stripe-checkout-url" }
			
			// For now, show what would be sent:
			console.log('Would create checkout with:', {
				fund_id: formData.fund_id,
				amount_cents: amountCents,
				guest_email: formData.guest_email,
				guest_name: formData.guest_name
			});

			// Temporary: Use existing authenticated endpoint (will fail without auth)
			const response = await fetch(`${API_URL}/api/giving/checkout`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					fund_id: formData.fund_id,
					amount_cents: amountCents,
					// NOTE: Existing endpoint expects person_id, not guest info
					// We need a new public endpoint that accepts guest details
				})
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText || 'Failed to create checkout');
			}

			const data = await response.json();
			
			// Redirect to Stripe Checkout
			if (data.url) {
				window.location.href = data.url;
			} else {
				throw new Error('No checkout URL returned');
			}
		} catch (err) {
			error = err.message || 'Failed to process donation. Please try again.';
			console.error('Donation error:', err);
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen flex items-center justify-center p-4 py-12">
	<div class="max-w-2xl w-full">
		{#if status === 'success'}
			<!-- Success Message -->
			<div class="bg-green-900 bg-opacity-30 border border-green-600 rounded-lg p-8 text-center">
				<div class="text-6xl mb-4">🎉</div>
				<h2 class="text-3xl font-bold text-green-400 mb-3">Thank You!</h2>
				<p class="text-xl text-gray-300 mb-6">
					Your generous donation has been received. You'll receive a confirmation email shortly.
				</p>
				<a
					href="/"
					class="inline-block px-6 py-3 bg-green-600 hover:bg-green-700 text-white rounded-md font-medium transition"
				>
					Return Home
				</a>
			</div>
		{:else if status === 'canceled'}
			<!-- Canceled Message -->
			<div class="bg-yellow-900 bg-opacity-30 border border-yellow-600 rounded-lg p-8 text-center mb-8">
				<div class="text-5xl mb-4">ℹ️</div>
				<h2 class="text-2xl font-bold text-yellow-400 mb-3">Donation Canceled</h2>
				<p class="text-gray-300 mb-4">
					Your donation was not completed. You can try again below.
				</p>
			</div>
		{/if}

		<!-- Header -->
		<div class="text-center mb-8">
			<h1 class="text-4xl md:text-5xl font-bold text-white mb-4">
				💚 Give Online
			</h1>
			<p class="text-xl text-gray-300">
				Thank you for your generosity! Your donation helps us fulfill our mission.
			</p>
		</div>

		<!-- Form -->
		<div class="bg-gray-800 bg-opacity-50 backdrop-blur rounded-lg shadow-2xl p-6 md:p-8">
			{#if error}
				<div class="bg-red-900 bg-opacity-30 border border-red-600 text-red-300 px-4 py-3 rounded-lg mb-6">
					{error}
				</div>
			{/if}

			<form on:submit|preventDefault={submitDonation} class="space-y-6">
				<!-- Fund Selection -->
				<div>
					<label for="fund-select" class="block text-sm font-medium text-gray-300 mb-2">
						Give To <span class="text-red-400">*</span>
					</label>
					{#if loadingFunds}
						<div class="text-gray-400">Loading funds...</div>
					{:else}
						<select
							id="fund-select"
							bind:value={formData.fund_id}
							required
							class="w-full px-4 py-3 bg-gray-900 bg-opacity-50 border border-gray-600 text-white rounded-md focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none"
						>
							<option value="">Select a fund...</option>
							{#each funds as fund}
								<option value={fund.id}>{fund.name}</option>
							{/each}
						</select>
						{#if formData.fund_id}
							{@const selectedFund = funds.find(f => f.id === formData.fund_id)}
							{#if selectedFund?.description}
								<p class="text-sm text-gray-400 mt-2">{selectedFund.description}</p>
							{/if}
						{/if}
					{/if}
				</div>

				<!-- Quick Amount Buttons -->
				<div>
					<div class="block text-sm font-medium text-gray-300 mb-3">
						Quick Select Amount
					</div>
					<div class="grid grid-cols-3 md:grid-cols-5 gap-2">
						{#each quickAmounts as amount}
							<button
								type="button"
								on:click={() => selectAmount(amount)}
								class="px-4 py-3 bg-gray-700 hover:bg-teal-600 text-white rounded-md font-medium transition"
								class:bg-teal-600={formData.amount === amount.toString()}
							>
								${amount}
							</button>
						{/each}
					</div>
				</div>

				<!-- Custom Amount -->
				<div>
					<label for="amount" class="block text-sm font-medium text-gray-300 mb-2">
						Amount <span class="text-red-400">*</span>
					</label>
					<div class="relative">
						<span class="absolute left-4 top-3.5 text-gray-400 text-lg">$</span>
						<input
							id="amount"
							type="number"
							step="0.01"
							min="1"
							bind:value={formData.amount}
							required
							placeholder="0.00"
							class="w-full pl-10 pr-4 py-3 bg-gray-900 bg-opacity-50 border border-gray-600 text-white text-lg rounded-md focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none placeholder-gray-500"
						/>
					</div>
				</div>

				<!-- Guest Information -->
				<div class="border-t border-gray-700 pt-6">
					<h3 class="text-lg font-semibold text-white mb-4">Your Information</h3>
					
					<div class="space-y-4">
						<div>
							<label for="guest-name" class="block text-sm font-medium text-gray-300 mb-2">
								Full Name <span class="text-red-400">*</span>
							</label>
							<input
								id="guest-name"
								type="text"
								bind:value={formData.guest_name}
								required
								placeholder="John Doe"
								class="w-full px-4 py-3 bg-gray-900 bg-opacity-50 border border-gray-600 text-white rounded-md focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none placeholder-gray-500"
							/>
						</div>

						<div>
							<label for="guest-email" class="block text-sm font-medium text-gray-300 mb-2">
								Email <span class="text-red-400">*</span>
							</label>
							<input
								id="guest-email"
								type="email"
								bind:value={formData.guest_email}
								required
								placeholder="john@example.com"
								class="w-full px-4 py-3 bg-gray-900 bg-opacity-50 border border-gray-600 text-white rounded-md focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none placeholder-gray-500"
							/>
							<p class="text-sm text-gray-400 mt-1">For your donation receipt</p>
						</div>
					</div>
				</div>

				<!-- Submit Button -->
				<button
					type="submit"
					disabled={loading || loadingFunds}
					class="w-full px-6 py-4 bg-teal-600 hover:bg-teal-700 text-white text-lg font-bold rounded-md transition disabled:opacity-50 disabled:cursor-not-allowed shadow-lg hover:shadow-xl"
				>
					{#if loading}
						Processing...
					{:else}
						Continue to Secure Checkout
					{/if}
				</button>

				<div class="flex items-center justify-center gap-2 text-gray-400 text-sm">
					<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
						<path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" />
					</svg>
					<span>Secure payment powered by Stripe</span>
				</div>
			</form>

			<!-- Development Note -->
			<div class="mt-6 p-4 bg-blue-900 bg-opacity-20 border border-blue-600 rounded-lg">
				<p class="text-sm text-blue-300">
					<strong>⚠️ Backend TODO:</strong> Create public checkout endpoint at 
					<code class="bg-gray-900 px-2 py-1 rounded">POST /api/giving/public/checkout</code>
					that accepts guest donations without authentication. See BACKEND_TODO_PUBLIC_PAGES.md for details.
				</p>
			</div>
		</div>

		<!-- Footer -->
		<div class="text-center mt-8">
			<p class="text-gray-500 text-sm">
				Powered by <span class="font-bold text-teal-500">Pews</span>
			</p>
		</div>
	</div>
</div>
