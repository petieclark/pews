<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let name = '';
	let channel = 'email';
	let subject = '';
	let body = '';
	let targetType = 'all';
	let targetId = '';
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

	function loadTemplate(event) {
		const templateId = event.target.value;
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

	async function createCampaign() {
		if (!name || !body) {
			error = 'Please fill in all required fields';
			return;
		}

		try {
			loading = true;
			error = '';

			const campaign = await api('/api/communication/campaigns', {
				method: 'POST',
				body: JSON.stringify({
					name,
					channel,
					subject: channel === 'email' ? subject : null,
					body,
					target_type: targetType,
					target_id: targetType === 'all' ? null : targetId
				})
			});

			goto(`/dashboard/communication/campaigns/${campaign.id}`);
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="mb-6">
		<h1 class="text-3xl font-bold" style="color: var(--text-primary)">New Campaign</h1>
	</div>

	{#if error}
		<div class="mb-6 px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">
			{error}
		</div>
	{/if}

	<div class="rounded-lg shadow border p-6 space-y-6" style="background: var(--surface); border-color: var(--border)">
		<!-- Template Selector -->
		<div>
			<label class="block text-sm font-medium mb-2" style="color: var(--text-primary)">Start from Template (Optional)</label>
			<select
				on:change={loadTemplate}
				class="w-full px-4 py-2 rounded-lg border"
				style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
			>
				<option value="">— Select Template —</option>
				{#each templates as template}
					<option value={template.id}>{template.name} ({template.channel})</option>
				{/each}
			</select>
		</div>

		<!-- Name -->
		<div>
			<label class="block text-sm font-medium mb-2" style="color: var(--text-primary)">Campaign Name *</label>
			<input
				type="text"
				bind:value={name}
				placeholder="Easter Sunday Invitation"
				class="w-full px-4 py-2 rounded-lg border"
				style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
			/>
		</div>

		<!-- Channel -->
		<div>
			<label class="block text-sm font-medium mb-2" style="color: var(--text-primary)">Channel *</label>
			<div class="flex gap-4">
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

		<!-- Subject (email only) -->
		{#if channel === 'email'}
			<div>
				<label class="block text-sm font-medium mb-2" style="color: var(--text-primary)">Subject Line *</label>
				<input
					type="text"
					bind:value={subject}
					placeholder="You're Invited to Easter Sunday!"
					class="w-full px-4 py-2 rounded-lg border"
					style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
				/>
			</div>
		{/if}

		<!-- Body -->
		<div>
			<div class="flex items-center justify-between mb-2">
				<label class="block text-sm font-medium" style="color: var(--text-primary)">Message Body *</label>
				<div class="flex gap-2">
					<button
						type="button"
						on:click={() => insertVariable('first_name')}
						class="text-xs px-2 py-1 rounded border"
						style="background: var(--surface); border-color: var(--border); color: var(--text-primary)"
					>
						+ First Name
					</button>
					<button
						type="button"
						on:click={() => insertVariable('church_name')}
						class="text-xs px-2 py-1 rounded border"
						style="background: var(--surface); border-color: var(--border); color: var(--text-primary)"
					>
						+ Church Name
					</button>
				</div>
			</div>
			<textarea
				bind:value={body}
				rows="10"
				placeholder="Hi {{first_name}}, we'd love to see you this Easter..."
				class="w-full px-4 py-2 rounded-lg border font-mono text-sm"
				style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
			></textarea>
		</div>

		<!-- Target -->
		<div>
			<label class="block text-sm font-medium mb-2" style="color: var(--text-primary)">Send To *</label>
			<select
				bind:value={targetType}
				class="w-full px-4 py-2 rounded-lg border mb-2"
				style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
			>
				<option value="all">All Members</option>
				<option value="tag">By Tag</option>
				<option value="group">By Group</option>
			</select>

			{#if targetType === 'tag'}
				<select
					bind:value={targetId}
					class="w-full px-4 py-2 rounded-lg border"
					style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
				>
					<option value="">— Select Tag —</option>
					{#each tags as tag}
						<option value={tag.name}>{tag.name}</option>
					{/each}
				</select>
			{:else if targetType === 'group'}
				<select
					bind:value={targetId}
					class="w-full px-4 py-2 rounded-lg border"
					style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
				>
					<option value="">— Select Group —</option>
					{#each groups as group}
						<option value={group.id}>{group.name}</option>
					{/each}
				</select>
			{/if}
		</div>

		<!-- Actions -->
		<div class="flex justify-end gap-3 pt-4">
			<button
				on:click={() => goto('/dashboard/communication/campaigns')}
				class="px-6 py-2 rounded-lg font-medium border"
				style="background: var(--surface); border-color: var(--border); color: var(--text-primary)"
			>
				Cancel
			</button>
			<button
				on:click={createCampaign}
				disabled={loading}
				class="px-6 py-2 rounded-lg font-medium"
				style="background: var(--teal); color: white; opacity: {loading ? 0.5 : 1}"
			>
				{loading ? 'Creating...' : 'Create Campaign'}
			</button>
		</div>
	</div>
</div>
