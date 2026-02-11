import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch }) => {
	const streamId = params.id;
	const API_URL = process.env.VITE_API_URL || 'http://localhost:8190';
	
	try {
		const response = await fetch(`${API_URL}/api/streaming/watch/${streamId}`);
		if (!response.ok) {
			return { stream: null };
		}
		const stream = await response.json();
		
		return {
			stream,
			meta: {
				title: stream.title || 'Watch Live',
				description: stream.description || 'Join us for our live service',
				image: stream.thumbnail_url || '/og-default.jpg',
				url: `https://yourchurch.com/watch/${streamId}`
			}
		};
	} catch (error) {
		return { stream: null };
	}
};
