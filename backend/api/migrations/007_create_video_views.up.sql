CREATE TABLE video_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id VARCHAR(255),
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    watch_seconds INTEGER
);

CREATE INDEX idx_video_views_video_id ON video_views (video_id);
CREATE INDEX idx_video_views_viewed_at ON video_views (viewed_at DESC);
CREATE INDEX idx_video_views_trending ON video_views (video_id, viewed_at DESC);
