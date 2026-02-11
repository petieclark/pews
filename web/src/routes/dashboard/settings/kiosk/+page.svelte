<script lang="ts">
	import { onMount } from 'svelte';

	let loading = true;
	let saving = false;
	let message = '';
	let messageType: 'success' | 'error' | '' = '';

	// Config state
	let enabled = false;
	let quickAmounts: number[] = [1000, 2500, 5000, 10000, 25000];
	let defaultFundId: string | null = null;
	let thankYouMessage = 'Thank you for your generous gift!';

	// Funds for dropdown
	let funds: any[] = [];

	// New amount input
	let newAmount = '';

	onMount(async () => {
		await loadConfig();
		await loadFunds();
		loading = false;
	});

	async function loadConfig() {
		try {
			const response = await fetch('/api/giving/kiosk', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});

			if (response.ok) {
				const config = await response.json();
				enabled = config.enabled;
				quickAmounts = config.quick_amounts || [1000, 2500, 5000, 10000, 25000];
				defaultFundId = config.default_fund_id;
				thankYouMessage = config.thank_you_message || 'Thank you for your generous gift!';
			}
		} catch (error) {
			console.error('Failed to load config:', error);
		}
	}

	async function loadFunds() {
		try {
			const response = await fetch('/api/giving/funds', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});

			if (response.ok) {
				funds = await response.json();
			}
		} catch (error) {
			console.error('Failed to load funds:', error);
		}
	}

	async function saveConfig() {
		saving = true;
		message = '';

		try {
			const response = await fetch('/api/giving/kiosk', {
				method: 'PUT',
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`,
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					enabled,
					quick_amounts: quickAmounts,
					default_fund_id: defaultFundId || null,
					thank_you_message: thankYouMessage
				})
			});

			if (response.ok) {
				message = 'Kiosk settings saved successfully!';
				messageType = 'success';
			} else {
				message = 'Failed to save settings';
				messageType = 'error';
			}
		} catch (error) {
			console.error('Failed to save:', error);
			message = 'An error occurred';
			messageType = 'error';
		} finally {
			saving = false;
			
			// Clear message after 3 seconds
			setTimeout(() => {
				message = '';
				messageType = '';
			}, 3000);
		}
	}

	function addQuickAmount() {
		const cents = Math.round(parseFloat(newAmount) * 100);
		if (cents > 0 && !quickAmounts.includes(cents)) {
			quickAmounts = [...quickAmounts, cents].sort((a, b) => a - b);
			newAmount = '';
		}
	}

	function removeQuickAmount(amount: number) {
		quickAmounts = quickAmounts.filter(a => a !== amount);
	}

	function formatCurrency(cents: number): string {
		return `$${(cents / 100).toFixed(2)}`;
	}

	function openKiosk() {
		// Get tenant ID from token or current user
		const url = `/giving-kiosk`; // Will need tenant param in production
		window.open(url, '_blank');
	}
</script>

<div class="p-6 max-w-4xl mx-auto">
	<div class="mb-6">
		<h1 class="text-3xl font-bold text-[#1B3A4B]">Kiosk Settings</h1>
		<p class="text-gray-600 mt-1">Configure your touch-screen giving kiosk</p>
	</div>

	{#if message}
		<div class="mb-6 p-4 rounded-lg {messageType === 'success' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}">
			{message}
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center items-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[#4A8B8C]"></div>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-lg p-8 space-y-8">
			<!-- Enable/Disable -->
			<div class="flex items-center justify-between pb-6 border-b">
				<div>
					<h2 class="text-xl font-semibold text-[#1B3A4B]">Enable Kiosk</h2>
					<p class="text-gray-600 text-sm mt-1">
						Turn the giving kiosk on or off
					</p>
				</div>
				<label class="relative inline-flex items-center cursor-pointer">
					<input type="checkbox" bind:checked={enabled} class="sr-only peer">
					<div class="w-14 h-8 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-6 peer-checked:after:border-white after:content-[''] after:absolute after:top-1 after:left-1 after:bg-white after:border-gray-300 after:border after:rounded-full after:h-6 after:w-6 after:transition-all peer-checked:bg-[#4A8B8C]"></div>
				</label>
			</div>

			<!-- Quick Amounts -->
			<div>
				<h2 class="text-xl font-semibold text-[#1B3A4B] mb-4">Quick Amount Buttons</h2>
				<p class="text-gray-600 text-sm mb-4">
					Configure the preset donation amounts shown on the kiosk
				</p>

				<div class="flex flex-wrap gap-3 mb-4">
					{#each quickAmounts as amount}
						<div class="inline-flex items-center gap-2 px-4 py-2 bg-[#4A8B8C] text-white rounded-lg">
							<span class="font-semibold">{formatCurrency(amount)}</span>
							<button
								on:click={() => removeQuickAmount(amount)}
								class="ml-2 text-white hover:text-red-200 transition"
							>
								✕
							</button>
						</div>
					{/each}
				</div>

				<div class="flex gap-3">
					<input
						type="number"
						bind:value={newAmount}
						placeholder="Enter amount (e.g., 50)"
						class="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:border-[#4A8B8C] focus:outline-none"
					/>
					<button
						on:click={addQuickAmount}
						disabled={!newAmount || parseFloat(newAmount) <= 0}
						class="px-6 py-2 bg-[#4A8B8C] text-white font-semibold rounded-lg hover:bg-[#3d7576] transition disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Add Amount
					</button>
				</div>
			</div>

			<!-- Default Fund -->
			<div>
				<h2 class="text-xl font-semibold text-[#1B3A4B] mb-4">Default Fund</h2>
				<p class="text-gray-600 text-sm mb-4">
					Pre-select a fund for faster donations (optional)
				</p>

				<select
					bind:value={defaultFundId}
					class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:border-[#4A8B8C] focus:outline-none"
				>
					<option value={null}>No default (user selects)</option>
					{#each funds.filter(f => f.is_active) as fund}
						<option value={fund.id}>{fund.name}</option>
					{/each}
				</select>
			</div>

			<!-- Thank You Message -->
			<div>
				<h2 class="text-xl font-semibold text-[#1B3A4B] mb-4">Thank You Message</h2>
				<p class="text-gray-600 text-sm mb-4">
					Customize the message shown after a successful donation
				</p>

				<textarea
					bind:value={thankYouMessage}
					rows="3"
					class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:border-[#4A8B8C] focus:outline-none resize-none"
					placeholder="Thank you for your generous gift!"
				></textarea>
			</div>

			<!-- Save Button -->
			<div class="pt-6 border-t flex gap-4">
				<button
					on:click={saveConfig}
					disabled={saving}
					class="flex-1 px-8 py-4 bg-[#4A8B8C] text-white text-lg font-semibold rounded-lg hover:bg-[#3d7576] transition disabled:opacity-50 disabled:cursor-not-allowed shadow-lg"
				>
					{saving ? 'Saving...' : 'Save Settings'}
				</button>

				{#if enabled}
					<button
						on:click={openKiosk}
						class="px-8 py-4 bg-blue-600 text-white text-lg font-semibold rounded-lg hover:bg-blue-700 transition shadow-lg"
					>
						Open Kiosk →
					</button>
				{/if}
			</div>
		</div>

		<!-- Info Box -->
		<div class="mt-6 p-6 bg-blue-50 border border-blue-200 rounded-lg">
			<div class="flex gap-3">
				<span class="text-blue-600 text-2xl">ℹ️</span>
				<div>
					<p class="font-semibold text-blue-900 mb-2">Kiosk Setup Tips</p>
					<ul class="text-sm text-blue-800 space-y-1">
						<li>• Set up an iPad or tablet in landscape orientation (1366x1024)</li>
						<li>• Enable full-screen mode in the browser</li>
						<li>• Enable Guided Access on iPad to prevent users from exiting</li>
						<li>• Place the kiosk in a visible, accessible location</li>
						<li>• The kiosk will auto-reset after 60 seconds of inactivity</li>
					</ul>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	/* Toggle switch styles are inline above */
</style>
