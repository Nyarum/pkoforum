DROP TABLE IF EXISTS comment_images;
DROP TABLE IF EXISTS comment_translations;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS threads;

-- Create threads table
CREATE TABLE IF NOT EXISTS threads (
    id VARCHAR(255) PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    created_at TIMESTAMP NOT NULL
);

-- Create comments table
CREATE TABLE IF NOT EXISTS comments (
    id VARCHAR(255) PRIMARY KEY,
    thread_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (thread_id) REFERENCES threads(id)
);

-- Create comment translations table
CREATE TABLE IF NOT EXISTS comment_translations (
    id VARCHAR(255) PRIMARY KEY,
    comment_id VARCHAR(255) NOT NULL,
    language VARCHAR(10) NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (comment_id) REFERENCES comments(id),
    UNIQUE (comment_id, language)
);

-- Create comment images table
CREATE TABLE IF NOT EXISTS comment_images (
    id VARCHAR(255) PRIMARY KEY,
    comment_id VARCHAR(255) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    filepath VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (comment_id) REFERENCES comments(id)
); 