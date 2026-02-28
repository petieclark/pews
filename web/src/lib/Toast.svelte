<script>
	import { toast } from './stores/toast';
	import { fade, fly } from 'svelte/transition';
	import { AlertTriangle, CheckCircle, Zap, Info } from 'lucide-svelte';

	$: toasts = $toast;
</script>

<div class="fixed top-4 right-4 z-50 flex flex-col gap-2 max-w-md">
	{#each toasts as { id, message, type } (id)}
		<div 
			class="toast toast-{type}"
			transition:fly="{{ y: -20, duration: 300 }}"
			role="alert"
		>
			<div class="flex items-start gap-3">
				{#if type === 'error'}
					<AlertTriangle size={20} />
				{:else if type === 'success'}
					<CheckCircle size={20} />
				{:else if type === 'warning'}
					<Zap size={20} />
				{:else}
					<Info size={20} />
				{/if}
				<div class="flex-1">
					<p class="text-sm font-medium">{message}</p>
				</div>
				<button 
					on:click={() => toast.dismiss(id)}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Dismiss"
				>
					×
				</button>
			</div>
		</div>
	{/each}
</div>

<style>
	.toast {
		background-color: var(--surface);
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		padding: 1rem;
		box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
		cursor: pointer;
		transition: all 0.2s;
	}

	.toast:hover {
		box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
	}

	.toast-error {
		background-color: var(--error-bg);
		border-color: var(--error-border);
	}

	.toast-success {
		background-color: #F0FDF4;
		border-color: #86EFAC;
	}

	:root.dark .toast-success {
		background-color: #14532D;
		border-color: #166534;
	}

	.toast-warning {
		background-color: #FFFBEB;
		border-color: #FDE68A;
	}

	:root.dark .toast-warning {
		background-color: #451A03;
		border-color: #78350F;
	}
</style>
