CREATE TABLE video_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    tag VARCHAR(100) NOT NULL
);

CREATE INDEX idx_video_tags_video_id ON video_tags (video_id);
CREATE INDEX idx_video_tags_tag ON video_tags (tag);
CREATE INDEX idx_video_tags_tag_trgm ON video_tags USING GIN (tag gin_trgm_ops);

ALTER TABLE video_tags ADD CONSTRAINT uq_video_tags UNIQUE (video_id, tag);
