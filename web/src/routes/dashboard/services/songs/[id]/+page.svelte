<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import { Music2, FileText } from 'lucide-svelte';

	const NOTES = ['C', 'C#', 'Db', 'D', 'D#', 'Eb', 'E', 'F', 'F#', 'Gb', 'G', 'G#', 'Ab', 'A', 'A#', 'Bb', 'B'];
	const ENHARMONIC = { 'C#': 'Db', 'Db': 'C#', 'D#': 'Eb', 'Eb': 'D#', 'F#': 'Gb', 'Gb': 'F#', 'G#': 'Ab', 'Ab': 'G#', 'A#': 'Bb', 'Bb': 'A#' };
	const SHARP_KEYS = ['C', 'G', 'D', 'A', 'E', 'B', 'F#', 'C#'];
	const FLAT_KEYS = ['F', 'Bb', 'Eb', 'Ab', 'Db', 'Gb'];

	let songId = '';
	let song = null;
	let usage = [];
	let attachments = [];
	let loading = true;
	let usageLoading = true;

	// Editing state
	let editing = null; // field name being edited
	let editValue = '';

	// Transpose
	let transposeKey = '';
	let showTranspose = false;

	// Attachments
	let uploadingFile = false;
	let dragOver = false;

	$: songId = $page.params.id;

	onMount(() => {
		loadSong();
		loadUsage();
		loadAttachments();
	});

	async function loadSong() {
		loading = true;
		try {
			song = await api(`/api/services/songs/${songId}`);
			transposeKey = song.default_key || '';
			// Ensure authors is always an array (in case backend returns null or string)
			if (!song.authors || !Array.isArray(song.authors)) {
				song.authors = [];
			}
		} catch (error) {
			console.error('Failed to load song:', error);
			goto('/dashboard/services/songs');
		} finally {
			loading = false;
		}
	}

	async function loadUsage() {
		usageLoading = true;
		try {
			const result = await api(`/api/services/songs/${songId}/usage`);
			usage = result || [];
		} catch (error) {
			console.error('Failed to load usage:', error);
			usage = [];
		} finally {
			usageLoading = false;
		}
	}

	async function loadAttachments() {
		try {
			attachments = await api(`/api/services/songs/${songId}/attachments`);
			if (!Array.isArray(attachments)) attachments = [];
		} catch (error) {
			attachments = [];
		}
	}

	function startEdit(field) {
		editing = field;
		editValue = song[field] || '';
	}

	async function saveField(field) {
		if (editValue === (song[field] || '')) {
			editing = null;
			return;
		}
		try {
			const payload = {
				title: song.title,
				artist: song.artist || '',
				default_key: song.default_key || '',
				tempo: song.tempo || 0,
				ccli_number: song.ccli_number || '',
				authors: song.authors || [], // authors is array in backend
				copyright_year: song.copyright_year || null,
				publisher: song.publisher || '',
				license_type: song.license_type || '',
				lyrics: song.lyrics || '',
				notes: song.notes || '',
				tags: song.tags || '',
				youtube_url: song.youtube_url || '',
				spotify_url: song.spotify_url || '',
				apple_music_url: song.apple_music_url || '',
				rehearsal_url: song.rehearsal_url || ''
			};
			
			if (field === 'tempo') {
				payload[field] = parseInt(editValue) || 0;
			} else if (field === 'authors') {
				// Convert comma-separated string to array
				payload[field] = editValue.split(',').map(a => a.trim()).filter(a => a);
			} else if (field === 'copyright_year') {
				payload[field] = editValue ? parseInt(editValue) || null : null;
			} else {
				payload[field] = editValue;
			}
			
			song = await api(`/api/services/songs/${songId}`, {
				method: 'PUT',
				body: JSON.stringify(payload)
			});
			if (field === 'default_key') transposeKey = song.default_key || '';
		} catch (error) {
			console.error('Failed to save:', error);
			alert('Failed to save: ' + error.message);
		}
		editing = null;
	}

	function handleEditKeydown(e, field) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			saveField(field);
		}
		if (e.key === 'Escape') editing = null;
	}

	// Transpose logic
	function getNoteIndex(note) {
		const idx = NOTES.indexOf(note);
		if (idx >= 0) return idx;
		if (ENHARMONIC[note]) return NOTES.indexOf(ENHARMONIC[note]);
		return -1;
	}

	function transposeNote(note, semitones, useFlats) {
		const idx = getNoteIndex(note);
		if (idx < 0) return note;
		const newIdx = ((idx + semitones) % 12 + 12) % 12;
		const result = NOTES[newIdx];
		if (useFlats && ENHARMONIC[result] && FLAT_KEYS.includes(ENHARMONIC[result])) {
			return ENHARMONIC[result];
		}
		return result;
	}

	function transposeLyrics(lyrics, fromKey, toKey) {
		if (!lyrics || !fromKey || !toKey || fromKey === toKey) return lyrics;
		const fromIdx = getNoteIndex(fromKey);
		const toIdx = getNoteIndex(toKey);
		if (fromIdx < 0 || toIdx < 0) return lyrics;
		const semitones = toIdx - fromIdx;
		const useFlats = FLAT_KEYS.includes(toKey);

		// Match chord patterns: letter + optional # or b + optional m/maj/dim/aug/sus/add/7/9/11/13
		return lyrics.replace(/\b([A-G][#b]?)(m|maj|min|dim|aug|sus[24]?|add[0-9]*|[0-9]*)\b/g, (match, root, suffix) => {
			return transposeNote(root, semitones, useFlats) + suffix;
		});
	}

	$: displayLyrics = showTranspose && transposeKey !== song?.default_key
		? transposeLyrics(song?.lyrics, song?.default_key, transposeKey)
		: song?.lyrics;

	function formatDate(dateStr) {
		if (!dateStr) return 'Never';
		return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	function formatFileSize(bytes) {
		if (bytes < 1024) return bytes + ' B';
		if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
		return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
	}

	function getTagColor(tag) {
		const colors = [
			'bg-teal/20 text-teal',
			'bg-blue-500/20 text-blue-400',
			'bg-purple-500/20 text-purple-400',
			'bg-amber-500/20 text-amber-400',
			'bg-rose-500/20 text-rose-400',
			'bg-green-500/20 text-green-400',
		];
		let hash = 0;
		for (let i = 0; i < tag.length; i++) hash = tag.charCodeAt(i) + ((hash << 5) - hash);
		return colors[Math.abs(hash) % colors.length];
	}

	async function handleFileDrop(e) {
		e.preventDefault();
		dragOver = false;
		const files = e.dataTransfer?.files;
		if (files?.length) await uploadFile(files[0]);
	}

	async function handleFileSelect(e) {
		const file = e.target.files[0];
		if (file) await uploadFile(file);
		e.target.value = '';
	}

	async function uploadFile(file) {
		const allowedTypes = {
			'application/pdf': true,
			'image/jpeg': true,
			'image/jpg': true,
			'image/png': true,
			'application/vnd.openxmlformats-officedocument.wordprocessingml.document': true
		};
		if (!allowedTypes[file.type]) {
			alert('File type not allowed. Allowed: PDF, PNG, JPG, DOCX');
			return;
		}
		if (file.size > 20 * 1024 * 1024) {
			alert('File size must be less than 20MB');
			return;
		}
		uploadingFile = true;
		try {
			const formData = new FormData();
			formData.append('file', file);
			const token = localStorage.getItem('token');
			const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8190';
			const response = await fetch(`${apiUrl}/api/services/songs/${songId}/attachments`, {
				method: 'POST',
				headers: { 'Authorization': `Bearer ${token}` },
				body: formData
			});
			if (!response.ok) throw new Error('Upload failed');
			await loadAttachments();
		} catch (error) {
			alert('Failed to upload: ' + error.message);
		} finally {
			uploadingFile = false;
		}
	}

	async function deleteAttachment(attachmentId) {
		if (!confirm('Delete this attachment?')) return;
		try {
			await api(`/api/services/songs/attachments/${attachmentId}`, { method: 'DELETE' });
			await loadAttachments();
		} catch (error) {
			alert('Failed to delete attachment');
		}
	}

	// Media URL helpers
	function extractYoutubeId(url) {
		if (!url) return null;
		const match = url.match(/(?:youtube\.com\/(?:watch\?v=|embed\/)|youtu\.be\/)([a-zA-Z0-9_-]{11})/);
		return match ? match[1] : null;
	}

	function extractSpotifyId(url) {
		if (!url) return null;
		const match = url.match(/spotify\.com\/track\/([a-zA-Z0-9]+)/);
		return match ? match[1] : null;
	}

	function isAudioUrl(url) {
		if (!url) return false;
		return /\.(mp3|wav|ogg|m4a|aac|webm)(\?|$)/i.test(url);
	}

	function searchYoutube() {
		const q = encodeURIComponent(`${song.title} ${song.artist || ''}`);
		window.open(`https://youtube.com/results?search_query=${q}`, '_blank');
	}

	async function deleteSong() {
		if (!confirm(`Delete "${song.title}"? This cannot be undone.`)) return;
		try {
			await api(`/api/services/songs/${songId}`, { method: 'DELETE' });
			goto('/dashboard/services/songs');
		} catch (error) {
			alert('Failed to delete song');
		}
	}
</script>

{#if loading}
	<div class="flex justify-center py-16">
		<div class="animate-spin rounded-full h-8 w-8 border-4 border-teal border-t-transparent"></div>
	</div>
{:else if song}
	<div class="space-y-6 max-w-5xl">
		<!-- Header -->
		<div class="flex items-start justify-between">
			<div>
				<button
					on:click={() => goto('/dashboard/services/songs')}
					class="text-teal hover:underline mb-3 text-sm flex items-center gap-1"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
					Song Library
				</button>

				<!-- Title (editable) -->
				{#if editing === 'title'}
					<input
						bind:value={editValue}
						on:blur={() => saveField('title')}
						on:keydown={(e) => handleEditKeydown(e, 'title')}
						class="text-3xl font-bold bg-[var(--bg)] border border-teal rounded px-2 py-1 text-[var(--text-primary)] w-full focus:outline-none"
						autofocus
					/>
				{:else}
					<h1
						on:click={() => startEdit('title')}
						class="text-3xl font-bold text-[var(--text-primary)] cursor-pointer hover:text-teal transition-colors"
						title="Click to edit"
					>
						{song.title}
					</h1>
				{/if}

				<!-- Artist (editable) -->
				{#if editing === 'artist'}
					<input
						bind:value={editValue}
						on:blur={() => saveField('artist')}
						on:keydown={(e) => handleEditKeydown(e, 'artist')}
						class="text-lg bg-[var(--bg)] border border-teal rounded px-2 py-1 text-[var(--text-secondary)] mt-1 focus:outline-none"
						placeholder="Artist name"
						autofocus
					/>
				{:else}
					<p
						on:click={() => startEdit('artist')}
						class="text-lg text-[var(--text-secondary)] mt-1 cursor-pointer hover:text-[var(--text-primary)] transition-colors"
						title="Click to edit"
					>
						{song.artist || 'Unknown Artist'}
					</p>
				{/if}
			</div>

			<button
				on:click={deleteSong}
				class="text-sm text-red-400 hover:text-red-300 px-3 py-1.5 rounded border border-red-400/30 hover:border-red-400/60 transition-colors"
			>
				Delete Song
			</button>
		</div>

		<!-- Badges Row -->
		<div class="flex flex-wrap gap-3">
			<!-- Key badge -->
			{#if editing === 'default_key'}
				<input
					bind:value={editValue}
					on:blur={() => saveField('default_key')}
					on:keydown={(e) => handleEditKeydown(e, 'default_key')}
					class="w-20 px-2 py-1 bg-[var(--bg)] border border-teal rounded text-sm text-[var(--text-primary)] focus:outline-none"
					placeholder="Key"
					autofocus
				/>
			{:else}
				<button
					on:click={() => startEdit('default_key')}
					class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-teal/15 text-teal text-sm font-semibold hover:bg-teal/25 transition-colors"
					title="Click to edit key"
				>
					<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3"/></svg>
					{song.default_key || 'No Key'}
				</button>
			{/if}

			<!-- Tempo badge -->
			{#if editing === 'tempo'}
				<input
					type="number"
					bind:value={editValue}
					on:blur={() => saveField('tempo')}
					on:keydown={(e) => handleEditKeydown(e, 'tempo')}
					class="w-24 px-2 py-1 bg-[var(--bg)] border border-teal rounded text-sm text-[var(--text-primary)] focus:outline-none"
					placeholder="BPM"
					autofocus
				/>
			{:else}
				<button
					on:click={() => startEdit('tempo')}
					class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-blue-500/15 text-blue-400 text-sm font-semibold hover:bg-blue-500/25 transition-colors"
					title="Click to edit tempo"
				>
					♩ {song.tempo ? `${song.tempo} BPM` : 'No Tempo'}
				</button>
			{/if}

			<!-- CCLI badge -->
			{#if editing === 'ccli_number'}
				<input
					bind:value={editValue}
					on:blur={() => saveField('ccli_number')}
					on:keydown={(e) => handleEditKeydown(e, 'ccli_number')}
					class="w-32 px-2 py-1 bg-[var(--bg)] border border-teal rounded text-sm text-[var(--text-primary)] focus:outline-none font-mono"
					placeholder="CCLI #"
					autofocus
				/>
			{:else}
				<button
					on:click={() => startEdit('ccli_number')}
					class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-purple-500/15 text-purple-400 text-sm font-mono hover:bg-purple-500/25 transition-colors"
					title="Click to edit CCLI number"
				>
					CCLI {song.ccli_number || '—'}
				</button>
			{/if}

			<!-- Stats -->
			<span class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-[var(--surface)] text-[var(--text-secondary)] text-sm border border-[var(--border)]">
				{song.times_used}× used
			</span>
			<span class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-[var(--surface)] text-[var(--text-secondary)] text-sm border border-[var(--border)]">
				Last: {formatDate(song.last_used)}
			</span>
			<span class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-[var(--surface)] text-[var(--text-secondary)] text-sm border border-[var(--border)]">
				Added: {formatDate(song.created_at)}
			</span>
		</div>

		<!-- Tags (editable) -->
		<div class="bg-[var(--surface)] rounded-lg border border-[var(--border)] p-4">
			<div class="flex items-center justify-between mb-2">
				<h3 class="text-xs font-semibold text-[var(--text-secondary)] uppercase tracking-wider">Tags</h3>
				{#if editing !== 'tags'}
					<button on:click={() => startEdit('tags')} class="text-xs text-teal hover:underline">Edit</button>
				{/if}
			</div>
			{#if editing === 'tags'}
				<div>
					<input
						bind:value={editValue}
						on:blur={() => saveField('tags')}
						on:keydown={(e) => handleEditKeydown(e, 'tags')}
						class="w-full px-3 py-2 bg-[var(--bg)] border border-teal rounded-lg text-sm text-[var(--text-primary)] focus:outline-none"
						placeholder="worship, fast, opener (comma-separated)"
						autofocus
					/>
					<p class="text-xs text-[var(--text-secondary)] mt-1">Comma-separated tags. Press Enter to save, Escape to cancel.</p>
				</div>
			{:else}
				<div class="flex flex-wrap gap-2">
					{#if song.tags}
						{#each song.tags.split(',') as tag}
							<span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium {getTagColor(tag.trim())}">
								{tag.trim()}
							</span>
						{/each}
					{:else}
						<span class="text-sm text-[var(--text-secondary)] italic">No tags</span>
					{/if}
				</div>
			{/if}
		</div>

		<!-- Licensing & Attribution (NEW - CCLI metadata) -->
		<div class="bg-[var(--surface)] rounded-lg border border-[var(--border)] p-4">
			<h3 class="text-sm font-semibold text-[var(--text-primary)] mb-4">Licensing &amp; Attribution</h3>
			<div class="space-y-4">
				<!-- Authors -->
				{#if editing === 'authors'}
					<div>
						<label class="block text-xs font-medium text-[var(--text-secondary)] mb-1">Authors / Songwriters</label>
						<input
							bind:value={editValue}
							on:blur={() => saveField('authors')}
							on:keydown={(e) => handleEditKeydown(e, 'authors')}
							class="w-full px-3 py-2 bg-[var(--bg)] border border-teal rounded-lg text-sm text-[var(--text-primary)] focus:outline-none"
							placeholder="John Smith, Jane Doe (comma-separated)"
							autofocus
						/>
					</div>
				{:else}
					<div>
						<label class="block text-xs font-medium text-[var(--text-secondary)] mb-1">Authors / Songwriters</label>
						<div class="flex items-center justify-between">
							{#if song.authors && song.authors.length > 0}
								<span class="text-sm text-[var(--text-primary)]">{song.authors.join(', ')}</span>
							{:else}
								<span class="text-sm text-[var(--text-secondary)] italic">Not specified</span>
							{/if}
							<button on:click={() => startEdit('authors')} class="text-xs text-teal hover:underline ml-2">Edit</button>
						</div>
					</div>
				{/if}

				<!-- Copyright Year -->
				{#if editing === 'copyright_year'}
					<div>
						<label class="block text-xs font-medium text-[var(--text-secondary)] mb-1">Copyright Year</label>
						<input
							type="number"
							bind:value={editValue}
							on:blur={() => saveField('copyright_year')}
							on:keydown={(e) => handleEditKeydown(e, 'copyright_year')}
							class="w-full px-3 py-2 bg-[var(--bg)] border border-teal rounded-lg text-sm text-[var(--text-primary)] focus:outline-none"
							placeholder="YYYY"
							autofocus
						/>
					</div>
				{:else}
					<div>
						<label class="block text-xs font-medium text-[var(--text-secondary)] mb-1">Copyright Year</label>
						<div class="flex items-center justify-between">
							{#if song.copyright_year}
								<span class="text-sm text-[var(--text-primary)]">{song.copyright_year}</span>
							{:else}
								<span class="text-sm text-[var(--text-secondary)] italic">Not specified</span>
							{/if}
							<button on:click={() => startEdit('copyright_year')} class="text-xs text-teal hover:underline ml-2">Edit</button>
						</div>
					</div>
				{/if}

				<!-- Publisher -->
				{#if editing === 'publisher'}
					<div>
						<label class="block text-xs font-medium text-[var(--text-secondary)] mb-1">Publisher</label>
						<input
							bind:value={editValue}
							on:blur={() => saveField('publisher')}
							on:keydown={(e) => handleEditKeydown(e, 'publisher')}
							class="w-full px-3 py-2 bg-[var(--bg)] border border-teal rounded-lg text-sm text-[var(--text-primary)] focus:outline-none"
							placeholder="Publisher name"
							autofocus
						/>
					</div>
				{:else}
					<div>
						<label class="block text-xs font-medium text-[var(--text-secondary)] mb-1">Publisher</label>
						<div class="flex items-center justify-between">
							{#if song.publisher}
								<span class="text-sm text-[var(--text-primary)]">{song.publisher}</span>
							{:else}
								<span class="text-sm text-[var(--text-secondary)] italic">Not specified</span>
							{/if}
							<button on:click={() => startEdit('publisher')} class="text-xs text-teal hover:underline ml-2">Edit</button>
						</div>
					</div>
				{/if}

				<!-- License Type -->
				{#if editing === 'license_type'}
					<div>
						<label class="block text-xs font-medium text-[var(--text-secondary)] mb-1">License Type</label>
						<select
							bind:value={editValue}
							on:blur={() => saveField('license_type')}
							class="w-full px-3 py-2 bg-[var(--bg)] border border-teal rounded-lg text-sm text-[var(--text-primary)] focus:outline-none"
						>
							<option value="">Select license type</option>
							<option value="CCLI">CCLI</option>
							<option value="PD">Public Domain</option>
							<option value="Custom">Custom License</option>
							<option value="ASCAP">ASCAP</option>
							<option value="BMI">BMI</option>
							<option value="SESAC">SESAC</option>
						</select>
					</div>
				{:else}
					<div>
						<label class="block text-xs font-medium text-[var(--text-secondary)] mb-1">License Type</label>
						<div class="flex items-center justify-between">
							{#if song.license_type}
								<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-500/15 text-purple-400">
									{song.license_type}
								</span>
							{:else}
								<span class="text-sm text-[var(--text-secondary)] italic">Not specified</span>
							{/if}
							<button on:click={() => startEdit('license_type')} class="text-xs text-teal hover:underline ml-2">Edit</button>
						</div>
					</div>
				{/if}
			</div>
		</div>

		<!-- Media Player Section -->
		{#if song.youtube_url || song.spotify_url || song.apple_music_url || song.rehearsal_url}
			<div class="bg-[var(--surface)] rounded-lg border border-[var(--border)]">
				<div class="p-4 border-b border-[var(--border)]">
					<h3 class="text-sm font-semibold text-[var(--text-primary)]"><Music2 size={16} class="inline" /> Media & Rehearsal</h3>
				</div>
				<div class="p-4 space-y-4">
					{#if song.youtube_url}
						{@const ytId = extractYoutubeId(song.youtube_url)}
						{#if ytId}
							<div class="aspect-video rounded-lg overflow-hidden bg-black">
								<iframe
									src="https://www.youtube.com/embed/{ytId}"
									class="w-full h-full"
									frameborder="0"
									allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
									allowfullscreen
									title="YouTube video"
								></iframe>
							</div>
						{:else}
							<a href={song.youtube_url} target="_blank" class="inline-flex items-center gap-2 px-3 py-2 rounded-lg bg-red-500/15 text-red-400 hover:bg-red-500/25 text-sm">
								▶ Watch on YouTube
							</a>
						{/if}
					{/if}

					{#if song.spotify_url}
						{@const spotifyId = extractSpotifyId(song.spotify_url)}
						{#if spotifyId}
							<div class="rounded-lg overflow-hidden">
								<iframe
									src="https://open.spotify.com/embed/track/{spotifyId}?theme=0"
									width="100%"
									height="80"
									frameborder="0"
									allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture"
									loading="lazy"
									title="Spotify player"
								></iframe>
							</div>
						{:else}
							<a href={song.spotify_url} target="_blank" class="inline-flex items-center gap-2 px-3 py-2 rounded-lg bg-green-500/15 text-green-400 hover:bg-green-500/25 text-sm">
								🟢 Listen on Spotify
							</a>
						{/if}
					{/if}

					{#if song.apple_music_url}
						<a href={song.apple_music_url} target="_blank" class="inline-flex items-center gap-2 px-3 py-2 rounded-lg bg-pink-500/15 text-pink-400 hover:bg-pink-500/25 text-sm">
							🍎 Listen on Apple Music
						</a>
					{/if}

					{#if song.rehearsal_url}
						{#if isAudioUrl(song.rehearsal_url)}
							<div>
								<p class="text-xs text-[var(--text-secondary)] mb-1">Rehearsal Track</p>
								<audio controls class="w-full" src={song.rehearsal_url}>
									Your browser does not support audio playback.
								</audio>
							</div>
						{:else}
							<a href={song.rehearsal_url} target="_blank" class="inline-flex items-center gap-2 px-3 py-2 rounded-lg bg-amber-500/15 text-amber-400 hover:bg-amber-500/25 text-sm">
								🎧 Rehearsal Track
							</a>
						{/if}
					{/if}
				</div>
			</div>
		{/if}

		<!-- Media URLs (editable) -->
		<div class="bg-[var(--surface)] rounded-lg border border-[var(--border)] p-4">
			<div class="flex items-center justify-between mb-3">
				<h3 class="text-xs font-semibold text-[var(--text-secondary)] uppercase tracking-wider">Media Links</h3>
				<button on:click={searchYoutube} class="text-xs text-red-400 hover:text-red-300 flex items-center gap-1">
					<svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="currentColor"><path d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814z"/><path fill="#fff" d="M9.545 15.568V8.432L15.818 12l-6.273 3.568z"/></svg>
					Search YouTube
				</button>
			</div>
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
				{#each [
					{ field: 'youtube_url', label: 'YouTube URL', icon: '▶', placeholder: 'https://youtube.com/watch?v=...' },
					{ field: 'spotify_url', label: 'Spotify URL', icon: '🟢', placeholder: 'https://open.spotify.com/track/...' },
					{ field: 'apple_music_url', label: 'Apple Music URL', icon: '🍎', placeholder: 'https://music.apple.com/...' },
					{ field: 'rehearsal_url', label: 'Rehearsal Track URL', icon: '🎧', placeholder: 'https://example.com/track.mp3' },
				] as { field, label, icon, placeholder }}
					<div>
						<label class="text-xs text-[var(--text-secondary)] mb-1 block">{icon} {label}</label>
						{#if editing === field}
							<input
								bind:value={editValue}
								on:blur={() => saveField(field)}
								on:keydown={(e) => handleEditKeydown(e, field)}
								class="w-full px-2 py-1.5 bg-[var(--bg)] border border-teal rounded text-xs text-[var(--text-primary)] focus:outline-none font-mono"
								placeholder={placeholder}
								autofocus
							/>
						{:else}
							<button
								on:click={() => startEdit(field)}
								class="w-full text-left px-2 py-1.5 rounded bg-[var(--bg)] border border-[var(--border)] text-xs font-mono truncate hover:border-teal/50 transition-colors {song[field] ? 'text-teal' : 'text-[var(--text-secondary)] italic'}"
								title="Click to edit"
							>
								{song[field] || 'Not set'}
							</button>
						{/if}
					</div>
				{/each}
			</div>
		</div>

		<!-- Chord Chart / Lyrics -->
		<div class="bg-[var(--surface)] rounded-lg border border-[var(--border)]">
			<div class="flex items-center justify-between p-4 border-b border-[var(--border)]">
				<h3 class="text-sm font-semibold text-[var(--text-primary)]">Chord Chart / Lyrics</h3>
				<div class="flex items-center gap-3">
					{#if song.lyrics && song.default_key}
						<button
							on:click={() => showTranspose = !showTranspose}
							class="text-xs px-3 py-1.5 rounded-lg {showTranspose ? 'bg-teal text-white' : 'bg-[var(--bg)] text-[var(--text-secondary)] border border-[var(--border)]'} hover:opacity-80 transition"
						>
							Transpose
						</button>
					{/if}
					{#if editing !== 'lyrics'}
						<button on:click={() => startEdit('lyrics')} class="text-xs text-teal hover:underline">Edit</button>
					{/if}
				</div>
			</div>

			{#if showTranspose && song.default_key}
				<div class="px-4 py-3 border-b border-[var(--border)] bg-[var(--bg)]/50">
					<div class="flex items-center gap-2 flex-wrap">
						<span class="text-xs text-[var(--text-secondary)]">Key:</span>
						{#each ['C', 'Db', 'D', 'Eb', 'E', 'F', 'F#', 'G', 'Ab', 'A', 'Bb', 'B'] as key}
							<button
								on:click={() => transposeKey = key}
								class="px-2 py-0.5 rounded text-xs font-semibold transition-colors {transposeKey === key ? 'bg-teal text-white' : 'bg-[var(--surface)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] border border-[var(--border)]'}"
							>
								{key}
							</button>
						{/each}
						{#if transposeKey !== song.default_key}
							<span class="text-xs text-amber-400 ml-2">
								Transposed from {song.default_key}
							</span>
						{/if}
					</div>
				</div>
			{/if}

			<div class="p-4">
				{#if editing === 'lyrics'}
					<textarea
						bind:value={editValue}
						on:blur={() => saveField('lyrics')}
						on:keydown={(e) => { if (e.key === 'Escape') editing = null; }}
						rows="20"
						class="w-full px-3 py-2 bg-[var(--bg)] border border-teal rounded-lg font-mono text-sm text-[var(--text-primary)] focus:outline-none resize-y"
						autofocus
					></textarea>
					<p class="text-xs text-[var(--text-secondary)] mt-1">Click outside to save. Escape to cancel.</p>
				{:else if song.lyrics}
					<pre class="font-mono text-sm text-[var(--text-primary)] whitespace-pre-wrap leading-relaxed">{displayLyrics}</pre>
				{:else}
					<p class="text-[var(--text-secondary)] italic">No lyrics or chord chart. Click Edit to add.</p>
				{/if}
			</div>
		</div>

		<!-- Notes -->
		<div class="bg-[var(--surface)] rounded-lg border border-[var(--border)] p-4">
			<div class="flex items-center justify-between mb-2">
				<h3 class="text-xs font-semibold text-[var(--text-secondary)] uppercase tracking-wider">Notes</h3>
				{#if editing !== 'notes'}
					<button on:click={() => startEdit('notes')} class="text-xs text-teal hover:underline">Edit</button>
				{/if}
			</div>
			{#if editing === 'notes'}
				<textarea
					bind:value={editValue}
					on:blur={() => saveField('notes')}
					on:keydown={(e) => { if (e.key === 'Escape') editing = null; }}
					rows="4"
					class="w-full px-3 py-2 bg-[var(--bg)] border border-teal rounded-lg text-sm text-[var(--text-primary)] focus:outline-none resize-y"
					autofocus
				></textarea>
			{:else if song.notes}
				<p class="text-sm text-[var(--text-primary)] whitespace-pre-wrap">{song.notes}</p>
			{:else}
				<p class="text-sm text-[var(--text-secondary)] italic">No notes.</p>
			{/if}
		</div>

		<!-- Attachments -->
		<div class="bg-[var(--surface)] rounded-lg border border-[var(--border)]">
			<div class="p-4 border-b border-[var(--border)]">
				<h3 class="text-sm font-semibold text-[var(--text-primary)]">PDF Attachments</h3>
				<p class="text-xs text-[var(--text-secondary)] mt-0.5">Chord charts, sheet music, lead sheets</p>
			</div>

			<div class="p-4 space-y-3">
				{#if attachments.length > 0}
					{#each attachments as att}
						<div class="flex items-center justify-between p-3 rounded-lg bg-[var(--bg)] border border-[var(--border)]">
							<div class="flex items-center gap-3">
								<FileText size={20} />
								<div>
									<a
										href="{import.meta.env.VITE_API_URL || 'http://localhost:8190'}/api/services/songs/attachments/{att.id}"
										target="_blank"
										class="text-sm font-medium text-[var(--text-primary)] hover:text-teal"
									>
										{att.original_name}
									</a>
									<div class="text-xs text-[var(--text-secondary)]">
										{formatFileSize(att.file_size)} • {formatDate(att.created_at)}
									</div>
								</div>
							</div>
							<div class="flex items-center gap-2">
								<a
									href="{import.meta.env.VITE_API_URL || 'http://localhost:8190'}/api/services/songs/attachments/{att.id}"
									target="_blank"
									class="text-xs text-teal hover:underline"
								>
									View
								</a>
								<button
									on:click={() => deleteAttachment(att.id)}
									class="text-xs text-red-400 hover:text-red-300"
								>
									Delete
								</button>
							</div>
						</div>
					{/each}
				{/if}

				<!-- Upload zone -->
				<div
					on:dragover|preventDefault={() => dragOver = true}
					on:dragleave={() => dragOver = false}
					on:drop={handleFileDrop}
					class="border-2 border-dashed rounded-lg p-6 text-center transition-colors {dragOver ? 'border-teal bg-teal/5' : 'border-[var(--border)] hover:border-[var(--text-secondary)]'}"
				>
					{#if uploadingFile}
						<div class="animate-spin rounded-full h-6 w-6 border-4 border-teal border-t-transparent mx-auto mb-2"></div>
						<p class="text-sm text-[var(--text-secondary)]">Uploading...</p>
					{:else}
						<svg class="mx-auto h-8 w-8 text-[var(--text-secondary)] mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/></svg>
						<label class="cursor-pointer">
							<span class="text-sm text-teal font-medium hover:underline">Upload Files</span>
							<span class="text-sm text-[var(--text-secondary)]"> or drag and drop</span>
							<input type="file" accept=".pdf,.png,.jpg,.jpeg,.docx" on:change={handleFileSelect} class="hidden" />
						</label>
						<p class="text-xs text-[var(--text-secondary)] mt-1">PDF, PNG, JPG, DOCX files allowed, max 20MB each</p>
					{/if}
				</div>
			</div>
		</div>

		<!-- Service History -->
		<div class="bg-[var(--surface)] rounded-lg border border-[var(--border)]">
			<div class="p-4 border-b border-[var(--border)]">
				<h3 class="text-sm font-semibold text-[var(--text-primary)]">Service History</h3>
				<p class="text-xs text-[var(--text-secondary)] mt-0.5">Services where this song was used</p>
			</div>

			{#if usageLoading}
				<div class="p-8 flex justify-center">
					<div class="animate-spin rounded-full h-6 w-6 border-4 border-teal border-t-transparent"></div>
				</div>
			{:else if usage.length === 0}
				<div class="p-8 text-center">
					<svg class="mx-auto h-8 w-8 text-[var(--text-secondary)] mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/></svg>
					<p class="text-sm text-[var(--text-secondary)]">Not used in any service yet</p>
				</div>
			{:else}
				<div class="divide-y divide-[var(--border)]">
					{#each usage as item}
						<button
							on:click={() => goto(`/dashboard/services/${item.service_id}`)}
							class="w-full flex items-center justify-between p-4 hover:bg-[var(--surface-hover)] transition-colors text-left"
						>
							<div>
								<div class="text-sm font-medium text-[var(--text-primary)]">{item.service_name || 'Service'}</div>
								<div class="text-xs text-[var(--text-secondary)]">{formatDate(item.service_date)}{item.service_time ? ` • ${item.service_time}` : ''}</div>
							</div>
							<div class="flex items-center gap-3">
								{#if item.song_key}
									<span class="text-xs font-semibold px-2 py-0.5 rounded bg-teal/15 text-teal">{item.song_key}</span>
								{/if}
								<span class="text-xs text-[var(--text-secondary)]">#{item.position}</span>
								<svg class="w-4 h-4 text-[var(--text-secondary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
							</div>
						</button>
					{/each}
				</div>
			{/if}
		</div>
	</div>
{/if}
