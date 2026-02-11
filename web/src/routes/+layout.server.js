/** @type {import('./$types').LayoutServerLoad} */
export async function load() {
	// Organization schema for church (used site-wide)
	const organizationSchema = {
		'@context': 'https://schema.org',
		'@type': 'Organization',
		name: 'Pews',
		description: 'Modern church management platform with live streaming, giving, and engagement tools',
		url: 'https://pews.church',
		logo: 'https://pews.church/logo.png',
		sameAs: [
			// Add social media URLs when available
		],
		contactPoint: {
			'@type': 'ContactPoint',
			contactType: 'Customer Support',
			email: 'support@pews.church'
		}
	};

	return {
		organizationSchema
	};
}
