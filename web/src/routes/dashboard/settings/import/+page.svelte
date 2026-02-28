<script>
	import { onMount } from 'svelte';
	import { ClipboardList } from 'lucide-svelte';

	// State
	let step = 1; // 1: Select type, 2: Upload/Preview, 3: Column mapping, 4: Import, 5: Results
	let importType = 'people'; // 'people' or 'songs'
	let file = null;
	let fileInput;
	let preview = null;
	let headers = [];
	let previewRows = [];
	let columnMappings = {};
	let suggestedMappings = {};
	let updateMode = 'skip'; // 'skip' or 'update'
	let uploading = false;
	let result = null;
	let error = '';

	// PCO Column Mapping templates
	const peopleFieldMap = {
		first_name: ['first_name', 'firstname', 'first name'],
		last_name: ['last_name', 'lastname', 'last name'],
		email: ['email', 'email_address', 'email address'],
		phone: ['phone', 'phone_number', 'mobile', 'cell', 'mobile_phone', 'phone number'],
		address_line1: ['street', 'address', 'address_line_1', 'address line 1'],
		city: ['city'],
		state: ['state'],
		zip: ['zip', 'postal_code', 'zipcode', 'postal code'],
		birthdate: ['birthdate', 'birthday', 'date_of_birth', 'dob', 'birth date'],
		gender: ['gender', 'sex'],
		membership_status: ['membership', 'membership_type', 'membership_status', 'status', 'member status']
	};

	const songsFieldMap = {
		title: ['title', 'song_title', 'name', 'song title'],
		artist: ['author', 'artist', 'writer'],
		ccli_number: ['ccli', 'ccli_#', 'ccli_number', 'ccli #', 'ccli number'],
		default_key: ['key', 'default_key', 'default key'],
		tempo: ['bpm', 'tempo'],
		tags: ['themes', 'tags', 'categories']
	};

	function normalizeHeader(header) {
		return header.toLowerCase().trim().replace(/[_\s-]+/g, '_');
	}

	function suggestMapping(headers, fieldMap) {
		const mappings = {};
		const normalizedHeaders = headers.map(h => normalizeHeader(h));

		for (const [field, patterns] of Object.entries(fieldMap)) {
			for (const pattern of patterns) {
				const normalized = normalizeHeader(pattern);
				const idx = normalizedHeaders.indexOf(normalized);
				if (idx !== -1) {
					mappings[headers[idx]] = field;
					break;
				}
			}
		}

		return mappings;
	}

	async function handleFileSelect(e) {
		const files = e.target.files;
		if (files.length === 0) return;

		file = files[0];
		error = '';
		result = null;

		// Parse CSV preview
		if (file.type === 'text/csv' || file.name.endsWith('.csv')) {
			const text = await file.text();
			const lines = text.trim().split('\n');

			if (lines.length > 0) {
				// Parse headers
				headers = parseCSVLine(lines[0]);

				// Parse first 5 data rows
				previewRows = [];
				for (let i = 1; i < Math.min(6, lines.length); i++) {
					if (lines[i].trim()) {
						previewRows.push(parseCSVLine(lines[i]));
					}
				}

				// Auto-suggest column mappings
				const fieldMap = importType === 'people' ? peopleFieldMap : songsFieldMap;
				suggestedMappings = suggestMapping(headers, fieldMap);
				columnMappings = { ...suggestedMappings };

				step = 2;
			}
		}
	}

	function parseCSVLine(line) {
		const result = [];
		let current = '';
		let inQuotes = false;

		for (let i = 0; i < line.length; i++) {
			const char = line[i];
			const nextChar = line[i + 1];

			if (char === '"') {
				if (inQuotes && nextChar === '"') {
					current += '"';
					i++;
				} else {
					inQuotes = !inQuotes;
				}
			} else if (char === ',' && !inQuotes) {
				result.push(current.trim());
				current = '';
			} else {
				current += char;
			}
		}

		result.push(current.trim());
		return result;
	}

	function clearFile() {
		file = null;
		preview = null;
		result = null;
		error = '';
		headers = [];
		previewRows = [];
		columnMappings = {};
		step = 1;
		if (fileInput) fileInput.value = '';
	}

	async function performImport() {
		if (!file) {
			error = 'No file selected';
			return;
		}

		uploading = true;
		error = '';
		result = null;
		step = 4;

		try {
			const formData = new FormData();
			formData.append('file', file);
			formData.append('update_mode', updateMode);

			const endpoint =
				importType === 'people' ? '/api/import/pco/people' : '/api/import/pco/songs';

			const response = await fetch(`${import.meta.env.VITE_API_URL}${endpoint}`, {
				method: 'POST',
				headers: {
					Authorization: `Bearer ${localStorage.getItem('token')}`
				},
				body: formData
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText || 'Failed to import data');
			}

			result = await response.json();
			step = 5;
		} catch (err) {
			error = err.message;
			step = 2;
		} finally {
			uploading = false;
		}
	}

	function startOver() {
		clearFile();
		step = 1;
	}
