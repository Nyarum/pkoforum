declare module '@skeletonlabs/skeleton-svelte/components/Select' {
  import type { SvelteComponent } from 'svelte';
  
  export class Select extends SvelteComponent<{
    value: any;
    options: Array<{ value: any; label: string }>;
    placeholder?: string;
  }> {}
}

declare module '@skeletonlabs/skeleton-svelte' {
  export * from '@skeletonlabs/skeleton-svelte/components/Select';
} 