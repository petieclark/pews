<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';
	import { Mail, ClipboardList } from 'lucide-svelte';

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

	// New step
	let showAddStep = false;
	let newStep = { step_type: 'send_email', delay_days: 3, delay_hours: 0, subject: '', body: '' };
	let addingStep = false;

	// Enroll
	let showEnroll = false;
	let enrollPersonId = '';
	let enrolling = false;

	const triggerOptions = [
		{ value: 'connection_card', label: 'Connection Card' },
		{ value: 'new_member', label: 'New Member' },
		{ value: 'first_visit', label: 'First Visit' },
		{ value: 'group_joined', label: 'Group Joined' },
		{ value: 'tag_added', label: 'Tag Added' },
		{ value: 'manual', label: 'Manual' },
		{ value: 'checkin_first_time', label: 'First Check-in' }
	];

	const mergeTags = [
		{ tag: '{{first_name}}', label: 'First Name' },
		{ tag: '{{last_name}}', label: 'Last Name' },
		{ tag: '{{email}}', label: 'Email' },
		{ tag: '{{church_name}}', label: 'Church Name' }
	];

	const stepIcons = { send_email: Mail, send_sms: null, wait: null, add_tag: null, add_to_group: ClipboardList };
	const stepLabels = { send_email: 'Send Email', send_sms: 'Send SMS', wait: 'Wait', add_tag: 'Add Tag', add_to_group: 'Add to Group' };

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
				body: JSON.stringify({ name: editName, description: editDescription, trigger_type: editTriggerType, trigger_value: editTriggerValue, is_active: journey.is_active })
			});
			editing = false;
			await loadJourney();
		} catch (err) {
			error = err.message;
		}
	}

	async function toggleActive() {
		try {
			await api(`/api/communication/journeys/${journeyId}/activate`, { method: 'PUT' });
			await loadJourney();
		} catch (err) {
			error = err.message;
		}
	}

	async function deleteJourney() {
		if (!confirm('Delete this journey and all its steps? This cannot be undone.')) return;
		try {
			await api(`/api/communication/journeys/${journeyId}`, { method: 'DELETE' });
			goto('/dashboard/communication/journeys');
		} catch (err) {
			error = err.message;
		}
	}

	function getStepConfig(step) {
		if (!step.config) return {};
		try { return typeof step.config === 'string' ? JSON.parse(step.config) : step.config; }
		catch { return {}; }
	}

	function computeDay(steps, index) {
		let total = 0;
		for (let i = 0; i <= index; i++) {
			total += steps[i].delay_days || 0;
		}
		return total;
	}

	async function addStep() {
		addingStep = true;
		try {
			const position = (journey.steps || []).length + 1;
			await api(`/api/communication/journeys/${journeyId}/steps`, {
				method: 'POST',
				body: JSON.stringify({
					position,
					step_type: newStep.step_type,
					delay_days: newStep.delay_days,
					delay_hours: newStep.delay_hours,
					config: { subject: newStep.subject, body: newStep.body }
				})
			});
			showAddStep = false;
			newStep = { step_type: 'send_email', delay_days: 3, delay_hours: 0, subject: '', body: '' };
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

	function formatDate(d) {
		if (!d) return '—';
		return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	function getNextStepLabel(enrollment) {
		if (enrollment.status === 'completed') return 'Completed';
		if (enrollment.status === 'paused') return 'Paused';
		if (!enrollment.next_step_at) return '—';
		return formatDate(enrollment.next_step_at);
	}
</script>

<div>
	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2" style="border-color: var(--teal)"></div>
		</div>
	{:else if error && !journey}
		<div class="px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">{error}</div>
	{:else if journey}
		<!-- Header -->
		<div class="flex items-center justify-between mb-6">
			<div>
				<a href="/dashboard/communication/journeys" class="text-sm font-medium" style="color: var(--teal)">← Journeys</a>
				{#if editing}
					<input bind:value={editName} class="text-3xl font-bold mt-1 bg-transparent border-b-2 outline-none w-full" style="color: var(--text-primary); border-color: var(--teal)" />
				{:else}
					<h1 class="text-3xl font-bold mt-1" style="color: var(--text-primary)">{journey.name}</h1>
				{/if}
			</div>
			<div class="flex gap-2">
				<button on:click={toggleActive} class="px-4 py-2 rounded-lg font-medium text-sm"
					style="background: {journey.is_active ? '#fee' : '#d1fae5'}; color: {journey.is_active ? '#c33' : '#065f46'}">
					{journey.is_active ? '⏸ Pause' : '▶ Activate'}
				</button>
				{#if editing}
					<button on:click={saveJourney} class="px-4 py-2 rounded-lg font-medium text-sm" style="background: var(--teal); color: white">Save</button>
					<button on:click={() => editing = false} class="px-4 py-2 rounded-lg font-medium text-sm border" style="background: var(--surface); border-color: var(--border); color: var(--text-primary)">Cancel</button>
				{:else}
					<button on:click={() => editing = true} class="px-4 py-2 rounded-lg font-medium text-sm border" style="background: var(--surface); border-color: var(--border); color: var(--text-primary)">Edit</button>
					<button on:click={deleteJourney} class="px-4 py-2 rounded-lg font-medium text-sm" style="background: #fee; color: #c33">Delete</button>
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
					<input bind:value={editDescription} type="text" class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Trigger</label>
						<select bind:value={editTriggerType} class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)">
							{#each triggerOptions as opt}
								<option value={opt.value}>{opt.label}</option>
							{/each}
						</select>
					</div>
					{#if editTriggerType === 'tag_added' || editTriggerType === 'group_joined'}
						<div>
							<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Trigger Value</label>
							<input bind:value={editTriggerValue} type="text" class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
						</div>
					{/if}
				</div>
			</div>
		{:else}
			<div class="mb-6 flex flex-wrap gap-3 items-center">
				<span class="px-3 py-1 rounded-full text-sm font-medium" style="background: {journey.is_active ? '#d1fae5' : '#e2e8f0'}; color: {journey.is_active ? '#065f46' : '#475569'}">
					{journey.is_active ? '● Active' : '○ Paused'}
				</span>
				<span class="px-3 py-1 rounded-full text-sm border" style="background: var(--bg); border-color: var(--border); color: var(--text-secondary)">
					{(triggerOptions.find(t => t.value === journey.trigger_type) || {}).label || journey.trigger_type}
					{#if journey.trigger_value} ({journey.trigger_value}){/if}
				</span>
				<span class="px-3 py-1 rounded-full text-sm border" style="background: var(--bg); border-color: var(--border); color: var(--text-secondary)">
					{(journey.steps || []).length} steps
				</span>
				<span class="px-3 py-1 rounded-full text-sm border" style="background: var(--bg); border-color: var(--border); color: var(--text-secondary)">
					{journey.enrollment_count || enrollments.length} enrolled
				</span>
				{#if journey.description}
					<span class="text-sm" style="color: var(--text-secondary)">— {journey.description}</span>
				{/if}
			</div>
		{/if}

		<!-- Visual Timeline -->
		{#if journey.steps && journey.steps.length > 0}
			<div class="mb-6 flex items-center gap-2 overflow-x-auto pb-2">
				{#each journey.steps as step, i}
					<div class="flex items-center gap-2">
						<div class="px-3 py-2 rounded-lg border text-center min-w-[80px]" style="background: var(--surface); border-color: var(--teal); border-width: 2px">
							<div class="text-xs font-medium" style="color: var(--text-secondary)">Day {computeDay(journey.steps, i)}</div>
							<div class="text-lg">{#if stepIcons[step.step_type]}<svelte:component this={stepIcons[step.step_type]} size={20} />{:else}•{/if}</div>
							<div class="text-xs" style="color: var(--text-secondary)">{stepLabels[step.step_type] || step.step_type}</div>
						</div>
						{#if i < journey.steps.length - 1}
							<div class="flex flex-col items-center">
								<div class="text-xs font-medium" style="color: var(--text-secondary)">{journey.steps[i + 1].delay_days}d</div>
								<div style="color: var(--text-secondary)">→</div>
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}

		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Steps -->
			<div class="lg:col-span-2">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold" style="color: var(--text-primary)">Steps</h2>
					<button on:click={() => showAddStep = !showAddStep} class="px-3 py-1 rounded-lg text-sm font-medium" style="background: var(--teal); color: white">+ Add Step</button>
				</div>

				{#if showAddStep}
					<div class="mb-4 rounded-lg shadow border p-4 space-y-3" style="background: var(--surface); border-color: var(--teal); border-width: 2px">
						<div class="text-sm font-semibold" style="color: var(--text-primary)">New Step</div>
						<div class="grid grid-cols-3 gap-3">
							<div>
								<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Channel</label>
								<select bind:value={newStep.step_type} class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)">
									<option value="send_email">Email</option>
									<option value="send_sms">SMS</option>
								</select>
							</div>
							<div>
								<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Delay (days)</label>
								<input bind:value={newStep.delay_days} type="number" min="0" class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
							</div>
							<div>
								<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Delay (hours)</label>
								<input bind:value={newStep.delay_hours} type="number" min="0" class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
							</div>
						</div>
						{#if newStep.step_type === 'send_email'}
							<div>
								<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Subject</label>
								<input bind:value={newStep.subject} type="text" placeholder="Email subject..." class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
							</div>
						{/if}
						<div>
							<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Body</label>
							<textarea bind:value={newStep.body} rows="3" placeholder="Message body..." class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"></textarea>
						</div>
						<div class="flex flex-wrap gap-1">
							<span class="text-xs" style="color: var(--text-secondary)">Merge tags:</span>
							{#each mergeTags as mt}
								<button on:click={() => newStep.body = (newStep.body || '') + mt.tag} class="px-2 py-0.5 rounded text-xs border" style="background: var(--bg); border-color: var(--border); color: var(--text-secondary)">{mt.label}</button>
							{/each}
						</div>
						<div class="flex gap-2">
							<button on:click={addStep} disabled={addingStep} class="px-3 py-1 rounded-lg text-sm font-medium" style="background: var(--teal); color: white">{addingStep ? 'Adding...' : 'Add Step'}</button>
							<button on:click={() => showAddStep = false} class="px-3 py-1 rounded-lg text-sm border" style="background: var(--surface); border-color: var(--border); color: var(--text-primary)">Cancel</button>
						</div>
					</div>
				{/if}

				<!-- Step List -->
				{#if !journey.steps || journey.steps.length === 0}
					<div class="rounded-lg border p-8 text-center" style="background: var(--surface); border-color: var(--border)">
						<p style="color: var(--text-secondary)">No steps yet. Add steps to build your journey.</p>
					</div>
				{:else}
					<div class="space-y-0">
						{#each journey.steps as step, i}
							{@const cfg = getStepConfig(step)}
							<div class="flex items-start gap-4">
								<div class="flex flex-col items-center">
									<div class="w-10 h-10 rounded-full flex items-center justify-center text-lg border-2" style="background: var(--surface); border-color: var(--teal)">
										{#if stepIcons[step.step_type]}<svelte:component this={stepIcons[step.step_type]} size={20} />{:else}•{/if}
									</div>
									{#if i < journey.steps.length - 1}
										<div class="w-0.5 h-16" style="background: var(--border)"></div>
									{/if}
								</div>
								<div class="flex-1 rounded-lg border p-4 mb-2" style="background: var(--surface); border-color: var(--border)">
									<div class="flex items-center justify-between mb-1">
										<div class="flex items-center gap-2">
											<span class="font-medium" style="color: var(--text-primary)">{stepLabels[step.step_type] || step.step_type}</span>
											<span class="text-xs px-2 py-0.5 rounded-full" style="background: var(--bg); color: var(--text-secondary)">
												Day {computeDay(journey.steps, i)}
												{#if step.delay_days > 0 || step.delay_hours > 0}
													(+{step.delay_days > 0 ? `${step.delay_days}d` : ''}{step.delay_hours > 0 ? ` ${step.delay_hours}h` : ''})
												{/if}
											</span>
										</div>
										<button on:click={() => deleteStep(step.id)} class="text-xs px-2 py-1 rounded hover:bg-red-50" style="color: #c33">✕</button>
									</div>
									{#if cfg.subject}
										<div class="text-sm font-medium mt-1" style="color: var(--text-primary)"><ClipboardList size={14} class="inline" /> {cfg.subject}</div>
									{/if}
									{#if cfg.body}
										<div class="text-sm mt-1 line-clamp-2" style="color: var(--text-secondary)">{@html cfg.body.replace(/<[^>]*>/g, ' ').substring(0, 150)}{cfg.body.length > 150 ? '...' : ''}</div>
									{/if}
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
					<h2 class="text-lg font-semibold" style="color: var(--text-primary)">Enrollments</h2>
					<button on:click={() => showEnroll = !showEnroll} class="px-3 py-1 rounded-lg text-sm font-medium" style="background: var(--teal); color: white">+ Enroll</button>
				</div>

				{#if showEnroll}
					<div class="mb-4 rounded-lg border p-4 space-y-3" style="background: var(--surface); border-color: var(--border)">
						<div>
							<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Person ID</label>
							<input bind:value={enrollPersonId} type="text" placeholder="Person UUID" class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
						</div>
						<div class="flex gap-2">
							<button on:click={enrollPerson} disabled={enrolling} class="px-3 py-1 rounded-lg text-sm font-medium" style="background: var(--teal); color: white">{enrolling ? '...' : 'Enroll'}</button>
							<button on:click={() => showEnroll = false} class="px-3 py-1 rounded-lg text-sm border" style="background: var(--surface); border-color: var(--border); color: var(--text-primary)">Cancel</button>
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
									<div class="font-medium text-sm" style="color: var(--text-primary)">{enrollment.person_name || enrollment.person_id}</div>
									<div class="flex items-center gap-2 mt-1 flex-wrap">
										<span class="text-xs px-2 py-0.5 rounded-full" style="background: {enrollment.status === 'active' ? '#d1fae5' : enrollment.status === 'completed' ? '#dbeafe' : '#e2e8f0'}; color: {enrollment.status === 'active' ? '#065f46' : enrollment.status === 'completed' ? '#1e40af' : '#475569'}">
											{enrollment.status}
										</span>
										<span class="text-xs" style="color: var(--text-secondary)">Step {enrollment.current_step}/{(journey.steps || []).length}</span>
										<span class="text-xs" style="color: var(--text-secondary)">Next: {getNextStepLabel(enrollment)}</span>
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
