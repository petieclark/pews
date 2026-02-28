<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { Building2, Users, CalendarDays } from 'lucide-svelte';

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8190';

	let loading = true;
	let error = '';
	let church = null;
	let serviceTimes = [];

	$: slug = $page.params.slug;

	onMount(async () => {
		try {
			const res = await fetch(`${API_URL}/api/public/church?tenant_slug=${slug}`);
			if (!res.ok) throw new Error('Church not found');
			const data = await res.json();
			church = data.church;
			serviceTimes = data.service_times || [];
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	function formatAddress(c) {
		const parts = [c.address_line1, c.address_line2, c.city, c.state, c.zip].filter(Boolean);
		if (c.city && c.state) {
			return [c.address_line1, c.address_line2, `${c.city}, ${c.state} ${c.zip}`].filter(Boolean).join('\n');
		}
		return parts.join(', ');
	}
</script>

<svelte:head>
	<title>{church ? church.name : 'Church'}</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center min-h-screen">
		<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[#4A8B8C]"></div>
	</div>
{:else if error}
	<div class="flex items-center justify-center min-h-screen">
		<div class="text-center">
			<p class="mb-4"><Building2 size={64} /></p>
			<h1 class="text-2xl font-bold text-[#1B3A4B] mb-2">Church Not Found</h1>
			<p class="text-gray-500">The page you're looking for doesn't exist.</p>
		</div>
	</div>
{:else}
	<!-- Hero -->
	<div class="bg-gradient-to-br from-[#1B3A4B] to-[#4A8B8C] text-white">
		<div class="max-w-4xl mx-auto px-4 py-16 text-center">
			{#if church.logo}
				<img src={church.logo} alt={church.name} class="w-24 h-24 rounded-full mx-auto mb-6 object-cover border-4 border-white/20" />
			{:else}
				<div class="w-24 h-24 rounded-full mx-auto mb-6 bg-white/10 flex items-center justify-center"><Building2 size={40} /></div>
			{/if}
			<h1 class="text-4xl font-bold mb-4">{church.name}</h1>
			{#if church.about}
				<p class="text-lg text-white/80 max-w-2xl mx-auto leading-relaxed">{church.about}</p>
			{/if}
		</div>
	</div>

	<!-- Quick Links -->
	<div class="max-w-4xl mx-auto px-4 -mt-8">
		<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
			<a href="/groups/{slug}" class="bg-white rounded-xl shadow-md p-6 text-center hover:shadow-lg transition-shadow group">
				<div class="mb-2 group-hover:scale-110 transition-transform"><Users size={28} /></div>
				<div class="font-semibold text-[#1B3A4B]">Groups</div>
				<div class="text-sm text-gray-500">Find community</div>
			</a>
			<a href="/events/{slug}" class="bg-white rounded-xl shadow-md p-6 text-center hover:shadow-lg transition-shadow group">
				<div class="mb-2 group-hover:scale-110 transition-transform"><CalendarDays size={28} /></div>
				<div class="font-semibold text-[#1B3A4B]">Events</div>
				<div class="text-sm text-gray-500">What's happening</div>
			</a>
			<a href="/connect?tenant={slug}" class="bg-white rounded-xl shadow-md p-6 text-center hover:shadow-lg transition-shadow group">
				<div class="text-3xl mb-2 group-hover:scale-110 transition-transform">💬</div>
				<div class="font-semibold text-[#1B3A4B]">Connect</div>
				<div class="text-sm text-gray-500">Get in touch</div>
			</a>
			<a href="/giving-kiosk?tenant={slug}" class="bg-white rounded-xl shadow-md p-6 text-center hover:shadow-lg transition-shadow group">
				<div class="text-3xl mb-2 group-hover:scale-110 transition-transform">💝</div>
				<div class="font-semibold text-[#1B3A4B]">Give</div>
				<div class="text-sm text-gray-500">Support the mission</div>
			</a>
		</div>
	</div>

	<!-- Service Times & Info -->
	<div class="max-w-4xl mx-auto px-4 py-12">
		<div class="grid md:grid-cols-2 gap-8">
			{#if serviceTimes.length > 0}
				<div class="bg-white rounded-xl shadow-sm p-8">
					<h2 class="text-xl font-bold text-[#1B3A4B] mb-6 flex items-center gap-2">
						🕐 Service Times
					</h2>
					<div class="space-y-4">
						{#each serviceTimes as st}
							<div class="flex items-center justify-between py-3 border-b border-gray-100 last:border-0">
								<div>
									<div class="font-medium text-[#1B3A4B]">{st.name}</div>
									{#if st.default_day}
										<div class="text-sm text-gray-500">{st.default_day}</div>
									{/if}
								</div>
								{#if st.default_time}
									<div class="text-[#4A8B8C] font-semibold">{st.default_time}</div>
								{/if}
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<div class="bg-white rounded-xl shadow-sm p-8">
				<h2 class="text-xl font-bold text-[#1B3A4B] mb-6 flex items-center gap-2">
					📍 Location & Contact
				</h2>
				<div class="space-y-4">
					{#if church.address_line1}
						<div>
							<div class="text-sm text-gray-500 mb-1">Address</div>
							<div class="text-[#1B3A4B] whitespace-pre-line">{formatAddress(church)}</div>
						</div>
					{/if}
					{#if church.phone}
						<div>
							<div class="text-sm text-gray-500 mb-1">Phone</div>
							<a href="tel:{church.phone}" class="text-[#4A8B8C] hover:underline">{church.phone}</a>
						</div>
					{/if}
					{#if church.email}
						<div>
							<div class="text-sm text-gray-500 mb-1">Email</div>
							<a href="mailto:{church.email}" class="text-[#4A8B8C] hover:underline">{church.email}</a>
						</div>
					{/if}
					{#if church.website}
						<div>
							<div class="text-sm text-gray-500 mb-1">Website</div>
							<a href={church.website} target="_blank" class="text-[#4A8B8C] hover:underline">{church.website}</a>
						</div>
					{/if}
				</div>
			</div>
		</div>
	</div>

	<!-- Footer -->
	<div class="text-center py-8 text-gray-400 text-sm">
		Powered by <span class="text-[#4A8B8C]">Pews</span>
	</div>
{/if}
