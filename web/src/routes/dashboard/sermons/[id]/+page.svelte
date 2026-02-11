<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';

	let form = {
		id: '',
		title: '',
		speaker: '',
		sermon_date: '',
		scripture_reference: '',
		notes_text: '',
		audio_url: '',
		video_url: '',
		series_name: '',
		audio_duration_seconds: '',
		published: false
	};

	let loading = true;
	let saving = false;

	$: sermonId = $page.params.id;

	onMount(async () => {
		await loadSermon();
		loading = false;
	});

	async function loadSermon() {
		try {
			const response = await fetch(`/api/sermons/${sermonId}`, {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});

			if (response.ok) {
				const sermon = await response.json();
				form = {
					...sermon,
					sermon_date: sermon.sermon_date.split('T')[0],
					audio_duration_seconds: sermon.audio_duration_seconds || ''
				};
			}
		} catch (error) {
			console.error('Failed to load sermon:', error);
		}
	}

	async function handleSubmit() {
		saving = true;
		try {
			const payload = { ...form };
			if (payload.audio_duration_seconds) {
				payload.audio_duration_seconds = parseInt(payload.audio_duration_seconds);
			} else {
				delete payload.audio_duration_seconds;
			}

			const response = await fetch(`/api/sermons/${sermonId}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				},
				body: JSON.stringify(payload)
			});

			if (response.ok) {
				goto('/dashboard/sermons');
			} else {
				alert('Failed to update sermon');
			}
		} catch (error) {
			console.error('Error updating sermon:', error);
			alert('Failed to update sermon');
		}
		saving = false;
	}
</script>

<div class="p-6 max-w-4xl mx-auto">
	{#if loading}
		<div class="flex justify-center items-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-teal-600"></div>
		</div>
	{:else}
		<div class="mb-6">
			<h1 class="text-3xl font-bold">Edit Sermon</h1>
			<p class="mt-1 text-gray-600">Update sermon details</p>
		</div>

		<form on:submit|preventDefault={handleSubmit} class="bg-white rounded-lg shadow p-6 space-y-4">
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<div>
					<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Title *</label>
					<input type="text" bind:value={form.title} required class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Speaker *</label>
					<input type="text" bind:value={form.speaker} required class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Date *</label>
					<input type="date" bind:value={form.sermon_date} required class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Scripture Reference</label>
					<input type="text" bind:value={form.scripture_reference} placeholder="John 3:16" class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Series Name</label>
					<input type="text" bind:value={form.series_name} class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Audio Duration (seconds)</label>
					<input type="number" bind:value={form.audio_duration_seconds} placeholder="2400" class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary" />
				</div>
				<div class="md:col-span-2">
					<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Audio URL</label>
					<input type="url" bind:value={form.audio_url} placeholder="https://..." class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary" />
				</div>
				<div class="md:col-span-2">
					<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Video URL</label>
					<input type="url" bind:value={form.video_url} placeholder="https://..." class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary" />
				</div>
			</div>

			<div>
				<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Sermon Notes (Markdown)</label>
				<textarea bind:value={form.notes_text} rows="12" class="w-full px-4 py-2 border border-custom rounded font-mono text-sm bg-surface text-primary"></textarea>
			</div>

			<div class="flex items-center gap-2">
				<input type="checkbox" bind:checked={form.published} id="published" class="w-4 h-4" />
				<label for="published" class="text-sm font-medium text-[var(--text-primary)]">Published</label>
			</div>

			<div class="flex gap-4 pt-4">
				<button type="submit" disabled={saving} class="bg-[var(--teal)] text-white px-6 py-2 rounded hover:opacity-90 disabled:opacity-50">
					{saving ? 'Updating...' : 'Update Sermon'}
				</button>
				<button type="button" on:click={() => goto('/dashboard/sermons')} class="px-6 py-2 border border-custom rounded hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800 text-[var(--text-primary)]">
					Cancel
				</button>
			</div>
		</form>
	{/if}
</div>
