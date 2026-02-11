<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let loading = true;
	let error = '';
	
	// Stats data
	let totalMembers = 0;
	let donationStatsThisMonth = 0;
	let upcomingServicesCount = 0;
	let activeGroupsCount = 0;
	
	// Module data
	let modules = [];
	let enabledModules = new Set();
	
	// Activity data
	let recentDonations = [];
	let recentCheckins = [];
	let upcomingServices = [];
	
	// Loading states
	let statsLoaded = false;
	let activitiesLoaded = false;

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		loading = true;
		error = '';
		
		try {
			// Load modules first to know what's enabled
			modules = await api('/api/tenant/modules');
			enabledModules = new Set(
				modules.filter(m => m.enabled).map(m => m.name)
			);
			
			// Load stats in parallel
			await Promise.all([
				loadStats(),
				loadActivities()
			]);
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	async function loadStats() {
		try {
			// People count
			if (enabledModules.has('people')) {
				try {
					const peopleData = await api('/api/people?per_page=1');
					totalMembers = peopleData.total || 0;
				} catch (e) {
					console.error('Failed to load people stats:', e);
				}
			}
			
			// Giving stats
			if (enabledModules.has('giving')) {
				try {
					const givingStats = await api('/api/giving/stats');
					donationStatsThisMonth = givingStats.total_this_month || 0;
				} catch (e) {
					console.error('Failed to load giving stats:', e);
				}
			}
			
			// Groups count
			if (enabledModules.has('groups')) {
				try {
					const groupsData = await api('/api/groups');
					activeGroupsCount = groupsData.filter(g => g.is_active).length;
				} catch (e) {
					console.error('Failed to load groups stats:', e);
				}
			}
			
			// Upcoming services
			if (enabledModules.has('services')) {
				try {
					const servicesData = await api('/api/services/upcoming');
					upcomingServices = servicesData.slice(0, 5);
					upcomingServicesCount = servicesData.length;
				} catch (e) {
					console.error('Failed to load services stats:', e);
					upcomingServices = [];
				}
			}
			
			statsLoaded = true;
		} catch (err) {
			console.error('Stats loading error:', err);
		}
	}

	async function loadActivities() {
		try {
			// Recent donations
			if (enabledModules.has('giving')) {
				try {
					const donationsData = await api('/api/giving/donations?per_page=5');
					recentDonations = donationsData.donations || [];
				} catch (e) {
					console.error('Failed to load recent donations:', e);
				}
			}
			
			// Today's check-ins
			if (enabledModules.has('checkins')) {
				try {
					const checkinsStats = await api('/api/checkins/stats');
					// Stats endpoint only returns counts, we'd need a separate endpoint for actual checkins
					// For now, just show the count
				} catch (e) {
					console.error('Failed to load checkin stats:', e);
				}
			}
			
			activitiesLoaded = true;
		} catch (err) {
			console.error('Activities loading error:', err);
		}
	}

	function formatCurrency(cents) {
		if (cents === 0) return '$0.00';
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(cents / 100);
	}

	function formatDate(dateString) {
		const date = new Date(dateString);
		return new Intl.DateTimeFormat('en-US', {
			weekday: 'short',
			month: 'short',
			day: 'numeric',
			hour: 'numeric',
			minute: '2-digit'
		}).format(date);
	}

	function formatDateOnly(dateString) {
		const date = new Date(dateString);
		return new Intl.DateTimeFormat('en-US', {
			weekday: 'short',
			month: 'short',
			day: 'numeric'
		}).format(date);
	}

	function formatRelativeTime(dateString) {
		const date = new Date(dateString);
		const now = new Date();
		const diffMs = now - date;
		const diffMins = Math.floor(diffMs / 60000);
		const diffHours = Math.floor(diffMins / 60);
		const diffDays = Math.floor(diffHours / 24);

		if (diffMins < 1) return 'Just now';
		if (diffMins < 60) return `${diffMins}m ago`;
		if (diffHours < 24) return `${diffHours}h ago`;
		if (diffDays < 7) return `${diffDays}d ago`;
		return formatDateOnly(dateString);
	}

	// Quick actions
	function handleAddMember() {
		goto('/dashboard/people?action=add');
	}

	function handleRecordDonation() {
		goto('/dashboard/giving?action=add');
	}

	function handlePlanService() {
		goto('/dashboard/services?action=add');
	}

	// Check if church is brand new (no data)
	$: isEmpty = statsLoaded && 
		totalMembers === 0 && 
		donationStatsThisMonth === 0 && 
		activeGroupsCount === 0 && 
		upcomingServicesCount === 0;
</script>

<div class="max-w-7xl mx-auto">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-6 gap-4">
		<h1 class="text-3xl font-bold text-primary">Dashboard</h1>
		
		{#if !loading && !isEmpty}
			<div class="flex flex-wrap gap-2">
				{#if enabledModules.has('people')}
					<button
						on:click={handleAddMember}
						class="btn-primary text-sm sm:text-base px-3 py-2 sm:px-4"
					>
						+ Add Member
					</button>
				{/if}
				{#if enabledModules.has('giving')}
					<button
						on:click={handleRecordDonation}
						class="btn-secondary text-sm sm:text-base px-3 py-2 sm:px-4"
					>
						+ Record Donation
					</button>
				{/if}
				{#if enabledModules.has('services')}
					<button
						on:click={handlePlanService}
						class="btn-secondary text-sm sm:text-base px-3 py-2 sm:px-4"
					>
						+ Plan Service
					</button>
				{/if}
			</div>
		{/if}
	</div>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
			<p class="mt-4 text-secondary">Loading dashboard...</p>
		</div>
	{:else if error}
		<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg">
			<strong>Error:</strong> {error}
		</div>
	{:else if isEmpty}
		<!-- Empty State - New Church -->
		<div class="bg-surface rounded-lg shadow-md p-8 border border-[var(--sage)] text-center">
			<div class="max-w-2xl mx-auto">
				<div class="text-6xl mb-4">🎉</div>
				<h2 class="text-2xl font-bold text-primary mb-4">Welcome to Pews!</h2>
				<p class="text-secondary mb-8">
					Your church management system is ready to go. Here are some steps to get started:
				</p>
				
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6 text-left">
					{#if enabledModules.has('people')}
						<div class="bg-[var(--background)] p-6 rounded-lg border border-custom">
							<div class="flex items-start gap-3">
								<div class="text-2xl">👥</div>
								<div class="flex-1">
									<h3 class="font-semibold text-primary mb-2">Add Your First Members</h3>
									<p class="text-sm text-secondary mb-3">
										Start building your church directory by adding members.
									</p>
									<button on:click={handleAddMember} class="btn-primary text-sm w-full">
										Add Member
									</button>
								</div>
							</div>
						</div>
					{/if}
					
					{#if enabledModules.has('services')}
						<div class="bg-[var(--background)] p-6 rounded-lg border border-custom">
							<div class="flex items-start gap-3">
								<div class="text-2xl">📅</div>
								<div class="flex-1">
									<h3 class="font-semibold text-primary mb-2">Plan Your Services</h3>
									<p class="text-sm text-secondary mb-3">
										Schedule your upcoming services and plan worship elements.
									</p>
									<button on:click={handlePlanService} class="btn-primary text-sm w-full">
										Plan Service
									</button>
								</div>
							</div>
						</div>
					{/if}
					
					{#if enabledModules.has('groups')}
						<div class="bg-[var(--background)] p-6 rounded-lg border border-custom">
							<div class="flex items-start gap-3">
								<div class="text-2xl">🤝</div>
								<div class="flex-1">
									<h3 class="font-semibold text-primary mb-2">Create Groups</h3>
									<p class="text-sm text-secondary mb-3">
										Set up small groups, ministries, or teams for your church.
									</p>
									<a href="/dashboard/groups" class="btn-primary text-sm w-full block text-center">
										Create Group
									</a>
								</div>
							</div>
						</div>
					{/if}
					
					{#if enabledModules.has('giving')}
						<div class="bg-[var(--background)] p-6 rounded-lg border border-custom">
							<div class="flex items-start gap-3">
								<div class="text-2xl">💰</div>
								<div class="flex-1">
									<h3 class="font-semibold text-primary mb-2">Setup Giving</h3>
									<p class="text-sm text-secondary mb-3">
										Configure donation funds and start tracking contributions.
									</p>
									<a href="/dashboard/giving" class="btn-primary text-sm w-full block text-center">
										Configure Giving
									</a>
								</div>
							</div>
						</div>
					{/if}
				</div>

				<div class="mt-8 pt-6 border-t border-custom">
					<p class="text-sm text-secondary">
						Need help? Check out our <a href="/docs" class="text-[var(--teal)] hover:underline">documentation</a>
						or enable additional modules in <a href="/dashboard/settings" class="text-[var(--teal)] hover:underline">settings</a>.
					</p>
				</div>
			</div>
		</div>
	{:else}
		<!-- Stats Cards -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6 mb-8">
			{#if enabledModules.has('people')}
				<div class="bg-surface rounded-lg shadow-md p-6 border border-custom hover:border-[var(--teal)] transition-colors">
					<div class="flex items-center justify-between mb-2">
						<h3 class="text-sm font-medium text-secondary">Total Members</h3>
						<span class="text-2xl">👥</span>
					</div>
					<p class="text-3xl font-bold text-primary">{totalMembers.toLocaleString()}</p>
					<a href="/dashboard/people" class="text-xs text-[var(--teal)] hover:underline mt-2 inline-block">
						View all →
					</a>
				</div>
			{/if}

			{#if enabledModules.has('giving')}
				<div class="bg-surface rounded-lg shadow-md p-6 border border-custom hover:border-[var(--sage)] transition-colors">
					<div class="flex items-center justify-between mb-2">
						<h3 class="text-sm font-medium text-secondary">Donations This Month</h3>
						<span class="text-2xl">💰</span>
					</div>
					<p class="text-3xl font-bold text-primary">{formatCurrency(donationStatsThisMonth)}</p>
					<a href="/dashboard/giving" class="text-xs text-[var(--teal)] hover:underline mt-2 inline-block">
						View details →
					</a>
				</div>
			{/if}

			{#if enabledModules.has('services')}
				<div class="bg-surface rounded-lg shadow-md p-6 border border-custom hover:border-[var(--coral)] transition-colors">
					<div class="flex items-center justify-between mb-2">
						<h3 class="text-sm font-medium text-secondary">Upcoming Services</h3>
						<span class="text-2xl">📅</span>
					</div>
					<p class="text-3xl font-bold text-primary">{upcomingServicesCount}</p>
					<a href="/dashboard/services" class="text-xs text-[var(--teal)] hover:underline mt-2 inline-block">
						Manage services →
					</a>
				</div>
			{/if}

			{#if enabledModules.has('groups')}
				<div class="bg-surface rounded-lg shadow-md p-6 border border-custom hover:border-[var(--sand)] transition-colors">
					<div class="flex items-center justify-between mb-2">
						<h3 class="text-sm font-medium text-secondary">Active Groups</h3>
						<span class="text-2xl">🤝</span>
					</div>
					<p class="text-3xl font-bold text-primary">{activeGroupsCount}</p>
					<a href="/dashboard/groups" class="text-xs text-[var(--teal)] hover:underline mt-2 inline-block">
						Manage groups →
					</a>
				</div>
			{/if}
		</div>

		<!-- Main Content Grid -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Left Column - Upcoming Services -->
			{#if enabledModules.has('services') && upcomingServices.length > 0}
				<div class="lg:col-span-2 bg-surface rounded-lg shadow-md p-6 border border-custom">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-xl font-semibold text-primary">Upcoming Services</h2>
						<a href="/dashboard/services" class="text-sm text-[var(--teal)] hover:underline">
							View all
						</a>
					</div>
					<div class="space-y-3">
						{#each upcomingServices as service}
							<a 
								href="/dashboard/services/{service.id}"
								class="block p-4 bg-[var(--background)] rounded-lg border border-custom hover:border-[var(--teal)] transition-colors"
							>
								<div class="flex items-start justify-between gap-4">
									<div class="flex-1 min-w-0">
										<h3 class="font-medium text-primary truncate">{service.name}</h3>
										<p class="text-sm text-secondary mt-1">
											{formatDate(service.scheduled_at)}
										</p>
										{#if service.service_type_name}
											<span class="inline-block mt-2 text-xs px-2 py-1 bg-[var(--teal)] bg-opacity-10 text-[var(--teal)] rounded">
												{service.service_type_name}
											</span>
										{/if}
									</div>
									<div class="text-right text-sm text-secondary whitespace-nowrap">
										{#if service.item_count > 0}
											<div>{service.item_count} items</div>
										{/if}
										{#if service.team_count > 0}
											<div>{service.team_count} team members</div>
										{/if}
									</div>
								</div>
							</a>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Right Column - Recent Activity -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom {!enabledModules.has('services') || upcomingServices.length === 0 ? 'lg:col-span-3' : ''}">
				<h2 class="text-xl font-semibold text-primary mb-4">Recent Activity</h2>
				
				{#if recentDonations.length === 0}
					<div class="text-center py-8 text-secondary text-sm">
						<p>No recent activity yet.</p>
						<p class="mt-2">Activity will appear here as things happen.</p>
					</div>
				{:else}
					<div class="space-y-3">
						{#each recentDonations as donation}
							<div class="flex items-start gap-3 p-3 bg-[var(--background)] rounded-lg border border-custom">
								<div class="text-xl">💰</div>
								<div class="flex-1 min-w-0">
									<div class="flex items-start justify-between gap-2">
										<p class="text-sm font-medium text-primary">
											{donation.person_name || 'Anonymous'} donated
										</p>
										<span class="text-sm font-semibold text-[var(--teal)] whitespace-nowrap">
											{formatCurrency(donation.amount_cents)}
										</span>
									</div>
									<p class="text-xs text-secondary mt-1">
										{donation.fund_name} • {formatRelativeTime(donation.donated_at)}
									</p>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Enable More Modules CTA -->
		{#if modules.filter(m => !m.enabled).length > 0}
			<div class="mt-8 bg-gradient-to-r from-[var(--teal)] to-[var(--sage)] bg-opacity-10 border border-[var(--sage)] rounded-lg p-6">
				<div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
					<div>
						<h2 class="text-lg font-semibold text-primary mb-2">Unlock More Features</h2>
						<p class="text-secondary text-sm">
							Enable additional modules to access more powerful tools for managing your church.
						</p>
					</div>
					<a
						href="/dashboard/settings"
						class="btn-primary whitespace-nowrap"
					>
						Manage Modules
					</a>
				</div>
			</div>
		{/if}
	{/if}
</div>

<style>
	.btn-primary {
		@apply bg-[var(--teal)] text-white py-2 px-6 rounded-lg font-medium hover:opacity-90 transition-opacity;
	}

	.btn-secondary {
		@apply bg-[var(--surface-hover)] text-primary py-2 px-4 rounded-lg font-medium hover:bg-[var(--sage)] hover:bg-opacity-20 transition-colors;
	}
</style>
