CREATE TYPE upload_status AS ENUM ('initiated', 'uploading', 'completed', 'failed', 'expired');

CREATE TABLE upload_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    object_key TEXT NOT NULL,
    upload_status upload_status NOT NULL DEFAULT 'initiated',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_upload_sessions_user_id ON upload_sessions (user_id);
CREATE INDEX idx_upload_sessions_video_id ON upload_sessions (video_id);
