<script>
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8190';

	$: streamId = $page.params.id;

	let stream = null;
	let chatMessages = [];
	let notes = '';
	let guestName = '';
	let chatInput = '';
	let viewerId = null;
	let showConnectionCard = false;
	
	// Connection card
	let connectionCard = {
		first_name: '',
		email: '',
		prayer_request: ''
	};

	let chatPollInterval;
	let isAuthenticated = false;

	onMount(async () => {
		// Check if user is authenticated
		const token = localStorage.getItem('token');
		isAuthenticated = !!token;

		// If not authenticated, ask for guest name
		if (!isAuthenticated) {
			guestName = localStorage.getItem('guest_name') || '';
			if (!guestName) {
				guestName = prompt('Enter your name to join the chat:') || 'Guest';
				localStorage.setItem('guest_name', guestName);
			}
		}

		await loadStream();
		await loadChat();
		await joinStream();

		// Poll chat every 3 seconds
		chatPollInterval = setInterval(async () => {
			await loadChat();
		}, 3000);

		// Handle page unload
		window.addEventListener('beforeunload', handleLeave);
	});

	onDestroy(() => {
		if (chatPollInterval) clearInterval(chatPollInterval);
		handleLeave();
	});

	async function loadStream() {
		try {
			const response = await fetch(`${API_URL}/api/streaming/watch/${streamId}`);
			if (!response.ok) throw new Error('Stream not found');
			stream = await response.json();
		} catch (error) {
			console.error('Failed to load stream:', error);
			alert('Stream not found');
		}
	}

	async function loadChat() {
		try {
			const lastMsgId = chatMessages.length > 0 ? chatMessages[chatMessages.length - 1].id : '';
			const url = lastMsgId 
				? `${API_URL}/api/streaming/${streamId}/chat?after=${lastMsgId}&limit=50`
				: `${API_URL}/api/streaming/${streamId}/chat?limit=50`;
			
			const response = await fetch(url);
			if (!response.ok) throw new Error('Failed to load chat');
			
			const data = await response.json();
			if (data.messages && data.messages.length > 0) {
				chatMessages = [...chatMessages, ...data.messages];
				
				// Auto-scroll chat
				setTimeout(() => {
					const chatContainer = document.getElementById('chat-container');
					if (chatContainer) {
						chatContainer.scrollTop = chatContainer.scrollHeight;
					}
				}, 100);
			}
		} catch (error) {
			console.error('Failed to load chat:', error);
		}
	}

	async function joinStream() {
		try {
			const response = await fetch(`${API_URL}/api/streaming/${streamId}/join`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ guest_name: guestName })
			});
			
			if (!response.ok) throw new Error('Failed to join stream');
			const data = await response.json();
			viewerId = data.id;
		} catch (error) {
			console.error('Failed to join stream:', error);
		}
	}

	async function handleLeave() {
		if (!viewerId) return;
		
		try {
			await fetch(`${API_URL}/api/streaming/${streamId}/leave`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ viewer_id: viewerId })
			});
		} catch (error) {
			console.error('Failed to leave stream:', error);
		}
	}

	async function sendMessage() {
		if (!chatInput.trim()) return;

		try {
			const response = await fetch(`${API_URL}/api/streaming/${streamId}/chat`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					guest_name: guestName,
					message: chatInput
				})
			});

			if (!response.ok) throw new Error('Failed to send message');
			
			chatInput = '';
			await loadChat();
		} catch (error) {
			console.error('Failed to send message:', error);
			alert('Failed to send message');
		}
	}

	async function submitConnectionCard() {
		if (!connectionCard.first_name || !connectionCard.email) {
			alert('Please provide at least your name and email');
			return;
		}

		try {
			const response = await fetch(`${API_URL}/api/communication/cards`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					first_name: connectionCard.first_name,
					email: connectionCard.email,
					message: connectionCard.prayer_request,
					source: 'live_stream'
				})
			});

			if (!response.ok) throw new Error('Failed to submit');
			
			alert('Thank you! We\'ll be in touch soon.');
			showConnectionCard = false;
			connectionCard = { first_name: '', email: '', prayer_request: '' };
		} catch (error) {
			console.error('Failed to submit connection card:', error);
			alert('Failed to submit. Please try again.');
		}
	}

	function formatTime(timestamp) {
		const date = new Date(timestamp);
		return date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
	}
</script>

