<script>
	import { api } from '$lib/api';
	import { Mail } from 'lucide-svelte';
	import { goto } from '$app/navigation';

	let name = '';
	let description = '';
	let triggerType = 'connection_card';
	let triggerValue = '';
	let isActive = false;
	let steps = [];
	let saving = false;
	let error = '';

	const triggerOptions = [
		{ value: 'connection_card', label: 'Connection Card Submitted' },
		{ value: 'new_member', label: 'New Member' },
		{ value: 'first_visit', label: 'First Visit' },
		{ value: 'group_joined', label: 'Group Joined' },
		{ value: 'tag_added', label: 'Tag Added' },
		{ value: 'manual', label: 'Manual' }
	];

	const mergeTags = [
		{ tag: '{{first_name}}', label: 'First Name' },
		{ tag: '{{last_name}}', label: 'Last Name' },
		{ tag: '{{email}}', label: 'Email' },
		{ tag: '{{church_name}}', label: 'Church Name' }
	];

	function addStep() {
		steps = [...steps, {
			position: steps.length + 1,
			step_type: 'send_email',
			delay_days: steps.length === 0 ? 0 : 3,
			delay_hours: 0,
			subject: '',
			body: '',
			template_id: null
		}];
	}

	function removeStep(index) {
		steps = steps.filter((_, i) => i !== index).map((s, i) => ({ ...s, position: i + 1 }));
	}

	function moveStep(index, direction) {
		if ((direction === -1 && index === 0) || (direction === 1 && index === steps.length - 1)) return;
		const newSteps = [...steps];
		const temp = newSteps[index];
		newSteps[index] = newSteps[index + direction];
		newSteps[index + direction] = temp;
		steps = newSteps.map((s, i) => ({ ...s, position: i + 1 }));
	}

	function insertMergeTag(stepIndex, field, tag) {
		const step = steps[stepIndex];
		step[field] = (step[field] || '') + tag;
		steps = [...steps];
	}

	function computeDay(index) {
		let total = 0;
		for (let i = 0; i <= index; i++) {
			total += steps[i].delay_days || 0;
		}
		return total;
	}

	async function save() {
		if (!name.trim()) { error = 'Name is required'; return; }
		saving = true;
		error = '';
		try {
			const journey = await api('/api/communication/journeys', {
				method: 'POST',
				body: JSON.stringify({ name, description, trigger_type: triggerType, trigger_value: triggerValue, is_active: isActive })
			});

			// Add steps sequentially
			for (const step of steps) {
				await api(`/api/communication/journeys/${journey.id}/steps`, {
					method: 'POST',
					body: JSON.stringify({
						position: step.position,
						step_type: step.step_type,
						delay_days: step.delay_days,
						delay_hours: step.delay_hours,
						config: { subject: step.subject, body: step.body }
					})
				});
			}

			goto(`/dashboard/communication/journeys/${journey.id}`);
		} catch (err) {
			error = err.message;
		} finally {
			saving = false;
		}
	}
</script>

