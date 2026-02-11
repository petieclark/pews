<script lang="ts">
	import { goto } from '$app/navigation';

	let form = {
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

	let saving = false;

	async function handleSubmit() {
		saving = true;
		try {
			const payload = { ...form };
			if (payload.audio_duration_seconds) {
				payload.audio_duration_seconds = parseInt(payload.audio_duration_seconds);
			} else {
				delete payload.audio_duration_seconds;
			}

			const response = await fetch('/api/sermons', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				},
				body: JSON.stringify(payload)
			});

			if (response.ok) {
				goto('/dashboard/sermons');
			} else {
				alert('Failed to create sermon');
			}
		} catch (error) {
			console.error('Error creating sermon:', error);
			alert('Failed to create sermon');
		}
		saving = false;
	}
</script>

<div class="p-6 max-w-4xl mx-auto">
	<div class="mb-6">
		<h1 class="text-3xl font-bold">New Sermon</h1>
		<p class="mt-1 text-gray-600">Create a new sermon note</p>
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
			<textarea bind:value={form.notes_text} rows="12" class="w-full px-4 py-2 border border-custom rounded font-mono text-sm bg-surface text-primary" placeholder="## Main Points

1. First point..."></textarea>
		</div>

		<div class="flex items-center gap-2">
			<input type="checkbox" bind:checked={form.published} id="published" class="w-4 h-4" />
			<label for="published" class="text-sm font-medium text-[var(--text-primary)]">Publish immediately</label>
		</div>

		<div class="flex gap-4 pt-4">
			<button type="submit" disabled={saving} class="bg-[var(--teal)] text-white px-6 py-2 rounded hover:opacity-90 disabled:opacity-50">
				{saving ? 'Creating...' : 'Create Sermon'}
			</button>
			<button type="button" on:click={() => goto('/dashboard/sermons')} class="px-6 py-2 border border-custom rounded hover:bg-[var(--surface-hover)] dark:hover:bg-gray-800 text-[var(--text-primary)]">
				Cancel
			</button>
		</div>
	</form>
</div>
