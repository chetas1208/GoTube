CREATE TABLE likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_likes_video_user UNIQUE (video_id, user_id)
);

CREATE INDEX idx_likes_video_id ON likes (video_id);
CREATE INDEX idx_likes_user_id ON likes (user_id);
