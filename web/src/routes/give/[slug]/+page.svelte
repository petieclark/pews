<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	let slug = $page.params.slug;
	let church = null;
	let loading = true;
	let error = '';
	let submitting = false;

	// Form state
	let selectedAmount = 5000; // $50 default
	let customAmount = '';
	let useCustom = false;
	let selectedFundId = '';
	let donorName = '';
	let donorEmail = '';
	let frequency = 'one-time';

	const presetAmounts = [2500, 5000, 10000, 25000];

	onMount(async () => {
		try {
			const res = await fetch(`/api/public/give/${slug}`);
			if (!res.ok) throw new Error('Church not found');
			church = await res.json();
			if (church.funds?.length > 0) {
				const defaultFund = church.funds.find(f => f.is_default);
				selectedFundId = defaultFund ? defaultFund.id : church.funds[0].id;
			}
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	});

	function selectPreset(amount) {
		selectedAmount = amount;
		useCustom = false;
		customAmount = '';
	}

	function enableCustom() {
		useCustom = true;
		selectedAmount = 0;
	}

	function getAmountCents() {
		if (useCustom) {
			const dollars = parseFloat(customAmount);
			return isNaN(dollars) ? 0 : Math.round(dollars * 100);
		}
		return selectedAmount;
	}

	function formatDollars(cents) {
		return '$' + (cents / 100).toFixed(0);
	}

	async function handleGive() {
		const amountCents = getAmountCents();
		if (amountCents < 100) {
			alert('Minimum donation is $1.00');
			return;
		}
		if (!selectedFundId) {
			alert('Please select a fund');
			return;
		}

		submitting = true;
		try {
			const res = await fetch(`/api/public/give`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					tenant_slug: slug,
					fund_id: selectedFundId,
					amount_cents: amountCents,
					donor_name: donorName,
					donor_email: donorEmail,
					frequency
				})
			});

			if (!res.ok) {
				const errText = await res.text();
				throw new Error(errText || 'Failed to create checkout');
			}

			const data = await res.json();
			window.location.href = data.url;
		} catch (err) {
			alert(err.message);
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>{church ? `Give to ${church.name}` : 'Online Giving'}</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 flex flex-col items-center justify-start pt-8 pb-16 px-4">
	{#if loading}
		<div class="flex justify-center items-center py-24">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-teal-600"></div>
		</div>
	{:else if error}
		<div class="text-center py-24">
			<div class="text-6xl mb-4">🏛️</div>
			<h1 class="text-2xl font-bold text-gray-800 mb-2">Church Not Found</h1>
			<p class="text-gray-500">This giving page doesn't exist or has been removed.</p>
		</div>
	{:else if !church.giving_enabled}
		<div class="text-center py-24">
			<div class="text-6xl mb-4">💒</div>
			<h1 class="text-2xl font-bold text-gray-800 mb-2">{church.name}</h1>
			<p class="text-gray-500">Online giving is not yet available for this church.</p>
		</div>
	{:else}
		<!-- Church Header -->
		<div class="text-center mb-8">
			{#if church.logo}
				<img src={church.logo} alt={church.name} class="w-20 h-20 rounded-full mx-auto mb-4 object-cover" />
			{:else}
				<div class="w-20 h-20 rounded-full mx-auto mb-4 bg-teal-100 flex items-center justify-center text-3xl">⛪</div>
			{/if}
			<h1 class="text-2xl font-bold text-gray-900">{church.name}</h1>
			<p class="text-gray-500 mt-1">Online Giving</p>
		</div>

		<!-- Give Form -->
		<div class="w-full max-w-md bg-white rounded-2xl shadow-lg p-6 space-y-6">
			<!-- Amount -->
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-3">Amount</label>
				<div class="grid grid-cols-2 gap-3 mb-3">
					{#each presetAmounts as amount}
						<button
							on:click={() => selectPreset(amount)}
							class="py-3 rounded-lg text-lg font-semibold border-2 transition
								{!useCustom && selectedAmount === amount
									? 'border-teal-600 bg-teal-50 text-teal-700'
									: 'border-gray-200 bg-white text-gray-700 hover:border-gray-300'}"
						>
							{formatDollars(amount)}
						</button>
					{/each}
				</div>
				<button
					on:click={enableCustom}
					class="w-full py-3 rounded-lg text-lg font-semibold border-2 transition mb-2
						{useCustom
							? 'border-teal-600 bg-teal-50 text-teal-700'
							: 'border-gray-200 bg-white text-gray-700 hover:border-gray-300'}"
				>
					Other Amount
				</button>
				{#if useCustom}
					<div class="relative mt-2">
						<span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 text-lg">$</span>
						<input
							type="number"
							bind:value={customAmount}
							placeholder="0.00"
							min="1"
							step="0.01"
							class="w-full pl-8 pr-4 py-3 border-2 border-gray-200 rounded-lg text-lg focus:border-teal-600 focus:outline-none"
						/>
					</div>
				{/if}
			</div>

			<!-- Fund -->
			{#if church.funds?.length > 1}
				<div>
					<label for="fund" class="block text-sm font-medium text-gray-700 mb-2">Fund</label>
					<select
						id="fund"
						bind:value={selectedFundId}
						class="w-full py-3 px-4 border-2 border-gray-200 rounded-lg text-gray-700 focus:border-teal-600 focus:outline-none"
					>
						{#each church.funds as fund}
							<option value={fund.id}>{fund.name}</option>
						{/each}
					</select>
				</div>
			{/if}

			<!-- Frequency -->
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-2">Frequency</label>
				<div class="grid grid-cols-2 gap-3">
					<button
						on:click={() => frequency = 'one-time'}
						class="py-3 rounded-lg font-medium border-2 transition
							{frequency === 'one-time'
								? 'border-teal-600 bg-teal-50 text-teal-700'
								: 'border-gray-200 bg-white text-gray-700 hover:border-gray-300'}"
					>
						One-time
					</button>
					<button
						on:click={() => frequency = 'monthly'}
						class="py-3 rounded-lg font-medium border-2 transition
							{frequency === 'monthly'
								? 'border-teal-600 bg-teal-50 text-teal-700'
								: 'border-gray-200 bg-white text-gray-700 hover:border-gray-300'}"
					>
						Monthly
					</button>
				</div>
			</div>

			<!-- Donor Info -->
			<div class="space-y-3">
				<div>
					<label for="name" class="block text-sm font-medium text-gray-700 mb-1">Name <span class="text-gray-400">(optional)</span></label>
					<input
						id="name"
						type="text"
						bind:value={donorName}
						placeholder="Your name"
						class="w-full py-3 px-4 border-2 border-gray-200 rounded-lg focus:border-teal-600 focus:outline-none"
					/>
				</div>
				<div>
					<label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email <span class="text-gray-400">(for receipt)</span></label>
					<input
						id="email"
						type="email"
						bind:value={donorEmail}
						placeholder="you@email.com"
						class="w-full py-3 px-4 border-2 border-gray-200 rounded-lg focus:border-teal-600 focus:outline-none"
					/>
				</div>
			</div>

			<!-- Give Button -->
			<button
				on:click={handleGive}
				disabled={submitting || getAmountCents() < 100}
				class="w-full py-4 bg-teal-600 text-white text-lg font-bold rounded-xl hover:bg-teal-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{#if submitting}
					Processing...
				{:else}
					Give {getAmountCents() >= 100 ? formatDollars(getAmountCents()) : ''}{frequency === 'monthly' ? '/month' : ''}
				{/if}
			</button>

			<p class="text-xs text-gray-400 text-center">
				Secure payment powered by Stripe. Your card details are never stored on our servers.
			</p>
		</div>
	{/if}
</div>
