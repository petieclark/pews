<script>
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	$: streamId = $page.params.id;

	let stream = null;
	let chatMessages = [];
	let viewers = [];
	let loading = false;
	let chatPollInterval;

	onMount(() => {
		loadStream();
		loadChat();
		loadViewers();
		
		// Poll chat every 3 seconds
		chatPollInterval = setInterval(() => {
			loadChat();
			if (stream?.status === 'live') {
				loadViewers();
			}
		}, 3000);
	});

	onDestroy(() => {
		if (chatPollInterval) clearInterval(chatPollInterval);
	});

	async function loadStream() {
		try {
			stream = await api(`/api/streaming/${streamId}`);
		} catch (error) {
			console.error('Failed to load stream:', error);
			alert('Stream not found');
			goto('/dashboard/streaming');
		}
	}

	async function loadChat() {
		try {
			const response = await api(`/api/streaming/${streamId}/chat?limit=100`);
			chatMessages = response.messages || [];
		} catch (error) {
			console.error('Failed to load chat:', error);
		}
	}

	async function loadViewers() {
		try {
			const response = await api(`/api/streaming/${streamId}/viewers`);
			viewers = response.viewers || [];
		} catch (error) {
			console.error('Failed to load viewers:', error);
		}
	}

	async function goLive() {
		if (!confirm('Mark this stream as LIVE?')) return;
		
		loading = true;
		try {
			stream = await api(`/api/streaming/${streamId}/go-live`, { method: 'POST' });
			alert('Stream is now LIVE!');
		} catch (error) {
			alert('Failed to go live: ' + error.message);
		} finally {
			loading = false;
		}
	}

	async function endStream() {
		if (!confirm('End this stream?')) return;
		
		loading = true;
		try {
			stream = await api(`/api/streaming/${streamId}/end`, { method: 'POST' });
			alert('Stream ended.');
		} catch (error) {
			alert('Failed to end stream: ' + error.message);
		} finally {
			loading = false;
		}
	}

	async function pinMessage(msgId) {
		try {
			await api(`/api/streaming/${streamId}/chat/${msgId}/pin`, { method: 'PUT' });
			await loadChat();
		} catch (error) {
			alert('Failed to pin message: ' + error.message);
		}
	}

	async function deleteMessage(msgId) {
		if (!confirm('Delete this message?')) return;
		
		try {
			await api(`/api/streaming/${streamId}/chat/${msgId}`, { method: 'DELETE' });
			await loadChat();
		} catch (error) {
			alert('Failed to delete message: ' + error.message);
		}
	}

	async function deleteStream() {
		if (!confirm('Delete this stream? This cannot be undone.')) return;
		
		loading = true;
		try {
			await api(`/api/streaming/${streamId}`, { method: 'DELETE' });
			goto('/dashboard/streaming');
		} catch (error) {
			alert('Failed to delete stream: ' + error.message);
			loading = false;
		}
	}

	function formatTime(timestamp) {
		const date = new Date(timestamp);
		return date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
	}

	function getStatusBadge(status) {
		const badges = {
			scheduled: 'bg-blue-100 text-blue-800',
			live: 'bg-red-100 text-red-800 animate-pulse',
			ended: 'bg-gray-100 text-gray-800',
			archived: 'bg-gray-100 text-gray-600'
		};
		return badges[status] || 'bg-gray-100 text-gray-800';
	}
</script>

