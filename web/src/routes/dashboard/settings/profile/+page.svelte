<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let tenant = { name: '', address: '', phone: '', email: '', website: '', timezone: '' };
	let logoPreview = '';
	let loading = true;
	let saving = false;
	let success = '';
	let error = '';
	let logoFile = null;

	const timezones = [
		'America/New_York', 'America/Chicago', 'America/Denver', 'America/Los_Angeles',
		'America/Phoenix', 'America/Anchorage', 'Pacific/Honolulu', 'America/Puerto_Rico',
		'Europe/London', 'Europe/Berlin', 'Asia/Tokyo', 'Australia/Sydney'
	];

	onMount(async () => {
		try {
			const data = await api('/api/tenant/profile');
			tenant = { ...tenant, ...data };
			logoPreview = data.logo || '';
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	});

	async function save() {
		saving = true;
		error = '';
		success = '';
		try {
			await api('/api/tenant/profile', {
				method: 'PUT',
				body: JSON.stringify(tenant)
			});

			if (logoFile) {
				const formData = new FormData();
				formData.append('logo', logoFile);
				await fetch('/api/tenant/profile/logo', {
					method: 'POST',
					headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` },
					body: formData
				});
			}

			success = 'Profile saved successfully';
		} catch (err) {
			error = err.message;
		} finally {
			saving = false;
		}
	}

	function handleLogoChange(e) {
		const file = e.target.files[0];
		if (file) {
			logoFile = file;
			const reader = new FileReader();
			reader.onload = (ev) => { logoPreview = ev.target.result; };
			reader.readAsDataURL(file);
		}
	}
</script>

<div class="max-w-2xl">
	<a href="/dashboard/settings" class="text-sm text-secondary hover:text-[var(--teal)] mb-4 inline-block">← Back to Settings</a>
	<h1 class="text-2xl sm:text-3xl font-bold text-primary mb-6">Church Profile</h1>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else}
		<form on:submit|preventDefault={save} class="space-y-5">
			<div class="bg-surface rounded-lg border border-custom p-6 space-y-5">
				<!-- Logo -->
				<div>
					<label class="block text-sm font-medium text-primary mb-2">Logo</label>
					<div class="flex items-center gap-4">
						{#if logoPreview}
							<img src={logoPreview} alt="Church logo" class="h-16 w-16 rounded-lg object-contain border border-custom" />
						{:else}
							<div class="h-16 w-16 rounded-lg bg-[var(--surface-hover)] border border-custom flex items-center justify-center text-secondary text-2xl">🏛️</div>
						{/if}
						<input type="file" accept="image/*" on:change={handleLogoChange} class="text-sm text-secondary" />
					</div>
				</div>

				<div>
					<label for="name" class="block text-sm font-medium text-primary mb-1">Church Name</label>
					<input id="name" type="text" bind:value={tenant.name} required
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary" />
				</div>

				<div>
					<label for="address" class="block text-sm font-medium text-primary mb-1">Address</label>
					<textarea id="address" bind:value={tenant.address} rows="2"
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary"></textarea>
				</div>

				<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
					<div>
						<label for="phone" class="block text-sm font-medium text-primary mb-1">Phone</label>
						<input id="phone" type="tel" bind:value={tenant.phone}
							class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary" />
					</div>
					<div>
						<label for="email" class="block text-sm font-medium text-primary mb-1">Email</label>
						<input id="email" type="email" bind:value={tenant.email}
							class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary" />
					</div>
				</div>

				<div>
					<label for="website" class="block text-sm font-medium text-primary mb-1">Website</label>
					<input id="website" type="url" bind:value={tenant.website} placeholder="https://"
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary" />
				</div>

				<div>
					<label for="timezone" class="block text-sm font-medium text-primary mb-1">Timezone</label>
					<select id="timezone" bind:value={tenant.timezone}
						class="w-full px-4 py-2 border input-border rounded-lg focus:ring-2 focus:ring-[var(--teal)] focus:border-transparent bg-[var(--input-bg)] text-primary">
						<option value="">Select timezone</option>
						{#each timezones as tz}
							<option value={tz}>{tz.replace(/_/g, ' ')}</option>
						{/each}
					</select>
				</div>
			</div>

			{#if error}
				<div class="bg-[var(--error-bg)] border border-[var(--error-border)] text-[var(--error-text)] px-4 py-3 rounded-lg">{error}</div>
			{/if}
			{#if success}
				<div class="bg-emerald-900/30 border border-emerald-700 text-emerald-300 px-4 py-3 rounded-lg">{success}</div>
			{/if}

			<button type="submit" disabled={saving}
				class="w-full sm:w-auto bg-[var(--teal)] text-white py-2 px-6 rounded-lg font-medium hover:opacity-90 disabled:opacity-50">
				{saving ? 'Saving...' : 'Save Profile'}
			</button>
		</form>
	{/if}
</div>
