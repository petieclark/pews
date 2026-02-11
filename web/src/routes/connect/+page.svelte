<script>
	import { onMount } from 'svelte';

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8190';

	let loading = false;
	let submitted = false;
	let error = '';

	let formData = {
		first_name: '',
		last_name: '',
		email: '',
		phone: '',
		is_first_visit: false,
		how_heard: '',
		prayer_request: '',
		interests: []
	};

	const hearAboutOptions = [
		'Friend or Family',
		'Social Media',
		'Google Search',
		'Drove By',
		'Community Event',
		'Other'
	];

	const interestOptions = [
		{ value: 'small_groups', label: 'Small Groups', icon: '👥' },
		{ value: 'volunteering', label: 'Volunteering', icon: '🤝' },
		{ value: 'membership', label: 'Membership', icon: '🏠' },
		{ value: 'youth_ministry', label: 'Youth Ministry', icon: '🧑‍🤝‍🧑' },
		{ value: 'childrens_ministry', label: "Children's Ministry", icon: '👶' },
		{ value: 'worship_team', label: 'Worship Team', icon: '🎵' },
		{ value: 'bible_study', label: 'Bible Study', icon: '📖' },
		{ value: 'outreach', label: 'Community Outreach', icon: '🌍' }
	];

	function toggleInterest(value) {
		if (formData.interests.includes(value)) {
			formData.interests = formData.interests.filter(i => i !== value);
		} else {
			formData.interests = [...formData.interests, value];
		}
	}

	async function submitCard() {
		if (!formData.first_name) {
			error = 'Please provide at least your first name';
			return;
		}

		loading = true;
		error = '';

		try {
			const response = await fetch(`${API_URL}/api/communication/cards`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					first_name: formData.first_name,
					last_name: formData.last_name,
					email: formData.email,
					phone: formData.phone,
					is_first_visit: formData.is_first_visit,
					how_heard: formData.how_heard,
					prayer_request: formData.prayer_request,
					interested_in: formData.interests.join(', ')
				})
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText || 'Failed to submit');
			}

			submitted = true;
			formData = {
				first_name: '', last_name: '', email: '', phone: '',
				is_first_visit: false, how_heard: '', prayer_request: '', interests: []
			};
		} catch (err) {
			error = err.message || 'Failed to submit. Please try again.';
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen flex items-center justify-center p-4 py-12">
	<div class="max-w-2xl w-full">
		<!-- Header -->
		<div class="text-center mb-8">
			<h1 class="text-4xl md:text-5xl font-bold text-white mb-4">
				👋 We'd Love to Connect!
			</h1>
			<p class="text-xl text-gray-300">
				Whether you're visiting for the first time or have been here for years, we'd love to hear from you.
			</p>
		</div>

		{#if submitted}
			<div class="bg-green-900 bg-opacity-30 border border-green-600 rounded-lg p-8 text-center">
				<div class="text-6xl mb-4">✅</div>
				<h2 class="text-2xl font-bold text-green-400 mb-3">Thank You!</h2>
				<p class="text-gray-300 mb-6">
					We've received your connection card and will be in touch soon.
				</p>
				<button
					on:click={() => submitted = false}
					class="px-6 py-2 bg-green-600 hover:bg-green-700 text-white rounded-md font-medium transition"
				>
					Submit Another
				</button>
			</div>
		{:else}
			<div class="bg-gray-800 bg-opacity-50 backdrop-blur rounded-lg shadow-2xl p-6 md:p-8">
				{#if error}
					<div class="bg-red-900 bg-opacity-30 border border-red-600 text-red-300 px-4 py-3 rounded-lg mb-6">{error}</div>
				{/if}

				<form on:submit|preventDefault={submitCard} class="space-y-6">
					<!-- Name Fields -->
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<label class="block text-sm font-medium text-gray-300 mb-2">
								First Name <span class="text-red-400">*</span>
							</label>
							<input type="text" bind:value={formData.first_name} required placeholder="John"
								class="w-full px-4 py-3 bg-gray-900 bg-opacity-50 border border-gray-600 text-white rounded-md focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none placeholder-gray-500" />
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-300 mb-2">Last Name</label>
							<input type="text" bind:value={formData.last_name} placeholder="Doe"
								class="w-full px-4 py-3 bg-gray-900 bg-opacity-50 border border-gray-600 text-white rounded-md focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none placeholder-gray-500" />
						</div>
					</div>

					<!-- Email & Phone -->
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<label class="block text-sm font-medium text-gray-300 mb-2">Email</label>
							<input type="email" bind:value={formData.email} placeholder="john@example.com"
								class="w-full px-4 py-3 bg-gray-900 bg-opacity-50 border border-gray-600 text-white rounded-md focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none placeholder-gray-500" />
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-300 mb-2">Phone Number</label>
							<input type="tel" bind:value={formData.phone} placeholder="(555) 123-4567"
								class="w-full px-4 py-3 bg-gray-900 bg-opacity-50 border border-gray-600 text-white rounded-md focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none placeholder-gray-500" />
						</div>
					</div>

					<!-- First Visit -->
					<div class="flex items-center gap-3 p-4 bg-yellow-900 bg-opacity-20 border border-yellow-700 rounded-lg">
						<input type="checkbox" id="first-visit" bind:checked={formData.is_first_visit}
							class="w-5 h-5 text-teal-600 bg-gray-700 border-gray-600 rounded focus:ring-teal-500 focus:ring-2" />
						<label for="first-visit" class="text-gray-200 font-medium cursor-pointer">
							This is my first time visiting
						</label>
					</div>

					<!-- How did you hear about us -->
					<div>
						<label class="block text-sm font-medium text-gray-300 mb-2">How did you hear about us?</label>
						<select bind:value={formData.how_heard}
							class="w-full px-4 py-3 bg-gray-900 bg-opacity-50 border border-gray-600 text-white rounded-md focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none">
							<option value="">Select an option...</option>
							{#each hearAboutOptions as option}
								<option value={option}>{option}</option>
							{/each}
						</select>
					</div>

					<!-- Interest Checkboxes -->
					<div>
						<label class="block text-sm font-medium text-gray-300 mb-3">I'm interested in...</label>
						<div class="grid grid-cols-2 gap-2">
							{#each interestOptions as opt}
								<button
									type="button"
									on:click={() => toggleInterest(opt.value)}
									class="flex items-center gap-2 p-3 rounded-lg border text-left transition-all {formData.interests.includes(opt.value) ? 'border-teal-500 bg-teal-900 bg-opacity-30' : 'border-gray-600 bg-gray-900 bg-opacity-30 hover:border-gray-500'}"
								>
									<span class="text-lg">{opt.icon}</span>
									<span class="text-sm font-medium {formData.interests.includes(opt.value) ? 'text-teal-300' : 'text-gray-300'}">{opt.label}</span>
								</button>
							{/each}
						</div>
					</div>

					<!-- Prayer Request -->
					<div>
						<label class="block text-sm font-medium text-gray-300 mb-2">Prayer Request or Message</label>
						<textarea bind:value={formData.prayer_request} rows="4"
							placeholder="Is there anything you'd like us to pray for?"
							class="w-full px-4 py-3 bg-gray-900 bg-opacity-50 border border-gray-600 text-white rounded-md focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none placeholder-gray-500 resize-none"></textarea>
					</div>

					<!-- Submit -->
					<button type="submit" disabled={loading}
						class="w-full px-6 py-4 bg-teal-600 hover:bg-teal-700 text-white text-lg font-bold rounded-md transition disabled:opacity-50 disabled:cursor-not-allowed shadow-lg hover:shadow-xl">
						{loading ? 'Submitting...' : 'Submit Connection Card'}
					</button>

					<p class="text-center text-sm text-gray-400">
						We respect your privacy and will never share your information.
					</p>
				</form>
			</div>
		{/if}

		<div class="text-center mt-8">
			<p class="text-gray-500 text-sm">
				Powered by <span class="font-bold text-teal-500">Pews</span>
			</p>
		</div>
	</div>
</div>
