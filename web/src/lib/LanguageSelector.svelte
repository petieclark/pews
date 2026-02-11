<script>
    import { locale, loadTranslations, localeNames } from './i18n.js';
    import { onMount } from 'svelte';

    let supportedLocales = ['en', 'es', 'pt', 'ko'];
    let currentLocale = 'en';

    // Subscribe to locale changes
    const unsubscribe = locale.subscribe(value => {
        currentLocale = value;
    });

    async function changeLocale(event) {
        const newLocale = event.target.value;
        await loadTranslations(newLocale);
    }

    onMount(() => {
        return unsubscribe;
    });
</script>

<div class="language-selector">
    <label for="locale-select">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="2" y1="12" x2="22" y2="12"></line>
            <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path>
        </svg>
    </label>
    <select id="locale-select" bind:value={currentLocale} on:change={changeLocale}>
        {#each supportedLocales as loc}
            <option value={loc}>{localeNames[loc] || loc}</option>
        {/each}
    </select>
</div>

<style>
    .language-selector {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    label {
        display: flex;
        align-items: center;
        color: var(--text-secondary, #666);
    }

    select {
        padding: 0.5rem;
        border: 1px solid var(--border-color, #ddd);
        border-radius: 4px;
        background: var(--bg-primary, #fff);
        color: var(--text-primary, #333);
        font-size: 0.9rem;
        cursor: pointer;
    }

    select:focus {
        outline: none;
        border-color: var(--primary-color, #007bff);
    }

    /* Dark mode support */
    @media (prefers-color-scheme: dark) {
        select {
            background: var(--bg-secondary, #2a2a2a);
            color: var(--text-primary, #e0e0e0);
            border-color: var(--border-color, #444);
        }
    }
</style>
