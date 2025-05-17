const translations = {
    en: {
        forum_title: "PKO Forum",
        threads: "Threads",
        create_thread: "Create New Thread",
        no_threads: "No threads yet. Be the first to create one!",
        select_thread: "Select a thread to view details and comments",
        thread_title: "Thread Title",
        thread_content: "Thread Content",
        cancel: "Cancel",
        create: "Create",
        comments: "Comments",
        no_comments: "No comments yet. Be the first to comment!",
        write_comment: "Write a comment...",
        add_comment: "Add Comment",
        created_at: "Created at",
        comments_count: "comments",
        upload_image: "Upload Image"
    },
    ru: {
        forum_title: "PKO Форум",
        threads: "Темы",
        create_thread: "Создать новую тему",
        no_threads: "Пока нет тем. Создайте первую!",
        select_thread: "Выберите тему для просмотра подробностей и комментариев",
        thread_title: "Заголовок темы",
        thread_content: "Содержание темы",
        cancel: "Отмена",
        create: "Создать",
        comments: "Комментарии",
        no_comments: "Пока нет комментариев. Оставьте первый!",
        write_comment: "Написать комментарий...",
        add_comment: "Добавить комментарий",
        created_at: "Создано",
        comments_count: "комментариев",
        upload_image: "Загрузить изображение"
    }
};

let currentLang = localStorage.getItem('pko_forum_lang') || 'en';

function setLanguage(lang) {
    currentLang = lang;
    localStorage.setItem('pko_forum_lang', lang);
    updatePageTranslations();
}

function t(key) {
    return translations[currentLang][key] || translations['en'][key] || key;
}

function getCommentContent(contentMap) {
    if (!contentMap) return '';
    return contentMap[currentLang] || contentMap['en'] || Object.values(contentMap)[0] || '';
}

function updatePageTranslations() {
    // Update all elements with data-i18n attribute
    document.querySelectorAll('[data-i18n]').forEach(element => {
        const key = element.getAttribute('data-i18n');
        if (key) {
            if (element.tagName === 'INPUT' || element.tagName === 'TEXTAREA') {
                element.placeholder = t(key);
            } else {
                element.textContent = t(key);
            }
        }
    });

    // Update document title
    document.title = t('forum_title');

    // Update comment contents
    document.querySelectorAll('[data-comment-content]').forEach(element => {
        const contentMap = JSON.parse(element.getAttribute('data-comment-content'));
        element.textContent = getCommentContent(contentMap);
    });
}

// Initialize translations when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    updatePageTranslations();
}); 