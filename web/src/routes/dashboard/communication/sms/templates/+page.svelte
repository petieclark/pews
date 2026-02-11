<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	let templates: any[] = [];
	let showForm = false;
	let editingId = '';
	let name = '';
	let body = '';
	let variables = '';
	let loading = false;
	let success = '';
	let error = '';

	onMount(async () => {
		await loadTemplates();
	});

	async function loadTemplates() {
		const res = await fetch('/api/sms/templates', {
			headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
		});
		if (res.ok) {
			templates = await res.json();
		}
	}

	async function saveTemplate() {
		loading = true;
		error = '';
		success = '';

		try {
			const method = editingId ? 'PUT' : 'POST';
			const url = editingId ? `/api/sms/templates/${editingId}` : '/api/sms/templates';

			const res = await fetch(url, {
				method,
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${localStorage.getItem('token')}`
				},
				body: JSON.stringify({ name, body, variables })
			});

			if (res.ok) {
				success = editingId ? 'Template updated!' : 'Template created!';
				showForm = false;
				resetForm();
				await loadTemplates();
			} else {
				error = await res.text();
			}
		} catch (err: any) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	async function deleteTemplate(id: string) {
		if (!confirm('Delete this template?')) return;

		const res = await fetch(`/api/sms/templates/${id}`, {
			method: 'DELETE',
			headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
		});

		if (res.ok) {
			success = 'Template deleted!';
			await loadTemplates();
		}
	}

	function editTemplate(template: any) {
		editingId = template.id;
		name = template.name;
		body = template.body;
		variables = template.variables || '';
		showForm = true;
	}

	function resetForm() {
		editingId = '';
		name = '';
		body = '';
		variables = '';
	}

	function cancelEdit() {
		showForm = false;
		resetForm();
	}
</script>

<div class="max-w-4xl mx-auto p-6">
	<div class="flex justify-between items-center mb-6">
		<h1 class="text-3xl font-bold">SMS Templates</h1>
		<div class="flex gap-2">
			<button
				on:click={() => goto('/dashboard/communication/sms')}
				class="bg-gray-200 hover:bg-gray-300 px-4 py-2 rounded"
			>
				← Back to SMS
			</button>
			<button
				on:click={() => (showForm = true)}
				class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded"
			>
				+ New Template
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

	<!-- Template Form -->
	{#if showForm}
		<div class="bg-white shadow rounded-lg p-6 mb-6">
			<h2 class="text-xl font-semibold mb-4">{editingId ? 'Edit' : 'Create'} Template</h2>

			<form on:submit|preventDefault={saveTemplate}>
				<div class="mb-4">
					<label class="block text-sm font-medium mb-2">Template Name</label>
					<input
						type="text"
						bind:value={name}
						placeholder="Welcome Message"
						class="w-full border rounded px-3 py-2"
						required
					/>
				</div>

				<div class="mb-4">
					<label class="block text-sm font-medium mb-2">Message Body</label>
					<textarea
						bind:value={body}
						placeholder="Hi {first_name}, welcome to {church_name}!"
						class="w-full border rounded px-3 py-2 h-32"
						required
					></textarea>
					<p class="text-sm text-gray-600 mt-1">
						Use merge fields: {'{first_name}'}, {'{last_name}'}, {'{church_name}'}
					</p>
				</div>

				<div class="mb-4">
					<label class="block text-sm font-medium mb-2">Variables (comma-separated)</label>
					<input
						type="text"
						bind:value={variables}
						placeholder="first_name,last_name,church_name"
						class="w-full border rounded px-3 py-2"
					/>
				</div>

				<div class="flex gap-2">
					<button
						type="submit"
						disabled={loading}
						class="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded disabled:opacity-50"
					>
						{loading ? 'Saving...' : 'Save Template'}
					</button>
					<button
						type="button"
						on:click={cancelEdit}
						class="bg-gray-200 hover:bg-gray-300 px-6 py-2 rounded"
					>
						Cancel
					</button>
				</div>
			</form>
		</div>
	{/if}

	<!-- Templates List -->
	<div class="bg-white shadow rounded-lg p-6">
		<h2 class="text-xl font-semibold mb-4">Your Templates</h2>

		{#if templates.length > 0}
			<div class="space-y-4">
				{#each templates as template}
					<div class="border rounded p-4 hover:bg-gray-50">
						<div class="flex justify-between items-start">
							<div class="flex-1">
								<h3 class="font-semibold text-lg">{template.name}</h3>
								<p class="text-gray-700 mt-2">{template.body}</p>
								{#if template.variables}
									<p class="text-sm text-gray-500 mt-2">
										Variables: {template.variables}
									</p>
								{/if}
								<p class="text-xs text-gray-400 mt-2">
									Created: {new Date(template.created_at).toLocaleDateString()}
								</p>
							</div>
							<div class="flex gap-2 ml-4">
								<button
									on:click={() => editTemplate(template)}
									class="text-blue-600 hover:underline text-sm"
								>
									Edit
								</button>
								<button
									on:click={() => deleteTemplate(template.id)}
									class="text-red-600 hover:underline text-sm"
								>
									Delete
								</button>
							</div>
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<p class="text-gray-600">No templates yet. Create your first template!</p>
		{/if}
	</div>
</div>