{#if stream}
	<div class="space-y-6">
		<!-- Header -->
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-4">
				<button
					on:click={() => goto('/dashboard/streaming')}
					class="text-secondary hover:text-primary"
				>
					← Back
				</button>
				<div>
					<h1 class="text-3xl font-bold" style="color: var(--navy)">{stream.title}</h1>
					<p class="text-secondary">{stream.description || 'No description'}</p>
				</div>
			</div>
			<span class="px-3 py-1 rounded text-sm font-medium {getStatusBadge(stream.status)}">
				{stream.status.toUpperCase()}
			</span>
		</div>

		<!-- Actions -->
		<div class="flex gap-3">
			{#if stream.status === 'scheduled'}
				<button
					on:click={goLive}
					disabled={loading}
					class="px-6 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50"
				>
					🔴 Go Live
				</button>
			{/if}
			
			{#if stream.status === 'live'}
				<button
					on:click={endStream}
					disabled={loading}
					class="px-6 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 disabled:opacity-50"
				>
					End Stream
				</button>
			{/if}

			<a
				href="/watch/{streamId}"
				target="_blank"
				class="px-6 py-2 rounded-md text-white"
				style="background-color: var(--teal)"
			>
				View as Viewer
			</a>

			<button
				on:click={() => goto(`/dashboard/streaming/${streamId}/edit`)}
				class="px-6 py-2 border rounded-md"
				style="border-color: var(--border); color: var(--text-primary)"
			>
				Edit
			</button>

			<button
				on:click={deleteStream}
				disabled={loading}
				class="px-6 py-2 border border-red-500 text-red-600 rounded-md hover:bg-red-50 disabled:opacity-50 ml-auto"
			>
				Delete
			</button>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Stream Preview (2/3) -->
			<div class="lg:col-span-2 space-y-4">
				<div class="rounded-lg overflow-hidden" style="background-color: var(--surface)">
					<div class="aspect-video bg-black">
						<iframe
							src={stream.embed_url}
							title="Stream preview"
							class="w-full h-full"
							frameborder="0"
							allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
							allowfullscreen
						></iframe>
					</div>
				</div>

				<!-- Stream Stats -->
				<div class="p-4 rounded-lg grid grid-cols-3 gap-4" style="background-color: var(--surface)">
					<div>
						<p class="text-sm text-secondary">Current Viewers</p>
						<p class="text-2xl font-bold" style="color: var(--text-primary)">{stream.viewer_count}</p>
					</div>
					<div>
						<p class="text-sm text-secondary">Peak Viewers</p>
						<p class="text-2xl font-bold" style="color: var(--text-primary)">{stream.peak_viewers}</p>
					</div>
					<div>
						<p class="text-sm text-secondary">Chat Messages</p>
						<p class="text-2xl font-bold" style="color: var(--text-primary)">{chatMessages.length}</p>
					</div>
				</div>
			</div>

			<!-- Chat Moderation (1/3) -->
			<div class="space-y-4">
				<div class="p-4 rounded-lg" style="background-color: var(--surface)">
					<h3 class="font-bold mb-3" style="color: var(--text-primary)">Live Chat</h3>
					<div class="space-y-2 max-h-96 overflow-y-auto">
						{#each chatMessages as msg}
							<div class="p-2 rounded border" style="border-color: var(--border)" class:bg-yellow-50={msg.is_pinned}>
								<div class="flex justify-between items-start mb-1">
									<span class="text-sm font-medium" style="color: var(--text-primary)">
										{msg.guest_name || 'Member'}
									</span>
									<span class="text-xs text-secondary">{formatTime(msg.created_at)}</span>
								</div>
								<p class="text-sm mb-2" style="color: var(--text-primary)">{msg.message}</p>
								<div class="flex gap-2">
									{#if !msg.is_pinned}
										<button
											on:click={() => pinMessage(msg.id)}
											class="text-xs text-blue-600 hover:underline"
										>
											Pin
										</button>
									{/if}
									<button
										on:click={() => deleteMessage(msg.id)}
										class="text-xs text-red-600 hover:underline"
									>
										Delete
									</button>
								</div>
							</div>
						{:else}
							<p class="text-sm text-secondary text-center py-4">No chat messages yet</p>
						{/each}
					</div>
				</div>

				<!-- Active Viewers -->
				{#if stream.status === 'live'}
					<div class="p-4 rounded-lg" style="background-color: var(--surface)">
						<h3 class="font-bold mb-3" style="color: var(--text-primary)">Active Viewers ({viewers.length})</h3>
						<div class="space-y-1 max-h-48 overflow-y-auto">
							{#each viewers as viewer}
								<div class="text-sm" style="color: var(--text-primary)">
									{viewer.guest_name || 'Anonymous'}
								</div>
							{:else}
								<p class="text-sm text-secondary text-center py-4">No viewers</p>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>
{:else}
	<div class="text-center py-12">
		<p class="text-secondary">Loading stream...</p>
	</div>
{/if}
