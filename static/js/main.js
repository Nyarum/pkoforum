// Show/hide new thread form
function showNewThreadForm() {
    document.getElementById('new-thread-modal').classList.remove('hidden');
}

function hideNewThreadForm() {
    document.getElementById('new-thread-modal').classList.add('hidden');
}

// Create new thread
async function createThread(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);

    try {
        const response = await fetch('/api/threads', {
            method: 'POST',
            body: formData
        });

        if (!response.ok) throw new Error('Failed to create thread');

        // Update threads list
        const threadsListHtml = await response.text();
        document.getElementById('threads-list').innerHTML = threadsListHtml;

        // Reset and hide form
        form.reset();
        hideNewThreadForm();
    } catch (error) {
        console.error('Error creating thread:', error);
        alert('Failed to create thread. Please try again.');
    }
}

// Load thread details
async function loadThread(threadId) {
    try {
        const response = await fetch(`/api/threads/${threadId}`);
        if (!response.ok) throw new Error('Failed to load thread');

        const threadHtml = await response.text();
        document.getElementById('thread-placeholder').classList.add('hidden');
        const selectedThreadElement = document.getElementById('selected-thread');
        selectedThreadElement.innerHTML = threadHtml;
        selectedThreadElement.classList.remove('hidden');

        // Get all comment elements
        const commentElements = selectedThreadElement.querySelectorAll('[data-comment-content]');
        commentElements.forEach(element => {
            const contentMap = JSON.parse(element.getAttribute('data-comment-content'));
            element.textContent = getCommentContent(contentMap);
        });

        // Update translations for other elements
        updatePageTranslations();

        // Update selected state in thread list
        document.querySelectorAll('.thread-item').forEach(item => {
            if (item.dataset.threadId === threadId) {
                item.classList.add('border-2', 'border-blue-500');
            } else {
                item.classList.remove('border-2', 'border-blue-500');
            }
        });
    } catch (error) {
        console.error('Error loading thread:', error);
        alert('Failed to load thread. Please try again.');
    }
}

// Add comment
async function addComment(event, threadId) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);

    try {
        const response = await fetch(`/api/threads/${threadId}/comments`, {
            method: 'POST',
            body: formData
        });

        if (!response.ok) throw new Error('Failed to add comment');

        // Update thread view with new comment
        const threadHtml = await response.text();
        document.getElementById('selected-thread').innerHTML = threadHtml;
        
        // Update translations for the newly added comment
        updatePageTranslations();

        // Reset form
        form.reset();
    } catch (error) {
        console.error('Error adding comment:', error);
        alert('Failed to add comment. Please try again.');
    }
} 