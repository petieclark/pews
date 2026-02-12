<script>
	import { page } from '$app/stores';

	let slug = $page.params.slug;
</script>

<svelte:head>
	<title>Thank You for Your Gift</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 flex flex-col items-center justify-center px-4">
	<div class="text-center max-w-md">
		<div class="text-7xl mb-6">🎉</div>
		<h1 class="text-3xl font-bold text-gray-900 mb-3">Thank You!</h1>
		<p class="text-gray-600 text-lg mb-2">
			Your generous gift has been received.
		</p>
		<p class="text-gray-500 mb-8">
			If you provided an email, you'll receive a receipt from Stripe shortly.
		</p>

		<a
			href="/give/{slug}"
			class="inline-block px-8 py-3 bg-teal-600 text-white font-semibold rounded-xl hover:bg-teal-700 transition"
		>
			Give Again
		</a>
	</div>
</div>
