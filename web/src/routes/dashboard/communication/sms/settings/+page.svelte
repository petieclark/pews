<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	let accountSID = '';
	let authToken = '';
	let fromNumber = '';
	let isEnabled = false;
	let lastTestedAt = '';
	let loading = false;
	let testing = false;
	let success = '';
	let error = '';

	onMount(async () => {
		await loadSettings();
	});

	async function loadSettings() {
		const res = await fetch('/api/sms/settings', {
			headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
		});
		if (res.ok) {
			const data = await res.json();
			accountSID = data.twilio_account_sid || '';
			fromNumber = data.twilio_from_number || '';
			isEnabled = data.is_enabled || false;
			lastTestedAt = data.last_tested_at || '';
		}
	}

	async function saveSettings() {
		loading = true;
		error = '';
		success = '';

		try {
			const res = await fetch('/api/sms/settings', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${localStorage.getItem('token')}`
				},
				body: JSON.stringify({
					twilio_account_sid: accountSID,
					twilio_auth_token: authToken,
					twilio_from_number: fromNumber
				})
			});

			if (res.ok) {
				success = 'Settings saved successfully!';
				authToken = ''; // Clear sensitive field
				await loadSettings();
			} else {
				error = await res.text();
			}
		} catch (err: any) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	async function testConnection() {
		testing = true;
		error = '';
		success = '';

		try {
			const res = await fetch('/api/sms/settings/test', {
				method: 'POST',
				headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
			});

			if (res.ok) {
				const data = await res.json();
				success = data.message || 'Connection test successful!';
				await loadSettings();
			} else {
				error = await res.text();
			}
		} catch (err: any) {
			error = err.message;
		} finally {
			testing = false;
		}
	}
</script>

<div class="max-w-4xl mx-auto p-6">
	<div class="flex justify-between items-center mb-6">
		<h1 class="text-3xl font-bold">SMS Settings</h1>
		<button
			on:click={() => goto('/dashboard/communication/sms')}
			class="bg-gray-200 hover:bg-gray-300 px-4 py-2 rounded"
		>
			← Back to SMS
		</button>
	</div>

	{#if success}
		<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded mb-4">
			{success}
		</div>
	{/if}

	{#if error}
		<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
			{error}
		</div>
	{/if}

	<!-- Settings Status -->
	<div class="bg-white shadow rounded-lg p-6 mb-6">
		<h2 class="text-xl font-semibold mb-4">Status</h2>
		<div class="flex items-center gap-4">
			<div>
				<span
					class="px-3 py-1 rounded text-sm font-medium {isEnabled
						? 'bg-green-100 text-green-800'
						: 'bg-gray-100 text-gray-800'}"
				>
					{isEnabled ? 'Enabled' : 'Not Configured'}
				</span>
			</div>
			{#if lastTestedAt}
				<p class="text-sm text-gray-600">
					Last tested: {new Date(lastTestedAt).toLocaleString()}
				</p>
			{/if}
		</div>
	</div>

	<!-- Twilio Configuration -->
	<div class="bg-white shadow rounded-lg p-6 mb-6">
		<h2 class="text-xl font-semibold mb-4">Twilio Configuration</h2>

		<div class="bg-blue-50 border border-blue-200 rounded p-4 mb-4">
			<h3 class="font-semibold text-blue-900 mb-2">How to get Twilio credentials:</h3>
			<ol class="list-decimal ml-5 text-sm text-blue-900 space-y-1">
				<li>Sign up for a Twilio account at <a href="https://www.twilio.com" target="_blank" class="underline">twilio.com</a></li>
				<li>Go to your Twilio Console Dashboard</li>
				<li>Copy your Account SID and Auth Token</li>
				<li>Purchase a phone number (or use your trial number)</li>
				<li>Paste the credentials below and save</li>
			</ol>
		</div>

		<form on:submit|preventDefault={saveSettings}>
			<div class="mb-4">
				<label class="block text-sm font-medium mb-2">Twilio Account SID</label>
				<input
					type="text"
					bind:value={accountSID}
					placeholder="ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
					class="w-full border rounded px-3 py-2 font-mono"
					required
				/>
			</div>

			<div class="mb-4">
				<label class="block text-sm font-medium mb-2">Twilio Auth Token</label>
				<input
					type="password"
					bind:value={authToken}
					placeholder="Enter your Twilio Auth Token"
					class="w-full border rounded px-3 py-2 font-mono"
					required={!isEnabled}
				/>
				<p class="text-sm text-gray-600 mt-1">
					{isEnabled ? 'Leave blank to keep existing token' : 'Required for first-time setup'}
				</p>
			</div>

			<div class="mb-4">
				<label class="block text-sm font-medium mb-2">From Phone Number</label>
				<input
					type="tel"
					bind:value={fromNumber}
					placeholder="+1234567890"
					class="w-full border rounded px-3 py-2"
					required
				/>
				<p class="text-sm text-gray-600 mt-1">Must be a valid Twilio phone number in E.164 format</p>
			</div>

			<div class="flex gap-2">
				<button
					type="submit"
					disabled={loading}
					class="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded disabled:opacity-50"
				>
					{loading ? 'Saving...' : 'Save Settings'}
				</button>

				{#if isEnabled}
					<button
						type="button"
						on:click={testConnection}
						disabled={testing}
						class="bg-green-600 hover:bg-green-700 text-white px-6 py-2 rounded disabled:opacity-50"
					>
						{testing ? 'Testing...' : 'Test Connection'}
					</button>
				{/if}
			</div>
		</form>
	</div>

	<!-- Important Notes -->
	<div class="bg-white shadow rounded-lg p-6">
		<h2 class="text-xl font-semibold mb-4">Important Notes</h2>
		<ul class="list-disc ml-5 text-sm space-y-2 text-gray-700">
			<li>Your church is responsible for its own Twilio account and SMS costs</li>
			<li>Twilio credentials are encrypted and stored securely</li>
			<li>Rate limiting: Maximum 1 SMS per second</li>
			<li>SMS segments: Each 160 characters counts as 1 segment</li>
			<li>Bulk sending to large groups may take time due to rate limits</li>
			<li>Test your connection before sending messages to ensure everything works</li>
		</ul>
	</div>
</div>
