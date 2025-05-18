import { writable } from 'svelte/store';

type Language = 'en' | 'ru';

export const language = writable<Language>('en'); 