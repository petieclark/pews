<script>
	import { onMount } from 'svelte';
	import { AlertTriangle } from 'lucide-svelte';
	import { api } from '$lib/api';

	let backups = [];
	let loading = true;
	let creating = false;
	let error = '';
	let success = '';
	let showRestoreDialog = false;
	let restoreFilename = '';
	let restoreConfirmation = '';
	let restoring = false;

	onMount(async () => {
		await loadBackups();
	});

	async function loadBackups() {
		loading = true;
		error = '';
		
		try {
			const data = await api('/api/admin/backups');
			backups = data.backups || [];
		} catch (err) {
			error = 'Failed to load backups: ' + err.message;
		} finally {
			loading = false;
		}
	}

	async function createBackup() {
		creating = true;
		error = '';
		success = '';
		
		try {
			await api('/api/admin/backup', { method: 'POST' });
			success = 'Backup created successfully';
			await loadBackups();
		} catch (err) {
			error = 'Failed to create backup: ' + err.message;
		} finally {
			creating = false;
		}
	}

	function openRestoreDialog(filename) {
		restoreFilename = filename;
		restoreConfirmation = '';
		showRestoreDialog = true;
	}

	function closeRestoreDialog() {
		showRestoreDialog = false;
		restoreFilename = '';
		restoreConfirmation = '';
	}

	async function confirmRestore() {
		if (restoreConfirmation !== 'RESTORE') {
			error = 'Please type RESTORE to confirm';
			return;
		}

		restoring = true;
		error = '';
		success = '';

		try {
			await api(`/api/admin/restore/${restoreFilename}`, {
				method: 'POST',
				body: JSON.stringify({ confirmation: 'RESTORE' })
			});
			success = 'Backup restored successfully. Please refresh the page.';
			closeRestoreDialog();
			await loadBackups();
		} catch (err) {
			error = 'Failed to restore backup: ' + err.message;
		} finally {
			restoring = false;
		}
	}

	async function downloadBackup(filename) {
		try {
			const token = localStorage.getItem('token');
			const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8190';
			
			const response = await fetch(`${API_URL}/api/admin/backups/${filename}/download`, {
				headers: {
					'Authorization': `Bearer ${token}`
				}
			});
			
			if (!response.ok) {
				throw new Error('Download failed');
			}
			
			const blob = await response.blob();
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = filename;
			document.body.appendChild(a);
			a.click();
			window.URL.revokeObjectURL(url);
			document.body.removeChild(a);
		} catch (err) {
			error = 'Failed to download backup: ' + err.message;
		}
	}

	async function deleteBackup(filename) {
		if (!confirm('Are you sure you want to delete this backup?')) {
			return;
		}

		try {
			await api(`/api/admin/backups/${filename}`, { method: 'DELETE' });
			success = 'Backup deleted successfully';
			await loadBackups();
		} catch (err) {
			error = 'Failed to delete backup: ' + err.message;
		}
	}

	function formatDate(dateString) {
		return new Date(dateString).toLocaleString();
	}

	function formatSize(bytes) {
		if (bytes < 1024) return bytes + ' B';
		if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
		return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
	}
</script>

