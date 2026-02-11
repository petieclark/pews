<script>
	export let value = '';
	export let required = false;
	export let placeholder = 'Select a key';

	const keys = [
		'C', 'C#', 'Db', 'D', 'D#', 'Eb', 'E', 'F', 'F#', 'Gb', 'G', 'G#', 'Ab', 'A', 'A#', 'Bb', 'B'
	];
	const modes = ['Major', 'Minor'];
	
	let selectedKey = '';
	let selectedMode = 'Major';
	
	// Parse existing value on mount
	$: {
		if (value) {
			const match = value.match(/^([A-G][#b]?)\s*(Major|Minor)?$/i);
			if (match) {
				selectedKey = match[1];
				selectedMode = match[2] || 'Major';
			}
		}
	}
	
	// Update value when selections change
	$: {
		if (selectedKey) {
			value = selectedMode === 'Major' ? selectedKey : `${selectedKey} ${selectedMode}`;
		} else {
			value = '';
		}
	}
</script>

<div class="flex gap-2">
	<select
		bind:value={selectedKey}
		{required}
		class="flex-1 px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
	>
		<option value="">{placeholder}</option>
		{#each keys as key}
			<option value={key}>{key}</option>
		{/each}
	</select>
	<select
		bind:value={selectedMode}
		class="px-3 py-2 border input-border bg-[var(--input-bg)] text-primary rounded-md focus:outline-none focus:ring-2 focus:ring-teal"
	>
		{#each modes as mode}
			<option value={mode}>{mode}</option>
		{/each}
	</select>
</div>

<style>
	:global(.ring-teal) {
		--tw-ring-color: #4a8b8c;
	}
</style>
