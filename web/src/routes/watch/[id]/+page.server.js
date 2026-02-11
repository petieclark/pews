import { error } from '@sveltejs/kit';

const API_URL = process.env.VITE_API_URL || 'http://localhost:8190';

/** @type {import('./$types').PageServerLoad} */
export async function load({ params, fetch }) {
	try {
		const response = await fetch(`${API_URL}/api/streaming/watch/${params.id}`);
		
		if (!response.ok) {
			throw error(404, 'Stream not found');
		}
		
		const stream = await response.json();
		
		// Generate SEO metadata
		const seoData = {
			title: `${stream.title} - Live Stream | Pews`,
			description: stream.description || `Watch ${stream.title} live on Pews church streaming platform.`,
			image: stream.thumbnail || '/og-image.png',
			type: 'video.other',
			
			// Structured data for search engines
			structuredData: {
				'@context': 'https://schema.org',
				'@type': 'VideoObject',
				name: stream.title,
				description: stream.description || stream.title,
				thumbnailUrl: stream.thumbnail || 'https://pews.church/og-image.png',
				uploadDate: stream.start_time || new Date().toISOString(),
				duration: stream.duration || 'PT0M',
				embedUrl: stream.embed_url,
				isLiveBroadcast: stream.status === 'live',
				publication: {
					'@type': 'BroadcastEvent',
					isLiveBroadcast: stream.status === 'live',
					startDate: stream.start_time,
					endDate: stream.end_time
				}
			}
		};
		
		return {
			stream,
			seo: seoData
		};
	} catch (err) {
		throw error(500, 'Failed to load stream');
	}
}
