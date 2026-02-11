<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	$: campaignId = $page.params.id;

	let campaign = null;
	let steps = [];
	let enrollments = [];
	let people = [];
	let loading = true;
	let error = '';
	let activeTab = 'steps'; // steps, enrollments

	// Step editor
	let showStepEditor = false;
	let editingStep = null;
	let stepForm = {
		step_order: 1,
		delay_days: 0,
		action_type: 'email',
		subject: '',
		body: ''
	};

	onMount(async () => {
		await loadCampaign();
		await loadSteps();
		await loadEnrollments();
		await loadPeople();
	});

	async function loadCampaign() {
		if (campaignId === 'new') {
			campaign = {
				name: '',
				trigger_event: 'new_member',
				is_active: true
			};
			loading = false;
			return;
		}

		try {
			loading = true;
			campaign = await api(`/api/drip/campaigns/${campaignId}`);
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	async function loadSteps() {
		if (campaignId === 'new') return;
		try {
			steps = await api(`/api/drip/campaigns/${campaignId}/steps`);
			steps.sort((a, b) => a.step_order - b.step_order);
		} catch (err) {
			error = err.message;
		}
	}

	async function loadEnrollments() {
		if (campaignId === 'new') return;
		try {
			enrollments = await api(`/api/drip/campaigns/${campaignId}/enrollments`);
		} catch (err) {
			error = err.message;
		}
	}

	async function loadPeople() {
		try {
			people = await api('/api/people');
		} catch (err) {
			console.error('Failed to load people:', err);
		}
	}

	async function saveCampaign() {
		try {
			if (campaignId === 'new') {
				const created = await api('/api/drip/campaigns', {
					method: 'POST',
					body: JSON.stringify(campaign)
				});
				goto(`/dashboard/communication/drip/${created.id}`);
			} else {
				await api(`/api/drip/campaigns/${campaignId}`, {
					method: 'PUT',
					body: JSON.stringify(campaign)
				});
				alert('Campaign saved!');
			}
		} catch (err) {
			error = err.message;
		}
	}

	function openStepEditor(step = null) {
		if (step) {
			editingStep = step;
			stepForm = { ...step };
		} else {
			editingStep = null;
			stepForm = {
				step_order: steps.length + 1,
				delay_days: 0,
				action_type: 'email',
				subject: '',
				body: ''
			};
		}
		showStepEditor = true;
	}

	async function saveStep() {
		try {
			if (editingStep) {
				await api(`/api/drip/campaigns/${campaignId}/steps/${editingStep.id}`, {
					method: 'PUT',
					body: JSON.stringify(stepForm)
				});
			} else {
				await api(`/api/drip/campaigns/${campaignId}/steps`, {
					method: 'POST',
					body: JSON.stringify(stepForm)
				});
			}
			showStepEditor = false;
			await loadSteps();
		} catch (err) {
			error = err.message;
		}
	}

	async function deleteStep(stepId) {
		if (!confirm('Are you sure you want to delete this step?')) return;
		try {
			await api(`/api/drip/campaigns/${campaignId}/steps/${stepId}`, {
				method: 'DELETE'
			});
			await loadSteps();
		} catch (err) {
			error = err.message;
		}
	}

	async function enrollPerson(personId) {
		try {
			await api(`/api/drip/campaigns/${campaignId}/enroll/${personId}`, {
				method: 'POST'
			});
			await loadEnrollments();
			alert('Person enrolled successfully!');
		} catch (err) {
			error = err.message;
		}
	}

	function getActionLabel(action) {
		const labels = {
			'email': 'Email',
			'sms': 'SMS',
			'follow_up': 'Follow-up Task'
		};
		return labels[action] || action;
	}
</script>

<div>
	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2" style="border-color: var(--teal)"></div>
		</div>
	{:else if campaign}
		<!-- Header -->
		<div class="mb-6">
			<button
				on:click={() => goto('/dashboard/communication/drip')}
				class="text-sm mb-2"
				style="color: var(--text-secondary)"
			>
				← Back to Campaigns
			</button>
			<h1 class="text-3xl font-bold" style="color: var(--text-primary)">
				{campaignId === 'new' ? 'Create Campaign' : campaign.name}
			</h1>
		</div>

		<!-- Campaign Settings -->
		<div class="rounded-lg shadow border p-6 mb-6" style="background: var(--surface); border-color: var(--border)">
			<h2 class="text-xl font-semibold mb-4" style="color: var(--text-primary)">Campaign Settings</h2>
			
			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-primary)">Name</label>
					<input
						type="text"
						bind:value={campaign.name}
						class="w-full px-3 py-2 rounded-lg border"
						style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
						placeholder="e.g., New Visitor Welcome"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-primary)">Trigger Event</label>
					<select
						bind:value={campaign.trigger_event}
						class="w-full px-3 py-2 rounded-lg border"
						style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
					>
						<option value="new_member">New Member</option>
						<option value="connection_card">Connection Card Submitted</option>
						<option value="first_visit">First Visit</option>
					</select>
				</div>

				<div class="flex items-center gap-2">
					<input
						type="checkbox"
						id="is_active"
						bind:checked={campaign.is_active}
						class="w-4 h-4 rounded"
						style="accent-color: var(--teal)"
					/>
					<label for="is_active" class="text-sm font-medium" style="color: var(--text-primary)">
						Campaign is active
					</label>
				</div>

				<button
					on:click={saveCampaign}
					class="px-6 py-2 rounded-lg font-medium"
					style="background: var(--teal); color: white"
				>
					Save Campaign
				</button>
			</div>
		</div>

		{#if campaignId !== 'new'}
			<!-- Tabs -->
			<div class="flex gap-4 mb-6 border-b" style="border-color: var(--border)">
				<button
					on:click={() => activeTab = 'steps'}
					class="px-4 py-2 font-medium border-b-2 -mb-px"
					style="color: {activeTab === 'steps' ? 'var(--teal)' : 'var(--text-secondary)'}; border-color: {activeTab === 'steps' ? 'var(--teal)' : 'transparent'}"
				>
					Steps ({steps.length})
				</button>
				<button
					on:click={() => activeTab = 'enrollments'}
					class="px-4 py-2 font-medium border-b-2 -mb-px"
					style="color: {activeTab === 'enrollments' ? 'var(--teal)' : 'var(--text-secondary)'}; border-color: {activeTab === 'enrollments' ? 'var(--teal)' : 'transparent'}"
				>
					Enrollments ({enrollments.length})
				</button>
			</div>

			<!-- Steps Tab -->
			{#if activeTab === 'steps'}
				<div class="rounded-lg shadow border p-6" style="background: var(--surface); border-color: var(--border)">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-xl font-semibold" style="color: var(--text-primary)">Campaign Steps</h2>
						<button
							on:click={() => openStepEditor()}
							class="px-4 py-2 rounded-lg font-medium"
							style="background: var(--teal); color: white"
						>
							Add Step
						</button>
					</div>

					{#if steps.length === 0}
						<div class="text-center py-8" style="color: var(--text-secondary)">
							No steps yet. Add your first step to get started.
						</div>
					{:else}
						<div class="space-y-4">
							{#each steps as step, i}
								<div class="border rounded-lg p-4" style="border-color: var(--border)">
									<div class="flex items-start justify-between">
										<div class="flex-1">
											<div class="flex items-center gap-2 mb-2">
												<span class="font-semibold text-lg" style="color: var(--text-primary)">
													Step {step.step_order}: Day {step.delay_days}
												</span>
												<span class="px-2 py-1 rounded text-xs font-medium" style="background: var(--surface-hover); color: var(--text-secondary)">
													{getActionLabel(step.action_type)}
												</span>
											</div>
											{#if step.subject}
												<div class="font-medium mb-1" style="color: var(--text-primary)">{step.subject}</div>
											{/if}
											<div class="text-sm" style="color: var(--text-secondary)">{step.body.substring(0, 150)}{step.body.length > 150 ? '...' : ''}</div>
										</div>
										<div class="flex gap-2 ml-4">
											<button
												on:click={() => openStepEditor(step)}
												class="px-3 py-1 rounded text-sm"
												style="background: var(--surface-hover); color: var(--text-primary)"
											>
												Edit
											</button>
											<button
												on:click={() => deleteStep(step.id)}
												class="px-3 py-1 rounded text-sm"
												style="background: var(--surface-hover); color: #c33"
											>
												Delete
											</button>
										</div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			{/if}

			<!-- Enrollments Tab -->
			{#if activeTab === 'enrollments'}
				<div class="rounded-lg shadow border p-6" style="background: var(--surface); border-color: var(--border)">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-xl font-semibold" style="color: var(--text-primary)">Enrolled People</h2>
						<div>
							<select
								on:change={(e) => enrollPerson(e.target.value)}
								class="px-3 py-2 rounded-lg border"
								style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
							>
								<option value="">Enroll someone...</option>
								{#each people as person}
									<option value={person.id}>{person.first_name} {person.last_name}</option>
								{/each}
							</select>
						</div>
					</div>

					{#if enrollments.length === 0}
						<div class="text-center py-8" style="color: var(--text-secondary)">
							No one enrolled yet.
						</div>
					{:else}
						<div class="space-y-2">
							{#each enrollments as enrollment}
								<div class="flex items-center justify-between p-3 border rounded-lg" style="border-color: var(--border)">
									<div>
										<div class="font-medium" style="color: var(--text-primary)">{enrollment.person_name}</div>
										<div class="text-sm" style="color: var(--text-secondary)">
											Enrolled {new Date(enrollment.enrolled_at).toLocaleDateString()}
											· Status: {enrollment.status}
										</div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		{/if}
	{/if}

	{#if error}
		<div class="mt-4 px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">
			{error}
		</div>
	{/if}
</div>

<!-- Step Editor Modal -->
{#if showStepEditor}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" on:click={() => showStepEditor = false}>
		<div class="rounded-lg shadow-xl p-6 max-w-2xl w-full mx-4" style="background: var(--surface)" on:click|stopPropagation>
			<h2 class="text-2xl font-bold mb-4" style="color: var(--text-primary)">
				{editingStep ? 'Edit Step' : 'Add Step'}
			</h2>

			<div class="space-y-4">
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-primary)">Step Order</label>
						<input
							type="number"
							bind:value={stepForm.step_order}
							min="1"
							class="w-full px-3 py-2 rounded-lg border"
							style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-primary)">Delay (days)</label>
						<input
							type="number"
							bind:value={stepForm.delay_days}
							min="0"
							class="w-full px-3 py-2 rounded-lg border"
							style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
						/>
					</div>
				</div>

				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-primary)">Action Type</label>
					<select
						bind:value={stepForm.action_type}
						class="w-full px-3 py-2 rounded-lg border"
						style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
					>
						<option value="email">Email</option>
						<option value="sms">SMS</option>
						<option value="follow_up">Follow-up Task</option>
					</select>
				</div>

				{#if stepForm.action_type === 'email'}
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-primary)">Subject</label>
						<input
							type="text"
							bind:value={stepForm.subject}
							class="w-full px-3 py-2 rounded-lg border"
							style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
							placeholder="Email subject line"
						/>
					</div>
				{/if}

				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-primary)">Message</label>
					<textarea
						bind:value={stepForm.body}
						rows="6"
						class="w-full px-3 py-2 rounded-lg border"
						style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
						placeholder="Message content..."
					></textarea>
					<p class="text-xs mt-1" style="color: var(--text-secondary)">
						You can use variables like: first_name, last_name, email
					</p>
				</div>

				<div class="flex gap-2 justify-end">
					<button
						on:click={() => showStepEditor = false}
						class="px-4 py-2 rounded-lg font-medium border"
						style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"
					>
						Cancel
					</button>
					<button
						on:click={saveStep}
						class="px-4 py-2 rounded-lg font-medium"
						style="background: var(--teal); color: white"
					>
						Save Step
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
