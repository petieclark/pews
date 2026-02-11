<script lang="ts">
	import { onMount } from 'svelte';

	let people = [];
	let loading = false;
	let selectedPerson = '';
	let selectedYear = new Date().getFullYear();
	let generatingFor = null;

	const currentYear = new Date().getFullYear();
	const years = Array.from({ length: 5 }, (_, i) => currentYear - i);

	onMount(async () => {
		await loadPeople();
	});

	async function loadPeople() {
		try {
			const response = await fetch('/api/people?per_page=200', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});
			if (response.ok) {
				const data = await response.json();
				people = data.people || [];
			}
		} catch (error) {
			console.error('Failed to load people:', error);
		}
	}

	async function generateStatement() {
		if (!selectedPerson) {
			alert('Please select a person');
			return;
		}

		generatingFor = selectedPerson;
		loading = true;

		try {
			// Call the PDF generation endpoint
			const response = await fetch(`/api/giving/statements/${selectedYear}/${selectedPerson}`, {
				method: 'GET',
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});

			if (response.ok) {
				// Get the PDF blob
				const blob = await response.blob();
				
				// Create a download link
				const url = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = `tax-statement-${selectedYear}-${selectedPerson.substring(0, 8)}.pdf`;
				document.body.appendChild(a);
				a.click();
				window.URL.revokeObjectURL(url);
				document.body.removeChild(a);
				
				alert('Statement downloaded successfully!');
			} else if (response.status === 404) {
				alert('No donations found for this person in the selected year.');
			} else {
				alert('Failed to generate statement');
			}
		} catch (error) {
			console.error('Failed to generate statement:', error);
			alert('An error occurred');
		} finally {
			loading = false;
			generatingFor = null;
		}
	}
	
	async function generateBatchStatements() {
		loading = true;
		
		try {
			const response = await fetch(`/api/giving/statements/${selectedYear}`, {
				method: 'GET',
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});

			if (response.ok) {
				const data = await response.json();
				alert(`Found ${data.donor_count} donors with contributions in ${selectedYear}. Use the individual statement generator to create PDFs for each donor.`);
			} else if (response.status === 404) {
				alert('No donations found for the selected year.');
			} else {
				alert('Failed to get donor list');
			}
		} catch (error) {
			console.error('Failed to generate batch statements:', error);
			alert('An error occurred');
		} finally {
			loading = false;
		}
	}

	function formatCurrency(cents: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(cents / 100);
	}
</script>

<div class="p-6 max-w-4xl mx-auto">
	<div class="mb-6">
		<h1 class="text-3xl font-bold text-[#1B3A4B]">Giving Statements</h1>
		<p class="text-gray-600 mt-1">Generate annual giving statements for tax purposes</p>
	</div>

	<div class="bg-white rounded-lg shadow p-6 mb-6">
		<h2 class="text-xl font-semibold text-[#1B3A4B] mb-4">Generate Statement</h2>
		
		<div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
			<div class="md:col-span-2">
				<label class="block text-sm font-medium text-gray-700 mb-1">
					Select Person <span class="text-red-500">*</span>
				</label>
				<select
					bind:value={selectedPerson}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent"
				>
					<option value="">Choose a donor...</option>
					{#each people as person}
						<option value={person.id}>{person.first_name} {person.last_name}</option>
					{/each}
				</select>
			</div>

			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">
					Year
				</label>
				<select
					bind:value={selectedYear}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent"
				>
					{#each years as year}
						<option value={year}>{year}</option>
					{/each}
				</select>
			</div>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			<button
				on:click={generateStatement}
				disabled={loading || !selectedPerson}
				class="px-4 py-2 bg-[#4A8B8C] text-white rounded-lg hover:bg-[#3d7576] transition disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{loading ? 'Generating...' : 'Download PDF Statement'}
			</button>
			
			<button
				on:click={generateBatchStatements}
				disabled={loading}
				class="px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
			>
				Check All Donors for {selectedYear}
			</button>
		</div>
	</div>

	<div class="bg-blue-50 border border-blue-200 rounded-lg p-6">
		<h3 class="font-semibold text-blue-900 mb-2">📄 About Giving Statements</h3>
		<ul class="text-sm text-blue-800 space-y-2">
			<li>• Giving statements show total donations for a calendar year</li>
			<li>• Only completed donations are included (pending/failed excluded)</li>
			<li>• Statements are required for tax-deductible contributions</li>
			<li>• Generate statements by January 31st for the previous tax year</li>
			<li>• PDFs include donor info, donation details, total, and tax disclaimer</li>
			<li>• Use "Check All Donors" to see how many people gave in a year</li>
		</ul>
	</div>
</div>
