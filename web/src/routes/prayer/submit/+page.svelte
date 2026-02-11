<script lang="ts">
	import { goto } from '$app/navigation';

	let name = '';
	let email = '';
	let requestText = '';
	let isPublic = false;
	let submitting = false;
	let success = false;

	async function submit() {
		if (!name || !requestText) {
			alert('Name and prayer request are required');
			return;
		}

		submitting = true;
		try {
			const response = await fetch('/api/prayer-requests', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'X-Tenant-ID': getTenantId()
				},
				body: JSON.stringify({ name, email: email || null, request_text: requestText, is_public: isPublic })
			});

			if (response.ok) {
				success = true;
				setTimeout(() => goto('/prayer'), 2000);
			} else {
				alert('Failed to submit');
			}
		} catch (error) {
			alert('Error occurred');
		}
		submitting = false;
	}

	function getTenantId() {
		return localStorage.getItem('tenant_id') || window.location.hostname.split('.')[0];
	}
</script>

<svelte:head><title>Submit Prayer Request</title></svelte:head>

<div class="min-h-screen bg-gradient-to-b from-[#1B3A4B] to-[#4A8B8C] py-12 px-4">
	<div class="max-w-2xl mx-auto">
		<div class="text-center mb-8">
			<h1 class="text-4xl font-bold text-white mb-4">Submit a Prayer Request</h1>
			<p class="text-xl text-white/90">We're here to pray with you</p>
		</div>

		<div class="bg-white rounded-lg shadow-xl p-8">
			{#if success}
				<div class="text-center py-12">
					<div class="mb-6"><svg class="w-20 h-20 mx-auto text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg></div>
					<h2 class="text-2xl font-bold text-[#1B3A4B] mb-4">Submitted!</h2>
					<p class="text-gray-600">Thank you. We'll be praying for you.</p>
				</div>
			{:else}
				<form on:submit|preventDefault={submit}>
					<div class="mb-6">
						<label class="block text-sm font-medium text-gray-700 mb-2">Your Name <span class="text-red-500">*</span></label>
						<input type="text" bind:value={name} required class="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-[#4A8B8C]" />
					</div>

					<div class="mb-6">
						<label class="block text-sm font-medium text-gray-700 mb-2">Email (optional)</label>
						<input type="email" bind:value={email} class="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-[#4A8B8C]" />
					</div>

					<div class="mb-6">
						<label class="block text-sm font-medium text-gray-700 mb-2">Prayer Request <span class="text-red-500">*</span></label>
						<textarea bind:value={requestText} required rows="6" class="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-[#4A8B8C]"></textarea>
					</div>

					<div class="mb-6">
						<label class="flex items-start gap-3 cursor-pointer">
							<input type="checkbox" bind:checked={isPublic} class="mt-1 h-5 w-5 text-[#4A8B8C] rounded" />
							<div>
								<span class="font-medium">Share on public prayer wall</span>
								<p class="text-sm text-gray-500">Allow others to pray for this request</p>
							</div>
						</label>
					</div>

					<div class="flex gap-4">
						<button type="submit" disabled={submitting} class="flex-1 px-6 py-3 bg-[#4A8B8C] text-white font-semibold rounded-lg hover:bg-[#3d7576] disabled:opacity-50">
							{submitting ? 'Submitting...' : 'Submit'}
						</button>
						<a href="/prayer" class="px-6 py-3 border text-gray-700 font-semibold rounded-lg hover:bg-gray-50">Back</a>
					</div>
				</form>
			{/if}
		</div>
	</div>
</div>
