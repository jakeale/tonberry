CREATE TABLE subscriptions (
    id          BIGSERIAL PRIMARY KEY,
    guild_id    TEXT        NOT NULL,
    channel_id  TEXT        NOT NULL,
    world_name  TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (guild_id, channel_id, world_name)
);

CREATE INDEX idx_subscriptions_world ON subscriptions (world_name);
