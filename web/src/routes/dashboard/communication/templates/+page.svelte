<script>
	import { onMount } from 'svelte';
	import { Mail, FileEdit } from 'lucide-svelte';
	import { api } from '$lib/api';

	let templates = [];
	let loading = true;
	let error = '';
	let channelFilter = 'all';

	// Create/Edit
	let showForm = false;
	let editId = '';
	let name = '';
	let subject = '';
	let body = '';
	let channel = 'email';
	let category = '';
	let variables = '';
	let saving = false;

	onMount(async () => {
		await loadTemplates();
	});

	async function loadTemplates() {
		try {
			loading = true;
			templates = await api('/api/communication/templates');
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	$: filteredTemplates = channelFilter === 'all' ? templates : templates.filter(t => t.channel === channelFilter);

	function openCreate() {
		editId = ''; name = ''; subject = ''; body = ''; channel = 'email'; category = ''; variables = '';
		showForm = true;
	}

	function openEdit(t) {
		editId = t.id; name = t.name; subject = t.subject || ''; body = t.body; channel = t.channel; category = t.category || ''; variables = t.variables || '';
		showForm = true;
	}

	async function saveTemplate() {
		if (!name || !body) { error = 'Name and body are required'; return; }
		saving = true;
		error = '';
		try {
			const payload = { name, subject, body, channel, category, variables };
			if (editId) {
				await api(`/api/communication/templates/${editId}`, { method: 'PUT', body: JSON.stringify(payload) });
			} else {
				await api('/api/communication/templates', { method: 'POST', body: JSON.stringify(payload) });
			}
			showForm = false;
			await loadTemplates();
		} catch (err) {
			error = err.message;
		} finally {
			saving = false;
		}
	}

	async function deleteTemplate(id) {
		if (!confirm('Delete this template?')) return;
		try {
			await api(`/api/communication/templates/${id}`, { method: 'DELETE' });
			await loadTemplates();
		} catch (err) {
			error = err.message;
		}
	}

	const categories = ['welcome', 'follow_up', 'announcement', 'giving', 'custom'];
</script>

<div>
	<div class="flex items-center justify-between mb-6">
		<div>
			<a href="/dashboard/communication" class="text-sm font-medium" style="color: var(--teal)">← Communication</a>
			<h1 class="text-3xl font-bold mt-1" style="color: var(--text-primary)">Message Templates</h1>
		</div>
		<button on:click={openCreate} class="px-4 py-2 rounded-lg font-medium" style="background: var(--teal); color: white">New Template</button>
	</div>

	{#if error}
		<div class="mb-4 px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">{error}</div>
	{/if}

	<!-- Filter -->
	<div class="mb-6 flex gap-2">
		{#each ['all', 'email', 'sms'] as ch}
			<button on:click={() => channelFilter = ch} class="px-4 py-2 rounded-lg font-medium transition capitalize" style="background: {channelFilter === ch ? 'var(--teal)' : 'var(--surface)'}; color: {channelFilter === ch ? 'white' : 'var(--text-primary)'}; border: 1px solid {channelFilter === ch ? 'var(--teal)' : 'var(--border)'}">
				{#if ch === 'all'}All{:else if ch === 'email'}<Mail size={14} class="inline" /> Email{:else}💬 SMS{/if}
			</button>
		{/each}
	</div>

	<!-- Create/Edit Form -->
	{#if showForm}
		<div class="mb-6 rounded-lg shadow border p-6" style="background: var(--surface); border-color: var(--border)">
			<h2 class="text-lg font-semibold mb-4" style="color: var(--text-primary)">{editId ? 'Edit' : 'New'} Template</h2>
			<div class="space-y-4">
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Name</label>
						<input bind:value={name} type="text" placeholder="Template name" class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Channel</label>
						<select bind:value={channel} class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)">
							<option value="email">Email</option>
							<option value="sms">SMS</option>
						</select>
					</div>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Category</label>
						<select bind:value={category} class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)">
							<option value="">None</option>
							{#each categories as cat}
								<option value={cat}>{cat.replace('_', ' ')}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Variables (comma-separated)</label>
						<input bind:value={variables} type="text" placeholder="first_name,church_name" class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
					</div>
				</div>
				{#if channel === 'email'}
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Subject</label>
						<input bind:value={subject} type="text" placeholder="Email subject" class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
					</div>
				{/if}
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Body</label>
					<textarea bind:value={body} rows="6" placeholder="Message body... Use {{variable}} for personalization" class="w-full px-3 py-2 rounded-lg border font-mono text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"></textarea>
				</div>
				<div class="flex gap-3">
					<button on:click={saveTemplate} disabled={saving} class="px-4 py-2 rounded-lg font-medium" style="background: var(--teal); color: white; opacity: {saving ? 0.5 : 1}">
						{saving ? 'Saving...' : editId ? 'Update' : 'Create'}
					</button>
					<button on:click={() => showForm = false} class="px-4 py-2 rounded-lg font-medium border" style="background: var(--surface); border-color: var(--border); color: var(--text-primary)">Cancel</button>
				</div>
			</div>
		</div>
	{/if}

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2" style="border-color: var(--teal)"></div>
		</div>
	{:else if filteredTemplates.length === 0}
		<div class="rounded-lg shadow border p-12 text-center" style="background: var(--surface); border-color: var(--border)">
			<div class="mb-4"><FileEdit size={48} /></div>
			<h2 class="text-xl font-semibold mb-2" style="color: var(--text-primary)">No templates yet</h2>
			<p class="mb-6" style="color: var(--text-secondary)">Create reusable templates for your campaigns and journeys.</p>
			<button on:click={openCreate} class="px-6 py-2 rounded-lg font-medium" style="background: var(--teal); color: white">Create Template</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each filteredTemplates as template}
				<div class="rounded-lg shadow border p-5 hover:shadow-md transition" style="background: var(--surface); border-color: var(--border)">
					<div class="flex items-start justify-between mb-2">
						<h3 class="font-semibold" style="color: var(--text-primary)">{template.name}</h3>
						{#if template.channel === 'email'}<Mail size={20} />{:else}<span class="text-lg">💬</span>{/if}
					</div>
					{#if template.category}
						<span class="px-2 py-0.5 rounded text-xs font-medium capitalize" style="background: var(--bg); color: var(--text-secondary)">{template.category.replace('_', ' ')}</span>
					{/if}
					{#if template.subject}
						<div class="text-sm mt-2 font-medium" style="color: var(--text-primary)">{template.subject}</div>
					{/if}
					<div class="text-sm mt-2 line-clamp-3" style="color: var(--text-secondary)">{template.body}</div>
					{#if template.variables}
						<div class="mt-2 flex flex-wrap gap-1">
							{#each template.variables.split(',') as v}
								<span class="px-1.5 py-0.5 rounded text-xs font-mono" style="background: var(--bg); color: var(--teal)">{`{{${v.trim()}}}`}</span>
							{/each}
						</div>
					{/if}
					<div class="mt-3 pt-3 border-t flex gap-2" style="border-color: var(--border)">
						<button on:click={() => openEdit(template)} class="text-xs px-2 py-1 rounded" style="color: var(--teal)">Edit</button>
						<button on:click={() => deleteTemplate(template.id)} class="text-xs px-2 py-1 rounded" style="color: #c33">Delete</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
