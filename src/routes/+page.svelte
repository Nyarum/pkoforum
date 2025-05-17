<script>
    import { onMount } from 'svelte';
    
    let threads = $state([]);
    let selectedThread = $state(null);
    let selectedThreadId = $state(null);
    let showNewThreadModal = $state(false);

    onMount(async () => {
        const response = await fetch('/api/threads');
        threads = await response.json();
    });

    async function createThread(event) {
        event.preventDefault();
        const formData = new FormData(event.target);
        const title = formData.get('title');
        const content = formData.get('content');

        const response = await fetch('/api/threads', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ title, content })
        });

        if (response.ok) {
            const newThread = await response.json();
            threads = [...threads, newThread];
            hideNewThreadForm();
        }
    }

    async function loadThread(id) {
        const response = await fetch(`/api/threads/${id}`);
        selectedThread = await response.json();
        selectedThreadId = id;
    }

    async function addComment(event, threadId) {
        event.preventDefault();
        const formData = new FormData(event.target);
        const content = formData.get('content');
        const image = formData.get('image');

        const data = new FormData();
        data.append('content', content);
        if (image.size > 0) {
            data.append('image', image);
        }

        const response = await fetch(`/api/threads/${threadId}/comments`, {
            method: 'POST',
            body: data
        });

        if (response.ok) {
            const updatedThread = await response.json();
            selectedThread = updatedThread;
            event.target.reset();
        }
    }

    function showNewThreadForm() {
        showNewThreadModal = true;
    }

    function hideNewThreadForm() {
        showNewThreadModal = false;
    }

    function insertEmoji(emoji, target) {
        const textarea = document.querySelector(`[name="${target === 'thread' ? 'content' : 'content'}"]`);
        if (textarea) {
            const start = textarea.selectionStart;
            const end = textarea.selectionEnd;
            const text = textarea.value;
            textarea.value = text.substring(0, start) + emoji + text.substring(end);
            textarea.selectionStart = textarea.selectionEnd = start + emoji.length;
            textarea.focus();
        }
    }

    function toggleEmojiPicker(id) {
        const picker = document.getElementById(id);
        if (picker) {
            picker.classList.toggle('hidden');
        }
    }
</script>

