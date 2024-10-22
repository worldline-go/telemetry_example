CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL INDEX UNIQUE,
    description TEXT NOT NULL,
    last_user VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
);