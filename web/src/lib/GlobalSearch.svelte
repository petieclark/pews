<script>
  import { goto } from '$app/navigation';
  
  let query = '';
  let showResults = false;
  
  async function handleSearch() {
    if (!query.trim()) return;
    // TODO: Implement full search
    goto(`/dashboard?search=${encodeURIComponent(query)}`);
    showResults = false;
  }
</script>

<div class="relative">
  <form on:submit|preventDefault={handleSearch}>
    <input
      type="text"
      bind:value={query}
      placeholder="Search..."
      class="w-64 px-3 py-1.5 text-sm border rounded-lg bg-[var(--bg-primary)] border-[var(--border-color)] text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]"
    />
  </form>
</div>
