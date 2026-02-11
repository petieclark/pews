<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let config = {
		enabled: false,
		theme: 'modern',
		hero_title: 'Welcome to Our Church',
		hero_subtitle: 'Join us this Sunday',
		hero_image_url: '',
		service_times: 'Sunday 9:00 AM & 11:00 AM',
		address: '',
		phone: '',
		email: '',
		sections: ['about', 'services', 'sermons', 'events', 'connect', 'give'],
		about_text: '',
		social_links: {
			facebook: '',
			instagram: '',
			youtube: ''
		},
		colors: {
			primary: '#1B3A4B',
			accent: '#4A8B8C'
		}
	};

	let loading = true;
	let saving = false;
	let error = '';
	let success = '';
	let previewWindow = null;
	let tenant = {};

	const availableSections = [
		{ id: 'about', name: 'About' },
		{ id: 'services', name: 'Service Times' },
		{ id: 'sermons', name: 'Sermons' },
		{ id: 'events', name: 'Events' },
		{ id: 'connect', name: 'Connect' },
		{ id: 'give', name: 'Give' }
	];

	onMount(async () => {
		try {
			const [configData, tenantData] = await Promise.all([
				api('/api/website/config'),
				api('/api/tenant')
			]);
			config = { ...config, ...configData };
			tenant = tenantData;
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	});

	async function saveConfig() {
		saving = true;
		error = '';
		success = '';

		try {
			await api('/api/website/config', {
				method: 'PUT',
				body: JSON.stringify(config)
			});
			success = 'Website settings saved successfully!';
			setTimeout(() => success = '', 3000);
		} catch (err) {
			error = err.message;
		} finally {
			saving = false;
		}
	}

	function toggleSection(sectionId) {
		if (config.sections.includes(sectionId)) {
			config.sections = config.sections.filter(s => s !== sectionId);
		} else {
			config.sections = [...config.sections, sectionId];
		}
	}

	function openPreview() {
		if (previewWindow && !previewWindow.closed) {
			previewWindow.focus();
		} else {
			previewWindow = window.open('/api/website/preview', '_blank');
		}
	}

	function openPublicSite() {
		if (tenant.slug) {
			window.open(`/${tenant.slug}`, '_blank');
		}
	}
</script>

<div class="max-w-6xl">
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-3xl font-bold text-primary">Website Builder</h1>
		<div class="flex gap-2">
			<button
				type="button"
				on:click={openPreview}
				class="btn btn-secondary"
				disabled={loading}
			>
				Preview
			</button>
			{#if config.enabled && tenant.slug}
			<button
				type="button"
				on:click={openPublicSite}
				class="btn btn-secondary"
			>
				View Live Site
			</button>
			{/if}
		</div>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--teal)]"></div>
		</div>
	{:else}
		<div class="space-y-6">
			{#if error}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
					{error}
				</div>
			{/if}

			{#if success}
				<div class="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded">
					{success}
				</div>
			{/if}

			<!-- Enable/Disable -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<div class="flex items-center justify-between">
					<div>
						<h2 class="text-xl font-semibold text-primary">Website Status</h2>
						<p class="text-sm text-secondary mt-1">
							{#if config.enabled}
								Your website is live at <a href="/{tenant.slug}" class="text-accent hover:underline" target="_blank">/{tenant.slug}</a>
							{:else}
								Enable your website to make it publicly accessible
							{/if}
						</p>
					</div>
					<label class="relative inline-flex items-center cursor-pointer">
						<input
							type="checkbox"
							bind:checked={config.enabled}
							class="sr-only peer"
						/>
						<div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-accent/20 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-accent"></div>
					</label>
				</div>
			</div>

			<!-- Hero Section -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<h2 class="text-xl font-semibold text-primary mb-4">Hero Section</h2>
				<div class="space-y-4">
					<div>
						<label for="hero_title" class="block text-sm font-medium text-primary mb-1">
							Hero Title
						</label>
						<input
							id="hero_title"
							type="text"
							bind:value={config.hero_title}
							class="input w-full"
							placeholder="Welcome to Our Church"
						/>
					</div>
					<div>
						<label for="hero_subtitle" class="block text-sm font-medium text-primary mb-1">
							Hero Subtitle
						</label>
						<input
							id="hero_subtitle"
							type="text"
							bind:value={config.hero_subtitle}
							class="input w-full"
							placeholder="Join us this Sunday"
						/>
					</div>
					<div>
						<label for="hero_image_url" class="block text-sm font-medium text-primary mb-1">
							Hero Image URL
						</label>
						<input
							id="hero_image_url"
							type="url"
							bind:value={config.hero_image_url}
							class="input w-full"
							placeholder="https://example.com/image.jpg"
						/>
						<p class="text-xs text-secondary mt-1">Enter a direct URL to an image</p>
					</div>
				</div>
			</div>

			<!-- Sections -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<h2 class="text-xl font-semibold text-primary mb-4">Sections</h2>
				<p class="text-sm text-secondary mb-4">Toggle which sections appear on your website</p>
				<div class="grid grid-cols-2 md:grid-cols-3 gap-4">
					{#each availableSections as section}
						<label class="flex items-center space-x-2 cursor-pointer">
							<input
								type="checkbox"
								checked={config.sections.includes(section.id)}
								on:change={() => toggleSection(section.id)}
								class="rounded border-gray-300 text-accent focus:ring-accent"
							/>
							<span class="text-primary">{section.name}</span>
						</label>
					{/each}
				</div>
			</div>

			<!-- About Section -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<h2 class="text-xl font-semibold text-primary mb-4">About Section</h2>
				<div>
					<label for="about_text" class="block text-sm font-medium text-primary mb-1">
						About Text
					</label>
					<textarea
						id="about_text"
						bind:value={config.about_text}
						rows="4"
						class="input w-full"
						placeholder="Tell visitors about your church..."
					></textarea>
				</div>
			</div>

			<!-- Service Times & Contact -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<h2 class="text-xl font-semibold text-primary mb-4">Service Times & Contact</h2>
				<div class="space-y-4">
					<div>
						<label for="service_times" class="block text-sm font-medium text-primary mb-1">
							Service Times
						</label>
						<input
							id="service_times"
							type="text"
							bind:value={config.service_times}
							class="input w-full"
							placeholder="Sunday 9:00 AM & 11:00 AM"
						/>
					</div>
					<div>
						<label for="address" class="block text-sm font-medium text-primary mb-1">
							Address
						</label>
						<input
							id="address"
							type="text"
							bind:value={config.address}
							class="input w-full"
							placeholder="123 Main St, City, State 12345"
						/>
					</div>
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<label for="phone" class="block text-sm font-medium text-primary mb-1">
								Phone
							</label>
							<input
								id="phone"
								type="tel"
								bind:value={config.phone}
								class="input w-full"
								placeholder="(555) 123-4567"
							/>
						</div>
						<div>
							<label for="email" class="block text-sm font-medium text-primary mb-1">
								Email
							</label>
							<input
								id="email"
								type="email"
								bind:value={config.email}
								class="input w-full"
								placeholder="info@church.com"
							/>
						</div>
					</div>
				</div>
			</div>

			<!-- Social Links -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<h2 class="text-xl font-semibold text-primary mb-4">Social Media Links</h2>
				<div class="space-y-4">
					<div>
						<label for="facebook" class="block text-sm font-medium text-primary mb-1">
							Facebook URL
						</label>
						<input
							id="facebook"
							type="url"
							bind:value={config.social_links.facebook}
							class="input w-full"
							placeholder="https://facebook.com/yourchurch"
						/>
					</div>
					<div>
						<label for="instagram" class="block text-sm font-medium text-primary mb-1">
							Instagram URL
						</label>
						<input
							id="instagram"
							type="url"
							bind:value={config.social_links.instagram}
							class="input w-full"
							placeholder="https://instagram.com/yourchurch"
						/>
					</div>
					<div>
						<label for="youtube" class="block text-sm font-medium text-primary mb-1">
							YouTube URL
						</label>
						<input
							id="youtube"
							type="url"
							bind:value={config.social_links.youtube}
							class="input w-full"
							placeholder="https://youtube.com/@yourchurch"
						/>
					</div>
				</div>
			</div>

			<!-- Colors -->
			<div class="bg-surface rounded-lg shadow-md p-6 border border-custom">
				<h2 class="text-xl font-semibold text-primary mb-4">Brand Colors</h2>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="primary_color" class="block text-sm font-medium text-primary mb-1">
							Primary Color
						</label>
						<div class="flex gap-2">
							<input
								id="primary_color"
								type="color"
								bind:value={config.colors.primary}
								class="h-10 w-20 rounded border border-custom cursor-pointer"
							/>
							<input
								type="text"
								bind:value={config.colors.primary}
								class="input flex-1"
								placeholder="#1B3A4B"
							/>
						</div>
					</div>
					<div>
						<label for="accent_color" class="block text-sm font-medium text-primary mb-1">
							Accent Color
						</label>
						<div class="flex gap-2">
							<input
								id="accent_color"
								type="color"
								bind:value={config.colors.accent}
								class="h-10 w-20 rounded border border-custom cursor-pointer"
							/>
							<input
								type="text"
								bind:value={config.colors.accent}
								class="input flex-1"
								placeholder="#4A8B8C"
							/>
						</div>
					</div>
				</div>
			</div>

			<!-- Save Button -->
			<div class="flex justify-end gap-3 pt-4">
				<button
					type="button"
					on:click={saveConfig}
					disabled={saving}
					class="btn btn-primary"
				>
					{#if saving}
						<span class="inline-block animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></span>
						Saving...
					{:else}
						Publish Changes
					{/if}
				</button>
			</div>
		</div>
	{/if}
</div>

<style>
	.input {
		@apply px-3 py-2 border border-custom rounded-md bg-surface text-primary;
	}
	.input:focus {
		@apply outline-none ring-2 ring-accent/50;
	}
	.btn {
		@apply px-4 py-2 rounded-md font-medium transition-colors;
	}
	.btn-primary {
		@apply bg-accent text-white hover:bg-accent/90;
	}
	.btn-primary:disabled {
		@apply opacity-50 cursor-not-allowed;
	}
	.btn-secondary {
		@apply bg-surface text-primary border border-custom hover:bg-gray-50;
	}
</style>
