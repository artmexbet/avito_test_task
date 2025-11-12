CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(50) PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    team_id VARCHAR(50) NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Накинул индекс на team_id, чтобы получение по команде было быстрее
-- Верю, что получать пользователей будем сильно чаще, чем добавлять/обновлять
CREATE INDEX IF NOT EXISTS idx_users_team_id ON users(team_id);
-- Этот индекс для быстрого поиска активных пользователей, которых можно закинуть как ревьюеров
CREATE INDEX IF NOT EXISTS idx_users_team_id_and_active ON users(team_id, is_active);