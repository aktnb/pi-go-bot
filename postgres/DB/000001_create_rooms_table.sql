-- Up
CREATE TABLE IF NOT EXISTS room (
    id SERIAL PRIMARY KEY,
    guild_id TEXT NOT NULL,
    voicechannel_id TEXT NOT NULL,
    textchannel_id TEXT NOT NULL
);
ALTER TABLE room ADD UNIQUE (voicechannel_id);
ALTER TABLE room ADD UNIQUE (textchannel_id);
