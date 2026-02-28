<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { PartyPopper, Users, BookOpen, Heart, Building2, User, HeartHandshake, Globe, Music } from 'lucide-svelte';

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8190';

	let loading = true;
	let error = '';
	let groups = [];
	let church = null;

	// Signup modal state
	let showModal = false;
	let selectedGroup = null;
	let submitting = false;
	let submitted = false;
	let submitError = '';
	let form = { first_name: '', last_name: '', email: '', phone: '', interest: '' };

	$: slug = $page.params.slug;

	onMount(async () => {
		try {
			const res = await fetch(`${API_URL}/api/public/groups?tenant_slug=${slug}`);
			if (!res.ok) throw new Error('Not found');
			const data = await res.json();
			groups = data.groups || [];
			church = data.church;
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	function openSignup(group) {
		selectedGroup = group;
		showModal = true;
		submitted = false;
		submitError = '';
		form = { first_name: '', last_name: '', email: '', phone: '', interest: '' };
	}

	function closeModal() {
		showModal = false;
		selectedGroup = null;
	}

	async function handleSignup() {
		if (!form.first_name || !form.last_name || !form.email) {
			submitError = 'Please fill in all required fields.';
			return;
		}
		submitting = true;
		submitError = '';
		try {
			const res = await fetch(`${API_URL}/api/public/groups/${selectedGroup.id}/signup`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(form)
			});
			if (!res.ok) {
				const data = await res.json();
				throw new Error(data.error || 'Signup failed');
			}
			submitted = true;
		} catch (e) {
			submitError = e.message;
		} finally {
			submitting = false;
		}
	}

	const typeIcons = {
		'small_group': Users, 'bible_study': BookOpen, 'prayer': Heart, 'service': Building2,
		'youth': Users, 'men': User, 'women': User, 'couples': Heart,
		'recovery': HeartHandshake, 'outreach': Globe, 'worship': Music
	};

	function getTypeIcon(type) {
		return typeIcons[type] || Users;
	}

	function formatDay(day) {
		if (!day) return '';
		return day.charAt(0).toUpperCase() + day.slice(1) + 's';
	}
</script>

<svelte:head>
	<title>{church ? `Groups - ${church.name}` : 'Find a Group'}</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center min-h-screen">
		<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[#4A8B8C]"></div>
	</div>
{:else if error}
	<div class="flex items-center justify-center min-h-screen">
		<div class="text-center">
			<p class="mb-4"><Users size={64} /></p>
			<h1 class="text-2xl font-bold text-[#1B3A4B] mb-2">Not Found</h1>
			<p class="text-gray-500">This page doesn't exist.</p>
		</div>
	</div>
{:else}
	<!-- Header -->
	<div class="bg-gradient-to-r from-[#1B3A4B] to-[#4A8B8C] text-white">
		<div class="max-w-5xl mx-auto px-4 py-12">
			<a href="/church/{slug}" class="text-white/70 hover:text-white text-sm mb-4 inline-block">← Back to {church?.name || 'Church'}</a>
			<h1 class="text-3xl font-bold mb-2">Find a Group</h1>
			<p class="text-white/70">Connect with others in a small group community</p>
		</div>
	</div>

	<!-- Groups Grid -->
	<div class="max-w-5xl mx-auto px-4 py-8">
		{#if groups.length === 0}
			<div class="text-center py-16">
				<p class="mb-4"><Users size={48} /></p>
				<h2 class="text-xl font-semibold text-[#1B3A4B] mb-2">No groups available right now</h2>
				<p class="text-gray-500">Check back soon for new group opportunities!</p>
			</div>
		{:else}
			<div class="grid md:grid-cols-2 gap-6">
				{#each groups as group}
					<div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden hover:shadow-md transition-shadow">
						<div class="p-6">
							<div class="flex items-start justify-between mb-3">
								<div class="flex items-center gap-3">
									{#if group.photo_url}
										<img src={group.photo_url} alt={group.name} class="w-12 h-12 rounded-lg object-cover" />
									{:else}
										<div class="w-12 h-12 rounded-lg bg-[#8FBCB0]/20 flex items-center justify-center">
											<svelte:component this={getTypeIcon(group.group_type)} size={24} />
										</div>
									{/if}
									<div>
										<h3 class="font-bold text-[#1B3A4B] text-lg">{group.name}</h3>
										<span class="text-xs bg-[#4A8B8C]/10 text-[#4A8B8C] px-2 py-0.5 rounded-full capitalize">{group.group_type.replace('_', ' ')}</span>
									</div>
								</div>
							</div>

							{#if group.description}
								<p class="text-gray-600 text-sm mb-4 line-clamp-2">{group.description}</p>
							{/if}

							<div class="space-y-2 text-sm text-gray-500 mb-4">
								{#if group.meeting_day || group.meeting_time}
									<div class="flex items-center gap-2">
										<span>🕐</span>
										<span>{[formatDay(group.meeting_day), group.meeting_time].filter(Boolean).join(' at ')}</span>
									</div>
								{/if}
								{#if group.meeting_location}
									<div class="flex items-center gap-2">
										<span>📍</span>
										<span>{group.meeting_location}</span>
									</div>
								{/if}
								<div class="flex items-center gap-2">
									<User size={14} />
									<span>{group.member_count} member{group.member_count !== 1 ? 's' : ''}</span>
									{#if group.leader_name}
										<span class="text-gray-300">·</span>
										<span>Led by {group.leader_name}</span>
									{/if}
								</div>
							</div>

							<button
								on:click={() => openSignup(group)}
								class="w-full bg-[#4A8B8C] text-white py-2.5 rounded-lg font-medium hover:bg-[#3d7374] transition-colors"
							>
								Sign Up
							</button>
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

<!-- Signup Modal -->
{#if showModal}
	<!-- svelte-ignore a11y-click-events-have-key-events -->
	<!-- svelte-ignore a11y-no-static-element-interactions -->
	<div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" on:click={closeModal}>
		<div class="bg-white rounded-2xl w-full max-w-md shadow-2xl" on:click|stopPropagation>
			{#if submitted}
				<div class="p-8 text-center">
					<div class="mb-4"><PartyPopper size={48} /></div>
					<h2 class="text-2xl font-bold text-[#1B3A4B] mb-2">You're signed up!</h2>
					<p class="text-gray-600 mb-6">Thanks! The group leader will reach out soon.</p>
					<button on:click={closeModal} class="bg-[#4A8B8C] text-white px-8 py-2.5 rounded-lg font-medium hover:bg-[#3d7374]">
						Done
					</button>
				</div>
			{:else}
				<div class="p-6 border-b border-gray-100">
					<div class="flex items-center justify-between">
						<h2 class="text-xl font-bold text-[#1B3A4B]">Join {selectedGroup?.name}</h2>
						<button on:click={closeModal} class="text-gray-400 hover:text-gray-600 text-2xl leading-none">&times;</button>
					</div>
				</div>
				<form on:submit|preventDefault={handleSignup} class="p-6 space-y-4">
					{#if submitError}
						<div class="bg-red-50 text-red-600 text-sm p-3 rounded-lg">{submitError}</div>
					{/if}

					<div class="grid grid-cols-2 gap-3">
						<div>
							<label for="fname" class="block text-sm font-medium text-gray-700 mb-1">First Name *</label>
							<input id="fname" bind:value={form.first_name} required
								class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent outline-none" />
						</div>
						<div>
							<label for="lname" class="block text-sm font-medium text-gray-700 mb-1">Last Name *</label>
							<input id="lname" bind:value={form.last_name} required
								class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent outline-none" />
						</div>
					</div>

					<div>
						<label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email *</label>
						<input id="email" type="email" bind:value={form.email} required
							class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent outline-none" />
					</div>

					<div>
						<label for="phone" class="block text-sm font-medium text-gray-700 mb-1">Phone</label>
						<input id="phone" type="tel" bind:value={form.phone}
							class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent outline-none" />
					</div>

					<div>
						<label for="interest" class="block text-sm font-medium text-gray-700 mb-1">Why are you interested?</label>
						<textarea id="interest" bind:value={form.interest} rows="3" placeholder="Tell us a bit about yourself..."
							class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-[#4A8B8C] focus:border-transparent outline-none resize-none"></textarea>
					</div>

					<button type="submit" disabled={submitting}
						class="w-full bg-[#4A8B8C] text-white py-2.5 rounded-lg font-medium hover:bg-[#3d7374] transition-colors disabled:opacity-50">
						{submitting ? 'Signing up...' : 'Sign Up'}
					</button>
				</form>
			{/if}
		</div>
	</div>
{/if}
