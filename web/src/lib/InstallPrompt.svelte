<script>
  import { onMount } from 'svelte';
  
  let deferredPrompt = null;
  let showPrompt = false;
  let isInstalled = false;
  
  onMount(() => {
    // Check if already installed
    if (window.matchMedia('(display-mode: standalone)').matches) {
      isInstalled = true;
      return;
    }
    
    // Check if we've already shown the prompt this session
    if (sessionStorage.getItem('pwa-prompt-shown')) {
      return;
    }
    
    // Listen for beforeinstallprompt event
    window.addEventListener('beforeinstallprompt', (e) => {
      // Prevent the default prompt
      e.preventDefault();
      
      // Store the event for later use
      deferredPrompt = e;
      
      // Show our custom prompt
      showPrompt = true;
      
      // Mark that we've shown the prompt this session
      sessionStorage.setItem('pwa-prompt-shown', 'true');
    });
    
    // Listen for successful install
    window.addEventListener('appinstalled', () => {
      isInstalled = true;
      showPrompt = false;
      deferredPrompt = null;
    });
  });
  
  async function handleInstall() {
    if (!deferredPrompt) return;
    
    // Show the install prompt
    deferredPrompt.prompt();
    
    // Wait for the user's response
    const { outcome } = await deferredPrompt.userChoice;
    
    if (outcome === 'accepted') {
      console.log('User accepted the install prompt');
    } else {
      console.log('User dismissed the install prompt');
    }
    
    // Clear the deferred prompt
    deferredPrompt = null;
    showPrompt = false;
  }
  
  function handleDismiss() {
    showPrompt = false;
    deferredPrompt = null;
  }
</script>

{#if showPrompt && !isInstalled}
  <div class="fixed bottom-0 left-0 right-0 bg-[#1B3A4B] text-white p-4 shadow-2xl z-50 animate-slide-up">
    <div class="max-w-4xl mx-auto flex items-center justify-between gap-4">
      <!-- Icon and Message -->
      <div class="flex items-center gap-3 flex-1">
        <img 
          src="/icons/icon-72x72.png" 
          alt="Pews" 
          class="w-12 h-12 rounded-lg"
        />
        <div>
          <h3 class="font-semibold text-lg">Install Pews</h3>
          <p class="text-sm text-gray-300">
            Add to your home screen for quick access
          </p>
        </div>
      </div>
      
      <!-- Action Buttons -->
      <div class="flex gap-2">
        <button
          on:click={handleDismiss}
          class="px-4 py-2 text-sm font-medium text-white hover:bg-white/10 rounded-lg transition-colors"
        >
          Not Now
        </button>
        <button
          on:click={handleInstall}
          class="px-4 py-2 text-sm font-medium bg-white text-[#1B3A4B] rounded-lg hover:bg-gray-100 transition-colors shadow-md"
        >
          Install
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  @keyframes slide-up {
    from {
      transform: translateY(100%);
    }
    to {
      transform: translateY(0);
    }
  }
  
  .animate-slide-up {
    animation: slide-up 0.3s ease-out;
  }
  
  /* Mobile responsive */
  @media (max-width: 640px) {
    .fixed {
      padding: 1rem;
    }
    
    .flex {
      flex-direction: column;
      align-items: stretch;
    }
    
    .flex.gap-2 {
      flex-direction: row;
    }
  }
</style>
