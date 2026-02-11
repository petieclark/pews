<script>
	import { onMount } from 'svelte';
	
	export let show = false;
	export let title = '';
	export let onClose = () => {};
	
	let modalElement;
	let previousFocus;
	
	$: if (show) {
		// Save current focus
		previousFocus = document.activeElement;
		// Focus modal after render
		setTimeout(() => {
			if (modalElement) {
				const firstFocusable = modalElement.querySelector('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
				if (firstFocusable) {
					firstFocusable.focus();
				}
			}
		}, 10);
	} else if (previousFocus) {
		// Restore focus when modal closes
		previousFocus.focus();
	}
	
	function handleKeydown(event) {
		if (event.key === 'Escape') {
			onClose();
		}
		
		// Trap focus within modal
		if (event.key === 'Tab') {
			const focusableElements = modalElement.querySelectorAll(
				'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
			);
			const firstElement = focusableElements[0];
			const lastElement = focusableElements[focusableElements.length - 1];
			
			if (event.shiftKey && document.activeElement === firstElement) {
				event.preventDefault();
				lastElement.focus();
			} else if (!event.shiftKey && document.activeElement === lastElement) {
				event.preventDefault();
				firstElement.focus();
			}
		}
	}
	
	function handleBackdropClick(event) {
		if (event.target === event.currentTarget) {
			onClose();
		}
	}
</script>

{#if show}
	<div
		class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
		on:click={handleBackdropClick}
		on:keydown={handleKeydown}
		role="presentation"
	>
		<div
			bind:this={modalElement}
			role="dialog"
			aria-modal="true"
			aria-labelledby="modal-title"
			class="bg-surface rounded-lg max-w-md w-full p-6 border border-custom"
		>
			<h2 id="modal-title" class="text-2xl font-bold text-primary mb-4">{title}</h2>
			<slot />
		</div>
	</div>
{/if}
