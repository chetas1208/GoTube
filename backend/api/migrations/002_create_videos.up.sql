CREATE TYPE video_status AS ENUM ('uploaded', 'queued', 'processing', 'ready', 'failed');
CREATE TYPE video_visibility AS ENUM ('public', 'unlisted', 'private');

CREATE TABLE videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status video_status NOT NULL DEFAULT 'uploaded',
    visibility video_visibility NOT NULL DEFAULT 'public',
    duration_seconds INTEGER,
    thumbnail_object_key TEXT,
    source_object_key TEXT NOT NULL,
    processed_object_key TEXT,
    views_count INTEGER NOT NULL DEFAULT 0,
    likes_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);

CREATE INDEX idx_videos_user_id ON videos (user_id);
CREATE INDEX idx_videos_status ON videos (status);
CREATE INDEX idx_videos_created_at ON videos (created_at DESC);
CREATE INDEX idx_videos_views_count ON videos (views_count DESC);
CREATE INDEX idx_videos_title_trgm ON videos USING GIN (title gin_trgm_ops);
