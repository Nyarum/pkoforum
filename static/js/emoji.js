// Handle emoji picker visibility
function toggleEmojiPicker(pickerId) {
    const picker = document.getElementById(pickerId);
    const allPickers = document.querySelectorAll('[id$="-emoji-picker"]');
    
    // Hide all other pickers
    allPickers.forEach(p => {
        if (p.id !== pickerId) {
            p.classList.add('hidden');
        }
    });

    // Toggle current picker
    picker.classList.toggle('hidden');
}

// Insert emoji into the textarea
function insertEmoji(emoji, type) {
    let textarea;
    if (type === 'thread') {
        textarea = document.querySelector('#new-thread-form textarea[name="content"]');
    } else {
        textarea = document.querySelector('#selected-thread textarea[name="content"]');
    }

    if (textarea) {
        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        const text = textarea.value;
        const before = text.substring(0, start);
        const after = text.substring(end);

        textarea.value = before + emoji + after;
        textarea.selectionStart = textarea.selectionEnd = start + emoji.length;
        textarea.focus();
    }

    // Hide the picker after selection
    const picker = document.getElementById(type + '-emoji-picker');
    if (picker) {
        picker.classList.add('hidden');
    }
}

// Close emoji pickers when clicking outside
document.addEventListener('click', function(event) {
    const isEmojiButton = event.target.closest('button[onclick*="toggleEmojiPicker"]');
    const isEmojiPicker = event.target.closest('[id$="-emoji-picker"]');
    
    if (!isEmojiButton && !isEmojiPicker) {
        const allPickers = document.querySelectorAll('[id$="-emoji-picker"]');
        allPickers.forEach(picker => picker.classList.add('hidden'));
    }
}); 