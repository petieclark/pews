<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { Music, Heart, BookOpen, Megaphone, MessageSquare } from 'lucide-svelte';

	let templates = [];
	let loading = true;
	let showCreateModal = false;
	let showEditModal = false;
	let editingTemplate = null;

	let newTemplate = {
		name: '',
		description: '',
		template_data: { items: [], roles: [] }
	};

	let newItemType = 'song';
	let newItemTitle = '';

	const itemTypeConfig = {
		song: { icon: Music, label: 'Song' },
		prayer: { icon: Heart, label: 'Prayer' },
		reading: { icon: BookOpen, label: 'Scripture' },
		sermon: { icon: Megaphone, label: 'Sermon' },
		announcement: { icon: MessageSquare, label: 'Announcement' },
		transition: { icon: null, label: 'Transition' },
		other: { icon: null, label: 'Other' }
	};

	onMount(loadTemplates);

	async function loadTemplates() {
		loading = true;
		try {
			templates = await api('/api/services/templates');
		} catch (error) {
			console.error('Failed to load templates:', error);
			templates = [];
		} finally {
			loading = false;
		}
	}

	function addItemToTemplate(form) {
		if (!newItemTitle && newItemType !== 'song') return;
		form.template_data.items = [...form.template_data.items, {
			item_type: newItemType,
			title: newItemTitle || `${itemTypeConfig[newItemType]?.label || 'Item'}`,
			position: form.template_data.items.length + 1,
			notes: ''
		}];
		newItemTitle = '';
		if (editingTemplate) editingTemplate = { ...editingTemplate };
		else newTemplate = { ...newTemplate };
	}

	function removeItemFromTemplate(form, index) {
		form.template_data.items = form.template_data.items.filter((_, i) => i !== index);
		// Reposition
		form.template_data.items.forEach((item, i) => { item.position = i + 1; });
		if (editingTemplate) editingTemplate = { ...editingTemplate };
		else newTemplate = { ...newTemplate };
	}

	async function createTemplate() {
		try {
			await api('/api/services/templates', {
				method: 'POST',
				body: JSON.stringify({
					name: newTemplate.name,
					description: newTemplate.description,
					template_data: newTemplate.template_data
				})
			});
			showCreateModal = false;
			newTemplate = { name: '', description: '', template_data: { items: [], roles: [] } };
			loadTemplates();
		} catch (error) {
			alert('Failed to create template');
		}
	}

	function openEditModal(template) {
		const data = typeof template.template_data === 'string' ? JSON.parse(template.template_data) : template.template_data;
		editingTemplate = {
			id: template.id,
			name: template.name,
			description: template.description || '',
			template_data: {
				items: data.items || [],
				roles: data.roles || []
			}
		};
		showEditModal = true;
	}

	async function updateTemplate() {
		try {
			await api(`/api/services/templates/${editingTemplate.id}`, {
				method: 'PUT',
				body: JSON.stringify({
					name: editingTemplate.name,
					description: editingTemplate.description,
					template_data: editingTemplate.template_data
				})
			});
			showEditModal = false;
			editingTemplate = null;
			loadTemplates();
		} catch (error) {
			alert('Failed to update template');
		}
	}

	async function deleteTemplate(id) {
		if (!confirm('Delete this template?')) return;
		try {
			await api(`/api/services/templates/${id}`, { method: 'DELETE' });
			loadTemplates();
		} catch (error) {
			alert('Failed to delete template');
		}
	}

	function getItemCount(template) {
		try {
			const data = typeof template.template_data === 'string' ? JSON.parse(template.template_data) : template.template_data;
			return data.items?.length || 0;
		} catch { return 0; }
	}

	function formatDate(dateStr) {
		return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}
</script>

