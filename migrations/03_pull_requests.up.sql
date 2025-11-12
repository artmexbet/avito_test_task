CREATE TABLE IF NOT EXISTS pull_requests (
    id VARCHAR(50) PRIMARY KEY,  -- Допустил, что идентификатор может быть любым строковым
    name varchar(300) NOT NULL,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    -- Статус не стал добавлять - можно будет определять по дате
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP WITHOUT TIME ZONE
);