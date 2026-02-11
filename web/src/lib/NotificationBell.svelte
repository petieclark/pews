<script>
	import { onMount, onDestroy } from 'svelte';
	import { api } from './api';

	let unreadCount = 0;
	let notifications = [];
	let showDropdown = false;
	let pollInterval;

	onMount(() => {
		fetchUnreadCount();
		fetchNotifications();
		
		// Poll for new notifications every 30 seconds
		pollInterval = setInterval(() => {
			fetchUnreadCount();
			if (showDropdown) {
				fetchNotifications();
			}
		}, 30000);
	});

	onDestroy(() => {
		if (pollInterval) {
			clearInterval(pollInterval);
		}
	});

	async function fetchUnreadCount() {
		try {
			const data = await api('/api/notifications/unread-count');
			unreadCount = data.count;
		} catch (err) {
			console.error('Failed to fetch unread count:', err);
		}
	}

	async function fetchNotifications() {
		try {
			const data = await api('/api/notifications?limit=10');
			notifications = data.notifications || [];
		} catch (err) {
			console.error('Failed to fetch notifications:', err);
		}
	}

	async function markAsRead(id) {
		try {
			await api(`/api/notifications/${id}/read`, { method: 'PUT' });
			await fetchUnreadCount();
			await fetchNotifications();
		} catch (err) {
			console.error('Failed to mark as read:', err);
		}
	}

	async function markAllAsRead() {
		try {
			await api('/api/notifications/read-all', { method: 'PUT' });
			unreadCount = 0;
			await fetchNotifications();
		} catch (err) {
			console.error('Failed to mark all as read:', err);
		}
	}

	function toggleDropdown() {
		showDropdown = !showDropdown;
		if (showDropdown) {
			fetchNotifications();
		}
	}

	function handleNotificationClick(notification) {
		if (!notification.read) {
			markAsRead(notification.id);
		}
		if (notification.link) {
			window.location.href = notification.link;
		}
		showDropdown = false;
	}

	function formatTime(timestamp) {
		const date = new Date(timestamp);
		const now = new Date();
		const diffMs = now - date;
		const diffMins = Math.floor(diffMs / 60000);
		const diffHours = Math.floor(diffMins / 60);
		const diffDays = Math.floor(diffHours / 24);

		if (diffMins < 1) return 'Just now';
		if (diffMins < 60) return `${diffMins}m ago`;
		if (diffHours < 24) return `${diffHours}h ago`;
		if (diffDays < 7) return `${diffDays}d ago`;
		return date.toLocaleDateString();
	}

	function getTypeColor(type) {
		switch (type) {
			case 'success': return 'text-green-600';
			case 'warning': return 'text-yellow-600';
			case 'info': return 'text-blue-600';
			default: return 'text-gray-600';
		}
	}
</script>

<div class="relative">
	<button
		on:click={toggleDropdown}
		class="relative p-2 text-secondary hover:text-primary transition-colors"
		aria-label="Notifications"
	>
		<!-- Bell Icon -->
		<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
		</svg>
		
		<!-- Badge -->
		{#if unreadCount > 0}
			<span class="absolute top-0 right-0 inline-flex items-center justify-center px-2 py-1 text-xs font-bold leading-none text-white bg-red-600 rounded-full">
				{unreadCount > 9 ? '9+' : unreadCount}
			</span>
		{/if}
	</button>

	<!-- Dropdown -->
	{#if showDropdown}
		<div class="absolute right-0 mt-2 w-96 bg-surface rounded-lg shadow-lg border border-custom z-50">
			<div class="p-4 border-b border-custom flex justify-between items-center">
				<h3 class="font-semibold text-[var(--text-primary)]">Notifications</h3>
				{#if unreadCount > 0}
					<button
						on:click={markAllAsRead}
						class="text-xs text-blue-600 hover:text-blue-800"
					>
						Mark all read
					</button>
				{/if}
			</div>

			<div class="max-h-96 overflow-y-auto">
				{#if notifications.length === 0}
					<div class="p-8 text-center text-secondary">
						<p>No notifications</p>
					</div>
				{:else}
					{#each notifications as notification}
						<button
							on:click={() => handleNotificationClick(notification)}
							class="w-full text-left p-4 border-b border-custom hover:bg-[var(--bg)] transition-colors {notification.read ? 'opacity-60' : ''}"
						>
							<div class="flex items-start space-x-3">
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium text-[var(--text-primary)] {getTypeColor(notification.type)}">
										{notification.title}
									</p>
									<p class="text-sm text-secondary mt-1">
										{notification.message}
									</p>
									<p class="text-xs text-secondary mt-2">
										{formatTime(notification.created_at)}
									</p>
								</div>
								{#if !notification.read}
									<div class="w-2 h-2 bg-blue-600 rounded-full flex-shrink-0 mt-1"></div>
								{/if}
							</div>
						</button>
					{/each}
				{/if}
			</div>
		</div>
	{/if}
</div>

<!-- Click outside to close -->
{#if showDropdown}
	<button
		class="fixed inset-0 z-40"
		on:click={() => showDropdown = false}
		aria-label="Close notifications"
	></button>
{/if}

<style>
	/* Custom scrollbar for notification list */
	.overflow-y-auto::-webkit-scrollbar {
		width: 6px;
	}
	.overflow-y-auto::-webkit-scrollbar-track {
		background: transparent;
	}
	.overflow-y-auto::-webkit-scrollbar-thumb {
		background: var(--border);
		border-radius: 3px;
	}
</style>
