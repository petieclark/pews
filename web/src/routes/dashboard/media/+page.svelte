<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { Image, FileText, FileEdit, Music, Paperclip, FolderOpen, Download, Trash2 } from 'lucide-svelte';

	let files = [];
	let folders = [];
	let loading = false;
	let selectedFolder = '';
	let selectedType = 'all';
	let selectedFiles = new Set();
	let showUploadModal = false;
	let showCreateFolderModal = false;
	let showPreviewModal = false;
	let previewFile = null;
	let uploadFiles = [];
	let uploadFolder = '';
	let uploadTags = '';
	let newFolderName = '';
	let dragOver = false;

	const mediaTypes = [
		{ value: 'all', label: 'All Files' },
		{ value: 'image', label: 'Images' },
		{ value: 'document', label: 'Documents' },
		{ value: 'audio', label: 'Audio' }
	];

	onMount(() => {
		loadFiles();
		loadFolders();
	});

	async function loadFiles() {
		loading = true;
		try {
			const params = new URLSearchParams();
			if (selectedType !== 'all') {
				params.append('type', selectedType);
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

	function handleFileSelect(event) {
		uploadFiles = Array.from(event.target.files);
	}

	function handleDrop(event) {
		event.preventDefault();
		dragOver = false;
		uploadFiles = Array.from(event.dataTransfer.files);
		showUploadModal = true;
	}

	function handleDragOver(event) {
		event.preventDefault();
		dragOver = true;
	}

	function handleDragLeave() {
		dragOver = false;
	}

	async function uploadFile() {
		if (uploadFiles.length === 0) return;

		try {
			for (const file of uploadFiles) {
				const formData = new FormData();
				formData.append('file', file);
				if (uploadFolder) {
					formData.append('folder', uploadFolder);
				}
				if (uploadTags) {
					formData.append('tags', uploadTags);
				}

				await fetch(`${import.meta.env.VITE_API_URL}/api/media/upload`, {
					method: 'POST',
					headers: {
						Authorization: `Bearer ${localStorage.getItem('token')}`
					},
					body: formData
				});
			}

			showUploadModal = false;
			uploadFiles = [];
			uploadFolder = '';
			uploadTags = '';
			loadFiles();
			loadFolders();
		} catch (error) {
			alert('Failed to upload file: ' + error.message);
		}
	}

	async function createFolder() {
		if (!newFolderName) return;
		uploadFolder = newFolderName;
		folders = [...folders, newFolderName];
		showCreateFolderModal = false;
		newFolderName = '';
	}

	async function deleteFile(id) {
		if (!confirm('Are you sure you want to delete this file?')) return;

		try {
			await api(`/api/media/${id}`, { method: 'DELETE' });
			loadFiles();
		} catch (error) {
			alert('Failed to delete file: ' + error.message);
		}
	}

	async function bulkDelete() {
		if (selectedFiles.size === 0) return;
		if (!confirm(`Delete ${selectedFiles.size} files?`)) return;

		try {
			for (const id of selectedFiles) {
				await api(`/api/media/${id}`, { method: 'DELETE' });
			}
			selectedFiles.clear();
			loadFiles();
		} catch (error) {
			alert('Failed to delete files: ' + error.message);
		}
	}

	function toggleFileSelection(id) {
		if (selectedFiles.has(id)) {
			selectedFiles.delete(id);
		} else {
			selectedFiles.add(id);
		}
		selectedFiles = selectedFiles;
	}

	function selectAll() {
		selectedFiles = new Set(files.map((f) => f.id));
	}

	function deselectAll() {
		selectedFiles.clear();
		selectedFiles = selectedFiles;
	}

	function preview(file) {
		previewFile = file;
		showPreviewModal = true;
	}

	function download(file) {
		window.open(`${import.meta.env.VITE_API_URL}${file.url}`, '_blank');
	}

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

	$: {
		loadFiles();
	}
</script>

<div class="media-library">
	<div class="header">
		<h1>Media Library</h1>
		<div class="actions">
			{#if selectedFiles.size > 0}
				<button on:click={deselectAll} class="btn-secondary">Deselect All</button>
				<button on:click={bulkDelete} class="btn-danger">Delete Selected ({selectedFiles.size})</button
				>
			{:else}
				<button on:click={() => (showCreateFolderModal = true)} class="btn-secondary"
					>New Folder</button
				>
				<button on:click={() => (showUploadModal = true)} class="btn-primary">Upload Files</button>
			{/if}
		</div>
	</div>

	<div class="filters">
		<select bind:value={selectedType} on:change={loadFiles}>
			{#each mediaTypes as type}
				<option value={type.value}>{type.label}</option>
			{/each}
		</select>

		<select bind:value={selectedFolder} on:change={loadFiles}>
			<option value="">All Folders</option>
			{#each folders as folder}
				<option value={folder}>{folder}</option>
			{/each}
		</select>

		{#if files.length > 0}
			<button on:click={selectAll} class="btn-link">Select All</button>
		{/if}
	</div>

	{#if loading}
		<div class="loading">Loading...</div>
	{:else if files.length === 0}
		<div
			class="empty-state"
			class:drag-over={dragOver}
			on:drop={handleDrop}
			on:dragover={handleDragOver}
			on:dragleave={handleDragLeave}
		>
			<div class="icon"><FolderOpen size={64} /></div>
			<h3>No files yet</h3>
			<p>Upload your first file or drag and drop files here</p>
			<button on:click={() => (showUploadModal = true)} class="btn-primary">Upload Files</button>
		</div>
	{:else}
		<div class="file-grid">
			{#each files as file}
				<div class="file-card" class:selected={selectedFiles.has(file.id)}>
					<div class="file-checkbox">
						<input
							type="checkbox"
							checked={selectedFiles.has(file.id)}
							on:change={() => toggleFileSelection(file.id)}
						/>
					</div>

					<button class="file-preview" on:click={() => preview(file)}>
						{#if file.content_type.startsWith('image/')}
							<img src={`${import.meta.env.VITE_API_URL}${file.url}`} alt={file.original_name} />
						{:else}
							<div class="file-icon"><svelte:component this={getFileIcon(file.content_type)} size={48} /></div>
						{/if}
					</button>

					<div class="file-info">
						<div class="file-name" title={file.original_name}>{file.original_name}</div>
						<div class="file-meta">
							{formatFileSize(file.size_bytes)} •
							{new Date(file.created_at).toLocaleDateString()}
						</div>
						{#if file.folder}
							<div class="file-folder"><FolderOpen size={14} class="inline" /> {file.folder}</div>
						{/if}
						{#if file.tags && file.tags.length > 0}
							<div class="file-tags">
								{#each file.tags as tag}
									<span class="tag">{tag}</span>
								{/each}
							</div>
						{/if}
					</div>

					<div class="file-actions">
						<button on:click={() => download(file)} class="btn-icon" title="Download"><Download size={16} /></button>
						<button on:click={() => deleteFile(file.id)} class="btn-icon" title="Delete"><Trash2 size={16} /></button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Upload Modal -->
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
					accept="image/*,.pdf,.docx,audio/*"
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
					placeholder="Leave empty or select/create folder"
				/>
				<datalist id="folders-list">
					{#each folders as folder}
						<option value={folder} />
					{/each}
				</datalist>
			</div>

			<div class="form-group">
				<label for="upload-tags">Tags (comma-separated)</label>
				<input
					id="upload-tags"
					type="text"
					bind:value={uploadTags}
					placeholder="e.g. sermon, worship, announcement"
				/>
			</div>

			<div class="modal-actions">
				<button on:click={() => (showUploadModal = false)} class="btn-secondary">Cancel</button>
				<button on:click={uploadFile} class="btn-primary" disabled={uploadFiles.length === 0}
					>Upload</button
				>
			</div>
		</div>
	</div>
{/if}

<!-- Create Folder Modal -->
{#if showCreateFolderModal}
	<div class="modal-backdrop" on:click={() => (showCreateFolderModal = false)}>
		<div class="modal" on:click|stopPropagation>
			<h2>Create Folder</h2>

			<div class="form-group">
				<label for="folder-name">Folder Name</label>
				<input id="folder-name" type="text" bind:value={newFolderName} placeholder="e.g. Sermons" />
			</div>

			<div class="modal-actions">
				<button on:click={() => (showCreateFolderModal = false)} class="btn-secondary">Cancel</button
				>
				<button on:click={createFolder} class="btn-primary" disabled={!newFolderName}>Create</button>
			</div>
		</div>
	</div>
{/if}

<!-- Preview Modal -->
{#if showPreviewModal && previewFile}
	<div class="modal-backdrop" on:click={() => (showPreviewModal = false)}>
		<div class="modal preview-modal" on:click|stopPropagation>
			<h2>{previewFile.original_name}</h2>

			<div class="preview-content">
				{#if previewFile.content_type.startsWith('image/')}
					<img src={`${import.meta.env.VITE_API_URL}${previewFile.url}`} alt={previewFile.original_name} />
				{:else if previewFile.content_type === 'application/pdf'}
					<iframe
						src={`${import.meta.env.VITE_API_URL}${previewFile.url}`}
						title={previewFile.original_name}
					></iframe>
				{:else if previewFile.content_type.startsWith('audio/')}
					<audio controls src={`${import.meta.env.VITE_API_URL}${previewFile.url}`}></audio>
				{:else}
					<p>Preview not available. <a href={`${import.meta.env.VITE_API_URL}${previewFile.url}`} target="_blank">Download file</a></p>
				{/if}
			</div>

			<div class="preview-info">
				<p><strong>Size:</strong> {formatFileSize(previewFile.size_bytes)}</p>
				<p><strong>Type:</strong> {previewFile.content_type}</p>
				<p><strong>Uploaded:</strong> {new Date(previewFile.created_at).toLocaleString()}</p>
				{#if previewFile.folder}
					<p><strong>Folder:</strong> {previewFile.folder}</p>
				{/if}
			</div>

			<div class="modal-actions">
				<button on:click={() => download(previewFile)} class="btn-secondary">Download</button>
				<button on:click={() => (showPreviewModal = false)} class="btn-primary">Close</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.media-library {
		padding: 2rem;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
	}

	.header h1 {
		margin: 0;
		font-size: 2rem;
	}

	.actions {
		display: flex;
		gap: 1rem;
	}

	.filters {
		display: flex;
		gap: 1rem;
		margin-bottom: 2rem;
		align-items: center;
	}

	.filters select {
		padding: 0.5rem;
		border: 1px solid #ddd;
		border-radius: 4px;
	}

	.loading {
		text-align: center;
		padding: 4rem;
		color: #666;
	}

	.empty-state {
		text-align: center;
		padding: 4rem;
		border: 2px dashed #ddd;
		border-radius: 8px;
		transition: border-color 0.3s;
	}

	.empty-state.drag-over {
		border-color: #4f46e5;
		background: #f5f3ff;
	}

	.empty-state .icon {
		font-size: 4rem;
		margin-bottom: 1rem;
	}

	.empty-state h3 {
		margin: 0 0 0.5rem;
		color: #333;
	}

	.empty-state p {
		color: #666;
		margin-bottom: 1.5rem;
	}

	.file-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 1.5rem;
	}

	.file-card {
		border: 1px solid #ddd;
		border-radius: 8px;
		overflow: hidden;
		position: relative;
		transition: all 0.2s;
	}

	.file-card:hover {
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
	}

	.file-card.selected {
		border-color: #4f46e5;
		box-shadow: 0 0 0 2px #4f46e5;
	}

	.file-checkbox {
		position: absolute;
		top: 0.5rem;
		left: 0.5rem;
		z-index: 10;
	}

	.file-checkbox input {
		width: 1.25rem;
		height: 1.25rem;
		cursor: pointer;
	}

	.file-preview {
		width: 100%;
		height: 150px;
		background: #f5f5f5;
		border: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.file-preview img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.file-icon {
		font-size: 3rem;
	}

	.file-info {
		padding: 0.75rem;
	}

	.file-name {
		font-weight: 500;
		margin-bottom: 0.25rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.file-meta {
		font-size: 0.875rem;
		color: #666;
	}

	.file-folder {
		font-size: 0.875rem;
		color: #4f46e5;
		margin-top: 0.25rem;
	}

	.file-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem;
		margin-top: 0.5rem;
	}

	.tag {
		background: #e0e7ff;
		color: #4f46e5;
		padding: 0.125rem 0.5rem;
		border-radius: 12px;
		font-size: 0.75rem;
	}

	.file-actions {
		display: flex;
		gap: 0.5rem;
		padding: 0.5rem 0.75rem;
		border-top: 1px solid #ddd;
		background: #fafafa;
	}

	.btn-icon {
		background: none;
		border: none;
		cursor: pointer;
		font-size: 1.25rem;
		padding: 0.25rem;
	}

	.btn-icon:hover {
		opacity: 0.7;
	}

	/* Modal Styles */
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
		max-width: 500px;
		width: 90%;
		max-height: 90vh;
		overflow-y: auto;
	}

	.preview-modal {
		max-width: 800px;
	}

	.modal h2 {
		margin: 0 0 1.5rem;
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-group label {
		display: block;
		margin-bottom: 0.5rem;
		font-weight: 500;
	}

	.form-group input,
	.form-group select {
		width: 100%;
		padding: 0.5rem;
		border: 1px solid #ddd;
		border-radius: 4px;
	}

	.modal-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		margin-top: 1.5rem;
	}

	.preview-content {
		margin: 1.5rem 0;
		text-align: center;
	}

	.preview-content img {
		max-width: 100%;
		max-height: 500px;
		object-fit: contain;
	}

	.preview-content iframe {
		width: 100%;
		height: 500px;
		border: 1px solid #ddd;
	}

	.preview-content audio {
		width: 100%;
	}

	.preview-info {
		background: #f5f5f5;
		padding: 1rem;
		border-radius: 4px;
		margin: 1rem 0;
	}

	.preview-info p {
		margin: 0.5rem 0;
	}

	/* Button Styles */
	button {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 4px;
		cursor: pointer;
		font-weight: 500;
		transition: all 0.2s;
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

	.btn-danger {
		background: #dc2626;
		color: white;
	}

	.btn-danger:hover {
		background: #b91c1c;
	}

	.btn-link {
		background: none;
		color: #4f46e5;
		padding: 0.25rem 0.5rem;
	}

	.btn-link:hover {
		text-decoration: underline;
	}
</style>
