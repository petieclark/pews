<script>
	import { onMount } from 'svelte';

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8190';

	let funds = [];
	let selectedFund = '';
	let amount = '';
	let loading = false;
	let church = null;

	onMount(async () => {
		await loadChurchInfo();
		await loadFunds();
	});

	async function loadChurchInfo() {
		try {
			// Get church info from the first available tenant
			// In a real deployment, this would be based on subdomain or config
			const response = await fetch(`${API_URL}/api/health`);
			church = {
				name: 'Your Church',
				logo: null
			};
		} catch (error) {
			console.error('Failed to load church info:', error);
		}
	}

	async function loadFunds() {
		try {
			// For now, we'll use hardcoded funds since the API requires auth
			// In production, this endpoint should be public
			funds = [
				{ id: 'general', name: 'General Fund', description: 'Support the overall ministry' },
				{ id: 'missions', name: 'Missions', description: 'Support our global outreach' },
				{ id: 'building', name: 'Building Fund', description: 'Help us expand our facilities' }
			];
			if (funds.length > 0) {
				selectedFund = funds[0].id;
			}
		} catch (error) {
			console.error('Failed to load funds:', error);
		}
	}

	async function handleGive() {
		if (!selectedFund || !amount || parseFloat(amount) <= 0) {
			alert('Please select a fund and enter a valid amount');
			return;
		}

		loading = true;
		
		try {
			// Create a Stripe checkout session
			const response = await fetch(`${API_URL}/api/giving/checkout`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					fund_id: selectedFund,
					amount: parseFloat(amount),
					currency: 'usd'
				})
			});

			if (!response.ok) throw new Error('Failed to create checkout session');
			
			const data = await response.json();
			
			// Redirect to Stripe checkout
			if (data.url) {
				window.location.href = data.url;
			}
		} catch (error) {
			console.error('Failed to process donation:', error);
			alert('Unable to process donation. Please try again or contact us.');
		} finally {
			loading = false;
		}
	}

	function setQuickAmount(value) {
		amount = value.toString();
	}
</script>

<svelte:head>
	<title>Give | {church?.name || 'Church'}</title>
	<meta name="description" content="Support our ministry with a one-time gift." />
	<meta property="og:title" content="Give | {church?.name || 'Church'}" />
	<meta property="og:description" content="Support our ministry with a one-time gift." />
	<meta property="og:type" content="website" />
</svelte:head>

<div class="min-h-screen" style="background-color: #0a0a0a; color: #ffffff">
	<div class="max-w-2xl mx-auto px-4 py-12">
		<!-- Header -->
		<div class="text-center mb-12">
			{#if church?.logo}
				<img src={church.logo} alt={church.name} class="h-16 mx-auto mb-4" />
			{/if}
			<h1 class="text-4xl font-bold text-white mb-3">Give</h1>
			<p class="text-gray-400 text-lg">
				Thank you for your generosity! Your gift helps us continue our mission and serve our community.
			</p>
		</div>

		<!-- Giving Form -->
		<div class="bg-gray-900 rounded-lg p-8 shadow-xl">
			<form on:submit|preventDefault={handleGive} class="space-y-6">
				<!-- Fund Selection -->
				<div>
					<label class="block text-sm font-medium text-gray-300 mb-2">
						Select Fund
					</label>
					<select
						bind:value={selectedFund}
						class="w-full px-4 py-3 bg-gray-800 border border-gray-700 text-white rounded-md focus:border-teal-500 focus:outline-none"
					>
						{#each funds as fund}
							<option value={fund.id}>{fund.name}</option>
						{/each}
					</select>
					{#if selectedFund}
						{@const fund = funds.find(f => f.id === selectedFund)}
						{#if fund?.description}
							<p class="text-sm text-gray-500 mt-1">{fund.description}</p>
						{/if}
					{/if}
				</div>

				<!-- Quick Amounts -->
				<div>
					<label class="block text-sm font-medium text-gray-300 mb-2">
						Quick Select
					</label>
					<div class="grid grid-cols-4 gap-2">
						<button
							type="button"
							on:click={() => setQuickAmount(25)}
							class="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-white rounded-md border border-gray-700"
						>
							$25
						</button>
						<button
							type="button"
							on:click={() => setQuickAmount(50)}
							class="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-white rounded-md border border-gray-700"
						>
							$50
						</button>
						<button
							type="button"
							on:click={() => setQuickAmount(100)}
							class="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-white rounded-md border border-gray-700"
						>
							$100
						</button>
						<button
							type="button"
							on:click={() => setQuickAmount(250)}
							class="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-white rounded-md border border-gray-700"
						>
							$250
						</button>
					</div>
				</div>

				<!-- Custom Amount -->
				<div>
					<label class="block text-sm font-medium text-gray-300 mb-2">
						Amount (USD)
					</label>
					<div class="relative">
						<span class="absolute left-4 top-3 text-gray-400">$</span>
						<input
							type="number"
							bind:value={amount}
							min="1"
							step="0.01"
							required
							placeholder="0.00"
							class="w-full pl-8 pr-4 py-3 bg-gray-800 border border-gray-700 text-white rounded-md focus:border-teal-500 focus:outline-none"
						/>
					</div>
				</div>

				<!-- Submit Button -->
				<button
					type="submit"
					disabled={loading}
					class="w-full px-6 py-4 bg-teal-600 hover:bg-teal-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white font-bold text-lg rounded-md transition-colors"
				>
					{#if loading}
						Processing...
					{:else}
						Continue to Payment
					{/if}
				</button>

				<!-- Security Note -->
				<p class="text-xs text-gray-500 text-center">
					🔒 Secure payment processed by Stripe. Your information is safe and encrypted.
				</p>
			</form>
		</div>

		<!-- Footer -->
		<div class="text-center mt-8 space-y-2">
			<p class="text-sm text-gray-500">
				Questions about giving? Contact us at giving@church.com
			</p>
			<p class="text-xs text-gray-600">
				Powered by <span class="font-bold text-teal-500">Pews</span>
			</p>
		</div>
	</div>
</div>