<div>
	<a href="/dashboard/communication/journeys" class="text-sm font-medium" style="color: var(--teal)">← Back to Journeys</a>
	<h1 class="text-3xl font-bold mt-2 mb-6" style="color: var(--text-primary)">Create Journey</h1>

	{#if error}
		<div class="mb-4 px-4 py-3 rounded-lg border" style="background: #fee; border-color: #fcc; color: #c33">{error}</div>
	{/if}

	<!-- Journey Settings -->
	<div class="rounded-lg shadow border p-6 mb-6 space-y-4" style="background: var(--surface); border-color: var(--border)">
		<div>
			<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Journey Name</label>
			<input bind:value={name} type="text" placeholder="e.g., Welcome Visitor" class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
		</div>
		<div>
			<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Description</label>
			<input bind:value={description} type="text" placeholder="Brief description..." class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
		</div>
		<div class="grid grid-cols-2 gap-4">
			<div>
				<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Trigger</label>
				<select bind:value={triggerType} class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)">
					{#each triggerOptions as opt}
						<option value={opt.value}>{opt.label}</option>
					{/each}
				</select>
			</div>
			{#if triggerType === 'tag_added' || triggerType === 'group_joined'}
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Trigger Value</label>
					<input bind:value={triggerValue} type="text" placeholder="Tag name or Group ID" class="w-full px-3 py-2 rounded-lg border" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
				</div>
			{/if}
		</div>
		<label class="flex items-center gap-2 cursor-pointer">
			<input bind:checked={isActive} type="checkbox" class="rounded" />
			<span class="text-sm" style="color: var(--text-primary)">Start active immediately</span>
		</label>
	</div>

	<!-- Steps Builder -->
	<div class="flex items-center justify-between mb-4">
		<h2 class="text-lg font-semibold" style="color: var(--text-primary)">Journey Steps</h2>
		<button on:click={addStep} class="px-4 py-2 rounded-lg font-medium text-sm" style="background: var(--teal); color: white">+ Add Step</button>
	</div>

	{#if steps.length === 0}
		<div class="rounded-lg border p-8 text-center mb-6" style="background: var(--surface); border-color: var(--border)">
			<p class="text-lg mb-2" style="color: var(--text-secondary)">No steps yet</p>
			<p class="text-sm mb-4" style="color: var(--text-secondary)">Add steps to build your automated journey</p>
			<button on:click={addStep} class="px-4 py-2 rounded-lg font-medium" style="background: var(--teal); color: white">Add First Step</button>
		</div>
	{:else}
		<!-- Visual Timeline -->
		<div class="mb-6 flex items-center gap-2 overflow-x-auto pb-2">
			{#each steps as step, i}
				<div class="flex items-center gap-2">
					<div class="px-3 py-2 rounded-lg border text-center min-w-[80px]" style="background: var(--surface); border-color: var(--teal); border-width: 2px">
						<div class="text-xs font-medium" style="color: var(--text-secondary)">Day {computeDay(i)}</div>
						<div class="text-sm font-bold" style="color: var(--text-primary)">{#if step.step_type === 'send_email'}<Mail size={16} />{:else if step.step_type === 'send_sms'}💬{:else}⏳{/if}</div>
					</div>
					{#if i < steps.length - 1}
						<div class="text-xs font-medium px-1" style="color: var(--text-secondary)">→</div>
					{/if}
				</div>
			{/each}
		</div>

		<!-- Step Cards -->
		<div class="space-y-4 mb-6">
			{#each steps as step, i}
				<div class="rounded-lg shadow border p-5" style="background: var(--surface); border-color: var(--border)">
					<div class="flex items-center justify-between mb-3">
						<div class="flex items-center gap-2">
							<span class="text-sm font-bold px-2 py-1 rounded" style="background: var(--teal); color: white">Step {i + 1}</span>
							<span class="text-sm" style="color: var(--text-secondary)">Day {computeDay(i)}</span>
						</div>
						<div class="flex items-center gap-1">
							<button on:click={() => moveStep(i, -1)} disabled={i === 0} class="px-2 py-1 rounded text-sm" style="color: var(--text-secondary)">↑</button>
							<button on:click={() => moveStep(i, 1)} disabled={i === steps.length - 1} class="px-2 py-1 rounded text-sm" style="color: var(--text-secondary)">↓</button>
							<button on:click={() => removeStep(i)} class="px-2 py-1 rounded text-sm" style="color: #c33">✕</button>
						</div>
					</div>

					<div class="grid grid-cols-3 gap-3 mb-3">
						<div>
							<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Channel</label>
							<select bind:value={step.step_type} class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)">
								<option value="send_email">Email</option>
								<option value="send_sms">SMS</option>
							</select>
						</div>
						<div>
							<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Delay (days)</label>
							<input bind:value={step.delay_days} type="number" min="0" class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
						</div>
						<div>
							<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Delay (hours)</label>
							<input bind:value={step.delay_hours} type="number" min="0" class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
						</div>
					</div>

					{#if step.step_type === 'send_email'}
						<div class="mb-3">
							<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Subject</label>
							<input bind:value={step.subject} type="text" placeholder="Email subject..." class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)" />
						</div>
					{/if}

					<div class="mb-2">
						<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Body</label>
						<textarea bind:value={step.body} rows="4" placeholder="Message body..." class="w-full px-3 py-2 rounded-lg border text-sm" style="background: var(--bg); border-color: var(--border); color: var(--text-primary)"></textarea>
					</div>

					<div class="flex flex-wrap gap-1">
						<span class="text-xs" style="color: var(--text-secondary)">Merge tags:</span>
						{#each mergeTags as mt}
							<button
								on:click={() => insertMergeTag(i, step.step_type === 'send_email' ? 'body' : 'body', mt.tag)}
								class="px-2 py-0.5 rounded text-xs border hover:border-teal-400"
								style="background: var(--bg); border-color: var(--border); color: var(--text-secondary)"
							>{mt.label}</button>
						{/each}
					</div>
				</div>
			{/each}
		</div>
	{/if}

	<!-- Save -->
	<div class="flex gap-3">
		<button on:click={save} disabled={saving} class="px-6 py-2 rounded-lg font-medium" style="background: var(--teal); color: white">
			{saving ? 'Creating...' : 'Create Journey'}
		</button>
		<a href="/dashboard/communication/journeys" class="px-6 py-2 rounded-lg font-medium border inline-flex items-center" style="background: var(--surface); border-color: var(--border); color: var(--text-primary)">
			Cancel
		</a>
	</div>
</div>
