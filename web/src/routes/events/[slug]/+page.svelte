<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8190';

	let loading = true;
	let error = '';
	let events = [];
	let church = null;

	$: slug = $page.params.slug;

	onMount(async () => {
		try {
			const res = await fetch(`${API_URL}/api/public/events?tenant_slug=${slug}`);
			if (!res.ok) throw new Error('Not found');
			const data = await res.json();
			events = data.events || [];
			church = data.church;
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	const typeColors = {
		service: 'bg-[#4A8B8C]', meeting: 'bg-[#1B3A4B]', class: 'bg-purple-500',
		social: 'bg-amber-500', outreach: 'bg-emerald-500', other: 'bg-gray-500'
	};

	function formatDate(dateStr) {
		const d = new Date(dateStr);
		return d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
	}

	function formatTime(dateStr) {
		const d = new Date(dateStr);
		return d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
	}

	function formatDateFull(dateStr) {
		const d = new Date(dateStr);
		return d.toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' });
	}

	// Group events by date
	$: groupedEvents = events.reduce((acc, event) => {
		const key = new Date(event.start_time).toDateString();
		if (!acc[key]) acc[key] = [];
		acc[key].push(event);
		return acc;
	}, {});
</script>

<svelte:head>
	<title>{church ? `Events - ${church.name}` : 'Events'}</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center min-h-screen">
		<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[#4A8B8C]"></div>
	</div>
{:else if error}
	<div class="flex items-center justify-center min-h-screen">
		<div class="text-center">
			<p class="text-6xl mb-4">📅</p>
			<h1 class="text-2xl font-bold text-[#1B3A4B] mb-2">Not Found</h1>
			<p class="text-gray-500">This page doesn't exist.</p>
		</div>
	</div>
{:else}
	<!-- Header -->
	<div class="bg-gradient-to-r from-[#1B3A4B] to-[#4A8B8C] text-white">
		<div class="max-w-4xl mx-auto px-4 py-12">
			<a href="/church/{slug}" class="text-white/70 hover:text-white text-sm mb-4 inline-block">← Back to {church?.name || 'Church'}</a>
			<h1 class="text-3xl font-bold mb-2">Upcoming Events</h1>
			<p class="text-white/70">See what's happening in the next 30 days</p>
		</div>
	</div>

	<!-- Events -->
	<div class="max-w-4xl mx-auto px-4 py-8">
		{#if events.length === 0}
			<div class="text-center py-16">
				<p class="text-5xl mb-4">📅</p>
				<h2 class="text-xl font-semibold text-[#1B3A4B] mb-2">No upcoming events</h2>
				<p class="text-gray-500">Check back soon!</p>
			</div>
		{:else}
			<div class="space-y-8">
				{#each Object.entries(groupedEvents) as [dateKey, dayEvents]}
					<div>
						<h2 class="text-lg font-bold text-[#1B3A4B] mb-3 sticky top-0 bg-[#F7FAFA] py-2">
							{formatDateFull(dayEvents[0].start_time)}
						</h2>
						<div class="space-y-3">
							{#each dayEvents as event}
								<div class="bg-white rounded-xl shadow-sm border border-gray-100 p-5 hover:shadow-md transition-shadow">
									<div class="flex items-start gap-4">
										<!-- Time block -->
										<div class="text-center min-w-[60px]">
											{#if event.all_day}
												<div class="text-sm font-medium text-[#4A8B8C]">All Day</div>
											{:else}
												<div class="text-sm font-bold text-[#1B3A4B]">{formatTime(event.start_time)}</div>
												<div class="text-xs text-gray-400">{formatTime(event.end_time)}</div>
											{/if}
										</div>

										<!-- Content -->
										<div class="flex-1">
											<div class="flex items-center gap-2 mb-1">
												<h3 class="font-semibold text-[#1B3A4B]">{event.title}</h3>
												<span class="text-xs text-white px-2 py-0.5 rounded-full capitalize {typeColors[event.event_type] || 'bg-gray-500'}">
													{event.event_type}
												</span>
											</div>
											{#if event.description}
												<p class="text-gray-600 text-sm mb-2">{event.description}</p>
											{/if}
											{#if event.location}
												<div class="flex items-center gap-1 text-sm text-gray-500">
													<span>📍</span> {event.location}
												</div>
											{/if}
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Footer -->
	<div class="text-center py-8 text-gray-400 text-sm">
		Powered by <span class="text-[#4A8B8C]">Pews</span>
	</div>
{/if}
