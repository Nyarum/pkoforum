import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';

// Create the language store
export const language = writable('en');

// Initialize language from localStorage if in browser
if (browser) {
    const storedLang = localStorage.getItem('language');
    if (storedLang) {
        language.set(storedLang);
    }
}

// Update language in localStorage when it changes
language.subscribe(value => {
    if (browser) {
        localStorage.setItem('language', value);
    }
});

// Translations
export const translations = {
    en: {
        switchLanguage: 'Switch to Russian',
        home: 'Home',
        threads: 'Threads',
        loading: 'Loading threads...',
        error: 'Error loading threads',
        noThreads: 'No threads found',
        createThread: 'Create New Thread',
        postedAt: 'Posted at',
        comments: 'Comments',
        writeComment: 'Write a comment',
        addImage: 'Add Image',
        submit: 'Submit',
        submitting: 'Submitting...',
        noComments: 'No comments yet',
        selectThread: 'Select a thread'
    },
    ru: {
        switchLanguage: 'Переключить на английский',
        home: 'Главная',
        threads: 'Обсуждения',
        loading: 'Загрузка обсуждений...',
        error: 'Ошибка загрузки обсуждений',
        noThreads: 'Обсуждения не найдены',
        createThread: 'Создать новое обсуждение',
        postedAt: 'Опубликовано',
        comments: 'Комментарии',
        writeComment: 'Написать комментарий',
        addImage: 'Добавить изображение',
        submit: 'Отправить',
        submitting: 'Отправка...',
        noComments: 'Пока нет комментариев',
        selectThread: 'Выберите обсуждение'
    }
};

// Create a derived store for the translation function
export const t = derived(
    language,
    $language => key => translations[$language][key] || key
); 