<div class="space-y-6 max-w-4xl mx-auto">
	<div class="flex items-center justify-between">
		<div>
			<button on:click={() => goto('/dashboard/services')} class="inline-flex items-center gap-1 text-sm text-[var(--teal)] hover:underline mb-2">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
				Services
			</button>
			<h1 class="text-3xl font-bold text-[var(--text-primary)]">Service Templates</h1>
			<p class="text-[var(--text-secondary)] mt-1">Reusable service order templates</p>
		</div>
		<button on:click={() => (showCreateModal = true)} class="btn-primary">
			<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
			New Template
		</button>
	</div>

	{#if loading}
		<div class="p-12 text-center text-[var(--text-secondary)]">
			<div class="inline-block w-6 h-6 border-2 border-[var(--teal)] border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if templates.length === 0}
		<div class="card p-12 text-center">
			<svg class="w-16 h-16 mx-auto text-[var(--text-secondary)] opacity-30 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z"/></svg>
			<p class="text-[var(--text-secondary)] text-lg">No templates yet</p>
			<p class="text-[var(--text-secondary)] text-sm mt-1">Create templates to quickly build service orders</p>
			<button on:click={() => (showCreateModal = true)} class="btn-primary mt-4">Create Template</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			{#each templates as template}
				<div class="card p-5 hover:border-[var(--teal)] transition-colors">
					<div class="flex items-start justify-between">
						<div class="flex-1">
							<h3 class="font-semibold text-[var(--text-primary)]">{template.name}</h3>
							{#if template.description}
								<p class="text-sm text-[var(--text-secondary)] mt-1">{template.description}</p>
							{/if}
							<div class="flex items-center gap-3 mt-3 text-xs text-[var(--text-secondary)]">
								<span>{getItemCount(template)} items</span>
								<span>·</span>
								<span>Created {formatDate(template.created_at)}</span>
							</div>
						</div>
						<div class="flex items-center gap-1 ml-4">
							<button on:click={() => openEditModal(template)} class="p-1.5 rounded hover:bg-[var(--surface-hover)] text-[var(--text-secondary)]" title="Edit">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
							</button>
							<button on:click={() => deleteTemplate(template.id)} class="p-1.5 rounded hover:bg-red-400/10 text-[var(--text-secondary)] hover:text-red-400" title="Delete">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Create/Edit Template Modal -->
{#if showCreateModal || showEditModal}
	{@const form = showEditModal ? editingTemplate : newTemplate}
	{@const isEdit = showEditModal}
	<div class="modal-overlay" on:click|self={() => { showCreateModal = false; showEditModal = false; }}>
		<div class="modal-content max-w-2xl">
			<div class="flex items-center justify-between mb-5">
				<h2 class="text-lg font-bold text-[var(--text-primary)]">{isEdit ? 'Edit' : 'Create'} Template</h2>
				<button on:click={() => { showCreateModal = false; showEditModal = false; }} class="close-btn">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
				</button>
			</div>

			<form on:submit|preventDefault={isEdit ? updateTemplate : createTemplate} class="space-y-5">
				<div>
					<label class="label">Template Name *</label>
					<input type="text" bind:value={form.name} required class="input-field" placeholder="e.g. Standard Sunday Morning" />
				</div>
				<div>
					<label class="label">Description</label>
					<textarea bind:value={form.description} rows="2" class="input-field" placeholder="Optional description..."></textarea>
				</div>

				<!-- Template Items -->
				<div>
					<label class="label mb-2">Service Order Items</label>
					{#if form.template_data.items.length > 0}
						<div class="space-y-1 mb-3">
							{#each form.template_data.items as item, i}
								<div class="flex items-center gap-2 p-2 bg-[var(--surface-hover)] rounded">
									<span class="text-sm w-5 text-center text-[var(--text-secondary)]">{i + 1}</span>
									<span>{#if itemTypeConfig[item.item_type]?.icon}<svelte:component this={itemTypeConfig[item.item_type].icon} size={16} />{:else}•{/if}</span>
									<span class="flex-1 text-sm text-[var(--text-primary)]">{item.title}</span>
									<span class="text-xs text-[var(--text-secondary)]">{itemTypeConfig[item.item_type]?.label}</span>
									<button type="button" on:click={() => removeItemFromTemplate(form, i)} class="p-1 text-[var(--text-secondary)] hover:text-red-400">
										<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
									</button>
								</div>
							{/each}
						</div>
					{/if}
					<div class="flex gap-2">
						<select bind:value={newItemType} class="input-field w-auto">
							{#each Object.entries(itemTypeConfig) as [key, config]}
								<option value={key}>{config.label}</option>
							{/each}
						</select>
						<input type="text" bind:value={newItemTitle} placeholder="Item title..." class="input-field flex-1"
							on:keydown={(e) => e.key === 'Enter' && (e.preventDefault(), addItemToTemplate(form))} />
						<button type="button" on:click={() => addItemToTemplate(form)} class="btn-secondary">Add</button>
					</div>
				</div>

				<div class="flex gap-3 pt-2">
					<button type="button" on:click={() => { showCreateModal = false; showEditModal = false; }} class="btn-secondary flex-1">Cancel</button>
					<button type="submit" class="btn-primary flex-1">{isEdit ? 'Save Changes' : 'Create Template'}</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
	.card {
		background: var(--surface);
		border-radius: 0.75rem;
		border: 1px solid var(--border);
	}

	.btn-primary {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 0.5rem 1rem;
		background: var(--teal);
		color: white;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
	}
	.btn-primary:hover { opacity: 0.9; }

	.btn-secondary {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 0.5rem 1rem;
		background: var(--surface);
		color: var(--text-primary);
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
	}
	.btn-secondary:hover { background: var(--surface-hover); }

	.input-field {
		display: block;
		width: 100%;
		padding: 0.5rem 0.75rem;
		background: var(--input-bg);
		color: var(--text-primary);
		border: 1px solid var(--input-border);
		border-radius: 0.5rem;
		font-size: 0.875rem;
		margin-top: 0.25rem;
	}
	.input-field:focus {
		outline: none;
		border-color: var(--teal);
		box-shadow: 0 0 0 2px rgba(74, 139, 140, 0.2);
	}

	.label {
		display: block;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text-primary);
	}

	.close-btn {
		padding: 0.25rem;
		border-radius: 0.375rem;
		color: var(--text-secondary);
	}
	.close-btn:hover { background: var(--surface-hover); }

	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.6);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
		z-index: 50;
		backdrop-filter: blur(4px);
	}
	.modal-content {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: 0.75rem;
		padding: 1.5rem;
		width: 100%;
		max-height: 90vh;
		overflow-y: auto;
	}
</style>
