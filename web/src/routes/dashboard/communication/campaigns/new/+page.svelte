<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let step = 1;
	let name = '';
	let channel = 'email';
	let subject = '';
	let body = '';
	let senderName = '';
	let targetType = 'all';
	let targetId = '';
	let templateId = '';
	let templates = [];
	let tags = [];
	let groups = [];
	let loading = false;
	let error = '';

	onMount(async () => {
		try {
			const [templatesData, tagsData, groupsData] = await Promise.all([
				api('/api/communication/templates'),
				api('/api/tags'),
				api('/api/groups')
			]);
			templates = templatesData;
			tags = tagsData;
			groups = groupsData;
		} catch (err) {
			error = err.message;
		}
	});

	function loadTemplate() {
		if (!templateId) return;
		const template = templates.find(t => t.id === templateId);
		if (template) {
			channel = template.channel;
			subject = template.subject || '';
			body = template.body;
		}
	}

	function insertVariable(variable) {
		body += `{{${variable}}}`;
	}

	function nextStep() {
		error = '';
		if (step === 1 && !targetType) { error = 'Select an audience'; return; }
		if (step === 2 && (!name || !body)) { error = 'Name and body are required'; return; }
		if (step === 2 && channel === 'email' && !subject) { error = 'Subject is required for email'; return; }
		step = Math.min(step + 1, 4);
	}

	function prevStep() {
		step = Math.max(step - 1, 1);
	}

	async function createAndSend(sendNow = false) {
		loading = true;
		error = '';
		try {
			const campaign = await api('/api/communication/campaigns', {
				method: 'POST',
				body: JSON.stringify({
					name,
					channel,
					subject: channel === 'email' ? subject : '',
					body,
					target_type: targetType,
					target_id: targetType === 'all' ? '' : targetId,
					template_id: templateId || null
				})
			});

			if (sendNow) {
				await api(`/api/communication/campaigns/${campaign.id}/send`, {
					method: 'POST',
					body: JSON.stringify({})
				});
			}

			goto(`/dashboard/communication/campaigns/${campaign.id}`);
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	function getTargetLabel() {
		if (targetType === 'all') return 'All Members';
		if (targetType === 'tag') return `Tag: ${targetId}`;
		if (targetType === 'group') {
			const g = groups.find(g => g.id === targetId);
			return `Group: ${g?.name || targetId}`;
		}
		return targetType;
	}

	function getPreviewBody() {
		return body
			.replace(/\{\{first_name\}\}/g, 'John')
			.replace(/\{\{last_name\}\}/g, 'Doe')
			.replace(/\{\{church_name\}\}/g, 'Your Church')
			.replace(/\{\{email\}\}/g, 'john@example.com');
	}

	const steps = ['Audience', 'Compose', 'Preview', 'Confirm'];
</script>

<div class="max-w-4xl mx-auto">
	<div class="mb-6">
		<a href="/dashboard/communication/campaigns" class="text-sm font-medium" style="color: var(--teal)">← Campaigns</a>
		<h1 class="text-3xl font-bold mt-1" style="color: var(--text-primary)">New Campaign</h1>
	</div>

	<!-- Step Indicator -->
	<div class="flex items-center gap-2 mb-8">
		{#each steps as s, i}
			<div class="flex items-center gap-2">
				<div class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold transition-all"
					style="background: {step > i + 1 ? 'var(--teal)' : step === i + 1 ? 'var(--teal)' : 'var(--surface)'}; color: {step >= i + 1 ? 'white' : 'var(--text-secondary)'}; border: 1px solid {step >= i + 1 ? 'var(--teal)' : 'var(--border)'}">
					{step > i + 1 ? '✓' : i + 1}
				</div>
				<span class="text-sm font-medium hidden sm:inline" style="color: {step === i + 1 ? 'var(--text-primary)' : 'var(--text-secondary)'}">{s}</span>
			</div>
			{#if i < steps.length - 1}
				<div class="flex-1 h-px" style="background: {step > i + 1 ? 'var(--teal)' : 'var(--border)'}"></div>
			{/if}
		{/each}
	</div>

	{#if error}
		<div class="mb-4 px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">{error}</div>
	{/if}

	<div class="rounded-lg shadow border p-6" style="background: var(--surface); border-color: var(--border)">
		<!-- Step 1: Audience -->
		{#if step === 1}
			<h2 class="text-xl font-semibold mb-4" style="color: var(--text-primary)">Choose Your Audience</h2>
			<div class="space-y-3">
				<button on:click={() => { targetType = 'all'; targetId = ''; }}
					class="w-full p-4 rounded-lg border text-left transition"
					style="background: {targetType === 'all' ? 'rgba(74,139,140,0.1)' : 'var(--bg)'}; border-color: {targetType === 'all' ? 'var(--teal)' : 'var(--border)'}">
					<div class="font-medium" style="color: var(--text-primary)">👥 All Members</div>
					<div class="text-sm" style="color: var(--text-secondary)">Send to everyone in your directory</div>
				</button>
				<button on:click={() => targetType = 'group'}
					class="w-full p-4 rounded-lg border text-left transition"
					style="background: {targetType === 'group' ? 'rgba(74,139,140,0.1)' : 'var(--bg)'}; border-color: {targetType === 'group' ? 'var(--teal)' : 'var(--border)'}">
					<div class="font-medium" style="color: var(--text-primary)">🏷️ Specific Group</div>
					<div class="text-sm" style="color: var(--text-secondary)">Send to members of a specific group</div>
				</button>
				<button on:click={() => targetType = 'tag'}
					class="w-full p-4 rounded-lg border text-left transition"
					style="background: {targetType === 'tag' ? 'rgba(74,139,140,0.1)' : 'var(--bg)'}; border-color: {targetType === 'tag' ? 'var(--teal)' : 'var(--border)'}">
					<div class="font-medium" style="color: var(--text-primary)">🔖 By Tag</div>
					<div class="text-sm" style="color: var(--text-secondary)">Filter by custom tags</div>
				</button>
			</div>

			{#if targetType === 'group'}
				<div class="mt-4">
					<select bind:value={targetId} class="w-full px-4 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)">
						<option value="">— Select Group —</option>
						{#each groups as group}
							<option value={group.id}>{group.name}</option>
						{/each}
					</select>
				</div>
			{:else if targetType === 'tag'}
				<div class="mt-4">
					<select bind:value={targetId} class="w-full px-4 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)">
						<option value="">— Select Tag —</option>
						{#each tags as tag}
							<option value={tag.name}>{tag.name}</option>
						{/each}
					</select>
				</div>
			{/if}

		<!-- Step 2: Compose -->
		{:else if step === 2}
			<h2 class="text-xl font-semibold mb-4" style="color: var(--text-primary)">Compose Your Message</h2>
			<div class="space-y-4">
				<!-- Template -->
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Start from Template</label>
					<select bind:value={templateId} on:change={loadTemplate} class="w-full px-4 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)">
						<option value="">— Blank —</option>
						{#each templates as t}
							<option value={t.id}>{t.name} ({t.channel})</option>
						{/each}
					</select>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Campaign Name *</label>
						<input type="text" bind:value={name} placeholder="Easter Sunday Invitation" class="w-full px-4 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Channel</label>
						<div class="flex gap-4 mt-2">
							<label class="flex items-center gap-2 cursor-pointer">
								<input type="radio" bind:group={channel} value="email" />
								<span style="color: var(--text-primary)">📧 Email</span>
							</label>
							<label class="flex items-center gap-2 cursor-pointer">
								<input type="radio" bind:group={channel} value="sms" />
								<span style="color: var(--text-primary)">💬 SMS</span>
							</label>
						</div>
					</div>
				</div>

				{#if channel === 'email'}
					<div class="grid grid-cols-2 gap-4">
						<div>
							<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Subject Line *</label>
							<input type="text" bind:value={subject} placeholder="You're Invited!" class="w-full px-4 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
						</div>
						<div>
							<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Sender Name</label>
							<input type="text" bind:value={senderName} placeholder="Pastor John" class="w-full px-4 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
						</div>
					</div>
				{/if}

				<div>
					<div class="flex items-center justify-between mb-1">
						<label class="text-sm font-medium" style="color: var(--text-secondary)">Message Body *</label>
						<div class="flex gap-1">
							{#each ['first_name', 'last_name', 'church_name'] as v}
								<button type="button" on:click={() => insertVariable(v)} class="text-xs px-2 py-1 rounded border" style="background: var(--bg); border-color: var(--border); color: var(--teal)">
									{`{${v}}`}
								</button>
							{/each}
						</div>
					</div>
					<textarea bind:value={body} rows="12" placeholder="Hi {{first_name}}, ..." class="w-full px-4 py-2 rounded-lg border font-mono text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"></textarea>
				</div>
			</div>

		<!-- Step 3: Preview -->
		{:else if step === 3}
			<h2 class="text-xl font-semibold mb-4" style="color: var(--text-primary)">Preview</h2>
			{#if channel === 'email'}
				<div class="rounded-lg border overflow-hidden" style="border-color: var(--border)">
					<!-- Email header -->
					<div class="p-4 border-b" style="background: var(--bg); border-color: var(--border)">
						<div class="text-xs mb-1" style="color: var(--text-secondary)">From: {senderName || 'Your Church'}</div>
						<div class="text-xs mb-2" style="color: var(--text-secondary)">To: {getTargetLabel()}</div>
						<div class="font-semibold" style="color: var(--text-primary)">{subject}</div>
					</div>
					<!-- Email body -->
					<div class="p-6" style="background: white; color: #333">
						<div class="whitespace-pre-wrap text-sm">{@html getPreviewBody().replace(/\n/g, '<br>')}</div>
					</div>
				</div>
			{:else}
				<div class="max-w-sm mx-auto">
					<div class="rounded-2xl p-4 shadow-lg" style="background: var(--bg); border: 1px solid var(--border)">
						<div class="text-xs mb-2 text-center" style="color: var(--text-secondary)">SMS Preview</div>
						<div class="rounded-xl p-3" style="background: var(--teal); color: white">
							<div class="text-sm whitespace-pre-wrap">{getPreviewBody()}</div>
						</div>
					</div>
				</div>
			{/if}

		<!-- Step 4: Confirm -->
		{:else if step === 4}
			<h2 class="text-xl font-semibold mb-4" style="color: var(--text-primary)">Confirm & Send</h2>
			<div class="space-y-4">
				<div class="grid grid-cols-2 gap-4">
					<div class="p-4 rounded-lg" style="background: var(--bg)">
						<div class="text-sm font-medium" style="color: var(--text-secondary)">Campaign</div>
						<div class="font-medium" style="color: var(--text-primary)">{name}</div>
					</div>
					<div class="p-4 rounded-lg" style="background: var(--bg)">
						<div class="text-sm font-medium" style="color: var(--text-secondary)">Channel</div>
						<div class="font-medium" style="color: var(--text-primary)">{channel === 'email' ? '📧 Email' : '💬 SMS'}</div>
					</div>
					<div class="p-4 rounded-lg" style="background: var(--bg)">
						<div class="text-sm font-medium" style="color: var(--text-secondary)">Audience</div>
						<div class="font-medium" style="color: var(--text-primary)">{getTargetLabel()}</div>
					</div>
					{#if subject}
						<div class="p-4 rounded-lg" style="background: var(--bg)">
							<div class="text-sm font-medium" style="color: var(--text-secondary)">Subject</div>
							<div class="font-medium" style="color: var(--text-primary)">{subject}</div>
						</div>
					{/if}
				</div>

				<div class="flex gap-3 pt-4">
					<button on:click={() => createAndSend(false)} disabled={loading}
						class="flex-1 px-6 py-3 rounded-lg font-medium border"
						style="background: var(--surface); border-color: var(--border); color: var(--text-primary); opacity: {loading ? 0.5 : 1}">
						{loading ? 'Saving...' : 'Save as Draft'}
					</button>
					<button on:click={() => createAndSend(true)} disabled={loading}
						class="flex-1 px-6 py-3 rounded-lg font-medium"
						style="background: var(--teal); color: white; opacity: {loading ? 0.5 : 1}">
						{loading ? 'Sending...' : '🚀 Send Now'}
					</button>
				</div>
			</div>
		{/if}
	</div>

	<!-- Navigation -->
	{#if step < 4}
		<div class="flex justify-between mt-6">
			<button on:click={prevStep} class="px-6 py-2 rounded-lg font-medium border" disabled={step === 1}
				style="background: var(--surface); border-color: var(--border); color: var(--text-primary); opacity: {step === 1 ? 0.3 : 1}">
				← Back
			</button>
			<button on:click={nextStep} class="px-6 py-2 rounded-lg font-medium"
				style="background: var(--teal); color: white">
				Next →
			</button>
		</div>
	{:else}
		<div class="mt-6">
			<button on:click={prevStep} class="px-6 py-2 rounded-lg font-medium border"
				style="background: var(--surface); border-color: var(--border); color: var(--text-primary)">
				← Back
			</button>
		</div>
	{/if}
</div>
