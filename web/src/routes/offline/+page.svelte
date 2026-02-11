<script>
  function retry() {
    window.location.reload();
  }
</script>

<div class="min-h-screen bg-[#F7FAFA] flex items-center justify-center px-4">
  <div class="max-w-md w-full text-center">
    <!-- Icon -->
    <div class="mb-8">
      <img 
        src="/icons/icon-192x192.png" 
        alt="Pews Logo" 
        class="w-24 h-24 mx-auto opacity-50"
      />
    </div>

    <!-- Message -->
    <h1 class="text-3xl font-bold text-[#1B3A4B] mb-4">
      You're Offline
    </h1>
    
    <p class="text-gray-600 mb-8">
      It looks like you've lost your internet connection. 
      Please check your network and try again.
    </p>

    <!-- Retry Button -->
    <button
      on:click={retry}
      class="bg-[#1B3A4B] text-white px-8 py-3 rounded-lg font-medium hover:bg-[#2d5568] transition-colors shadow-lg"
    >
      Retry Connection
    </button>

    <!-- Additional Info -->
    <div class="mt-12 text-sm text-gray-500">
      <p>Some features may be available offline.</p>
      <p class="mt-2">We'll sync your changes once you're back online.</p>
    </div>
  </div>
</div>

<style>
  /* Ensure full height even when offline */
  :global(body) {
    margin: 0;
    padding: 0;
  }
</style>
