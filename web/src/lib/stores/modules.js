import { writable } from 'svelte/store';
import { api } from '$lib/api';

export const enabledModules = writable([]);

export async function refreshModules() {
	try {
		const modules = await api('/api/tenant/modules').catch(() => []);
		const enabled = (modules || []).filter(m => m.enabled).map(m => m.name);
		enabledModules.set(enabled);
		return enabled;
	} catch {
		return [];
	}
}
