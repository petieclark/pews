<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let cards = [];
	let loading = true;
	let error = '';
	let processing = {};
	let showProcessed = false;

	onMount(async () => {
		await loadCards();
	});

	async function loadCards() {
		try {
			loading = true;
			cards = await api('/api/communication/cards');
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	$: filteredCards = showProcessed ? cards : cards.filter(c => !c.processed);

	async function processCard(card) {
		processing[card.id] = true;
		try {
			// First create person from card data
			const person = await api('/api/people', {
				method: 'POST',
				body: JSON.stringify({
					first_name: card.first_name,
					last_name: card.last_name || '',
					email: card.email || '',
					phone: card.phone || '',
					membership_status: 'visitor',
					notes: [
						card.how_heard ? `How heard: ${card.how_heard}` : '',
						card.interested_in ? `Interested in: ${card.interested_in}` : '',
						card.prayer_request ? `Prayer request: ${card.prayer_request}` : '',
						card.is_first_visit ? 'First-time visitor' : ''
					].filter(Boolean).join('\n')
				})
			});

			// Mark card as processed
			await api(`/api/communication/cards/${card.id}/process`, {
				method: 'POST',
				body: JSON.stringify({ person_id: person.id })
			});

			await loadCards();
		} catch (err) {
			error = err.message;
		} finally {
			processing[card.id] = false;
		}
	}
</script>

<div>
	<div class="flex items-center justify-between mb-6">
		<div>
			<a href="/dashboard/communication" class="text-sm font-medium" style="color: var(--teal)">← Communication</a>
			<h1 class="text-3xl font-bold mt-1" style="color: var(--text)">Connection Cards</h1>
		</div>
		<label class="flex items-center gap-2 cursor-pointer">
			<input bind:checked={showProcessed} type="checkbox" class="rounded" />
			<span class="text-sm" style="color: var(--text-secondary)">Show processed</span>
		</label>
	</div>

	{#if error}
		<div class="mb-4 px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">{error}</div>
	{/if}

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2" style="border-color: var(--teal)"></div>
		</div>
	{:else if filteredCards.length === 0}
		<div class="rounded-lg shadow border p-12 text-center" style="background: var(--surface); border-color: var(--border)">
			<div class="text-5xl mb-4">💳</div>
			<h2 class="text-xl font-semibold mb-2" style="color: var(--text)">
				{showProcessed ? 'No connection cards' : 'All caught up!'}
			</h2>
			<p style="color: var(--text-secondary)">
				{showProcessed ? 'Connection cards submitted by visitors will appear here.' : 'No unprocessed cards. Great job!'}
			</p>
		</div>
	{:else}
		<div class="space-y-4">
			{#each filteredCards as card}
				<div class="rounded-lg shadow border p-6" style="background: var(--surface); border-color: var(--border)">
					<div class="flex items-start justify-between">
						<div class="flex-1">
							<div class="flex items-center gap-3 mb-2">
								<h3 class="text-lg font-semibold" style="color: var(--text)">
									{card.first_name} {card.last_name || ''}
								</h3>
								{#if card.is_first_visit}
									<span class="px-2 py-0.5 rounded-full text-xs font-medium" style="background: #dbeafe; color: #1e40af">First Visit</span>
								{/if}
								{#if card.processed}
									<span class="px-2 py-0.5 rounded-full text-xs font-medium" style="background: #d1fae5; color: #065f46">Processed</span>
								{/if}
							</div>
							<div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm" style="color: var(--text-secondary)">
								{#if card.email}
									<div>📧 {card.email}</div>
								{/if}
								{#if card.phone}
									<div>📱 {card.phone}</div>
								{/if}
								{#if card.how_heard}
									<div>🔍 How heard: {card.how_heard}</div>
								{/if}
								{#if card.interested_in}
									<div>⭐ Interested in: {card.interested_in}</div>
								{/if}
							</div>
							{#if card.prayer_request}
								<div class="mt-3 p-3 rounded-lg text-sm" style="background: var(--bg); color: var(--text)">
									<span class="font-medium">🙏 Prayer Request:</span> {card.prayer_request}
								</div>
							{/if}
							<div class="mt-2 text-xs" style="color: var(--text-secondary)">
								Submitted {new Date(card.submitted_at).toLocaleString()}
							</div>
						</div>
						{#if !card.processed}
							<button
								on:click={() => processCard(card)}
								disabled={processing[card.id]}
								class="ml-4 px-4 py-2 rounded-lg font-medium text-sm whitespace-nowrap"
								style="background: var(--teal); color: white; opacity: {processing[card.id] ? 0.5 : 1}"
							>
								{processing[card.id] ? 'Processing...' : '✓ Process & Create Person'}
							</button>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
