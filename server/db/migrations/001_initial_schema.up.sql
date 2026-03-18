-- Migration 001: Initial Schema (users and characters tables)
-- Created: 2026-03-16

-- Drop tables if they exist (for development)
DROP TABLE IF EXISTS characters CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Users table
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_login TIMESTAMPTZ,
    is_banned BOOLEAN DEFAULT FALSE
);

-- Characters table
CREATE TABLE characters (
    character_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    class_type VARCHAR(20) DEFAULT 'warrior',
    level INTEGER DEFAULT 1,
    exp BIGINT DEFAULT 0,
    current_zone VARCHAR(50) DEFAULT 'starter_zone',
    position_x FLOAT DEFAULT 0,
    position_y FLOAT DEFAULT 0,
    position_z FLOAT DEFAULT 0,
    orientation FLOAT DEFAULT 0,
    is_online BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_characters_user_id ON characters(user_id);
CREATE INDEX idx_characters_name ON characters(name);

-- Constraints
ALTER TABLE characters ADD CONSTRAINT valid_class_type 
    CHECK (class_type IN ('warrior', 'rogue', 'mage'));

ALTER TABLE characters ADD CONSTRAINT valid_name_length 
    CHECK (LENGTH(name) >= 3 AND LENGTH(name) <= 30);

ALTER TABLE users ADD CONSTRAINT valid_username_length 
    CHECK (LENGTH(username) >= 3 AND LENGTH(username) <= 50);

-- Seed data for testing (password is "test123" hashed with bcrypt cost 10)
INSERT INTO users (username, email, password_hash, created_at) VALUES
    ('testuser1', 'test1@example.com', '$2a$10$rLS5X9bT7bLjJZ7K5Z5L0OqJ3K5Z5L0OqJ3K5Z5L0OqJ3K5Z5L0Oq', NOW()),
    ('testuser2', 'test2@example.com', '$2a$10$rLS5X9bT7bLjJZ7K5Z5L0OqJ3K5Z5L0OqJ3K5Z5L0OqJ3K5Z5L0Oq', NOW());

INSERT INTO characters (user_id, name, class_type, position_x, position_y, position_z) VALUES
    ((SELECT user_id FROM users WHERE username = 'testuser1'), 'Hero1', 'warrior', 0, 0, 0),
    ((SELECT user_id FROM users WHERE username = 'testuser2'), 'Hero2', 'mage', 5, 0, 5);
