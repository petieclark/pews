<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let settings = {
		enabled: true,
		send_day: 'monday',
		recipients: []
	};
	let digestData = null;
	let loading = true;
	let saving = false;
	let previewLoading = false;
	let showPreview = false;
	let error = '';
	let success = '';
	let newRecipient = '';

	const days = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];

	onMount(async () => {
		await loadSettings();
	});

	async function loadSettings() {
		loading = true;
		try {
			settings = await api('/api/digest/settings');
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	async function saveSettings() {
		saving = true;
		error = '';
		success = '';

		try {
			await api('/api/digest/settings', {
				method: 'PUT',
				body: JSON.stringify(settings)
			});
			success = 'Digest settings saved successfully';
		} catch (err) {
			error = err.message;
		} finally {
			saving = false;
		}
	}

	function addRecipient() {
		if (newRecipient && !settings.recipients.includes(newRecipient)) {
			settings.recipients = [...settings.recipients, newRecipient];
			newRecipient = '';
		}
	}

	function removeRecipient(email) {
		settings.recipients = settings.recipients.filter(r => r !== email);
	}

	async function previewDigest() {
		previewLoading = true;
		error = '';

		try {
			digestData = await api('/api/digest/data');
			showPreview = true;
		} catch (err) {
			error = err.message;
		} finally {
			previewLoading = false;
		}
	}

	async function openPreview() {
		window.open('/api/digest/preview', '_blank');
	}
</script>

<div class="max-w-4xl">
	<div class="mb-6">
		<a href="/dashboard/settings" class="text-[var(--teal)] hover:underline">
			← Back to Settings
		</a>
	</div>

	<h1 class="text-3xl font-bold text-primary mb-6">Weekly Digest</h1>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else}
		<div class="bg-surface rounded-lg shadow-md p-6 mb-6 border border-custom">
			<h2 class="text-xl font-semibold text-primary mb-4">Digest Settings</h2>
			<p class="text-secondary mb-6">
				Configure your weekly church activity digest. This email summarizes attendance, giving,
				new members, and upcoming events.
			</p>

			<form on:submit|preventDefault={saveSettings} class="space-y-6">
				<!-- Enable/Disable Toggle -->
				<div class="flex items-center justify-between p-4 border border-custom rounded-lg">
					<div>
						<h3 class="font-medium text-primary">Enable Weekly Digest</h3>
						<p class="text-sm text-secondary">Send a weekly summary email to admins</p>
					</div>
					<label class="relative inline-flex items-center cursor-pointer">
						<input
							type="checkbox"
							bind:checked={settings.enabled}
							class="sr-only peer"
						/>
						<div class="w-11 h-6 bg-[var(--surface-hover)] peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-[var(--teal)]/30 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-[var(--border)] after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-[var(--teal)]"></div>
					</label>
				</div>

				<!-- Send Day -->
				<div>
					<label for="send_day" class="block text-sm font-medium text-primary mb-2">
						Send Day
					</label>
					<select
						id="send_day"
						bind:value={settings.send_day}
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary capitalize"
					>
						{#each days as day}
							<option value={day} class="capitalize">{day}</option>
						{/each}
					</select>
					<p class="text-sm text-secondary mt-1">
						Default is Monday (recaps Sunday activity)
					</p>
				</div>

				<!-- Recipients -->
				<div>
					<label class="block text-sm font-medium text-primary mb-2">
						Email Recipients
					</label>
					<div class="space-y-2">
						{#each settings.recipients as recipient}
							<div class="flex items-center justify-between p-3 border border-custom rounded-lg bg-[var(--surface-hover)]">
								<span class="text-primary">{recipient}</span>
								<button
									type="button"
									on:click={() => removeRecipient(recipient)}
									class="text-red-500 hover:text-red-700"
								>
									Remove
								</button>
							</div>
						{/each}

						<div class="flex gap-2">
							<input
								type="email"
								bind:value={newRecipient}
								placeholder="Add email address"
								class="flex-1 px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"
							/>
							<button
								type="button"
								on:click={addRecipient}
								class="bg-[var(--text-secondary)] text-white py-2 px-4 rounded-lg font-medium hover:opacity-90"
							>
								Add
							</button>
						</div>
					</div>
					<p class="text-sm text-secondary mt-1">
						If no recipients are specified, all admins will receive the digest
					</p>
				</div>

				{#if error}
					<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg">
						{error}
					</div>
				{/if}

				{#if success}
					<div class="success-box">
						{success}
					</div>
				{/if}

				<div class="flex gap-3">
					<button
						type="submit"
						disabled={saving}
						class="bg-[var(--teal)] text-white py-2 px-6 rounded-lg font-medium hover:opacity-90 disabled:opacity-50"
					>
						{saving ? 'Saving...' : 'Save Settings'}
					</button>

					<button
						type="button"
						on:click={openPreview}
						disabled={previewLoading}
						class="bg-[var(--text-secondary)] text-white py-2 px-6 rounded-lg font-medium hover:opacity-90 disabled:opacity-50"
					>
						{previewLoading ? 'Loading...' : 'Preview Digest'}
					</button>
				</div>
			</form>
		</div>

		<!-- Digest Preview Data -->
		{#if showPreview && digestData}
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<h2 class="text-xl font-semibold text-primary mb-4">Digest Preview Data</h2>
				
				<div class="grid grid-cols-2 gap-4 mb-6">
					<div class="p-4 border border-custom rounded-lg">
						<p class="text-sm text-secondary">Attendance This Week</p>
						<p class="text-2xl font-bold text-primary">{digestData.attendance.this_week}</p>
						{#if digestData.attendance.change !== 0}
							<p class="text-sm {digestData.attendance.change > 0 ? 'text-green-500' : 'text-red-500'}">
								{digestData.attendance.change > 0 ? '↑' : '↓'} {Math.abs(digestData.attendance.change)} from last week
							</p>
						{/if}
					</div>

					<div class="p-4 border border-custom rounded-lg">
						<p class="text-sm text-secondary">New Members</p>
						<p class="text-2xl font-bold text-primary">{digestData.members.new_this_week}</p>
					</div>

					<div class="p-4 border border-custom rounded-lg">
						<p class="text-sm text-secondary">Giving This Week</p>
						<p class="text-2xl font-bold text-primary">{digestData.giving.this_week_display}</p>
					</div>

					<div class="p-4 border border-custom rounded-lg">
						<p class="text-sm text-secondary">Year to Date</p>
						<p class="text-2xl font-bold text-primary">{digestData.giving.year_to_date_display}</p>
					</div>
				</div>

				{#if digestData.upcoming_services.length > 0}
					<div class="mb-4">
						<h3 class="font-semibold text-primary mb-2">Upcoming Services</h3>
						<ul class="space-y-2">
							{#each digestData.upcoming_services as service}
								<li class="text-secondary">
									{service.name} - {new Date(service.service_date).toLocaleDateString()}
								</li>
							{/each}
						</ul>
					</div>
				{/if}

				<p class="text-sm text-secondary mt-4">
					Open the full HTML preview in a new tab to see the complete email design.
				</p>
			</div>
		{/if}
	{/if}
</div>

<style>
	.success-box {
		background-color: #D1FAE5;
		border: 1px solid #6EE7B7;
		color: #065F46;
		padding: 1rem;
		border-radius: 0.5rem;
	}
	:global(.dark) .success-box {
		background-color: #064E3B;
		border-color: #059669;
		color: #6EE7B7;
	}
</style>
