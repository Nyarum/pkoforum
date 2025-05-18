import { derived } from 'svelte/store';
import { language } from '../stores/language';

export type Translation = {
  [key: string]: string | Translation;
};

const translations: Record<string, Translation> = {
  en: {
    // Navigation & UI
    home: 'Home',
    switchLanguage: 'Switch Language',
    
    // Thread related
    threads: 'Threads',
    createThread: 'Create Thread',
    noThreads: 'No threads yet',
    selectThread: 'Select a thread to view',
    
    // Comments
    comments: 'Comments',
    noComments: 'No comments yet',
    writeComment: 'Write a comment...',
    postedAt: 'Posted at',
    addComment: 'Add Comment',
    
    // Categories
    categories: {
      general: 'General',
      help: 'Help',
      discussion: 'Discussion',
      announcement: 'Announcement'
    },

    // Modal
    cancel: 'Cancel',
    create: 'Create'
  },
  ru: {
    // Navigation & UI
    home: 'Главная',
    switchLanguage: 'Сменить язык',
    
    // Thread related
    threads: 'Обсуждения',
    createThread: 'Создать обсуждение',
    noThreads: 'Пока нет обсуждений',
    selectThread: 'Выберите обсуждение для просмотра',
    
    // Comments
    comments: 'Комментарии',
    noComments: 'Пока нет комментариев',
    writeComment: 'Напишите комментарий...',
    postedAt: 'Опубликовано',
    addComment: 'Добавить комментарий',
    
    // Categories
    categories: {
      general: 'Общее',
      help: 'Помощь',
      discussion: 'Обсуждение',
      announcement: 'Объявление'
    },

    // Modal
    cancel: 'Отмена',
    create: 'Создать'
  }
};

export const t = derived(language, ($language) => {
  return function translate(key: string): string {
    const keys = key.split('.');
    let value: any = translations[$language];
    
    for (const k of keys) {
      if (value === undefined) return key;
      value = value[k];
    }
    
    return typeof value === 'string' ? value : key;
  };
}); 