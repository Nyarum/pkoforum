export type Category = 'general' | 'help' | 'discussion' | 'announcement';

export interface Comment {
    id: string;
    content: string | Record<string, string>;
    created_at: string;
    image_url?: string;
    image_path?: string;
}

export interface Thread {
    id: string;
    title: string;
    content: string;
    category: Category;
    created_at: string;
    comments: Comment[];
}

export interface CategoryOption {
    value: Category;
    label: string;
}

export interface CategoryResponse {
    value: Category;
    label: string;
} 