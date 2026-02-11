import { toast } from './stores/toast';
import { goto } from '$app/navigation';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8190';

export async function api(endpoint, options = {}) {
	const token = localStorage.getItem('token');
	const headers = {
		'Content-Type': 'application/json',
		...options.headers
	};

	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	try {
		const response = await fetch(`${API_URL}${endpoint}`, {
			...options,
			headers
		});

		// Handle 401 - JWT expired or invalid
		if (response.status === 401) {
			clearToken();
			toast.show('Your session has expired. Please login again.', 'warning');
			goto('/login');
			throw new Error('Unauthorized - redirecting to login');
		}

		if (!response.ok) {
			const error = await response.text();
			const errorMessage = error || 'Request failed';
			
			// Show toast for all API errors unless explicitly silenced
			if (!options.silent) {
				toast.show(errorMessage, 'error');
			}
			
			throw new Error(errorMessage);
		}

		return response.json();
	} catch (error) {
		// Network errors or other fetch failures
		if (!error.message.includes('Unauthorized')) {
			if (!options.silent) {
				toast.show(error.message || 'Network error. Please check your connection.', 'error');
			}
		}
		throw error;
	}
}

export function setToken(token) {
	localStorage.setItem('token', token);
}

export function clearToken() {
	localStorage.removeItem('token');
}

export function getToken() {
	return localStorage.getItem('token');
}
