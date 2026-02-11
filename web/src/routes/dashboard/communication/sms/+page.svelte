<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	let toPhone = '';
	let body = '';
	let templateId = '';
	let targetType = 'person_ids';
	let targetIds: string[] = [];
	let templates: any[] = [];
	let people: any[] = [];
	let groups: any[] = [];
	let history: any[] = [];
	let loading = false;
	let success = '';
	let error = '';
	let showHistory = false;
	let isBulk = false;

	onMount(async () => {
		await loadTemplates();
		await loadPeople();
		await loadGroups();
	});

	async function loadTemplates() {
		const res = await fetch('/api/sms/templates', {
			headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
		});
		if (res.ok) {
			templates = await res.json();
		}
	}

	async function loadPeople() {
		const res = await fetch('/api/people', {
			headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
		});
		if (res.ok) {
			people = await res.json();
		}
	}

	async function loadGroups() {
		const res = await fetch('/api/groups', {
			headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
		});
		if (res.ok) {
			groups = await res.json();
		}
	}

	async function loadHistory() {
		const res = await fetch('/api/sms/history?limit=50', {
			headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
		});
		if (res.ok) {
			history = await res.json();
			showHistory = true;
		}
	}

	async function sendSMS() {
		loading = true;
		error = '';
		success = '';

		try {
			const endpoint = isBulk ? '/api/sms/bulk' : '/api/sms/send';
			const payload = isBulk
				? {
						body,
						template_id: templateId || undefined,
						target_type: targetType,
						target_ids: targetType === 'person_ids' ? targetIds : targetType === 'group_id' ? [targetIds[0]] : []
				  }
				: {
						to_phone: toPhone,
						body,
						template_id: templateId || undefined
				  };

			const res = await fetch(endpoint, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${localStorage.getItem('token')}`
				},
				body: JSON.stringify(payload)
			});

			if (res.ok) {
				success = isBulk ? 'Bulk SMS sent successfully!' : 'SMS sent successfully!';
				toPhone = '';
				body = '';
				templateId = '';
				targetIds = [];
				await loadHistory();
			} else {
				const data = await res.text();
				error = data || 'Failed to send SMS';
			}
		} catch (err: any) {
			error = err.message || 'An error occurred';
		} finally {
			loading = false;
		}
	}

	function selectTemplate() {
		if (templateId) {
			const template = templates.find((t) => t.id === templateId);
			if (template) {
				body = template.body;
			}
		}
	}

	function countSegments(text: string): number {
		return Math.ceil(text.length / 160);
	}

	$: segments = countSegments(body);
	$: charCount = body.length;
</script>

<div class="max-w-4xl mx-auto p-6">
	<div class="flex justify-between items-center mb-6">
		<h1 class="text-3xl font-bold">SMS Messaging</h1>
		<div class="flex gap-2">
			<button
				on:click={() => goto('/dashboard/communication/sms/templates')}
				class="bg-gray-200 hover:bg-gray-300 px-4 py-2 rounded"
			>
				Manage Templates
			</button>
			<button
				on:click={() => goto('/dashboard/communication/sms/settings')}
				class="bg-gray-200 hover:bg-gray-300 px-4 py-2 rounded"
			>
				Settings
			</button>
		</div>
	</div>

	<!-- Mode Toggle -->
	<div class="mb-6">
		<div class="flex gap-4">
			<button
				on:click={() => (isBulk = false)}
				class="px-4 py-2 rounded {!isBulk ? 'bg-blue-600 text-white' : 'bg-gray-200'}"
			>
				Single Message
			</button>
			<button
				on:click={() => (isBulk = true)}
				class="px-4 py-2 rounded {isBulk ? 'bg-blue-600 text-white' : 'bg-gray-200'}"
			>
				Bulk Message
			</button>
		</div>
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

	<div class="bg-white shadow rounded-lg p-6 mb-6">
		<h2 class="text-xl font-semibold mb-4">Compose Message</h2>

		<form on:submit|preventDefault={sendSMS}>
			<!-- Template Selector -->
			<div class="mb-4">
				<label class="block text-sm font-medium mb-2">Template (optional)</label>
				<select bind:value={templateId} on:change={selectTemplate} class="w-full border rounded px-3 py-2">
					<option value="">-- No Template --</option>
					{#each templates as template}
						<option value={template.id}>{template.name}</option>
					{/each}
				</select>
			</div>

			{#if !isBulk}
				<!-- Single Message: Phone Number -->
				<div class="mb-4">
					<label class="block text-sm font-medium mb-2">Phone Number</label>
					<input
						type="tel"
						bind:value={toPhone}
						placeholder="+1234567890"
						class="w-full border rounded px-3 py-2"
						required
					/>
				</div>
			{:else}
				<!-- Bulk Message: Recipient Selection -->
				<div class="mb-4">
					<label class="block text-sm font-medium mb-2">Recipients</label>
					<select bind:value={targetType} class="w-full border rounded px-3 py-2 mb-2">
						<option value="person_ids">Select People</option>
						<option value="group_id">Select Group</option>
						<option value="all">All People</option>
					</select>

					{#if targetType === 'person_ids'}
						<select bind:value={targetIds} multiple class="w-full border rounded px-3 py-2 h-32">
							{#each people as person}
								<option value={person.id}>{person.first_name} {person.last_name} - {person.phone || 'No phone'}</option>
							{/each}
						</select>
						<p class="text-sm text-gray-600 mt-1">Hold Ctrl/Cmd to select multiple people</p>
					{:else if targetType === 'group_id'}
						<select bind:value={targetIds[0]} class="w-full border rounded px-3 py-2">
							<option value="">-- Select Group --</option>
							{#each groups as group}
								<option value={group.id}>{group.name}</option>
							{/each}
						</select>
					{/if}
				</div>
			{/if}

			<!-- Message Body -->
			<div class="mb-4">
				<label class="block text-sm font-medium mb-2">Message</label>
				<textarea
					bind:value={body}
					placeholder="Type your message here... Use {first_name}, {last_name}, {church_name} for merge fields"
					class="w-full border rounded px-3 py-2 h-32"
					required
				></textarea>
				<div class="flex justify-between text-sm text-gray-600 mt-1">
					<span>{charCount} characters</span>
					<span>{segments} segment{segments !== 1 ? 's' : ''} (160 chars/segment)</span>
				</div>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded disabled:opacity-50"
			>
				{loading ? 'Sending...' : isBulk ? 'Send Bulk SMS' : 'Send SMS'}
			</button>
		</form>
	</div>

	<!-- Message History -->
	<div class="bg-white shadow rounded-lg p-6">
		<div class="flex justify-between items-center mb-4">
			<h2 class="text-xl font-semibold">Message History</h2>
			<button
				on:click={loadHistory}
				class="bg-gray-200 hover:bg-gray-300 px-4 py-2 rounded text-sm"
			>
				Refresh
			</button>
		</div>

		{#if showHistory && history.length > 0}
			<div class="overflow-x-auto">
				<table class="w-full">
					<thead>
						<tr class="border-b">
							<th class="text-left py-2">To</th>
							<th class="text-left py-2">Message</th>
							<th class="text-left py-2">Status</th>
							<th class="text-left py-2">Sent At</th>
						</tr>
					</thead>
					<tbody>
						{#each history as msg}
							<tr class="border-b hover:bg-gray-50">
								<td class="py-2">{msg.to_phone}</td>
								<td class="py-2 max-w-xs truncate">{msg.body}</td>
								<td class="py-2">
									<span
										class="px-2 py-1 rounded text-xs
										{msg.status === 'delivered'
											? 'bg-green-100 text-green-800'
											: msg.status === 'sent'
											? 'bg-blue-100 text-blue-800'
											: msg.status === 'failed'
											? 'bg-red-100 text-red-800'
											: 'bg-gray-100 text-gray-800'}"
									>
										{msg.status}
									</span>
								</td>
								<td class="py-2 text-sm text-gray-600">
									{msg.sent_at ? new Date(msg.sent_at).toLocaleString() : '-'}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{:else if showHistory}
			<p class="text-gray-600">No messages sent yet.</p>
		{:else}
			<button on:click={loadHistory} class="text-blue-600 hover:underline">
				Click to load message history
			</button>
		{/if}
	</div>
</div>
