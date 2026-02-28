<script>
	// Generic /give page - redirects to appropriate slug-based giving page
	// The actual giving pages are at /give/[slug]
	import { Building2 } from 'lucide-svelte';
</script>

<div class="min-h-screen flex items-center justify-center p-4">
	<div class="text-center max-w-md">
		<div class="mb-4 flex justify-center"><Building2 size={64} /></div>
		<h1 class="text-2xl font-bold text-gray-800 mb-2">Online Giving</h1>
		<p class="text-gray-500 mb-6">
			To make a donation, please visit your church's specific giving page.
		</p>
		<p class="text-sm text-gray-400">
			Powered by <span class="font-bold text-teal-600">Pews</span>
		</p>
	</div>
</div>
