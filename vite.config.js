import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import path from 'path';

export default defineConfig({
  plugins: [sveltekit()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080'
    },
    fs: {
      allow: [
        // Default allowed directories
        'src',
        'node_modules',
        '.svelte-kit',
        // Add static directory to allowed list
        'static'
      ]
    }
  },
  resolve: {
    alias: {
      $static: path.resolve('./static'),
      '@skeletonlabs/skeleton-svelte': path.resolve('./node_modules/@skeletonlabs/skeleton-svelte'),
      '$lib': path.resolve('./src/lib')
    }
  },
  optimizeDeps: {
    include: ['@skeletonlabs/skeleton-svelte']
  }
}); 