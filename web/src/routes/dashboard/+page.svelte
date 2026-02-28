<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { User, Pencil, Trash2, DollarSign, CheckCircle, ClipboardList } from 'lucide-svelte';

	let kpis = null;
	let atRisk = [];
	let activities = [];
	let upcomingEvents = [];
	let loading = true;
	let error = '';

	onMount(async () => {
		try {
			const [kpiData, atRiskData, activityData, eventsData] = await Promise.allSettled([
				api('/api/dashboard/kpis'),
				api('/api/engagement/at-risk'),
				api('/api/dashboard/activity'),
				api('/api/events/upcoming')
			]);
			kpis = kpiData.status === 'fulfilled' ? kpiData.value : null;
			atRisk = atRiskData.status === 'fulfilled' ? atRiskData.value : [];
			activities = activityData.status === 'fulfilled' ? activityData.value : [];
			upcomingEvents = eventsData.status === 'fulfilled' ? (eventsData.value || []).slice(0, 5) : [];
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	});

	function formatCurrency(cents) {
		return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(cents / 100);
	}

	function formatPercent(value) {
		return (value >= 0 ? '+' : '') + value.toFixed(1) + '%';
	}

	function timeAgo(date) {
		const seconds = Math.floor((new Date() - new Date(date)) / 1000);
		if (seconds < 60) return 'just now';
		if (seconds < 3600) return Math.floor(seconds / 60) + 'm ago';
		if (seconds < 86400) return Math.floor(seconds / 3600) + 'h ago';
		return Math.floor(seconds / 86400) + 'd ago';
	}

	function formatEventDate(d) {
		return new Date(d).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
	}

	function getActivityIcon(type) {
		const icons = {
			'user-plus': User, 'user-edit': Pencil, 'user-minus': Trash2,
			'dollar': DollarSign, 'check-circle': CheckCircle, 'activity': ClipboardList
		};
		return icons[type] || ClipboardList;
	}

	const quickActions = [
		{ href: '/dashboard/people', label: 'Add Person', icon: 'M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z', color: '#4A8B8C' },
		{ href: '/dashboard/giving/donations/new', label: 'Record Donation', icon: 'M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z', color: '#10b981' },
		{ href: '/dashboard/services', label: 'Plan Service', icon: 'M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3', color: '#8b5cf6' },
		{ href: '/dashboard/communication', label: 'Send Message', icon: 'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z', color: '#3b82f6' },
	];
</script>

<div class="space-y-8">
	<div>
		<h1 class="text-3xl font-bold text-primary">Dashboard</h1>
		<p class="text-secondary text-sm mt-1">Welcome back. Here's what's happening.</p>
	</div>

	{#if loading}
		<div class="flex justify-center py-16">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else if error}
		<div class="bg-red-500 bg-opacity-10 border border-red-500 text-red-400 px-4 py-3 rounded-xl">{error}</div>
	{:else}
		<!-- KPI Cards -->
		{#if kpis}
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
				<!-- Active Members -->
				<div class="kpi-card group">
					<div class="kpi-accent" style="--accent: #4A8B8C"></div>
					<div class="p-5">
						<div class="flex items-center justify-between">
							<div>
								<p class="text-xs font-medium text-secondary uppercase tracking-wider">Active Members</p>
								<p class="text-3xl font-bold text-primary mt-1">{kpis.total_active_members}</p>
							</div>
							<div class="w-11 h-11 rounded-xl bg-[#4A8B8C] bg-opacity-10 flex items-center justify-center group-hover:scale-110 transition-transform">
								<svg class="w-5 h-5 text-[#4A8B8C]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
								</svg>
							</div>
						</div>
					</div>
				</div>

				<!-- Avg Attendance -->
				<div class="kpi-card group">
					<div class="kpi-accent" style="--accent: #8FBCB0"></div>
					<div class="p-5">
						<div class="flex items-center justify-between">
							<div>
								<p class="text-xs font-medium text-secondary uppercase tracking-wider">Avg Attendance</p>
								<p class="text-3xl font-bold text-primary mt-1">{Math.round(kpis.average_attendance)}</p>
								<p class="text-xs text-secondary mt-1">4-week average</p>
							</div>
							<div class="w-11 h-11 rounded-xl bg-[#8FBCB0] bg-opacity-10 flex items-center justify-center group-hover:scale-110 transition-transform">
								<svg class="w-5 h-5 text-[#8FBCB0]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
								</svg>
							</div>
						</div>
					</div>
				</div>

				<!-- Giving This Month -->
				<div class="kpi-card group">
					<div class="kpi-accent" style="--accent: #10b981"></div>
					<div class="p-5">
						<div class="flex items-center justify-between">
							<div>
								<p class="text-xs font-medium text-secondary uppercase tracking-wider">Giving This Month</p>
								<p class="text-3xl font-bold text-primary mt-1">{formatCurrency(kpis.giving_this_month_cents)}</p>
								<p class="text-xs mt-1" class:text-green-400={kpis.giving_percent_change >= 0} class:text-red-400={kpis.giving_percent_change < 0}>
									{formatPercent(kpis.giving_percent_change)} vs last month
								</p>
							</div>
							<div class="w-11 h-11 rounded-xl bg-green-500 bg-opacity-10 flex items-center justify-center group-hover:scale-110 transition-transform">
								<svg class="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
								</svg>
							</div>
						</div>
					</div>
				</div>

				<!-- New Visitors -->
				<div class="kpi-card group">
					<div class="kpi-accent" style="--accent: #3b82f6"></div>
					<div class="p-5">
						<div class="flex items-center justify-between">
							<div>
								<p class="text-xs font-medium text-secondary uppercase tracking-wider">New Visitors</p>
								<p class="text-3xl font-bold text-primary mt-1">{kpis.new_visitors_this_month}</p>
								<p class="text-xs text-secondary mt-1">This month</p>
							</div>
							<div class="w-11 h-11 rounded-xl bg-blue-500 bg-opacity-10 flex items-center justify-center group-hover:scale-110 transition-transform">
								<svg class="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
								</svg>
							</div>
						</div>
					</div>
				</div>
			</div>
		{/if}

		<!-- Quick Actions -->
		<div>
			<h2 class="text-lg font-semibold text-primary mb-3">Quick Actions</h2>
			<div class="grid grid-cols-2 lg:grid-cols-4 gap-3">
				{#each quickActions as action}
					<a href={action.href}
						class="bg-surface border border-custom rounded-xl p-4 hover:border-[var(--teal)] hover:shadow-md transition-all flex items-center gap-3 group">
						<div class="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 group-hover:scale-110 transition-transform"
							style="background-color: {action.color}15">
							<svg class="w-5 h-5" style="color: {action.color}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={action.icon} />
							</svg>
						</div>
						<span class="text-sm font-medium text-primary">{action.label}</span>
					</a>
				{/each}
			</div>
		</div>

		<!-- Activity Feed + Sidebar -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Recent Activity -->
			<div class="lg:col-span-2 bg-surface rounded-xl shadow-sm border border-custom p-6">
				<h2 class="text-lg font-semibold text-primary mb-4">Recent Activity</h2>
				{#if activities.length === 0}
					<div class="text-center py-12">
						<svg class="w-12 h-12 mx-auto mb-3 text-secondary opacity-40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<p class="text-secondary">No recent activity yet.</p>
						<p class="text-xs text-secondary mt-1">Activity will appear here as people are added, donations recorded, and check-ins happen.</p>
					</div>
				{:else}
					<div class="space-y-3">
						{#each activities as activity}
							<a href={activity.link || '#'} class="flex items-start gap-3 p-3 rounded-lg hover:bg-[var(--surface-hover)] transition-colors">
								<span class="flex-shrink-0 mt-0.5"><svelte:component this={getActivityIcon(activity.icon)} size={20} /></span>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium text-primary">{activity.title}</p>
									<p class="text-xs text-secondary truncate">{activity.description}</p>
								</div>
								<span class="text-xs text-secondary whitespace-nowrap flex-shrink-0">{timeAgo(activity.timestamp)}</span>
							</a>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Sidebar: At-Risk + Upcoming -->
			<div class="space-y-6">
				<!-- At-Risk Members -->
				<div class="bg-surface rounded-xl shadow-sm border border-custom p-6">
					<div class="flex items-center justify-between mb-3">
						<h2 class="text-lg font-semibold text-primary">At-Risk</h2>
						{#if atRisk.length > 0}
							<span class="px-2 py-0.5 text-xs font-medium rounded-full bg-red-500 bg-opacity-15 text-red-400">{atRisk.length}</span>
						{/if}
					</div>
					{#if atRisk.length === 0}
						<div class="text-center py-6">
							<svg class="w-10 h-10 mx-auto mb-2 text-green-500 opacity-60" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
							<p class="text-sm text-secondary">No at-risk members</p>
						</div>
					{:else}
						<div class="space-y-2">
							{#each atRisk.slice(0, 5) as person}
								<div class="flex items-center justify-between p-2 rounded-lg bg-[var(--surface-hover)]">
									<div class="min-w-0">
										<p class="text-sm font-medium text-primary truncate">{person.person_name}</p>
										<p class="text-xs text-secondary truncate">{person.person_email || ''}</p>
									</div>
									<span class="text-xs font-semibold text-red-400 flex-shrink-0">{person.percent_change.toFixed(0)}%</span>
								</div>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Upcoming Events -->
				<div class="bg-surface rounded-xl shadow-sm border border-custom p-6">
					<h2 class="text-lg font-semibold text-primary mb-3">Upcoming Events</h2>
					{#if upcomingEvents.length === 0}
						<div class="text-center py-6">
							<svg class="w-10 h-10 mx-auto mb-2 text-secondary opacity-40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
							</svg>
							<p class="text-sm text-secondary">No upcoming events.</p>
							<a href="/dashboard/calendar" class="text-xs text-[var(--teal)] hover:underline mt-1 inline-block">Create an event →</a>
						</div>
					{:else}
						<div class="space-y-2">
							{#each upcomingEvents as event}
								<a href="/dashboard/calendar" class="flex items-center gap-3 p-2 rounded-lg hover:bg-[var(--surface-hover)] transition-colors">
									<div class="w-10 h-10 rounded-lg bg-[var(--teal)] bg-opacity-10 flex items-center justify-center flex-shrink-0">
										<span class="text-xs font-bold text-[var(--teal)]">{new Date(event.start_time || event.date).getDate()}</span>
									</div>
									<div class="min-w-0">
										<p class="text-sm font-medium text-primary truncate">{event.title || event.name}</p>
										<p class="text-xs text-secondary">{formatEventDate(event.start_time || event.date)}</p>
									</div>
								</a>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Engagement Chart placeholder -->
				{#if kpis && kpis.engagement_distribution}
					{@const dist = kpis.engagement_distribution}
					{@const total = (dist.high || 0) + (dist.medium || 0) + (dist.low || 0) + (dist.inactive || 0)}
					<div class="bg-surface rounded-xl shadow-sm border border-custom p-6">
						<h2 class="text-lg font-semibold text-primary mb-3">Engagement</h2>
						{#if total === 0}
							<div class="text-center py-6">
								<p class="text-sm text-secondary">No engagement data yet.</p>
								<a href="/dashboard/checkins" class="text-xs text-[var(--teal)] hover:underline mt-1 inline-block">Set up check-ins →</a>
							</div>
						{:else}
							<div class="space-y-2">
								{#each [
									{ label: 'High', count: dist.high, color: '#4A8B8C' },
									{ label: 'Medium', count: dist.medium, color: '#8FBCB0' },
									{ label: 'Low', count: dist.low, color: '#f59e0b' },
									{ label: 'Inactive', count: dist.inactive, color: '#ef4444' },
								] as seg}
									<div class="flex items-center gap-2">
										<div class="w-2 h-2 rounded-full flex-shrink-0" style="background: {seg.color}"></div>
										<span class="text-xs text-secondary flex-1">{seg.label}</span>
										<span class="text-xs font-medium text-primary">{seg.count}</span>
										<div class="w-16 bg-[var(--surface-hover)] rounded-full h-1.5">
											<div class="h-1.5 rounded-full" style="width: {(seg.count / total) * 100}%; background: {seg.color}"></div>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	.kpi-card {
		position: relative;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: 0.75rem;
		overflow: hidden;
		transition: all 0.2s;
	}
	.kpi-card:hover {
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
		transform: translateY(-1px);
	}
	.kpi-accent {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 3px;
		background: var(--accent);
	}
</style>
