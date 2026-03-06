<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { Timer, User, Music, ChevronRight, Attachment } from 'lucide-svelte';

	let token = '';
	let plan = null;
	let items = [];
	let loading = false;
	let error = null;

	$: token = $page.params.token;

	onMount(async () => {
		await loadPlan();
	});

	async function loadPlan() {
		loading = true;
		error = null;
		try {
			const response = await fetch(`/api/worship/plans/shared/${token}`);
			if (!response.ok) {
				throw new Error('Plan not found or invalid token');
			}
			plan = await response.json();
			items = plan.items || [];
		} catch (err) {
			console.error('Failed to load public plan:', err);
			error = 'This service plan is not available or the link is invalid.';
		} finally {
			loading = false;
		}
	}

	function formatItemType(type) {
		const types = {
			song: 'Song',
			scripture: 'Scripture',
			prayer: 'Prayer',
			announcement: 'Announcement',
			video: 'Video',
			other: 'Other'
		};
		return types[type] || type;
	}

	function printPlan() {
		window.print();
	}
</script>

{#if loading}
	<div class="min-h-screen flex items-center justify-center bg-gray-50">
		<div class="text-center">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-navy mx-auto mb-4"></div>
			<p class="text-gray-600">Loading service plan...</p>
		</div>
	</div>

{:else if error}
	<div class="min-h-screen flex items-center justify-center bg-gray-50 p-4">
		<div class="max-w-md text-center">
			<h1 class="text-2xl font-bold text-navy mb-4">Service Plan Not Available</h1>
			<p class="text-gray-600 mb-6">{error}</p>
			<a href="/" class="inline-block px-6 py-3 bg-navy text-white rounded-md hover:bg-navy-dark">Return Home</a>
		</div>
	</div>

{:else if plan}
	<div class="min-h-screen bg-gray-50 print:bg-white">
		<!-- Printable container -->
		<div class="print-container max-w-3xl mx-auto p-4 md:p-8 bg-white shadow-sm rounded-lg my-4 print:my-0 print:shadow-none print:rounded-none">
			<!-- Header -->
			<div class="border-b border-gray-200 pb-4 mb-6">
				<h1 class="text-3xl font-bold text-navy mb-2">{plan.service_name || 'Service Plan'}</h1>
				<p class="text-gray-600">Published Service Order</p>
			</div>

			<!-- Notes -->
			{#if plan.notes}
				<div class="bg-blue-50 border-l-4 border-blue-500 p-4 mb-6">
					<h3 class="font-semibold text-blue-900 mb-1">Plan Notes</h3>
					<p class="text-blue-800">{plan.notes}</p>
				</div>
			{/if}

			<!-- Items -->
			<div class="space-y-4">
				{#each items as item, i (item.id)}
					<div class="flex gap-4 p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors print:bg-white print:border print:border-gray-300">
						<!-- Order number -->
						<div class="flex-shrink-0 text-center w-12 pt-1">
							<span class="text-3xl font-bold text-gray-400">{item.item_order}</span>
						</div>

						<!-- Content -->
						<div class="flex-1 min-w-0">
							<div class="flex items-start gap-2 mb-2">
								<span class="px-2 py-1 bg-navy text-white text-xs font-medium rounded uppercase tracking-wide whitespace-nowrap">
									{formatItemType(item.item_type)}
								</span>
								<h2 class="text-lg font-semibold text-gray-900 break-words">{item.title}</h2>
							</div>

							<!-- Key display (prominent for songs) -->
							{#if item.key}
								<div class="mb-3 flex items-center gap-2">
									<span class="text-sm text-gray-500">Key:</span>
									<span class="px-3 py-1 bg-green-100 text-green-800 font-mono font-semibold rounded">{item.key}</span>
								</div>
							{/if}

							<!-- Duration and assignee -->
							<div class="flex flex-wrap gap-4 text-sm text-gray-600 mb-2">
								{#if item.duration_minutes}
									<span class="flex items-center gap-1">
										<Timer size={14} />
										<strong>{item.duration_minutes}</strong> min
									</span>
								{/if}
								{#if item.assigned_to_name}
									<span class="flex items-center gap-1">
										<User size={14} />
										{item.assigned_to_name}
									</span>
								{/if}
								{#if item.song_title && item.title !== item.song_title}
									<span class="flex items-center gap-1">
										<Music size={14} />
										{item.song_title}
									</span>
								{/if}
							</div>

							<!-- Notes -->
							{#if item.notes}
								<p class="text-sm text-gray-600 italic mt-2 pl-3 border-l-2 border-gray-300">
									{item.notes}
								</p>
							{/if}

							<!-- Attachments (PDFs, images for songs) -->
							{#if item.attachments && item.attachments.length > 0}
								<div class="mt-4 pt-3 border-t border-gray-200">
									<p class="text-xs font-semibold text-gray-700 mb-2 uppercase tracking-wide">Attachments</p>
									{#each item.attachments as attachment, j}
										<a
											href={`/api/services/songs/attachments/public/${attachment.id}`}
											download
											class="flex items-center gap-2 px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded text-sm text-navy transition-colors"
										>
											<Attachment size={14} />
											<span class="truncate">{attachment.original_name || attachment.filename}</span>
											{#if j < item.attachments.length - 1}
												<br/>
											{/if}
										</a>
									{/each}
								</div>
							{/if}

							<!-- Separator (except for last item) -->
							{#if i < items.length - 1}
								<div class="mt-4 flex items-center gap-2 text-gray-400">
									<ChevronRight size={16} />
									<span class="text-sm">Next</span>
									<ChevronRight size={16} transform="rotate=180" />
								</div>
							{/if}
						</div>
					</div>
				{/each}
			</div>

			<!-- Footer -->
			<div class="mt-8 pt-4 border-t border-gray-200 text-center text-sm text-gray-500">
				<p>Total duration: <strong>{items.reduce((sum, item) => sum + (item.duration_minutes || 0), 0)} minutes</strong></p>
				<p class="mt-1">Generated from Pews Service Planning System</p>
			</div>
		</div>

		<!-- Print button (hidden on print) -->
		<div class="fixed bottom-6 right-6 print:hidden">
			<button
				on:click={printPlan}
				class="px-6 py-3 bg-navy text-white font-medium rounded-full shadow-lg hover:bg-navy-dark transition-all flex items-center gap-2"
			>
				🖨️ Print Plan
			</button>
		</div>
	</div>

	<style>
		@media print {
			body > *:not(.print-container) { display: none !important; }
			.print-container { 
				display: block !important; 
				margin: 0 !important; 
				padding: 20px !important;
			}
			h1 { font-size: 24pt !important; }
			h2 { font-size: 16pt !important; }
			body { background: white !important; }
		}

		.text-navy { color: #1e3a8a; }
		.bg-navy { background-color: #1e3a8a; }
	</style>
{/if}
