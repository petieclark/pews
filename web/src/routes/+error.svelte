<script>
	import { page } from '$app/stores';
	
	$: status = $page.status;
	$: message = $page.error?.message || 'An error occurred';
</script>

<svelte:head>
	<title>{status} - Pews</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-[var(--bg)] px-4">
	<div class="max-w-md w-full text-center">
		<!-- Pews Logo/Branding -->
		<div class="mb-8">
			<h1 class="text-6xl font-bold" style="color: var(--teal);">
				{status === 404 ? '🪑' : '⚠️'}
			</h1>
		</div>

		<!-- Error Message -->
		<div class="mb-8">
			<h2 class="text-4xl font-bold mb-4" style="color: var(--navy);">
				{#if status === 404}
					Page Not Found
				{:else if status === 500}
					Something Went Wrong
				{:else}
					Error {status}
				{/if}
			</h2>
			
			<p class="text-lg text-[var(--text-secondary)] mb-2">
				{#if status === 404}
					This pew doesn't exist. Maybe it's at a different church?
				{:else if status === 500}
					Our servers are having a moment. Please try again.
				{:else}
					{message}
				{/if}
			</p>
		</div>

		<!-- Actions -->
		<div class="flex flex-col sm:flex-row gap-4 justify-center">
			{#if status !== 404}
				<button 
					on:click={() => window.location.reload()}
					class="btn-primary"
				>
					Try Again
				</button>
			{/if}
			
			<a href="/dashboard" class="btn-secondary">
				Go to Dashboard
			</a>
		</div>

		<!-- Additional Info -->
		{#if status === 404}
			<div class="mt-8 text-sm text-[var(--text-secondary)]">
				<p>Lost? Check the URL or head back to the dashboard.</p>
			</div>
		{/if}
	</div>
</div>

<style>
	.btn-primary {
		background-color: var(--teal);
		color: white;
		padding: 0.75rem 1.5rem;
		border-radius: 0.5rem;
		font-weight: 600;
		transition: all 0.2s;
		border: none;
		cursor: pointer;
		display: inline-block;
		text-decoration: none;
	}

	.btn-primary:hover {
		opacity: 0.9;
		transform: translateY(-1px);
		box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
	}

	.btn-secondary {
		background-color: var(--surface);
		color: var(--navy);
		padding: 0.75rem 1.5rem;
		border-radius: 0.5rem;
		font-weight: 600;
		border: 2px solid var(--border);
		transition: all 0.2s;
		cursor: pointer;
		display: inline-block;
		text-decoration: none;
	}

	.btn-secondary:hover {
		background-color: var(--surface-hover);
		transform: translateY(-1px);
	}
</style>
