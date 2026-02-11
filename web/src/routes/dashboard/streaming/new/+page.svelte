<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let services = [];
	let loading = false;

	let formData = {
		title: '',
		description: '',
		service_id: null,
		stream_type: 'youtube',
		stream_url: '',
		embed_url: '',
		scheduled_start: '',
		chat_enabled: true,
		giving_enabled: true,
		connection_card_enabled: true,
		status: 'scheduled'
	};

	onMount(() => {
		loadServices();
	});

	async function loadServices() {
		try {
			const response = await api('/api/services?limit=100');
			services = response.services || [];
		} catch (error) {
			console.error('Failed to load services:', error);
		}
	}

	async function handleSubmit() {
		if (!formData.title || !formData.embed_url) {
			alert('Please provide a title and embed URL');
			return;
		}

		loading = true;
		try {
			const created = await api('/api/streaming', {
				method: 'POST',
				body: JSON.stringify({
					...formData,
					service_id: formData.service_id || null
				})
			});
			
			goto(`/dashboard/streaming/${created.id}`);
		} catch (error) {
			alert('Failed to create stream: ' + error.message);
		} finally {
			loading = false;
		}
	}

	function extractYouTubeEmbed() {
		if (formData.stream_type !== 'youtube' || !formData.stream_url) return;
		
		// Try to extract video ID from various YouTube URL formats
		const patterns = [
			/(?:youtube\.com\/watch\?v=|youtu\.be\/)([^&\s]+)/,
			/youtube\.com\/embed\/([^&\s]+)/,
			/youtube\.com\/live\/([^&\s]+)/
		];

		for (const pattern of patterns) {
			const match = formData.stream_url.match(pattern);
			if (match && match[1]) {
				formData.embed_url = `https://www.youtube.com/embed/${match[1]}`;
				return;
			}
		}
	}
</script>

