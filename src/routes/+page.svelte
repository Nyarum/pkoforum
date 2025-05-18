<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { language } from '$lib/stores/language';
    import { t } from '$lib/i18n';
    import type { Thread, Comment, Category, CategoryOption, CategoryResponse } from '$lib/types/forum';
    
    let threads = $state<Thread[]>([]);
    let selectedThread = $state<Thread | null>(null);
    let selectedThreadId = $state<string | null>(null);
    let showNewThreadModal = $state(false);
    let pollInterval = $state<ReturnType<typeof setInterval> | null>(null);
    let selectedCategory = $state<Category>('general');
    let categories = $state<CategoryOption[]>([]);
    let isLoading = $state(false);

    async function loadCategories() {
        try {
            const response = await fetch(`/api/categories?lang=${$language}`);
            if (!response.ok) throw new Error('Failed to load categories');
            const data: CategoryResponse[] = await response.json();
            categories = data.map(cat => ({
                value: cat.value,
                label: cat.label
            }));
            
            // If we have categories and no selected category yet, select the first one
            if (categories.length > 0 && !selectedCategory) {
                selectedCategory = categories[0].value;
            }
        } catch (error) {
            console.error('Error loading categories:', error);
            // Fallback to default categories if API fails
            categories = [
                { value: 'general', label: 'General' },
                { value: 'help', label: 'Help' },
                { value: 'discussion', label: 'Discussion' },
                { value: 'announcement', label: 'Announcement' }
            ];
        }
    }

    onMount(async () => {
        await loadCategories();
        await loadThreads();
    });

    // Reload data when language changes
    $effect(() => {
        console.log('Language changed to:', $language);
        // Use Promise.all to load categories and threads in parallel
        Promise.all([
            loadCategories(),
            loadThreads()
        ]).then(() => {
            // After both are loaded, reload the selected thread if any
            if (selectedThreadId) {
                loadThread(selectedThreadId);
            }
        }).catch(error => {
            console.error('Error reloading data after language change:', error);
        });
    });

    // Function to poll for updates
    async function pollForUpdates() {
        if (!selectedThreadId) return;
        
        try {
            const response = await fetch(`/api/threads/${selectedThreadId}?lang=${$language}`);
            if (response.ok) {
                const updatedThread: Thread = await response.json();
                // Initialize comments array if it doesn't exist
                updatedThread.comments = updatedThread.comments || [];
                // Ensure we keep the same order by sorting by creation time
                updatedThread.comments = updatedThread.comments.sort((a: Comment, b: Comment) => 
                    new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
                );
                selectedThread = updatedThread;
            }
        } catch (error) {
            console.error('Error polling for updates:', error);
        }
    }

    // Start polling when a thread is loaded
    async function loadThread(id: string) {
        try {
            console.log('Loading thread:', id);
            // Clear existing interval if any
            if (pollInterval) {
                clearInterval(pollInterval);
                pollInterval = null;
            }

            const response = await fetch(`/api/threads/${id}?lang=${$language}`);
            console.log('Thread API response status:', response.status);
            
            if (!response.ok) {
                throw new Error('Failed to load thread');
            }
            
            const thread: Thread = await response.json();
            console.log('Received thread data:', thread);
            
            // Initialize comments array if it doesn't exist
            thread.comments = thread.comments || [];
            
            // Ensure comments are sorted by creation time in descending order
            thread.comments = thread.comments.sort((a: Comment, b: Comment) => 
                new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
            );
            
            console.log('Setting selectedThread:', thread);
            selectedThread = thread;
            selectedThreadId = id;

            // Start polling for updates every 2 seconds
            pollInterval = setInterval(pollForUpdates, 2000);
            
            console.log('Thread loaded successfully, selectedThread:', selectedThread);
        } catch (error) {
            console.error('Error loading thread:', error);
            selectedThread = null;
            selectedThreadId = null;
        }
    }

    // Clean up interval when component is destroyed
    onDestroy(() => {
        if (pollInterval) {
            clearInterval(pollInterval);
            pollInterval = null;
        }
    });

    async function createThread(event: SubmitEvent) {
        event.preventDefault();
        const formData = new FormData(event.target as HTMLFormElement);
        const title = formData.get('title') as string;
        const content = formData.get('content') as string;
        
        try {
            const response = await fetch('/api/threads', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    title,
                    content,
                    category: selectedCategory
                }),
            });
            
            if (!response.ok) throw new Error('Failed to create thread');
            
            const newThread: Thread = await response.json();
            hideNewThreadForm();
            await loadThreads(); // Reload all threads first
            await loadThread(newThread.id); // Then load the new thread details
        } catch (error) {
            console.error('Error creating thread:', error);
            // Optionally show an error message to the user here
        }
    }

    async function addComment(event: SubmitEvent, threadId: string) {
        event.preventDefault();
        const formData = new FormData(event.target as HTMLFormElement);
        const content = formData.get('content') as string;
        const image = formData.get('image') as File;

        const data = new FormData();
        data.append('content', content);
        if (image && image.size > 0) {
            data.append('image', image);
        }

        const response = await fetch(`/api/threads/${threadId}/comments`, {
            method: 'POST',
            body: data
        });

        if (response.ok && selectedThread) {
            const newComment: Comment = await response.json();
            // Ensure we have the correct content structure
            const commentContent = typeof newComment.content === 'string' 
                ? newComment.content 
                : (newComment.content[$language] || Object.values(newComment.content)[0] || '');
            
            // Update the selectedThread with the new comment
            selectedThread = {
                ...selectedThread,
                comments: [
                    ...selectedThread.comments,
                    {
                        ...newComment,
                        content: commentContent
                    }
                ].sort((a: Comment, b: Comment) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
            };
            (event.target as HTMLFormElement).reset();
        }
    }

    function showNewThreadForm() {
        showNewThreadModal = true;
    }

    function hideNewThreadForm() {
        showNewThreadModal = false;
    }

    function insertEmoji(emoji: string, target: string) {
        const textarea = document.querySelector(`[name="${target === 'thread' ? 'content' : 'content'}"]`) as HTMLTextAreaElement;
        if (textarea) {
            const start = textarea.selectionStart;
            const end = textarea.selectionEnd;
            const text = textarea.value;
            textarea.value = text.substring(0, start) + emoji + text.substring(end);
            textarea.selectionStart = textarea.selectionEnd = start + emoji.length;
            textarea.focus();
        }
    }

    function toggleEmojiPicker(id: string) {
        const picker = document.getElementById(id);
        if (picker) {
            picker.classList.toggle('hidden');
        }
    }

    function handleKeyDown(event: KeyboardEvent, action: () => void) {
        if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault();
            action();
        }
    }

    async function loadThreads() {
        try {
            isLoading = true;
            const response = await fetch(`/api/threads?category=${selectedCategory}`);
            if (!response.ok) throw new Error('Failed to load threads');
            const data: Thread[] = await response.json();
            
            // Ensure data is an array
            if (!Array.isArray(data)) {
                console.error('Received invalid data format:', data);
                threads = [];
                return;
            }
            
            // Ensure each thread has a comments array
            threads = data.map(thread => ({
                ...thread,
                comments: Array.isArray(thread.comments) ? thread.comments : []
            }));
            
            // Clear selected thread if it's not in the current category
            if (selectedThread && selectedThread.category !== selectedCategory) {
                selectedThread = null;
                selectedThreadId = null;
                if (pollInterval) {
                    clearInterval(pollInterval);
                    pollInterval = null;
                }
            }
        } catch (error) {
            console.error('Error loading threads:', error);
            threads = [];
        } finally {
            isLoading = false;
        }
    }

    // Add category change handler
    $effect(() => {
        console.log('Category changed to:', selectedCategory);
        loadThreads();
    });

    $effect(() => {
        console.log('selectedThread changed:', selectedThread);
    });

    $effect(() => {
        console.log('selectedThreadId changed:', selectedThreadId);
    });
