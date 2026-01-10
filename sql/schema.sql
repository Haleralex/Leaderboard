CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS scores (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score BIGINT NOT NULL CHECK (score >= 0),
    season TEXT NOT NULL DEFAULT 'global',
    metadata JSONB,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_user_season UNIQUE (user_id, season)
);

CREATE INDEX IF NOT EXISTS idx_scores_user_id ON scores(user_id);
CREATE INDEX IF NOT EXISTS idx_scores_season ON scores(season);
CREATE INDEX IF NOT EXISTS idx_scores_season_score ON scores(season, score DESC);
CREATE INDEX IF NOT EXISTS idx_scores_timestamp ON scores(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- View for leaderboard with user info
CREATE OR REPLACE VIEW leaderboard_view AS
SELECT
    DENSE_RANK() OVER (PARTITION BY s.season ORDER BY s.score DESC, s.timestamp ASC) as rank,
    s.id,
    s.user_id,
    u.name as user_name,
    s.score,
    s.season,
    s.timestamp
FROM scores s
JOIN users u ON s.user_id = u.id
ORDER BY s.season, s.score DESC, s.timestamp ASC;

COMMENT ON TABLE users IS 'Game players with authentication';
COMMENT ON TABLE scores IS 'Player scores with seasonal support and metadata';
COMMENT ON VIEW leaderboard_view IS 'Materialized view for fast leaderboard queries';
