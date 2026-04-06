CREATE INDEX IF NOT EXISTS idx_video_views_video_user_recent
ON video_views (video_id, user_id, viewed_at DESC)
WHERE user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_video_views_video_session_recent
ON video_views (video_id, session_id, viewed_at DESC)
WHERE session_id IS NOT NULL;