<div class="max-w-6xl">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-3xl font-bold text-primary">Database Backups</h1>
			<p class="text-secondary mt-1">Create, download, and restore backups of your church data</p>
		</div>
		
		<button
			on:click={createBackup}
			disabled={creating}
			class="bg-[var(--teal)] text-white py-2 px-6 rounded-lg font-medium hover:opacity-90 disabled:opacity-50 flex items-center gap-2"
		>
			{#if creating}
				<div class="inline-block animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
			{/if}
			{creating ? 'Creating...' : 'Create Backup Now'}
		</button>
	</div>

	{#if error}
		<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg mb-4">
			{error}
		</div>
	{/if}

	{#if success}
		<div class="success-box mb-4">
			{success}
		</div>
	{/if}

	<!-- Backups List -->
	<div class="bg-surface rounded-lg shadow-md border border-custom overflow-hidden">
		<div class="p-6 border-b border-custom">
			<h2 class="text-xl font-semibold text-primary">Available Backups</h2>
			<p class="text-sm text-secondary mt-1">Backups older than 30 days are automatically deleted</p>
		</div>

		{#if loading}
			<div class="text-center py-12">
				<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
			</div>
		{:else if backups.length === 0}
			<div class="text-center py-12">
				<svg class="mx-auto h-12 w-12 text-secondary opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
				</svg>
				<p class="mt-2 text-secondary">No backups available yet</p>
				<p class="text-sm text-secondary mt-1">Create your first backup to get started</p>
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full">
					<thead class="bg-[var(--surface-hover)]">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
								Created
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
								Filename
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-secondary uppercase tracking-wider">
								Size
							</th>
							<th class="px-6 py-3 text-right text-xs font-medium text-secondary uppercase tracking-wider">
								Actions
							</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-custom">
						{#each backups as backup}
							<tr class="hover:bg-[var(--surface-hover)]">
								<td class="px-6 py-4 whitespace-nowrap text-sm text-primary">
									{formatDate(backup.created_at)}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm font-mono text-primary">
									{backup.filename}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-secondary">
									{formatSize(backup.size_bytes)}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
									<button
										on:click={() => downloadBackup(backup.filename)}
										class="text-[var(--teal)] hover:opacity-80"
										title="Download"
									>
										Download
									</button>
									<button
										on:click={() => openRestoreDialog(backup.filename)}
										class="text-[var(--orange)] hover:opacity-80"
										title="Restore"
									>
										Restore
									</button>
									<button
										on:click={() => deleteBackup(backup.filename)}
										class="text-[var(--error-text)] hover:opacity-80"
										title="Delete"
									>
										Delete
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>

	<!-- Safety Notice -->
	<div class="mt-6 p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-700 rounded-lg">
		<div class="flex">
			<svg class="h-5 w-5 text-yellow-400 mr-3" fill="currentColor" viewBox="0 0 20 20">
				<path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
			</svg>
			<div>
				<h3 class="text-sm font-medium text-yellow-800 dark:text-yellow-200">Important Safety Information</h3>
				<div class="mt-2 text-sm text-yellow-700 dark:text-yellow-300">
					<ul class="list-disc list-inside space-y-1">
						<li>Restoring a backup will replace ALL current data for your church</li>
						<li>An automatic safety backup will be created before any restore</li>
						<li>Backups are stored for 30 days, then automatically deleted</li>
						<li>Download important backups to keep them permanently</li>
					</ul>
				</div>
			</div>
		</div>
	</div>
</div>

<!-- Restore Confirmation Dialog -->
{#if showRestoreDialog}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" on:click={closeRestoreDialog}>
		<div class="bg-surface rounded-lg p-6 max-w-md w-full mx-4 border border-custom" on:click|stopPropagation>
			<h2 class="text-2xl font-bold text-[var(--error-text)] mb-4"><AlertTriangle size={24} class="inline" /> Restore Backup?</h2>
			
			<div class="mb-4 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-700 rounded-lg">
				<p class="text-sm text-red-800 dark:text-red-200 font-medium mb-2">This action will:</p>
				<ul class="text-sm text-red-700 dark:text-red-300 list-disc list-inside space-y-1">
					<li>Delete ALL current data</li>
					<li>Replace it with data from this backup</li>
					<li>Create an automatic safety backup first</li>
				</ul>
			</div>

			<p class="text-sm text-secondary mb-4">
				Restoring: <span class="font-mono text-primary">{restoreFilename}</span>
			</p>

			<div class="mb-4">
				<label for="restore-confirm" class="block text-sm font-medium text-primary mb-2">
					Type <span class="font-bold text-[var(--error-text)]">RESTORE</span> to confirm:
				</label>
				<input
					id="restore-confirm"
					type="text"
					bind:value={restoreConfirmation}
					placeholder="RESTORE"
					class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--error-text)] focus:border-transparent bg-[var(--input-bg)] text-primary font-bold"
				/>
			</div>

			<div class="flex justify-end space-x-3">
				<button
					on:click={closeRestoreDialog}
					disabled={restoring}
					class="px-4 py-2 border border-custom rounded-lg text-primary hover:bg-[var(--surface-hover)] disabled:opacity-50"
				>
					Cancel
				</button>
				<button
					on:click={confirmRestore}
					disabled={restoring || restoreConfirmation !== 'RESTORE'}
					class="px-4 py-2 bg-[var(--error-text)] text-white rounded-lg font-medium hover:opacity-90 disabled:opacity-50"
				>
					{restoring ? 'Restoring...' : 'Restore Backup'}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.success-box {
		background-color: #D1FAE5;
		border: 1px solid #6EE7B7;
		color: #065F46;
		padding: 1rem;
		border-radius: 0.5rem;
	}
	:global(.dark) .success-box {
		background-color: #064E3B;
		border-color: #059669;
		color: #6EE7B7;
	}
</style>