</script>

<div class="max-w-5xl mx-auto py-8 px-4">
	<div class="mb-6">
		<a href="/dashboard/settings" class="text-[var(--teal)] hover:underline mb-4 inline-block">
			← Back to Settings
		</a>
		<h1 class="text-3xl font-bold" style="color: var(--text-primary)">Import from Planning Center</h1>
		<p class="mt-2" style="color: var(--text-secondary)">
			Import your people or songs data from PCO. All custom fields are preserved.
		</p>
	</div>

	<!-- Progress Steps -->
	<div class="mb-8">
		<div class="flex items-center justify-between">
			{#each ['Select Type', 'Upload File', 'Review', 'Import', 'Results'] as label, i}
				<div class="flex flex-col items-center flex-1">
					<div
						class="w-10 h-10 rounded-full flex items-center justify-center text-sm font-semibold mb-2 {step >
						i + 1
							? 'bg-[var(--teal)] text-white'
							: step === i + 1
								? 'bg-[var(--teal)] text-white'
								: 'bg-[var(--surface)] text-[var(--text-secondary)] border-2'}"
						style="border-color: var(--border)"
					>
						{i + 1}
					</div>
					<span
						class="text-xs text-center {step >= i + 1
							? 'text-[var(--text-primary)] font-medium'
							: 'text-[var(--text-secondary)]'}"
					>
						{label}
					</span>
				</div>
				{#if i < 4}
					<div
						class="flex-1 h-1 mx-2 rounded {step > i + 1
							? 'bg-[var(--teal)]'
							: 'bg-[var(--border)]'}"
					></div>
				{/if}
			{/each}
		</div>
	</div>

	<!-- Step 1: Select Import Type -->
	{#if step === 1}
		<div class="rounded-lg shadow-md p-8 border" style="background-color: var(--surface); border-color: var(--border)">
			<h2 class="text-2xl font-semibold mb-6" style="color: var(--text-primary)">
				What would you like to import?
			</h2>

			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<button
					type="button"
					on:click={() => (importType = 'people')}
					class="p-6 rounded-lg border-2 transition-all text-left {importType === 'people'
						? 'border-[var(--teal)] bg-[var(--teal)]/5'
						: 'hover:border-[var(--teal)]/30'}"
					style="border-color: {importType === 'people'
						? 'var(--teal)'
						: 'var(--border)'}; background-color: {importType === 'people'
						? 'rgba(var(--teal-rgb), 0.05)'
						: 'var(--bg)'}"
				>
					<div class="flex items-start">
						<svg class="w-8 h-8 mr-4 text-[var(--teal)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
						</svg>
						<div>
							<h3 class="text-lg font-semibold mb-2" style="color: var(--text-primary)">People / Members</h3>
							<p class="text-sm" style="color: var(--text-secondary)">
								Import your congregation with names, emails, addresses, birthdays, and all custom fields
							</p>
						</div>
					</div>
				</button>

				<button
					type="button"
					on:click={() => (importType = 'songs')}
					class="p-6 rounded-lg border-2 transition-all text-left {importType === 'songs'
						? 'border-[var(--teal)] bg-[var(--teal)]/5'
						: 'hover:border-[var(--teal)]/30'}"
					style="border-color: {importType === 'songs'
						? 'var(--teal)'
						: 'var(--border)'}; background-color: {importType === 'songs'
						? 'rgba(var(--teal-rgb), 0.05)'
						: 'var(--bg)'}"
				>
					<div class="flex items-start">
						<svg class="w-8 h-8 mr-4 text-[var(--teal)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
						</svg>
						<div>
							<h3 class="text-lg font-semibold mb-2" style="color: var(--text-primary)">Songs</h3>
							<p class="text-sm" style="color: var(--text-secondary)">
								Import worship songs with CCLI numbers, keys, tempos, and metadata
							</p>
						</div>
					</div>
				</button>
			</div>

			<div class="mt-8">
				<button
					type="button"
					on:click={() => (step = 2)}
					class="w-full bg-[var(--teal)] text-white py-3 px-6 rounded-lg font-medium hover:opacity-90 transition-opacity"
				>
					Continue
				</button>
			</div>
		</div>
	{/if}

	<!-- Step 2: Upload and Preview -->
	{#if step === 2}
		<div class="rounded-lg shadow-md p-8 border" style="background-color: var(--surface); border-color: var(--border)">
			<h2 class="text-2xl font-semibold mb-4" style="color: var(--text-primary)">
				Upload PCO {importType === 'people' ? 'People' : 'Songs'} CSV
			</h2>

			<div class="mb-6 p-4 rounded-lg" style="background-color: var(--bg); border: 1px solid var(--border)">
				<h3 class="font-semibold mb-2 flex items-center gap-1" style="color: var(--text-primary)"><ClipboardList size={16} /> How to export from PCO:</h3>
				<ol class="text-sm space-y-1 list-decimal list-inside" style="color: var(--text-secondary)">
					{#if importType === 'people'}
						<li>Log into Planning Center Online</li>
						<li>Go to <strong>People</strong> → <strong>All People</strong></li>
						<li>Click <strong>Export</strong> and select <strong>CSV</strong></li>
						<li>Include all fields you want to preserve (custom fields will be saved automatically)</li>
					{:else}
						<li>Log into Planning Center Online</li>
						<li>Go to <strong>Services</strong> → <strong>Songs</strong></li>
						<li>Click <strong>Export</strong> and select <strong>CSV</strong></li>
						<li>Include Title, Author, CCLI, Key, BPM, Themes/Tags</li>
					{/if}
				</ol>
			</div>

			{#if !file}
				<div
					class="border-2 border-dashed rounded-lg p-12 text-center cursor-pointer transition-colors hover:border-[var(--teal)]"
					style="border-color: var(--border)"
					on:click={() => fileInput.click()}
					on:dragover|preventDefault
					on:drop|preventDefault={(e) => {
						const files = e.dataTransfer.files;
						if (files.length > 0) {
							handleFileSelect({ target: { files } });
						}
					}}
					role="button"
					tabindex="0"
					on:keypress={(e) => e.key === 'Enter' && fileInput.click()}
				>
					<input
						bind:this={fileInput}
						type="file"
						accept=".csv"
						on:change={handleFileSelect}
						class="hidden"
					/>

					<svg class="w-16 h-16 mx-auto mb-4" style="color: var(--text-secondary)" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
					</svg>

					<p class="font-medium mb-2" style="color: var(--text-primary)">
						Drop your CSV file here or click to browse
					</p>
					<p class="text-sm" style="color: var(--text-secondary)">Maximum file size: 50 MB</p>
				</div>
			{:else}
				<div class="mb-6">
					<div class="flex items-center justify-between p-4 rounded-lg border" style="background-color: var(--bg); border-color: var(--border)">
						<div class="flex items-center">
							<svg class="w-8 h-8 mr-3 text-[var(--teal)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
							</svg>
							<div>
								<p class="font-medium" style="color: var(--text-primary)">{file.name}</p>
								<p class="text-sm" style="color: var(--text-secondary)">
									{(file.size / 1024).toFixed(2)} KB • {headers.length} columns • {previewRows.length} preview rows
								</p>
							</div>
						</div>
						<button
							type="button"
							on:click={clearFile}
							class="text-red-600 hover:text-red-700 font-medium text-sm"
						>
							Remove
						</button>
					</div>
				</div>

				<!-- Preview Table -->
				{#if headers.length > 0}
					<div class="mb-6">
						<h3 class="font-semibold mb-3" style="color: var(--text-primary)">Preview (first 5 rows)</h3>
						<div class="overflow-x-auto rounded-lg border" style="border-color: var(--border)">
							<table class="w-full text-sm">
								<thead style="background-color: var(--bg)">
									<tr>
										{#each headers as header}
											<th class="px-4 py-3 text-left font-medium" style="color: var(--text-primary); border-bottom: 1px solid var(--border)">
												{header}
											</th>
										{/each}
									</tr>
								</thead>
								<tbody style="background-color: var(--surface)">
									{#each previewRows as row, i}
										<tr style="border-bottom: 1px solid var(--border)">
											{#each row as cell}
												<td class="px-4 py-2" style="color: var(--text-secondary)">
													{cell || '—'}
												</td>
											{/each}
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</div>
				{/if}

				<!-- Duplicate Handling -->
				<div class="mb-6">
					<h3 class="font-semibold mb-3" style="color: var(--text-primary)">Duplicate Handling</h3>
					<div class="flex gap-4">
						<label class="flex items-center cursor-pointer">
							<input
								type="radio"
								name="updateMode"
								value="skip"
								bind:group={updateMode}
								class="mr-2"
							/>
							<span style="color: var(--text-primary)">Skip duplicates</span>
						</label>
						<label class="flex items-center cursor-pointer">
							<input
								type="radio"
								name="updateMode"
								value="update"
								bind:group={updateMode}
								class="mr-2"
							/>
							<span style="color: var(--text-primary)">Update existing records</span>
						</label>
					</div>
					<p class="text-xs mt-2" style="color: var(--text-secondary)">
						Duplicates are matched by {importType === 'people' ? 'email address' : 'CCLI number or title + artist'}
					</p>
				</div>

				{#if error}
					<div class="mb-6 px-4 py-3 rounded-lg border" style="background-color: rgba(239, 68, 68, 0.1); border-color: rgb(239, 68, 68); color: rgb(239, 68, 68)">
						{error}
					</div>
				{/if}

				<div class="flex gap-4">
					<button
						type="button"
						on:click={() => (step = 1)}
						class="flex-1 py-3 px-6 rounded-lg font-medium border transition-colors"
						style="border-color: var(--border); color: var(--text-primary); background-color: var(--bg)"
					>
						Back
					</button>
					<button
						type="button"
						on:click={performImport}
						class="flex-1 bg-[var(--teal)] text-white py-3 px-6 rounded-lg font-medium hover:opacity-90 transition-opacity"
					>
						Start Import
					</button>
				</div>
			{/if}
		</div>
	{/if}

	<!-- Step 4: Importing (Progress) -->
	{#if step === 4}
		<div class="rounded-lg shadow-md p-12 border text-center" style="background-color: var(--surface); border-color: var(--border)">
			<div class="animate-spin rounded-full h-16 w-16 border-b-4 mx-auto mb-6" style="border-color: var(--teal)"></div>
			<h2 class="text-2xl font-semibold mb-2" style="color: var(--text-primary)">
				Importing {importType === 'people' ? 'People' : 'Songs'}...
			</h2>
			<p style="color: var(--text-secondary)">Please wait while we process your CSV file</p>
		</div>
	{/if}

	<!-- Step 5: Results -->
	{#if step === 5 && result}
		<div class="rounded-lg shadow-md p-8 border" style="background-color: var(--surface); border-color: var(--border)">
			<div class="text-center mb-8">
				<div class="w-16 h-16 bg-green-100 dark:bg-green-900/20 rounded-full flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
				</div>
				<h2 class="text-2xl font-semibold mb-2" style="color: var(--text-primary)">Import Complete!</h2>
				<p style="color: var(--text-secondary)">Your {importType} data has been imported</p>
			</div>

			<!-- Stats Grid -->
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
				<div class="p-4 rounded-lg text-center" style="background-color: var(--bg); border: 1px solid var(--border)">
					<div class="text-3xl font-bold text-green-600 dark:text-green-400 mb-1">
						{result.imported || 0}
					</div>
					<div class="text-sm" style="color: var(--text-secondary)">Imported</div>
				</div>
				<div class="p-4 rounded-lg text-center" style="background-color: var(--bg); border: 1px solid var(--border)">
					<div class="text-3xl font-bold text-blue-600 dark:text-blue-400 mb-1">
						{result.updated || 0}
					</div>
					<div class="text-sm" style="color: var(--text-secondary)">Updated</div>
				</div>
				<div class="p-4 rounded-lg text-center" style="background-color: var(--bg); border: 1px solid var(--border)">
					<div class="text-3xl font-bold text-yellow-600 dark:text-yellow-400 mb-1">
						{result.skipped || 0}
					</div>
					<div class="text-sm" style="color: var(--text-secondary)">Skipped</div>
				</div>
				<div class="p-4 rounded-lg text-center" style="background-color: var(--bg); border: 1px solid var(--border)">
					<div class="text-3xl font-bold text-red-600 dark:text-red-400 mb-1">
						{result.errors?.length || 0}
					</div>
					<div class="text-sm" style="color: var(--text-secondary)">Errors</div>
				</div>
			</div>

			<!-- Errors List -->
			{#if result.errors && result.errors.length > 0}
				<div class="mb-8 p-4 rounded-lg border" style="background-color: rgba(239, 68, 68, 0.05); border-color: rgb(239, 68, 68)">
					<h3 class="font-semibold mb-3" style="color: rgb(239, 68, 68)">
						{result.errors.length} Error{result.errors.length === 1 ? '' : 's'}
					</h3>
					<ul class="text-sm space-y-1 max-h-64 overflow-y-auto">
						{#each result.errors as err}
							<li style="color: rgb(239, 68, 68)">{err}</li>
						{/each}
					</ul>
				</div>
			{/if}

			<div class="flex gap-4">
				<button
					type="button"
					on:click={startOver}
					class="flex-1 py-3 px-6 rounded-lg font-medium border transition-colors"
					style="border-color: var(--border); color: var(--text-primary); background-color: var(--bg)"
				>
					Import More Data
				</button>
				<a
					href="/dashboard/{importType === 'people' ? 'people' : 'services/songs'}"
					class="flex-1 bg-[var(--teal)] text-white py-3 px-6 rounded-lg font-medium hover:opacity-90 transition-opacity text-center"
				>
					View Imported {importType === 'people' ? 'People' : 'Songs'}
				</a>
			</div>
		</div>
	{/if}
</div>

<style>
	:global(:root) {
		--teal-rgb: 20, 184, 166;
	}
</style>
