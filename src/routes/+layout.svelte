<script>
    import '../app.css';
    import { language, t } from '$lib/stores/language';
    import { browser } from '$app/environment';

    // Function to toggle language
    function toggleLanguage() {
        language.update(current => current === 'en' ? 'ru' : 'en');
    }

    // Update HTML lang attribute when language changes
    $: if (browser && $language) {
        document.documentElement.lang = $language;
    }
</script>

<div class="min-h-screen bg-gray-100">
    <nav class="bg-white shadow-lg">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="flex justify-between h-16">
                <div class="flex">
                    <div class="flex-shrink-0 flex items-center">
                        <a href="/" class="text-xl font-bold text-gray-800">PKO Forum</a>
                    </div>
                    <div class="hidden sm:ml-6 sm:flex sm:space-x-8">
                        <a href="/" class="text-gray-900 inline-flex items-center px-1 pt-1 border-b-2 border-transparent hover:border-gray-300">
                            {$t('home')}
                        </a>
                    </div>
                </div>
                <div class="flex items-center">
                    <button
                        on:click={toggleLanguage}
                        class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                    >
                        {$t('switchLanguage')}
                    </button>
                </div>
            </div>
        </div>
    </nav>

    <main class="container mx-auto px-4 py-8">
        <slot />
    </main>
</div>

<style>
  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen,
      Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  }
</style> 