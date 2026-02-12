<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let funds: any[] = [];
	let people: any[] = [];
	let filteredPeople: any[] = [];
	let loading = false;
	let error = '';
	let personSearch = '';
	let showPeopleDropdown = false;
	let selectedPersonName = '';

	let formData = {
		person_id: null as string | null,
		fund_id: '',
		amount: '',
		payment_method: 'cash',
		memo: '',
		donated_at: new Date().toISOString().split('T')[0]
	};

	onMount(async () => {
		await Promise.all([loadFunds(), loadPeople()]);
	});

	async function loadFunds() {
		try {
			const response = await fetch('/api/giving/funds', {
				headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
			});
			if (response.ok) {
				funds = await response.json();
				if (funds.length > 0 && !formData.fund_id) {
					const defaultFund = funds.find((f: any) => f.is_default) || funds[0];
					formData.fund_id = defaultFund.id;
				}
			}
		} catch (err) { console.error('Failed to load funds:', err); }
	}

	async function loadPeople() {
		try {
			const response = await fetch('/api/people?per_page=500', {
				headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
			});
			if (response.ok) {
				const data = await response.json();
				people = data.people || [];
			}
		} catch (err) { console.error('Failed to load people:', err); }
	}

	function handlePersonSearch() {
		if (personSearch.length === 0) {
			filteredPeople = [];
			showPeopleDropdown = false;
			return;
		}
		const q = personSearch.toLowerCase();
		filteredPeople = people.filter((p: any) =>
			`${p.first_name} ${p.last_name}`.toLowerCase().includes(q)
		).slice(0, 10);
		showPeopleDropdown = filteredPeople.length > 0;
	}

	function selectPerson(person: any) {
		formData.person_id = person.id;
		selectedPersonName = `${person.first_name} ${person.last_name}`;
		personSearch = selectedPersonName;
		showPeopleDropdown = false;
	}

	function clearPerson() {
		formData.person_id = null;
		selectedPersonName = '';
		personSearch = '';
		showPeopleDropdown = false;
	}

	async function submitDonation() {
		error = '';
		loading = true;

		try {
			// Convert amount to cents
			const amountCents = Math.round(parseFloat(formData.amount) * 100);
			
			if (isNaN(amountCents) || amountCents <= 0) {
				error = 'Please enter a valid amount';
				loading = false;
				return;
			}

			if (!formData.fund_id) {
				error = 'Please select a fund';
				loading = false;
				return;
			}

			const payload = {
				person_id: formData.person_id || null,
				fund_id: formData.fund_id,
				amount_cents: amountCents,
				payment_method: formData.payment_method,
				memo: formData.memo || '',
				donated_at: new Date(formData.donated_at).toISOString()
			};

			const response = await fetch('/api/giving/donations', {
				method: 'POST',
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`,
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(payload)
			});

			if (response.ok) {
				goto('/dashboard/giving/donations');
			} else {
				const data = await response.text();
				error = data || 'Failed to record donation';
			}
		} catch (err) {
			error = 'An error occurred while recording the donation';
			console.error(err);
		} finally {
			loading = false;
		}
	}
</script>

<div class="p-6 max-w-2xl mx-auto">
	<div class="mb-6">
		<h1 class="text-3xl font-bold text-primary">Record Donation</h1>
		<p class="text-secondary mt-1">Manually record a cash or check donation</p>
	</div>

	{#if error}
		<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg mb-6">
			{error}
		</div>
	{/if}

	<form on:submit|preventDefault={submitDonation} class="bg-surface rounded-lg shadow p-6 space-y-6 border border-custom">
		<!-- Donor -->
		<div class="relative">
			<label class="block text-sm font-medium text-primary mb-1">
				Donor (Optional)
			</label>
			<div class="flex gap-2">
				<input
					type="text"
					bind:value={personSearch}
					on:input={handlePersonSearch}
					on:focus={() => { if (personSearch) handlePersonSearch(); }}
					placeholder="Search by name..."
					class="flex-1 px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent"
				/>
				{#if formData.person_id}
					<button
						type="button"
						on:click={clearPerson}
						class="px-3 py-2 text-sm border border-custom text-secondary rounded-lg hover:bg-[var(--surface-hover)]"
					>
						✕
					</button>
				{/if}
			</div>
			{#if showPeopleDropdown}
				<div class="absolute z-10 w-full mt-1 bg-surface border border-custom rounded-lg shadow-lg max-h-48 overflow-y-auto">
					{#each filteredPeople as person}
						<button
							type="button"
							on:click={() => selectPerson(person)}
							class="w-full px-3 py-2 text-left text-primary hover:bg-[var(--surface-hover)] text-sm"
						>
							{person.first_name} {person.last_name}
							{#if person.email}
								<span class="text-secondary ml-2">({person.email})</span>
							{/if}
						</button>
					{/each}
				</div>
			{/if}
			<p class="text-sm text-secondary mt-1">
				{#if formData.person_id}
					Selected: <strong>{selectedPersonName}</strong>
				{:else}
					Leave empty for anonymous donation
				{/if}
			</p>
		</div>

		<!-- Fund -->
		<div>
			<label class="block text-sm font-medium text-primary mb-1">
				Fund <span class="text-red-500">*</span>
			</label>
			<select
				bind:value={formData.fund_id}
				required
				class="w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
			>
				<option value="">Select a fund</option>
				{#each funds as fund}
					{#if fund.is_active}
						<option value={fund.id}>{fund.name}</option>
					{/if}
				{/each}
			</select>
		</div>

		<!-- Amount -->
		<div>
			<label class="block text-sm font-medium text-primary mb-1">
				Amount <span class="text-red-500">*</span>
			</label>
			<div class="relative">
				<span class="absolute left-3 top-2 text-secondary">$</span>
				<input
					type="number"
					step="0.01"
					min="0.01"
					bind:value={formData.amount}
					required
					placeholder="0.00"
					class="w-full pl-7 pr-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
				/>
			</div>
		</div>

		<!-- Payment Method -->
		<div>
			<label class="block text-sm font-medium text-primary mb-1">
				Payment Method
			</label>
			<select
				bind:value={formData.payment_method}
				class="w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
			>
				<option value="cash">Cash</option>
				<option value="check">Check</option>
				<option value="card">Card</option>
				<option value="ach">ACH</option>
			</select>
		</div>

		<!-- Date -->
		<div>
			<label class="block text-sm font-medium text-primary mb-1">
				Donation Date <span class="text-red-500">*</span>
			</label>
			<input
				type="date"
				bind:value={formData.donated_at}
				required
				class="w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
			/>
		</div>

		<!-- Memo -->
		<div>
			<label class="block text-sm font-medium text-primary mb-1">
				Memo (Optional)
			</label>
			<textarea
				bind:value={formData.memo}
				rows="3"
				placeholder="Add any notes about this donation..."
				class="w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
			></textarea>
		</div>

		<!-- Actions -->
		<div class="flex gap-4 pt-4">
			<button
				type="submit"
				disabled={loading}
				class="flex-1 px-4 py-2 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 transition disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{loading ? 'Recording...' : 'Record Donation'}
			</button>
			<a
				href="/dashboard/giving/donations"
				class="flex-1 px-4 py-2 border border-custom text-primary rounded-lg hover:bg-[var(--surface-hover)] transition text-center"
			>
				Cancel
			</a>
		</div>
	</form>
</div>
