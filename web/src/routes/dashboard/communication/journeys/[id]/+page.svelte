<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let journey = null;
	let enrollments = [];
	let templates = [];
	let loading = true;
	let error = '';

	// Edit state
	let editing = false;
	let editName = '';
	let editDescription = '';
	let editTriggerType = '';
	let editTriggerValue = '';
	let editIsActive = true;

	// New step
	let showAddStep = false;
	let stepType = 'send_email';
	let stepDelayDays = 0;
	let stepDelayHours = 0;
	let stepTemplateId = '';
	let stepConfig = '{}';
	let addingStep = false;

	// Enroll
	let showEnroll = false;
	let enrollPersonId = '';
	let enrolling = false;

	$: journeyId = $page.params.id;

	onMount(async () => {
		await loadJourney();
	});

	async function loadJourney() {
		try {
			loading = true;
			const [j, e, t] = await Promise.all([
				api(`/api/communication/journeys/${journeyId}`),
				api(`/api/communication/journeys/${journeyId}/enrollments`),
				api('/api/communication/templates')
			]);
			journey = j;
			enrollments = e || [];
			templates = t || [];
			editName = j.name;
			editDescription = j.description || '';
			editTriggerType = j.trigger_type;
			editTriggerValue = j.trigger_value || '';
			editIsActive = j.is_active;
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	async function saveJourney() {
		try {
			await api(`/api/communication/journeys/${journeyId}`, {
				method: 'PUT',
				body: JSON.stringify({ name: editName, description: editDescription, trigger_type: editTriggerType, trigger_value: editTriggerValue, is_active: editIsActive })
			});
			editing = false;
			await loadJourney();
		} catch (err) {
			error = err.message;
		}
	}

	async function addStep() {
		addingStep = true;
		try {
			const position = (journey.steps || []).length + 1;
			let config;
			try { config = JSON.parse(stepConfig); } catch { config = {}; }
			await api(`/api/communication/journeys/${journeyId}/steps`, {
				method: 'POST',
				body: JSON.stringify({ position, step_type: stepType, delay_days: stepDelayDays, delay_hours: stepDelayHours, template_id: stepTemplateId || undefined, config })
			});
			showAddStep = false;
			stepType = 'send_email'; stepDelayDays = 0; stepDelayHours = 0; stepTemplateId = ''; stepConfig = '{}';
			await loadJourney();
		} catch (err) {
			error = err.message;
		} finally {
			addingStep = false;
		}
	}

	async function deleteStep(stepId) {
		if (!confirm('Delete this step?')) return;
		try {
			await api(`/api/communication/journeys/${journeyId}/steps/${stepId}`, { method: 'DELETE' });
			await loadJourney();
		} catch (err) {
			error = err.message;
		}
	}

	async function enrollPerson() {
		if (!enrollPersonId) return;
		enrolling = true;
		try {
			await api(`/api/communication/journeys/${journeyId}/enroll`, {
				method: 'POST',
				body: JSON.stringify({ person_id: enrollPersonId })
			});
			showEnroll = false;
			enrollPersonId = '';
			await loadJourney();
		} catch (err) {
			error = err.message;
		} finally {
			enrolling = false;
		}
	}

	const stepIcons = { send_email: '📧', send_sms: '💬', wait: '⏳', add_tag: '🏷️', add_to_group: '📋' };
	const stepLabels = { send_email: 'Send Email', send_sms: 'Send SMS', wait: 'Wait', add_tag: 'Add Tag', add_to_group: 'Add to Group' };
	const triggerLabels = { manual: 'Manual', first_visit: 'First Visit', tag_added: 'Tag Added', group_joined: 'Group Joined', checkin_first_time: 'First Check-in' };
</script>

<div>
	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2" style="border-color: var(--teal)"></div>
		</div>
	{:else if error && !journey}
		<div class="px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">{error}</div>
	{:else if journey}
		<div class="flex items-center justify-between mb-6">
			<div>
				<a href="/dashboard/communication/journeys" class="text-sm font-medium" style="color: var(--teal)">← Journeys</a>
				{#if editing}
					<input bind:value={editName} class="text-3xl font-bold mt-1 bg-transparent border-b-2 outline-none" style="color: var(--text); border-color: var(--teal)" />
				{:else}
					<h1 class="text-3xl font-bold mt-1" style="color: var(--text)">{journey.name}</h1>
				{/if}
			</div>
			<div class="flex gap-3">
				{#if editing}
					<button on:click={saveJourney} class="px-4 py-2 rounded-lg font-medium" style="background: var(--teal); color: white">Save</button>
					<button on:click={() => editing = false} class="px-4 py-2 rounded-lg font-medium border" style="background: var(--surface); border-color: var(--border); color: var(--text)">Cancel</button>
				{:else}
					<button on:click={() => editing = true} class="px-4 py-2 rounded-lg font-medium border" style="background: var(--surface); border-color: var(--border); color: var(--text)">Edit</button>
				{/if}
			</div>
		</div>

		{#if error}
			<div class="mb-4 px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">{error}</div>
		{/if}

		<!-- Journey Info -->
		{#if editing}
			<div class="mb-6 rounded-lg shadow border p-6 space-y-4" style="background: var(--surface); border-color: var(--border)">
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Description</label>
					<input bind:value={editDescription} type="text" class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text)" />
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Trigger</label>
						<select bind:value={editTriggerType} class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text)">
							<option value="manual">Manual</option>
							<option value="first_visit">First Visit</option>
							<option value="tag_added">Tag Added</option>
							<option value="group_joined">Group Joined</option>
							<option value="checkin_first_time">First Check-in</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Trigger Value</label>
						<input bind:value={editTriggerValue} type="text" class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text)" />
					</div>
				</div>
				<label class="flex items-center gap-2 cursor-pointer">
					<input bind:checked={editIsActive} type="checkbox" class="rounded" />
					<span class="text-sm" style="color: var(--text)">Active</span>
				</label>
			</div>
		{:else}
			<div class="mb-6 flex flex-wrap gap-3">
				<span class="px-3 py-1 rounded-full text-sm font-medium" style="background: {journey.is_active ? '#d1fae5' : '#e2e8f0'}; color: {journey.is_active ? '#065f46' : '#475569'}">
					{journey.is_active ? 'Active' : 'Inactive'}
				</span>
				<span class="px-3 py-1 rounded-full text-sm border" style="background: var(--bg); border-color: var(--border); color: var(--text-secondary)">
					Trigger: {triggerLabels[journey.trigger_type] || journey.trigger_type}
					{#if journey.trigger_value}({journey.trigger_value}){/if}
				</span>
				{#if journey.description}
					<span class="text-sm" style="color: var(--text-secondary)">{journey.description}</span>
				{/if}
			</div>
		{/if}

		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Steps Timeline -->
			<div class="lg:col-span-2">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold" style="color: var(--text)">Steps</h2>
					<button on:click={() => showAddStep = !showAddStep} class="px-3 py-1 rounded-lg text-sm font-medium" style="background: var(--teal); color: white">+ Add Step</button>
				</div>

				{#if showAddStep}
					<div class="mb-4 rounded-lg shadow border p-4 space-y-3" style="background: var(--surface); border-color: var(--border)">
						<div class="grid grid-cols-2 gap-3">
							<div>
								<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Step Type</label>
								<select bind:value={stepType} class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text)">
									<option value="send_email">Send Email</option>
									<option value="send_sms">Send SMS</option>
									<option value="wait">Wait</option>
									<option value="add_tag">Add Tag</option>
									<option value="add_to_group">Add to Group</option>
								</select>
							</div>
							{#if stepType === 'send_email' || stepType === 'send_sms'}
								<div>
									<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Template</label>
									<select bind:value={stepTemplateId} class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text)">
										<option value="">Select template...</option>
										{#each templates.filter(t => stepType === 'send_email' ? t.channel === 'email' : t.channel === 'sms') as t}
											<option value={t.id}>{t.name}</option>
										{/each}
									</select>
								</div>
							{/if}
						</div>
						<div class="grid grid-cols-2 gap-3">
							<div>
								<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Delay (days)</label>
								<input bind:value={stepDelayDays} type="number" min="0" class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text)" />
							</div>
							<div>
								<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Delay (hours)</label>
								<input bind:value={stepDelayHours} type="number" min="0" class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text)" />
							</div>
						</div>
						{#if stepType === 'add_tag' || stepType === 'add_to_group'}
							<div>
								<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Config (JSON)</label>
								<input bind:value={stepConfig} type="text" placeholder='{"tag_name": "followed-up"}' class="w-full px-3 py-2 rounded-lg border text-sm font-mono" style="background: var(--bg); border-color: var(--border); color: var(--text)" />
							</div>
						{/if}
						<div class="flex gap-2">
							<button on:click={addStep} disabled={addingStep} class="px-3 py-1 rounded-lg text-sm font-medium" style="background: var(--teal); color: white">{addingStep ? 'Adding...' : 'Add Step'}</button>
							<button on:click={() => showAddStep = false} class="px-3 py-1 rounded-lg text-sm border" style="background: var(--surface); border-color: var(--border); color: var(--text)">Cancel</button>
						</div>
					</div>
				{/if}

				<!-- Timeline -->
				{#if !journey.steps || journey.steps.length === 0}
					<div class="rounded-lg border p-8 text-center" style="background: var(--surface); border-color: var(--border)">
						<p style="color: var(--text-secondary)">No steps yet. Add steps to build your journey.</p>
					</div>
				{:else}
					<div class="space-y-0">
						{#each journey.steps as step, i}
							<div class="flex items-start gap-4">
								<!-- Timeline connector -->
								<div class="flex flex-col items-center">
									<div class="w-10 h-10 rounded-full flex items-center justify-center text-lg border-2" style="background: var(--surface); border-color: var(--teal)">
										{stepIcons[step.step_type] || '•'}
									</div>
									{#if i < journey.steps.length - 1}
										<div class="w-0.5 h-12" style="background: var(--border)"></div>
									{/if}
								</div>
								<!-- Step card -->
								<div class="flex-1 rounded-lg border p-4 mb-2" style="background: var(--surface); border-color: var(--border)">
									<div class="flex items-center justify-between">
										<div>
											<span class="font-medium" style="color: var(--text)">{stepLabels[step.step_type] || step.step_type}</span>
											{#if step.delay_days > 0 || step.delay_hours > 0}
												<span class="text-sm ml-2" style="color: var(--text-secondary)">
													after {step.delay_days > 0 ? `${step.delay_days}d` : ''}{step.delay_hours > 0 ? ` ${step.delay_hours}h` : ''}
												</span>
											{/if}
										</div>
										<button on:click={() => deleteStep(step.id)} class="text-xs px-2 py-1 rounded hover:bg-red-50" style="color: #c33">×</button>
									</div>
									{#if step.template_id}
										{@const tmpl = templates.find(t => t.id === step.template_id)}
										{#if tmpl}
											<div class="text-sm mt-1" style="color: var(--text-secondary)">Template: {tmpl.name}</div>
										{/if}
									{/if}
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Enrollments -->
			<div>
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold" style="color: var(--text)">Enrollments</h2>
					<button on:click={() => showEnroll = !showEnroll} class="px-3 py-1 rounded-lg text-sm font-medium" style="background: var(--teal); color: white">+ Enroll</button>
				</div>

				{#if showEnroll}
					<div class="mb-4 rounded-lg border p-4 space-y-3" style="background: var(--surface); border-color: var(--border)">
						<div>
							<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Person ID</label>
							<input bind:value={enrollPersonId} type="text" placeholder="Person UUID" class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text)" />
						</div>
						<div class="flex gap-2">
							<button on:click={enrollPerson} disabled={enrolling} class="px-3 py-1 rounded-lg text-sm font-medium" style="background: var(--teal); color: white">{enrolling ? '...' : 'Enroll'}</button>
							<button on:click={() => showEnroll = false} class="px-3 py-1 rounded-lg text-sm border" style="background: var(--surface); border-color: var(--border); color: var(--text)">Cancel</button>
						</div>
					</div>
				{/if}

				<div class="rounded-lg shadow border" style="background: var(--surface); border-color: var(--border)">
					{#if enrollments.length === 0}
						<div class="p-6 text-center text-sm" style="color: var(--text-secondary)">No enrollments yet</div>
					{:else}
						<div class="divide-y" style="border-color: var(--border)">
							{#each enrollments as enrollment}
								<div class="px-4 py-3">
									<div class="font-medium text-sm" style="color: var(--text)">{enrollment.person_name || enrollment.person_id}</div>
									<div class="flex items-center gap-2 mt-1">
										<span class="text-xs px-2 py-0.5 rounded-full" style="background: {enrollment.status === 'active' ? '#d1fae5' : enrollment.status === 'completed' ? '#dbeafe' : '#e2e8f0'}; color: {enrollment.status === 'active' ? '#065f46' : enrollment.status === 'completed' ? '#1e40af' : '#475569'}">
											{enrollment.status}
										</span>
										<span class="text-xs" style="color: var(--text-secondary)">Step {enrollment.current_step}</span>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>
