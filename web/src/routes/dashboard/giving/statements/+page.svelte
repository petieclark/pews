<script lang="ts">
	import { onMount } from 'svelte';
	import { FileText } from 'lucide-svelte';

	let people: any[] = [];
	let loading = false;
	let selectedPerson = '';
	let selectedYear = new Date().getFullYear();
	let generatingPDF = false;

	const currentYear = new Date().getFullYear();
	const years = Array.from({ length: 5 }, (_, i) => currentYear - i);

	onMount(async () => {
		await loadPeople();
	});

	async function loadPeople() {
		try {
			const response = await fetch('/api/people?per_page=200', {
				headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
			});
			if (response.ok) {
				const data = await response.json();
				people = data.people || [];
			}
		} catch (error) { console.error('Failed to load people:', error); }
	}

	async function downloadPDF() {
		if (!selectedPerson) {
			alert('Please select a person');
			return;
		}

		generatingPDF = true;
		try {
			const response = await fetch(`/api/giving/statements/${selectedYear}/${selectedPerson}`, {
				headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
			});

			if (response.ok) {
				const blob = await response.blob();
				const url = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = `giving-statement-${selectedYear}.pdf`;
				document.body.appendChild(a);
				a.click();
				document.body.removeChild(a);
				window.URL.revokeObjectURL(url);
			} else {
				const text = await response.text();
				alert(text || 'Failed to generate statement. This person may have no donations for this year.');
			}
		} catch (error) {
			console.error('Failed to generate statement:', error);
			alert('An error occurred');
		} finally {
			generatingPDF = false;
		}
	}

	function getPersonName(id: string): string {
		const p = people.find(p => p.id === id);
		return p ? `${p.first_name} ${p.last_name}` : '';
	}
</script>

<div class="p-6 max-w-4xl mx-auto">
	<div class="mb-6">
		<h1 class="text-3xl font-bold text-primary">Giving Statements</h1>
		<p class="text-secondary mt-1">Generate annual giving statements for tax purposes</p>
	</div>

	<div class="bg-surface rounded-lg shadow p-6 mb-6 border border-custom">
		<h2 class="text-xl font-semibold text-primary mb-4">Generate PDF Statement</h2>
		
		<div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
			<div class="md:col-span-2">
				<label class="block text-sm font-medium text-primary mb-1">
					Select Person <span class="text-red-500">*</span>
				</label>
				<select
					bind:value={selectedPerson}
					class="w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent"
				>
					<option value="">Choose a donor...</option>
					{#each people as person}
						<option value={person.id}>{person.first_name} {person.last_name}</option>
					{/each}
				</select>
			</div>

			<div>
				<label class="block text-sm font-medium text-primary mb-1">Year</label>
				<select
					bind:value={selectedYear}
					class="w-full px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent"
				>
					{#each years as year}
						<option value={year}>{year}</option>
					{/each}
				</select>
			</div>
		</div>

		<button
			on:click={downloadPDF}
			disabled={generatingPDF || !selectedPerson}
			class="w-full px-4 py-3 bg-[var(--teal)] text-white rounded-lg hover:opacity-90 transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
		>
			{#if generatingPDF}
				<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
				Generating PDF...
			{:else}
				<FileText size={14} class="inline" /> Download PDF Statement
			{/if}
		</button>
	</div>

	<div class="bg-surface rounded-lg shadow p-6 border border-custom">
		<h3 class="font-semibold text-primary mb-3"><FileText size={16} class="inline" /> About Giving Statements</h3>
		<ul class="text-sm text-secondary space-y-2">
			<li>• Statements include all completed donations for a calendar year</li>
			<li>• PDF includes church info, donor info, and itemized contributions</li>
			<li>• Statements are required for tax-deductible contributions over $250</li>
			<li>• Generate by January 31st for the previous tax year</li>
		</ul>
	</div>
</div>
