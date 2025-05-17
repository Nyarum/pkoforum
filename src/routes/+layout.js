import { language } from '$lib/stores/language';
import { browser } from '$app/environment';

export const ssr = true;
export const prerender = false;

// Get the initial language from localStorage or default to 'en'
const getInitialLang = () => {
    if (browser) {
        return localStorage.getItem('language') || 'en';
    }
    return 'en';
};

/** @type {import('@sveltejs/kit').Load} */
export async function load() {
    return {
        lang: getInitialLang()
    };
} 