</script>

<div class="flex gap-8">
    <!-- Left Side - Threads List -->
    <div class="w-1/2">
        <div class="flex justify-between items-center mb-4">
            <h2 class="text-2xl font-semibold">{$t('threads')}</h2>
            <div class="flex items-center gap-4">
                <select
                    bind:value={selectedCategory}
                    class="select w-40 rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                >
                    {#each categories as category}
                        <option value={category.value}>{category.label}</option>
                    {/each}
                </select>
                <button 
                    onclick={showNewThreadForm}
                    class="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
                >
                    {$t('createThread')}
                </button>
            </div>
        </div>

        <!-- Threads List -->
        <div id="threads-list" class="space-y-4">
            {#if isLoading}
                <div class="text-center text-gray-500 p-6 bg-white rounded-lg">
                    {$t('loading')}
                </div>
            {:else if !threads.length}
                <div class="text-center text-gray-500 p-6 bg-white rounded-lg">
                    {$t('noThreads')}
                </div>
            {:else}
                {#each threads as thread (thread.id)}
                    <div 
                        role="button"
                        tabindex="0"
                        class="bg-white p-6 rounded-lg shadow-md cursor-pointer hover:shadow-lg transition-shadow thread-item"
                        class:border-2={selectedThreadId === thread.id}
                        class:border-blue-500={selectedThreadId === thread.id}
                        onclick={() => loadThread(thread.id)}
                        onkeydown={(e) => handleKeyDown(e, () => loadThread(thread.id))}
                    >
                        <h3 class="text-xl font-semibold mb-2">{thread.title}</h3>
                        <p class="text-gray-600 mb-4">{thread.content}</p>
                        <div class="flex justify-between items-center">
                            <span class="text-sm text-gray-500">
                                {$t('postedAt')} {new Date(thread.created_at).toLocaleDateString($language)}
                            </span>
                            <span class="text-sm text-blue-500">{thread.comments.length} {$t('comments')}</span>
                        </div>
                    </div>
                {/each}
            {/if}
        </div>
    </div>

    <!-- Right Side - Thread Details and Comments -->
    <div class="w-1/2">
        {#if !selectedThread}
            <div class="text-center text-gray-500 p-6 bg-white rounded-lg">
                {$t('selectThread')}
            </div>
        {:else}
            <div class="bg-white rounded-lg shadow-md p-6">
                <h2 class="text-2xl font-semibold mb-4">{selectedThread.title}</h2>
                <p class="text-gray-600 mb-6">{selectedThread.content}</p>
                
                {#if (selectedThread.comments || []).length > 0}
                    <div class="mt-6">
                        <h3 class="text-xl font-semibold mb-4">{$t('comments')} ({(selectedThread.comments || []).length})</h3>
                        {#each (selectedThread.comments || []) as comment}
                            <div class="bg-gray-50 p-4 rounded-lg mb-4">
                                <p class="text-gray-700">{comment.content}</p>
                                {#if comment.image_path}
                                    <img src={comment.image_path} alt="Comment attachment" class="mt-2 max-w-full h-auto rounded" />
                                {/if}
                                <span class="text-sm text-gray-500 mt-2 block">
                                    {new Date(comment.created_at).toLocaleDateString($language)}
                                </span>
                            </div>
                        {/each}
                    </div>
                {:else}
                    <div class="text-center text-gray-500 p-4">
                        {$t('noComments')}
                    </div>
                {/if}

                <form 
                    class="mt-6"
                    onsubmit={(e) => {
                        e.preventDefault();
                        if (selectedThread) {
                            addComment(e, selectedThread.id);
                        }
                    }}
                >
                    <div class="relative mb-4">
                        <textarea 
                            name="content"
                            placeholder={$t('writeComment')}
                            class="w-full p-2 border rounded h-24"
                            required
                        ></textarea>
                        <button 
                            type="button"
                            onclick={() => toggleEmojiPicker('comment-emoji-picker')}
                            class="absolute right-2 bottom-2 text-gray-500 hover:text-gray-700"
                            aria-label="Open emoji picker"
                        >
                            😊
                        </button>
                        <div id="comment-emoji-picker" class="hidden absolute bottom-full right-0 bg-white border rounded-lg shadow-lg p-2 mb-2">
                            <div class="grid grid-cols-8 gap-1 max-h-48 overflow-y-auto">
                                <button type="button" onclick={() => insertEmoji('😊', 'comment')} class="emoji-btn" aria-label="Insert smile emoji">😊</button>
                                <button type="button" onclick={() => insertEmoji('👍', 'comment')} class="emoji-btn" aria-label="Insert thumbs up emoji">👍</button>
                                <button type="button" onclick={() => insertEmoji('❤️', 'comment')} class="emoji-btn" aria-label="Insert heart emoji">❤️</button>
                                <button type="button" onclick={() => insertEmoji('😂', 'comment')} class="emoji-btn" aria-label="Insert laugh emoji">😂</button>
                                <button type="button" onclick={() => insertEmoji('🎉', 'comment')} class="emoji-btn" aria-label="Insert party emoji">🎉</button>
                                <button type="button" onclick={() => insertEmoji('🤔', 'comment')} class="emoji-btn" aria-label="Insert thinking emoji">🤔</button>
                                <button type="button" onclick={() => insertEmoji('👏', 'comment')} class="emoji-btn" aria-label="Insert clap emoji">👏</button>
                                <button type="button" onclick={() => insertEmoji('🙌', 'comment')} class="emoji-btn" aria-label="Insert raised hands emoji">🙌</button>
                            </div>
                        </div>
                    </div>
                    <div class="flex items-center gap-4 mb-4">
                        <input 
                            type="file" 
                            name="image" 
                            accept="image/*"
                            class="block w-full text-sm text-gray-500
                            file:mr-4 file:py-2 file:px-4
                            file:rounded file:border-0
                            file:text-sm file:font-semibold
                            file:bg-blue-50 file:text-blue-700
                            hover:file:bg-blue-100"
                        />
                    </div>
                    <button 
                        type="submit"
                        class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
                    >
                        {$t('addComment')}
                    </button>
                </form>
            </div>
        {/if}
    </div>
</div>

<!-- New Thread Modal -->
{#if showNewThreadModal}
    <div class="fixed inset-0 modal-backdrop flex items-center justify-center z-40">
        <div class="bg-white rounded-lg shadow-xl w-full max-w-2xl mx-4 z-50">
            <div class="p-6">
                <div class="flex justify-between items-center mb-4">
                    <h2 class="text-2xl font-semibold">{$t('createThread')}</h2>
                    <button 
                        onclick={hideNewThreadForm}
                        class="text-gray-500 hover:text-gray-700"
                        aria-label="Close modal"
                    >
                        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                        </svg>
                    </button>
                </div>
                <form onsubmit={createThread}>
                    <div class="mb-4">
                        <label for="thread-category" class="block text-sm font-medium text-gray-700 mb-1">
                            {$t('category')}
                        </label>
                        <select 
                            id="thread-category"
                            bind:value={selectedCategory}
                            class="w-full p-2 border rounded"
                            required
                        >
                            {#each categories as category}
                                <option value={category.value}>{category.label}</option>
                            {/each}
                        </select>
                    </div>
                    <input 
                        type="text" 
                        name="title"
                        placeholder={$t('createThread')}
                        class="w-full mb-4 p-2 border rounded"
                        required
                    >
                    <div class="relative mb-4">
                        <textarea 
                            name="content"
                            placeholder={$t('writeComment')}
                            class="w-full p-2 border rounded h-32"
                            required
                        ></textarea>
                        <button 
                            type="button"
                            onclick={() => toggleEmojiPicker('thread-emoji-picker')}
                            class="absolute right-2 bottom-2 text-gray-500 hover:text-gray-700"
                            aria-label="Open emoji picker"
                        >
                            😊
                        </button>
                        <div id="thread-emoji-picker" class="hidden absolute bottom-full right-0 bg-white border rounded-lg shadow-lg p-2 mb-2">
                            <div class="grid grid-cols-8 gap-1 max-h-48 overflow-y-auto">
                                <button type="button" onclick={() => insertEmoji('😊', 'thread')} class="emoji-btn" aria-label="Insert smile emoji">😊</button>
                                <button type="button" onclick={() => insertEmoji('👍', 'thread')} class="emoji-btn" aria-label="Insert thumbs up emoji">👍</button>
                                <button type="button" onclick={() => insertEmoji('❤️', 'thread')} class="emoji-btn" aria-label="Insert heart emoji">❤️</button>
                                <button type="button" onclick={() => insertEmoji('😂', 'thread')} class="emoji-btn" aria-label="Insert laugh emoji">😂</button>
                                <button type="button" onclick={() => insertEmoji('🎉', 'thread')} class="emoji-btn" aria-label="Insert party emoji">🎉</button>
                                <button type="button" onclick={() => insertEmoji('🤔', 'thread')} class="emoji-btn" aria-label="Insert thinking emoji">🤔</button>
                                <button type="button" onclick={() => insertEmoji('👏', 'thread')} class="emoji-btn" aria-label="Insert clap emoji">👏</button>
                                <button type="button" onclick={() => insertEmoji('🙌', 'thread')} class="emoji-btn" aria-label="Insert raised hands emoji">🙌</button>
                            </div>
                        </div>
                    </div>
                    <div class="flex justify-end gap-2">
                        <button 
                            type="button"
                            onclick={hideNewThreadForm}
                            class="px-4 py-2 rounded text-gray-600 hover:bg-gray-100"
                        >
                            {$t('cancel')}
                        </button>
                        <button 
                            type="submit"
                            class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
                        >
                            {$t('create')}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>
{/if}

<style>
.emoji-btn {
    padding: 4px;
    font-size: 1.25rem;
    cursor: pointer;
    transition: transform 0.1s;
    border-radius: 4px;
}
.emoji-btn:hover {
    background-color: #f3f4f6;
    transform: scale(1.1);
}
.modal-backdrop {
    background-color: rgba(0, 0, 0, 0.5);
}
</style> 