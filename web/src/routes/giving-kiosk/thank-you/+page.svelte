<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	let countdown = 10;
	let message = "Thank you for your generous gift!";

	onMount(async () => {
		// Try to load custom thank you message from config
		const urlParams = new URLSearchParams(window.location.search);
		const tenantId = urlParams.get('tenant') || '';
		
		if (tenantId) {
			try {
				const response = await fetch(`/api/giving/kiosk/config?tenant_id=${tenantId}`);
				if (response.ok) {
					const data = await response.json();
					if (data.config.thank_you_message) {
						message = data.config.thank_you_message;
					}
				}
			} catch (err) {
				console.error('Failed to load message:', err);
			}
		}

		// Start countdown
		const interval = setInterval(() => {
			countdown--;
			if (countdown <= 0) {
				clearInterval(interval);
				goto('/giving-kiosk' + (tenantId ? `?tenant=${tenantId}` : ''));
			}
		}, 1000);

		return () => clearInterval(interval);
	});
</script>

<div class="min-h-screen bg-gradient-to-br from-green-50 to-blue-50 flex items-center justify-center p-8">
	<div class="bg-white rounded-3xl shadow-2xl p-16 max-w-3xl text-center">
		<!-- Success Animation -->
		<div class="mb-8">
			<div class="inline-block">
				<div class="text-9xl animate-bounce">✓</div>
			</div>
		</div>

		<h1 class="text-5xl font-bold text-green-600 mb-6">
			Thank You!
		</h1>

		<p class="text-3xl text-gray-700 mb-12 leading-relaxed">
			{message}
		</p>

		<div class="mb-8">
			<div class="inline-block px-8 py-4 bg-gray-100 rounded-2xl">
				<p class="text-xl text-gray-600 mb-2">Returning to start in</p>
				<p class="text-6xl font-bold text-gray-800">{countdown}</p>
			</div>
		</div>

		<p class="text-lg text-gray-500">
			Your receipt will be emailed to you if you provided an email address.
		</p>
	</div>
</div>

<style>
	@keyframes bounce {
		0%, 100% {
			transform: translateY(0);
		}
		50% {
			transform: translateY(-20px);
		}
	}

	.animate-bounce {
		animation: bounce 2s infinite;
	}
</style>
