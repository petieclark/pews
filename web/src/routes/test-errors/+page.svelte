<script>
	import { toast } from '$lib/stores/toast';
	import { api, clearToken } from '$lib/api';
	import { goto } from '$app/navigation';

	async function testErrorToast() {
		toast.show('This is a test error message', 'error');
	}

	async function testSuccessToast() {
		toast.show('Success! Everything worked', 'success');
	}

	async function testWarningToast() {
		toast.show('Warning: You should probably check this', 'warning');
	}

	async function testInfoToast() {
		toast.show('Just FYI, this is some info', 'info');
	}

	async function test401() {
		// Clear the token and try to make an API call
		clearToken();
		try {
			await api('/api/tenant');
		} catch (e) {
			// Should redirect to login
		}
	}

	async function testAPIError() {
		try {
			await api('/api/nonexistent-endpoint');
		} catch (e) {
			// Should show error toast
		}
	}
</script>

<svelte:head>
	<title>Error Testing - Pews</title>
</svelte:head>

<div class="min-h-screen bg-[var(--bg)] p-8">
	<div class="max-w-2xl mx-auto">
		<h1 class="text-3xl font-bold mb-8" style="color: var(--navy);">
			Error Handling Tests
		</h1>

		<div class="space-y-4">
			<div class="bg-surface p-6 rounded-lg border border-custom">
				<h2 class="text-xl font-semibold mb-4">Toast Notifications</h2>
				<div class="flex flex-wrap gap-3">
					<button on:click={testErrorToast} class="btn btn-error">
						Show Error Toast
					</button>
					<button on:click={testSuccessToast} class="btn btn-success">
						Show Success Toast
					</button>
					<button on:click={testWarningToast} class="btn btn-warning">
						Show Warning Toast
					</button>
					<button on:click={testInfoToast} class="btn btn-info">
						Show Info Toast
					</button>
				</div>
			</div>

			<div class="bg-surface p-6 rounded-lg border border-custom">
				<h2 class="text-xl font-semibold mb-4">API Error Handling</h2>
				<div class="flex flex-wrap gap-3">
					<button on:click={test401} class="btn btn-error">
						Test 401 (Should redirect to login)
					</button>
					<button on:click={testAPIError} class="btn btn-error">
						Test API Error (Should show toast)
					</button>
				</div>
			</div>

			<div class="bg-surface p-6 rounded-lg border border-custom">
				<h2 class="text-xl font-semibold mb-4">Error Pages</h2>
				<div class="flex flex-wrap gap-3">
					<a href="/nonexistent" class="btn btn-secondary">
						Visit 404 Page
					</a>
					<a href="/dashboard" class="btn btn-secondary">
						Back to Dashboard
					</a>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.btn {
		padding: 0.5rem 1rem;
		border-radius: 0.375rem;
		font-weight: 500;
		transition: all 0.2s;
		border: none;
		cursor: pointer;
		text-decoration: none;
		display: inline-block;
	}

	.btn-error {
		background-color: var(--error-bg);
		color: var(--error-text);
		border: 1px solid var(--error-border);
	}

	.btn-error:hover {
		opacity: 0.8;
	}

	.btn-success {
		background-color: #10B981;
		color: white;
	}

	.btn-success:hover {
		background-color: #059669;
	}

	.btn-warning {
		background-color: #F59E0B;
		color: white;
	}

	.btn-warning:hover {
		background-color: #D97706;
	}

	.btn-info {
		background-color: var(--teal);
		color: white;
	}

	.btn-info:hover {
		opacity: 0.9;
	}

	.btn-secondary {
		background-color: var(--surface);
		color: var(--navy);
		border: 2px solid var(--border);
	}

	.btn-secondary:hover {
		background-color: var(--surface-hover);
	}
</style>