<div class="flex gap-8">
    <!-- Left Side - Threads List -->
    <div class="w-1/2">
        <div class="flex justify-between items-center mb-4">
            <h2 class="text-2xl font-semibold">Threads</h2>
            <button 
                on:click={showNewThreadForm}
                class="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
            >
                Create New Thread
            </button>
        </div>

        <!-- Threads List -->
        <div id="threads-list" class="space-y-4">
            {#if !threads.length}
                <div class="text-center text-gray-500 p-6 bg-white rounded-lg">
                    No threads yet. Be the first to create one!
                </div>
            {:else}
                {#each threads as thread}
                    <div 
                        class="bg-white p-6 rounded-lg shadow-md cursor-pointer hover:shadow-lg transition-shadow thread-item"
                        class:border-2={selectedThreadId === thread.id}
                        class:border-blue-500={selectedThreadId === thread.id}
                        on:click={() => loadThread(thread.id)}
                    >
                        <h3 class="text-xl font-semibold mb-2">{thread.title}</h3>
                        <p class="text-gray-600 mb-4">{thread.content}</p>
                        <div class="flex justify-between items-center">
                            <span class="text-sm text-gray-500">
                                Created at {new Date(thread.created_at).toLocaleString()}
                            </span>
                            <span class="text-sm text-blue-500">{thread.comments?.length || 0} comments</span>
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
                Select a thread to view details and comments
            </div>
        {:else}
            <div class="bg-white rounded-lg shadow-md p-6">
                <h2 class="text-2xl font-semibold mb-4">{selectedThread.title}</h2>
                <p class="text-gray-600 mb-6">{selectedThread.content}</p>
                <div class="border-t pt-6">
                    <h3 class="text-xl font-semibold mb-4">Comments</h3>
                    <div class="space-y-4">
                        {#if !selectedThread.comments?.length}
                            <div class="text-gray-500 text-center">
                                No comments yet. Be the first to comment!
                            </div>
                        {:else}
                            {#each selectedThread.comments as comment}
                                <div class="bg-gray-50 p-4 rounded-lg">
                                    <p class="text-gray-700 mb-2">{comment.content}</p>
                                    {#if comment.image_path}
                                        <div class="mb-3">
                                            <img src={comment.image_path} alt="Comment image" class="max-w-full h-auto rounded-lg shadow-sm">
                                        </div>
                                    {/if}
                                    <span class="text-sm text-gray-500">
                                        Created at {new Date(comment.created_at).toLocaleString()}
                                    </span>
                                </div>
                            {/each}
                        {/if}
                    </div>
                    <div class="mt-6">
                        <form on:submit={(e) => addComment(e, selectedThread.id)}>
                            <div class="relative mb-4">
                                <textarea 
                                    name="content"
                                    placeholder="Write a comment..." 
                                    class="w-full p-2 border rounded h-24"
                                    required
                                ></textarea>
                                <button 
                                    type="button"
                                    on:click={() => toggleEmojiPicker('comment-emoji-picker')}
                                    class="absolute right-2 bottom-2 text-gray-500 hover:text-gray-700"
                                >
                                    😊
                                </button>
                                <div id="comment-emoji-picker" class="hidden absolute bottom-full right-0 bg-white border rounded-lg shadow-lg p-2 mb-2">
                                    <div class="grid grid-cols-8 gap-1 max-h-48 overflow-y-auto">
                                        <button type="button" on:click={() => insertEmoji('😊', 'comment')} class="emoji-btn">😊</button>
                                        <button type="button" on:click={() => insertEmoji('👍', 'comment')} class="emoji-btn">👍</button>
                                        <button type="button" on:click={() => insertEmoji('❤️', 'comment')} class="emoji-btn">❤️</button>
                                        <button type="button" on:click={() => insertEmoji('😂', 'comment')} class="emoji-btn">😂</button>
                                        <button type="button" on:click={() => insertEmoji('🎉', 'comment')} class="emoji-btn">🎉</button>
                                        <button type="button" on:click={() => insertEmoji('🤔', 'comment')} class="emoji-btn">🤔</button>
                                        <button type="button" on:click={() => insertEmoji('👏', 'comment')} class="emoji-btn">👏</button>
                                        <button type="button" on:click={() => insertEmoji('🙌', 'comment')} class="emoji-btn">🙌</button>
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
                                Add Comment
                            </button>
                        </form>
                    </div>
                </div>
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
                    <h2 class="text-2xl font-semibold">Create New Thread</h2>
                    <button 
                        on:click={hideNewThreadForm}
                        class="text-gray-500 hover:text-gray-700"
                    >
                        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                        </svg>
                    </button>
                </div>
                <form on:submit={createThread}>
                    <input 
                        type="text" 
                        name="title"
                        placeholder="Thread Title" 
                        class="w-full mb-4 p-2 border rounded"
                        required
                    >
                    <div class="relative mb-4">
                        <textarea 
                            name="content"
                            placeholder="Thread Content" 
                            class="w-full p-2 border rounded h-32"
                            required
                        ></textarea>
                        <button 
                            type="button"
                            on:click={() => toggleEmojiPicker('thread-emoji-picker')}
                            class="absolute right-2 bottom-2 text-gray-500 hover:text-gray-700"
                        >
                            😊
                        </button>
                        <div id="thread-emoji-picker" class="hidden absolute bottom-full right-0 bg-white border rounded-lg shadow-lg p-2 mb-2">
                            <div class="grid grid-cols-8 gap-1 max-h-48 overflow-y-auto">
                                <button type="button" on:click={() => insertEmoji('😊', 'thread')} class="emoji-btn">😊</button>
                                <button type="button" on:click={() => insertEmoji('👍', 'thread')} class="emoji-btn">👍</button>
                                <button type="button" on:click={() => insertEmoji('❤️', 'thread')} class="emoji-btn">❤️</button>
                                <button type="button" on:click={() => insertEmoji('😂', 'thread')} class="emoji-btn">😂</button>
                                <button type="button" on:click={() => insertEmoji('🎉', 'thread')} class="emoji-btn">🎉</button>
                                <button type="button" on:click={() => insertEmoji('🤔', 'thread')} class="emoji-btn">🤔</button>
                                <button type="button" on:click={() => insertEmoji('👏', 'thread')} class="emoji-btn">👏</button>
                                <button type="button" on:click={() => insertEmoji('🙌', 'thread')} class="emoji-btn">🙌</button>
                            </div>
                        </div>
                    </div>
                    <div class="flex justify-end gap-2">
                        <button 
                            type="button"
                            on:click={hideNewThreadForm}
                            class="px-4 py-2 rounded text-gray-600 hover:bg-gray-100"
                        >
                            Cancel
                        </button>
                        <button 
                            type="submit"
                            class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
                        >
                            Create
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