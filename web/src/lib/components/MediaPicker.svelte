<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { createEventDispatcher } from 'svelte';

	export let type = 'all'; // 'all', 'image', 'document', 'audio'
	export let value = null; // Selected file object
	export let label = 'Select Media';
	export let allowUpload = true;

	const dispatch = createEventDispatcher();

	let showPicker = false;
	let files = [];
	let loading = false;
	let searchQuery = '';
	let selectedFolder = '';
	let folders = [];
	let uploadFiles = [];
	let uploadFolder = '';
	let uploadTags = '';
	let showUploadModal = false;

	onMount(() => {
		loadFolders();
	});

	async function loadFiles() {
		loading = true;
		try {
			const params = new URLSearchParams();
			if (type !== 'all') {
				params.append('type', type);
			}
			if (selectedFolder) {
				params.append('folder', selectedFolder);
			}
			const response = await api(`/api/media?${params}`);
			files = response || [];
		} catch (error) {
			console.error('Failed to load files:', error);
			files = [];
		} finally {
			loading = false;
		}
	}

	async function loadFolders() {
		try {
			folders = await api('/api/media/folders');
		} catch (error) {
			console.error('Failed to load folders:', error);
			folders = [];
		}
	}

	function openPicker() {
		showPicker = true;
		loadFiles();
	}

	function selectFile(file) {
		value = file;
		showPicker = false;
		dispatch('select', file);
	}

	function clearSelection() {
		value = null;
		dispatch('select', null);
	}

	function handleFileSelect(event) {
		uploadFiles = Array.from(event.target.files);
	}

	async function uploadFile() {
		if (uploadFiles.length === 0) return;

		try {
			let uploadedFile = null;
			for (const file of uploadFiles) {
				const formData = new FormData();
				formData.append('file', file);
				if (uploadFolder) {
					formData.append('folder', uploadFolder);
				}
				if (uploadTags) {
					formData.append('tags', uploadTags);
				}

				const response = await fetch(`${import.meta.env.VITE_API_URL}/api/media/upload`, {
					method: 'POST',
					headers: {
						Authorization: `Bearer ${localStorage.getItem('token')}`
					},
					body: formData
				});

				const data = await response.json();
				uploadedFile = data.file;
			}

			showUploadModal = false;
			uploadFiles = [];
			uploadFolder = '';
			uploadTags = '';
			loadFiles();
			loadFolders();

			// Auto-select the uploaded file
			if (uploadedFile) {
				selectFile(uploadedFile);
			}
		} catch (error) {
			alert('Failed to upload file: ' + error.message);
		}
	}

	import { Image, FileText, FileEdit, Music, Paperclip } from 'lucide-svelte';

	function getFileIcon(contentType) {
		if (contentType.startsWith('image/')) return Image;
		if (contentType === 'application/pdf') return FileText;
		if (contentType.includes('word')) return FileEdit;
		if (contentType.startsWith('audio/')) return Music;
		return Paperclip;
	}

	function formatFileSize(bytes) {
		if (bytes < 1024) return bytes + ' B';
		if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
		return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
	}

	$: filteredFiles = files.filter((file) =>
		file.original_name.toLowerCase().includes(searchQuery.toLowerCase())
	);
</script>

