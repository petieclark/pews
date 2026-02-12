<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api';

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
	let viewMode = 'view'; // 'view' or 'edit'
	let showPreview = false;

	$: sermonId = $page.params.id;

	onMount(async () => {
		await loadSermon();
		loading = false;
	});

	async function loadSermon() {
		try {
			const sermon = await api(`/api/sermons/${sermonId}`);
			form = {
				...sermon,
				sermon_date: sermon.sermon_date.split('T')[0],
				audio_duration_seconds: sermon.audio_duration_seconds || ''
			};
		} catch (error) {
			console.error('Failed to load sermon:', error);
			goto('/dashboard/sermons');
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

			await api(`/api/sermons/${sermonId}`, {
				method: 'PUT',
				body: JSON.stringify(payload)
			});
			viewMode = 'view';
			await loadSermon();
		} catch (error) {
			console.error('Error updating sermon:', error);
		}
		saving = false;
	}

	async function togglePublished() {
		try {
			const payload = { ...form, published: !form.published };
			payload.sermon_date = form.sermon_date;
			if (payload.audio_duration_seconds) {
				payload.audio_duration_seconds = parseInt(payload.audio_duration_seconds);
			} else {
				delete payload.audio_duration_seconds;
			}
			await api(`/api/sermons/${sermonId}`, {
				method: 'PUT',
				body: JSON.stringify(payload)
			});
			await loadSermon();
		} catch (error) {
			console.error('Error toggling published:', error);
		}
	}

	function simpleMarkdown(text) {
		if (!text) return '<p class="text-secondary italic">No notes yet.</p>';
		return text
			.replace(/^### (.+)$/gm, '<h3 class="text-lg font-semibold mt-4 mb-2 text-[var(--text-primary)]">$1</h3>')
			.replace(/^## (.+)$/gm, '<h2 class="text-xl font-bold mt-5 mb-2 text-[var(--text-primary)]">$1</h2>')
			.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
			.replace(/^\d+\. (.+)$/gm, '<li class="ml-4 list-decimal text-[var(--text-primary)]">$1</li>')
			.replace(/^- (.+)$/gm, '<li class="ml-4 list-disc text-[var(--text-primary)]">$1</li>')
			.replace(/\n\n/g, '<br/><br/>')
			.replace(/\n/g, '<br/>');
	}

	function formatDate(dateString) {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function getYouTubeId(url) {
		if (!url) return null;
		const match = url.match(/(?:youtube\.com\/(?:watch\?v=|embed\/)|youtu\.be\/)([a-zA-Z0-9_-]{11})/);
		return match ? match[1] : null;
	}
</script>

<div class="p-6 max-w-4xl mx-auto">
	{#if loading}
		<div class="flex justify-center items-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else if viewMode === 'view'}
		<!-- View Mode -->
		<button on:click={() => goto('/dashboard/sermons')} class="text-[var(--teal)] hover:underline text-sm flex items-center gap-1 mb-4">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
			Back to Sermons
		</button>

		<div class="bg-surface border border-custom rounded-xl shadow-sm overflow-hidden">
			<!-- Header -->
			<div class="p-6 border-b border-custom">
				<div class="flex flex-col sm:flex-row justify-between items-start gap-4">
					<div>
						<div class="flex items-center gap-3 mb-2">
							{#if form.series_name}
								<span class="px-2.5 py-0.5 text-xs font-medium rounded-full bg-[var(--teal)] bg-opacity-15 text-[var(--teal)]">{form.series_name}</span>
							{/if}
							<span class="px-2 py-0.5 text-xs rounded-full {form.published ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100' : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100'}">
								{form.published ? 'Published' : 'Draft'}
							</span>
						</div>
						<h1 class="text-2xl font-bold text-[var(--text-primary)]">{form.title}</h1>
						<div class="flex items-center gap-4 mt-2 text-sm text-secondary">
							<span>{form.speaker}</span>
							<span>•</span>
							<span>{formatDate(form.sermon_date)}</span>
							{#if form.scripture_reference}
								<span>•</span>
								<span class="text-[var(--teal)] font-medium">{form.scripture_reference}</span>
							{/if}
						</div>
					</div>
					<div class="flex gap-2 flex-shrink-0">
						<button on:click={togglePublished} class="px-3 py-1.5 text-sm border border-custom rounded-lg hover:bg-[var(--surface-hover)] text-[var(--text-primary)]">
							{form.published ? 'Unpublish' : 'Publish'}
						</button>
						<button on:click={() => viewMode = 'edit'} class="px-4 py-1.5 text-sm bg-[var(--teal)] text-white rounded-lg hover:opacity-90">
							Edit
						</button>
					</div>
				</div>
			</div>

			<!-- Media -->
			{#if form.video_url || form.audio_url}
				<div class="p-6 border-b border-custom">
					{#if form.video_url}
						{@const ytId = getYouTubeId(form.video_url)}
						{#if ytId}
							<div class="aspect-video rounded-lg overflow-hidden bg-black">
								<iframe src="https://www.youtube.com/embed/{ytId}" title={form.title} class="w-full h-full" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
							</div>
						{:else}
							<div class="flex items-center gap-2 text-sm">
								<span class="text-secondary">Video:</span>
								<a href={form.video_url} target="_blank" rel="noopener" class="text-[var(--teal)] hover:underline">{form.video_url}</a>
							</div>
						{/if}
					{/if}
					{#if form.audio_url}
						<div class="mt-3">
							<audio controls class="w-full" src={form.audio_url}>
								<track kind="captions" />
							</audio>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Notes -->
			<div class="p-6">
				<div class="prose prose-sm dark:prose-invert max-w-none text-[var(--text-primary)]">
					{@html simpleMarkdown(form.notes_text)}
				</div>
			</div>
		</div>

	{:else}
		<!-- Edit Mode -->
		<div class="mb-6">
			<button on:click={() => viewMode = 'view'} class="text-[var(--teal)] hover:underline text-sm flex items-center gap-1 mb-4">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
				Back to View
			</button>
			<h1 class="text-3xl font-bold text-[var(--text-primary)]">Edit Sermon</h1>
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
					<textarea bind:value={form.notes_text} rows="12" class="w-full px-4 py-2 border border-custom rounded font-mono text-sm bg-surface text-primary focus:ring-2 focus:ring-[var(--teal)] focus:outline-none"></textarea>
				{/if}
			</div>

			<div class="flex items-center gap-2">
				<input type="checkbox" bind:checked={form.published} id="published" class="w-4 h-4 accent-[var(--teal)]" />
				<label for="published" class="text-sm font-medium text-[var(--text-primary)]">Published</label>
			</div>

			<div class="flex gap-4 pt-4">
				<button type="submit" disabled={saving} class="bg-[var(--teal)] text-white px-6 py-2 rounded hover:opacity-90 disabled:opacity-50">
					{saving ? 'Updating...' : 'Update Sermon'}
				</button>
				<button type="button" on:click={() => viewMode = 'view'} class="px-6 py-2 border border-custom rounded hover:bg-[var(--surface-hover)] text-[var(--text-primary)]">
					Cancel
				</button>
			</div>
		</form>
	{/if}
</div>
