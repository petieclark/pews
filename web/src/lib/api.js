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

	const response = await fetch(`${API_URL}${endpoint}`, {
		...options,
		headers
	});

	if (!response.ok) {
		const error = await response.text();
		throw new Error(error || 'Request failed');
	}

	return response.json();
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
