<script>
    let { children } = $props();
</script>

<div class="app">
    {@render children()}
</div>

<style>
    .app {
        max-width: 1280px;
        margin: 0 auto;
        padding: 2rem;
    }
</style> 