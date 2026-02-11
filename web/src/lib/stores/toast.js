import { writable } from 'svelte/store';

function createToastStore() {
	const { subscribe, update } = writable([]);
	let idCounter = 0;

	return {
		subscribe,
		show: (message, type = 'info', duration = 4000) => {
			const id = idCounter++;
			const toast = { id, message, type };
			
			update(toasts => [...toasts, toast]);
			
			if (duration > 0) {
				setTimeout(() => {
					update(toasts => toasts.filter(t => t.id !== id));
				}, duration);
			}
			
			return id;
		},
		dismiss: (id) => {
			update(toasts => toasts.filter(t => t.id !== id));
		},
		clear: () => {
			update(() => []);
		}
	};
}

export const toast = createToastStore();
