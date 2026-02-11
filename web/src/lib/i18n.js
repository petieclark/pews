import { writable, derived } from 'svelte/store';

// Store for current locale
export const locale = writable('en');

// Store for translations
const translations = writable({});

// Load translations from API
export async function loadTranslations(lang) {
    try {
        const response = await fetch(`/api/i18n/${lang}`);
        if (!response.ok) {
            throw new Error(`Failed to load translations for ${lang}`);
        }
        const data = await response.json();
        translations.set(data);
        locale.set(lang);
        
        // Save preference to localStorage
        if (typeof window !== 'undefined') {
            localStorage.setItem('locale', lang);
            // Set HTML dir attribute for RTL support
            document.documentElement.setAttribute('lang', lang);
            // Set direction based on locale (for future RTL support)
            const rtlLocales = ['ar', 'he'];
            document.documentElement.setAttribute('dir', rtlLocales.includes(lang) ? 'rtl' : 'ltr');
        }
    } catch (error) {
        console.error('Error loading translations:', error);
        // Fallback to English
        if (lang !== 'en') {
            await loadTranslations('en');
        }
    }
}

// Initialize locale from localStorage or default to 'en'
export function initLocale() {
    if (typeof window !== 'undefined') {
        const savedLocale = localStorage.getItem('locale') || 'en';
        return loadTranslations(savedLocale);
    }
    return loadTranslations('en');
}

// Translation function
export const t = derived(
    translations,
    ($translations) => (key, fallback = key) => {
        return $translations[key] || fallback;
    }
);

// Helper to get supported locales
export async function getSupportedLocales() {
    try {
        const response = await fetch('/api/i18n/locales');
        if (!response.ok) {
            throw new Error('Failed to load supported locales');
        }
        const data = await response.json();
        return data.locales || ['en'];
    } catch (error) {
        console.error('Error loading supported locales:', error);
        return ['en', 'es', 'pt', 'ko'];
    }
}

// Locale metadata for display
export const localeNames = {
    'en': 'English',
    'es': 'Español',
    'pt': 'Português',
    'ko': '한국어'
};
