<script>
	import { api } from '$lib/api';
	import { onMount } from 'svelte';

	let importType = 'people';
	let importSource = 'pco';
	let file = null;
	let fileInput;
	let uploading = false;
	let error = '';
	let success = '';
	let result = null;
	let preview = null;
	let isDragging = false;

	const importTypes = [
		{ value: 'people', label: 'People / Members' },
		{ value: 'groups', label: 'Groups' },
		{ value: 'songs', label: 'Songs' },
		{ value: 'giving', label: 'Donations' }
	];

	const importSources = [
		{ value: 'pco', label: 'Planning Center Online (PCO)' },
		{ value: 'generic', label: 'Generic CSV' }
	];

	function handleDragOver(e) {
		e.preventDefault();
		isDragging = true;
	}

	function handleDragLeave(e) {
		e.preventDefault();
		isDragging = false;
	}

	function handleDrop(e) {
		e.preventDefault();
		isDragging = false;
		const files = e.dataTransfer.files;
		if (files.length > 0) {
			handleFileSelect({ target: { files } });
		}
	}

	async function handleFileSelect(e) {
		const files = e.target.files;
		if (files.length === 0) return;

		file = files[0];
		error = '';
		success = '';
		result = null;

		// Show preview of first 5 rows
		if (file.type === 'text/csv' || file.name.endsWith('.csv')) {
			const text = await file.text();
			const lines = text.split('\n').slice(0, 6); // Header + 5 rows
			preview = lines.join('\n');
		}
	}

	function clearFile() {
		file = null;
		preview = null;
		result = null;
		error = '';
		success = '';
		if (fileInput) fileInput.value = '';
	}

	async function uploadFile() {
		if (!file) {
			error = 'Please select a file to upload';
			return;
		}

		uploading = true;
		error = '';
		success = '';
		result = null;

		try {
			const endpoint = importSource === 'pco' ? '/api/import/pco' : `/api/import/${importType}`;
			
			const response = await fetch(`${import.meta.env.VITE_API_URL}${endpoint}`, {
				method: 'POST',
				headers: {
					'Content-Type': 'text/csv',
					Authorization: `Bearer ${localStorage.getItem('token')}`
				},
				body: file
			});

			if (!response.ok) {
				const errorData = await response.text();
				throw new Error(errorData || 'Failed to import data');
			}

			result = await response.json();
			success = `Successfully imported ${result.created} records!`;
			
			if (result.skipped > 0) {
				success += ` (${result.skipped} skipped as duplicates)`;
			}

			// Clear file after successful import
			setTimeout(() => {
				clearFile();
			}, 3000);
		} catch (err) {
			error = err.message;
		} finally {
			uploading = false;
		}
	}

	function triggerFileInput() {
		fileInput.click();
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="mb-6">
		<a href="/dashboard/settings" class="text-[var(--teal)] hover:underline mb-4 inline-block">
			← Back to Settings
		</a>
		<h1 class="text-3xl font-bold text-primary">Import Data</h1>
		<p class="text-secondary mt-2">
			Import people, groups, songs, or donation data from CSV files
		</p>
	</div>

	<!-- Import Configuration -->
	<div class="bg-surface rounded-lg shadow-md p-6 mb-6 border border-custom">
		<h2 class="text-xl font-semibold text-primary mb-4">Import Configuration</h2>

		<div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
			<div>
				<label for="importType" class="block text-sm font-medium text-primary mb-1">
					Import Type
				</label>
				<select
					id="importType"
					bind:value={importType}
					on:change={clearFile}
					class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
				>
					{#each importTypes as type}
						<option value={type.value}>{type.label}</option>
					{/each}
				</select>
			</div>

			<div>
				<label for="importSource" class="block text-sm font-medium text-primary mb-1">
					Import Source
				</label>
				<select
					id="importSource"
					bind:value={importSource}
					on:change={clearFile}
					class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
				>
					{#each importSources as source}
						<option value={source.value}>{source.label}</option>
					{/each}
				</select>
			</div>
		</div>

		<!-- Format instructions -->
		{#if importSource === 'pco' && importType === 'people'}
			<div class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 mb-4">
				<h3 class="font-semibold text-primary mb-2">PCO People Export Format</h3>
				<p class="text-sm text-secondary mb-2">
					Expected columns: <code class="bg-[var(--surface-hover)] px-1 rounded">First Name, Last Name, Email, Phone, Member Status, Tags, Birthdate, Address, City, State, Zip</code>
				</p>
				<p class="text-xs text-secondary">
					Export your people data from Planning Center Online and upload the CSV file here.
				</p>
			</div>
		{/if}
	</div>

	<!-- File Upload -->
	<div class="bg-surface rounded-lg shadow-md p-6 mb-6 border border-custom">
		<h2 class="text-xl font-semibold text-primary mb-4">Upload CSV File</h2>

		<div
			class="border-2 border-dashed rounded-lg p-8 text-center transition-colors {isDragging
				? 'border-[var(--teal)] bg-[var(--teal)]/5'
				: 'border-custom'}"
			on:dragover={handleDragOver}
			on:dragleave={handleDragLeave}
			on:drop={handleDrop}
			role="button"
			tabindex="0"
			on:click={triggerFileInput}
			on:keypress={(e) => e.key === 'Enter' && triggerFileInput()}
		>
			<input
				bind:this={fileInput}
				type="file"
				accept=".csv"
				on:change={handleFileSelect}
				class="hidden"
			/>

			{#if file}
				<div class="flex items-center justify-center mb-4">
					<svg
						class="w-12 h-12 text-[var(--teal)]"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
						/>
					</svg>
				</div>
				<p class="text-primary font-medium mb-2">{file.name}</p>
				<p class="text-sm text-secondary mb-4">
					{(file.size / 1024).toFixed(2)} KB
				</p>
				<button
					type="button"
					on:click|stopPropagation={clearFile}
					class="text-red-600 dark:text-red-400 text-sm hover:underline"
				>
					Remove file
				</button>
			{:else}
				<div class="flex items-center justify-center mb-4">
					<svg
						class="w-12 h-12 text-secondary"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
						/>
					</svg>
				</div>
				<p class="text-primary font-medium mb-2">
					Drag and drop your CSV file here, or click to browse
				</p>
				<p class="text-sm text-secondary">Maximum file size: 10 MB</p>
			{/if}
		</div>

		{#if preview}
			<div class="mt-4">
				<h3 class="text-sm font-semibold text-primary mb-2">Preview (first 5 rows):</h3>
				<pre
					class="bg-[var(--surface-hover)] border border-custom rounded-lg p-4 text-xs overflow-x-auto text-primary">{preview}</pre>
			</div>
		{/if}

		{#if error}
			<div class="mt-4 bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg">
				{error}
			</div>
		{/if}

		{#if success}
			<div class="mt-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-green-800 dark:text-green-200 px-4 py-3 rounded-lg">
				{success}
			</div>
		{/if}

		{#if result}
			<div class="mt-4 bg-[var(--surface-hover)] border border-custom rounded-lg p-4">
				<h3 class="font-semibold text-primary mb-2">Import Results</h3>
				<div class="grid grid-cols-3 gap-4 text-center">
					<div>
						<div class="text-2xl font-bold text-green-600 dark:text-green-400">
							{result.created}
						</div>
						<div class="text-xs text-secondary">Created</div>
					</div>
					<div>
						<div class="text-2xl font-bold text-yellow-600 dark:text-yellow-400">
							{result.skipped}
						</div>
						<div class="text-xs text-secondary">Skipped</div>
					</div>
					<div>
						<div class="text-2xl font-bold text-red-600 dark:text-red-400">
							{result.errors?.length || 0}
						</div>
						<div class="text-xs text-secondary">Errors</div>
					</div>
				</div>

				{#if result.errors && result.errors.length > 0}
					<div class="mt-4">
						<h4 class="text-sm font-semibold text-primary mb-2">Errors:</h4>
						<ul class="list-disc list-inside text-sm text-secondary space-y-1">
							{#each result.errors as err}
								<li>{err}</li>
							{/each}
						</ul>
					</div>
				{/if}
			</div>
		{/if}

		<div class="mt-6 flex gap-4">
			<button
				type="button"
				on:click={uploadFile}
				disabled={!file || uploading}
				class="flex-1 bg-[var(--teal)] text-white py-2 px-6 rounded-lg font-medium hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{uploading ? 'Importing...' : 'Import Data'}
			</button>
		</div>
	</div>

	<!-- Help Section -->
	<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
		<h2 class="text-xl font-semibold text-primary mb-4">Import Help</h2>
		
		<div class="space-y-4 text-sm text-secondary">
			<div>
				<h3 class="font-semibold text-primary mb-1">People Import</h3>
				<p>Import church members with names, contact info, addresses, and membership status. Duplicates (by email) are automatically skipped.</p>
			</div>
			
			<div>
				<h3 class="font-semibold text-primary mb-1">Groups Import</h3>
				<p>Import small groups, ministries, or teams with meeting details and member lists.</p>
			</div>
			
			<div>
				<h3 class="font-semibold text-primary mb-1">Songs Import</h3>
				<p>Import worship songs with lyrics, CCLI numbers, keys, and tempo information.</p>
			</div>
			
			<div>
				<h3 class="font-semibold text-primary mb-1">Giving Import</h3>
				<p>Import historical donation records. Requires donor email, fund name, and amount.</p>
			</div>
		</div>
	</div>
</div>

<style>
	code {
		font-size: 0.875rem;
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
	}
</style>