<div class="media-picker">
	{#if value}
		<div class="selected-file">
			<div class="file-preview">
				{#if value.content_type.startsWith('image/')}
					<img src={`${import.meta.env.VITE_API_URL}${value.url}`} alt={value.original_name} />
				{:else}
					<div class="file-icon"><svelte:component this={getFileIcon(value.content_type)} size={32} /></div>
				{/if}
			</div>
			<div class="file-info">
				<div class="file-name">{value.original_name}</div>
				<div class="file-meta">{formatFileSize(value.size_bytes)}</div>
			</div>
			<button on:click={clearSelection} class="btn-clear" type="button">✕</button>
		</div>
	{:else}
		<button on:click={openPicker} class="btn-select" type="button">{label}</button>
	{/if}
</div>

{#if showPicker}
	<div class="modal-backdrop" on:click={() => (showPicker = false)}>
		<div class="modal" on:click|stopPropagation>
			<h2>Select Media</h2>

			<div class="picker-header">
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Search files..."
					class="search-input"
				/>

				<select bind:value={selectedFolder} on:change={loadFiles}>
					<option value="">All Folders</option>
					{#each folders as folder}
						<option value={folder}>{folder}</option>
					{/each}
				</select>

				{#if allowUpload}
					<button on:click={() => (showUploadModal = true)} class="btn-primary" type="button"
						>Upload</button
					>
				{/if}
			</div>

			{#if loading}
				<div class="loading">Loading...</div>
			{:else if filteredFiles.length === 0}
				<div class="empty">
					<p>No files found</p>
					{#if allowUpload}
						<button on:click={() => (showUploadModal = true)} class="btn-primary" type="button"
							>Upload Files</button
						>
					{/if}
				</div>
			{:else}
				<div class="file-list">
					{#each filteredFiles as file}
						<button class="file-item" on:click={() => selectFile(file)} type="button">
							<div class="file-preview-small">
								{#if file.content_type.startsWith('image/')}
									<img src={`${import.meta.env.VITE_API_URL}${file.url}`} alt={file.original_name} />
								{:else}
									<div class="file-icon-small"><svelte:component this={getFileIcon(file.content_type)} size={24} /></div>
								{/if}
							</div>
							<div class="file-info">
								<div class="file-name">{file.original_name}</div>
								<div class="file-meta">
									{formatFileSize(file.size_bytes)} •
									{new Date(file.created_at).toLocaleDateString()}
								</div>
							</div>
						</button>
					{/each}
				</div>
			{/if}

			<div class="modal-actions">
				<button on:click={() => (showPicker = false)} class="btn-secondary" type="button"
					>Cancel</button
				>
			</div>
		</div>
	</div>
{/if}

{#if showUploadModal}
	<div class="modal-backdrop" on:click={() => (showUploadModal = false)}>
		<div class="modal" on:click|stopPropagation>
			<h2>Upload Files</h2>

			<div class="form-group">
				<label for="upload-files">Select Files</label>
				<input
					id="upload-files"
					type="file"
					multiple
					on:change={handleFileSelect}
					accept={type === 'image'
						? 'image/*'
						: type === 'audio'
							? 'audio/*'
							: type === 'document'
								? '.pdf,.docx'
								: 'image/*,.pdf,.docx,audio/*'}
				/>
				{#if uploadFiles.length > 0}
					<p>{uploadFiles.length} file(s) selected</p>
				{/if}
			</div>

			<div class="form-group">
				<label for="upload-folder">Folder</label>
				<input
					id="upload-folder"
					type="text"
					bind:value={uploadFolder}
					list="folders-list"
					placeholder="Optional"
				/>
				<datalist id="folders-list">
					{#each folders as folder}
						<option value={folder} />
					{/each}
				</datalist>
			</div>

			<div class="form-group">
				<label for="upload-tags">Tags (comma-separated)</label>
				<input id="upload-tags" type="text" bind:value={uploadTags} placeholder="Optional" />
			</div>

			<div class="modal-actions">
				<button on:click={() => (showUploadModal = false)} class="btn-secondary" type="button"
					>Cancel</button
				>
				<button
					on:click={uploadFile}
					class="btn-primary"
					disabled={uploadFiles.length === 0}
					type="button">Upload</button
				>
			</div>
		</div>
	</div>
{/if}

<style>
	.media-picker {
		margin: 1rem 0;
	}

	.selected-file {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		border: 1px solid #ddd;
		border-radius: 4px;
		background: #fafafa;
	}

	.file-preview {
		width: 80px;
		height: 80px;
		background: #fff;
		border: 1px solid #ddd;
		border-radius: 4px;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}

	.file-preview img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.file-icon {
		font-size: 2rem;
	}

	.file-info {
		flex: 1;
	}

	.file-name {
		font-weight: 500;
		margin-bottom: 0.25rem;
	}

	.file-meta {
		font-size: 0.875rem;
		color: #666;
	}

	.btn-clear {
		background: #dc2626;
		color: white;
		border: none;
		border-radius: 4px;
		width: 32px;
		height: 32px;
		cursor: pointer;
		font-size: 1.25rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.btn-clear:hover {
		background: #b91c1c;
	}

	.btn-select {
		padding: 0.5rem 1rem;
		background: #4f46e5;
		color: white;
		border: none;
		border-radius: 4px;
		cursor: pointer;
		font-weight: 500;
	}

	.btn-select:hover {
		background: #4338ca;
	}

	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal {
		background: white;
		padding: 2rem;
		border-radius: 8px;
		max-width: 600px;
		width: 90%;
		max-height: 90vh;
		overflow-y: auto;
	}

	.modal h2 {
		margin: 0 0 1.5rem;
	}

	.picker-header {
		display: flex;
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.search-input {
		flex: 1;
		padding: 0.5rem;
		border: 1px solid #ddd;
		border-radius: 4px;
	}

	.picker-header select {
		padding: 0.5rem;
		border: 1px solid #ddd;
		border-radius: 4px;
	}

	.loading,
	.empty {
		text-align: center;
		padding: 2rem;
		color: #666;
	}

	.file-list {
		max-height: 400px;
		overflow-y: auto;
		border: 1px solid #ddd;
		border-radius: 4px;
	}

	.file-item {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		border: none;
		border-bottom: 1px solid #ddd;
		background: white;
		cursor: pointer;
		text-align: left;
		transition: background 0.2s;
	}

	.file-item:last-child {
		border-bottom: none;
	}

	.file-item:hover {
		background: #f5f5f5;
	}

	.file-preview-small {
		width: 50px;
		height: 50px;
		background: #fafafa;
		border: 1px solid #ddd;
		border-radius: 4px;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		flex-shrink: 0;
	}

	.file-preview-small img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.file-icon-small {
		font-size: 1.5rem;
	}

	.modal-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		margin-top: 1.5rem;
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-group label {
		display: block;
		margin-bottom: 0.5rem;
		font-weight: 500;
	}

	.form-group input {
		width: 100%;
		padding: 0.5rem;
		border: 1px solid #ddd;
		border-radius: 4px;
	}

	button {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 4px;
		cursor: pointer;
		font-weight: 500;
	}

	.btn-primary {
		background: #4f46e5;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #4338ca;
	}

	.btn-primary:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-secondary {
		background: #e5e7eb;
		color: #374151;
	}

	.btn-secondary:hover {
		background: #d1d5db;
	}
</style>