<div class="max-w-3xl mx-auto space-y-6">
	<div class="flex items-center gap-4">
		<button
			on:click={() => goto('/dashboard/streaming')}
			class="text-secondary hover:text-primary"
		>
			← Back
		</button>
		<h1 class="text-3xl font-bold" style="color: var(--navy)">Schedule Stream</h1>
	</div>

	<form on:submit|preventDefault={handleSubmit} class="space-y-6 p-6 rounded-lg" style="background-color: var(--surface)">
		<!-- Basic Info -->
		<div>
			<label class="block text-sm font-medium mb-2" style="color: var(--text-primary)">
				Title *
			</label>
			<input
				type="text"
				bind:value={formData.title}
				required
				class="w-full px-3 py-2 border rounded-md"
				style="background-color: var(--bg); border-color: var(--border); color: var(--text-primary)"
				placeholder="Sunday Service - Week 1"
			/>
		</div>

		<div>
			<label class="block text-sm font-medium mb-2" style="color: var(--text-primary)">
				Description
			</label>
			<textarea
				bind:value={formData.description}
				rows="3"
				class="w-full px-3 py-2 border rounded-md"
				style="background-color: var(--bg); border-color: var(--border); color: var(--text-primary)"
				placeholder="Optional description for this stream"
			></textarea>
		</div>

		<!-- Link to Service -->
		<div>
			<label class="block text-sm font-medium mb-2" style="color: var(--text-primary)">
				Link to Service (Optional)
			</label>
			<select
				bind:value={formData.service_id}
				class="w-full px-3 py-2 border rounded-md"
				style="background-color: var(--bg); border-color: var(--border); color: var(--text-primary)"
			>
				<option value={null}>No service linked</option>
				{#each services as service}
					<option value={service.id}>
						{new Date(service.service_date).toLocaleDateString()} - {service.service_type_name || 'Service'}
					</option>
				{/each}
			</select>
		</div>

		<!-- Stream Type -->
		<div>
			<label class="block text-sm font-medium mb-2" style="color: var(--text-primary)">
				Stream Type
			</label>
			<div class="grid grid-cols-2 md:grid-cols-4 gap-2">
				{#each ['youtube', 'facebook', 'vimeo', 'rtmp_custom'] as type}
					<label class="flex items-center gap-2 p-3 border rounded-md cursor-pointer hover:border-teal-500 transition" style="border-color: var(--border)">
						<input
							type="radio"
							bind:group={formData.stream_type}
							value={type}
							class="text-teal-600"
						/>
						<span class="capitalize text-sm" style="color: var(--text-primary)">{type.replace('_', ' ')}</span>
					</label>
				{/each}
			</div>
		</div>

		<!-- Stream URL -->
		{#if formData.stream_type !== 'rtmp_custom'}
			<div>
				<label class="block text-sm font-medium mb-2" style="color: var(--text-primary)">
					Stream URL
				</label>
				<input
					type="url"
					bind:value={formData.stream_url}
					on:blur={extractYouTubeEmbed}
					class="w-full px-3 py-2 border rounded-md"
					style="background-color: var(--bg); border-color: var(--border); color: var(--text-primary)"
					placeholder="https://youtube.com/watch?v=..."
				/>
				<p class="text-xs text-secondary mt-1">
					Paste your YouTube/Facebook live link. We'll extract the embed URL.
				</p>
			</div>
		{/if}

		<!-- Embed URL -->
		<div>
			<label class="block text-sm font-medium mb-2" style="color: var(--text-primary)">
				Embed URL *
			</label>
			<input
				type="url"
				bind:value={formData.embed_url}
				required
				class="w-full px-3 py-2 border rounded-md"
				style="background-color: var(--bg); border-color: var(--border); color: var(--text-primary)"
				placeholder="https://www.youtube.com/embed/VIDEO_ID"
			/>
			<p class="text-xs text-secondary mt-1">
				This is what viewers will see embedded on the watch page.
			</p>
		</div>

		<!-- Schedule Time -->
		<div>
			<label class="block text-sm font-medium mb-2" style="color: var(--text-primary)">
				Scheduled Start Time
			</label>
			<input
				type="datetime-local"
				bind:value={formData.scheduled_start}
				class="w-full px-3 py-2 border rounded-md"
				style="background-color: var(--bg); border-color: var(--border); color: var(--text-primary)"
			/>
		</div>

		<!-- Features -->
		<div>
			<label class="block text-sm font-medium mb-2" style="color: var(--text-primary)">
				Features
			</label>
			<div class="space-y-2">
				<label class="flex items-center gap-2">
					<input
						type="checkbox"
						bind:checked={formData.chat_enabled}
						class="rounded"
					/>
					<span class="text-sm" style="color: var(--text-primary)">Enable live chat</span>
				</label>
				<label class="flex items-center gap-2">
					<input
						type="checkbox"
						bind:checked={formData.giving_enabled}
						class="rounded"
					/>
					<span class="text-sm" style="color: var(--text-primary)">Show "Give Now" button</span>
				</label>
				<label class="flex items-center gap-2">
					<input
						type="checkbox"
						bind:checked={formData.connection_card_enabled}
						class="rounded"
					/>
					<span class="text-sm" style="color: var(--text-primary)">Show connection card prompt</span>
				</label>
			</div>
		</div>

		<!-- Actions -->
		<div class="flex gap-3 pt-4">
			<button
				type="submit"
				disabled={loading}
				class="px-6 py-2 rounded-md text-white hover:opacity-90 disabled:opacity-50"
				style="background-color: var(--teal)"
			>
				{loading ? 'Creating...' : 'Schedule Stream'}
			</button>
			<button
				type="button"
				on:click={() => goto('/dashboard/streaming')}
				class="px-6 py-2 border rounded-md"
				style="border-color: var(--border); color: var(--text-primary)"
			>
				Cancel
			</button>
		</div>
	</form>
</div>