{#if stream}
	<div class="max-w-7xl mx-auto">
		<!-- Video Section -->
		<div class="w-full bg-black">
			<div class="aspect-video">
				<iframe
					src={stream.embed_url}
					title={stream.title}
					class="w-full h-full"
					frameborder="0"
					allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
					allowfullscreen
				></iframe>
			</div>
		</div>

		<!-- Content Below Video -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 p-6">
			<!-- Left: Stream Info & Notes -->
			<div class="lg:col-span-2 space-y-6">
				<!-- Stream Info -->
				<div>
					<h1 class="text-3xl font-bold text-white mb-2">{stream.title}</h1>
					{#if stream.description}
						<p class="text-gray-400 mb-4">{stream.description}</p>
					{/if}
					<div class="flex items-center gap-4 text-sm text-gray-500">
						<span>👥 {stream.viewer_count} watching</span>
						{#if stream.status === 'live'}
							<span class="px-2 py-1 bg-red-600 text-white text-xs font-bold rounded">LIVE</span>
						{/if}
					</div>
				</div>

				<!-- Sermon Notes -->
				<div class="p-4 bg-gray-900 rounded-lg">
					<h3 class="text-lg font-bold text-white mb-3">📝 Sermon Notes</h3>
					<textarea
						bind:value={notes}
						rows="8"
						class="w-full px-3 py-2 bg-gray-800 border border-gray-700 text-white rounded-md focus:border-teal-500 focus:outline-none"
						placeholder="Take notes during the service..."
					></textarea>
					<p class="text-xs text-gray-500 mt-2">Your notes are saved locally in your browser.</p>
				</div>

				<!-- Give Now Button -->
				{#if stream.giving_enabled}
					<a
						href="/giving"
						class="block w-full px-6 py-3 bg-teal-600 hover:bg-teal-700 text-white text-center font-bold rounded-md"
					>
						💚 Give Now
					</a>
				{/if}

				<!-- Connection Card -->
				{#if stream.connection_card_enabled}
					<div class="p-4 bg-yellow-900 bg-opacity-30 border border-yellow-700 rounded-lg">
						<h3 class="text-lg font-bold text-yellow-400 mb-2">👋 First time here?</h3>
						<p class="text-gray-300 mb-3">We'd love to connect with you!</p>
						<button
							on:click={() => showConnectionCard = true}
							class="px-6 py-2 bg-yellow-600 hover:bg-yellow-700 text-white font-bold rounded-md"
						>
							Let us know you're here
						</button>
					</div>
				{/if}
			</div>

			<!-- Right: Live Chat -->
			{#if stream.chat_enabled}
				<div class="bg-gray-900 rounded-lg flex flex-col h-[600px]">
					<div class="p-4 border-b border-gray-700">
						<h3 class="text-lg font-bold text-white">💬 Live Chat</h3>
					</div>

					<!-- Messages -->
					<div id="chat-container" class="flex-1 overflow-y-auto p-4 space-y-3">
						{#each chatMessages as msg}
							<div class="text-sm p-2 rounded" class:bg-yellow-900={msg.is_pinned} class:bg-opacity-30={msg.is_pinned}>
								{#if msg.is_pinned}
									<div class="text-xs text-yellow-400 mb-1">📌 Pinned</div>
								{/if}
								<div class="flex justify-between items-start mb-1">
									<span class="font-medium text-teal-400">{msg.guest_name || 'Member'}</span>
									<span class="text-xs text-gray-500">{formatTime(msg.created_at)}</span>
								</div>
								<p class="text-gray-200">{msg.message}</p>
							</div>
						{:else}
							<p class="text-center text-gray-500 py-8">No messages yet. Be the first to say hello!</p>
						{/each}
					</div>

					<!-- Chat Input -->
					<div class="p-4 border-t border-gray-700">
						<form on:submit|preventDefault={sendMessage} class="flex gap-2">
							<input
								type="text"
								bind:value={chatInput}
								placeholder="Send a message..."
								class="flex-1 px-3 py-2 bg-gray-800 border border-gray-700 text-white rounded-md focus:border-teal-500 focus:outline-none"
							/>
							<button
								type="submit"
								class="px-4 py-2 bg-teal-600 hover:bg-teal-700 text-white font-bold rounded-md"
							>
								Send
							</button>
						</form>
					</div>
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<div class="text-center py-6 text-gray-500 text-sm">
			Powered by <span class="font-bold text-teal-500">Pews</span>
		</div>
	</div>

	<!-- Connection Card Modal -->
	{#if showConnectionCard}
		<div class="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center p-4 z-50" on:click={() => showConnectionCard = false} role="button" tabindex="0" on:keydown={(e) => e.key === 'Escape' && (showConnectionCard = false)}>
			<div class="bg-gray-900 rounded-lg p-6 max-w-md w-full" on:click|stopPropagation role="button" tabindex="0" on:keydown={(e) => e.key === 'Enter' && submitConnectionCard()}>
				<h2 class="text-2xl font-bold text-white mb-4">Connect With Us</h2>
				<form on:submit|preventDefault={submitConnectionCard} class="space-y-4">
					<div>
						<label class="block text-sm font-medium text-gray-300 mb-1">First Name *</label>
						<input
							type="text"
							bind:value={connectionCard.first_name}
							required
							class="w-full px-3 py-2 bg-gray-800 border border-gray-700 text-white rounded-md"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-300 mb-1">Email *</label>
						<input
							type="email"
							bind:value={connectionCard.email}
							required
							class="w-full px-3 py-2 bg-gray-800 border border-gray-700 text-white rounded-md"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-300 mb-1">Prayer Request (Optional)</label>
						<textarea
							bind:value={connectionCard.prayer_request}
							rows="3"
							class="w-full px-3 py-2 bg-gray-800 border border-gray-700 text-white rounded-md"
						></textarea>
					</div>
					<div class="flex gap-3">
						<button
							type="submit"
							class="flex-1 px-6 py-2 bg-teal-600 hover:bg-teal-700 text-white font-bold rounded-md"
						>
							Submit
						</button>
						<button
							type="button"
							on:click={() => showConnectionCard = false}
							class="px-6 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md"
						>
							Cancel
						</button>
					</div>
				</form>
			</div>
		</div>
	{/if}
{:else}
	<div class="min-h-screen flex items-center justify-center">
		<p class="text-gray-500">Loading stream...</p>
	</div>
{/if}
