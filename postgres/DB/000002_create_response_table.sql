-- Up
CREATE TABLE IF NOT EXISTS response (
    id SERIAL PRIMARY KEY,
    guild_id TEXT NOT NULL,
    response TEXT NOT NULL,
    keyword TEXT NOT NULL
);
ALTER TABLE response ADD UNIQUE (guild_id, keyword);
