<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

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
	let showPreview = false;

	async function handleSubmit() {
		saving = true;
		try {
			const payload = { ...form };
			if (payload.audio_duration_seconds) {
				payload.audio_duration_seconds = parseInt(payload.audio_duration_seconds);
			} else {
				delete payload.audio_duration_seconds;
			}

			await api('/api/sermons', {
				method: 'POST',
				body: JSON.stringify(payload)
			});
			goto('/dashboard/sermons');
		} catch (error) {
			console.error('Error creating sermon:', error);
		}
		saving = false;
	}

	function simpleMarkdown(text) {
		if (!text) return '';
		return text
			.replace(/^### (.+)$/gm, '<h3 class="text-lg font-semibold mt-4 mb-2 text-primary">$1</h3>')
			.replace(/^## (.+)$/gm, '<h2 class="text-xl font-bold mt-4 mb-2 text-primary">$1</h2>')
			.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
			.replace(/^\d+\. (.+)$/gm, '<li class="ml-4 list-decimal text-primary">$1</li>')
			.replace(/^- (.+)$/gm, '<li class="ml-4 list-disc text-primary">$1</li>')
			.replace(/\n\n/g, '<br/><br/>')
			.replace(/\n/g, '<br/>');
	}
</script>

<div class="p-6 max-w-4xl mx-auto">
	<div class="mb-6">
		<button on:click={() => goto('/dashboard/sermons')} class="text-[var(--teal)] hover:underline text-sm flex items-center gap-1 mb-4">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
			Back to Sermons
		</button>
		<h1 class="text-3xl font-bold text-[var(--text-primary)]">New Sermon</h1>
		<p class="mt-1 text-secondary">Create a new sermon note</p>
	</div>

	<form on:submit|preventDefault={handleSubmit} class="bg-surface border border-custom rounded-lg shadow p-6 space-y-4">
		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			<div>
				<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Title *</label>
				<input type="text" bind:value={form.title} required class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Speaker *</label>
				<input type="text" bind:value={form.speaker} required class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Date *</label>
				<input type="date" bind:value={form.sermon_date} required class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Scripture Reference</label>
				<input type="text" bind:value={form.scripture_reference} placeholder="John 3:16" class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Series Name</label>
				<input type="text" bind:value={form.series_name} class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Audio Duration (seconds)</label>
				<input type="number" bind:value={form.audio_duration_seconds} placeholder="2400" class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
			</div>
			<div class="md:col-span-2">
				<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Audio URL</label>
				<input type="url" bind:value={form.audio_url} placeholder="https://..." class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
			</div>
			<div class="md:col-span-2">
				<label class="block text-sm font-medium mb-1 text-[var(--text-primary)]">Video URL</label>
				<input type="url" bind:value={form.video_url} placeholder="https://..." class="w-full px-4 py-2 border border-custom rounded bg-surface text-primary focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" />
			</div>
		</div>

		<div>
			<div class="flex items-center justify-between mb-1">
				<label class="text-sm font-medium text-[var(--text-primary)]">Sermon Notes (Markdown)</label>
				<button type="button" on:click={() => showPreview = !showPreview} class="text-xs text-[var(--teal)] hover:underline">
					{showPreview ? 'Edit' : 'Preview'}
				</button>
			</div>
			{#if showPreview}
				<div class="w-full px-4 py-3 border border-custom rounded bg-surface text-primary min-h-[300px] prose prose-sm dark:prose-invert max-w-none">
					{@html simpleMarkdown(form.notes_text)}
				</div>
			{:else}
				<textarea bind:value={form.notes_text} rows="12" class="w-full px-4 py-2 border border-custom rounded font-mono text-sm bg-surface text-primary focus:ring-2 focus:ring-[var(--teal)] focus:outline-none" placeholder="## Main Points

1. First point..."></textarea>
			{/if}
		</div>

		<div class="flex items-center gap-2">
			<input type="checkbox" bind:checked={form.published} id="published" class="w-4 h-4 accent-[var(--teal)]" />
			<label for="published" class="text-sm font-medium text-[var(--text-primary)]">Publish immediately</label>
		</div>

		<div class="flex gap-4 pt-4">
			<button type="submit" disabled={saving} class="bg-[var(--teal)] text-white px-6 py-2 rounded hover:opacity-90 disabled:opacity-50">
				{saving ? 'Creating...' : 'Create Sermon'}
			</button>
			<button type="button" on:click={() => goto('/dashboard/sermons')} class="px-6 py-2 border border-custom rounded hover:bg-[var(--surface-hover)] text-[var(--text-primary)]">
				Cancel
			</button>
		</div>
	</form>
</div>
