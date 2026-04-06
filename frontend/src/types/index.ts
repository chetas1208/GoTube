export interface User {
  id: string;
  username: string;
  email: string;
  avatar_url?: string;
}

export interface AuthResponse {
  access_token: string;
  user: User;
}

export interface Video {
  id: string;
  user_id: string;
  username: string;
  title: string;
  description: string;
  status: VideoStatus;
  visibility: string;
  duration_seconds?: number;
  thumbnail_url?: string;
  views_count: number;
  likes_count: number;
  tags: string[];
  created_at: string;
  published_at?: string;
  user_has_liked: boolean;
}

export type VideoStatus = "uploaded" | "queued" | "processing" | "ready" | "failed";

export interface VideoListResponse {
  videos: Video[];
  total_count: number;
  page: number;
  per_page: number;
}

export interface Comment {
  id: string;
  video_id: string;
  user_id: string;
  username: string;
  parent_id?: string;
  body: string;
  created_at: string;
}

export interface CommentListResponse {
  comments: Comment[];
  total_count: number;
  page: number;
  per_page: number;
}

export interface InitiateUploadResponse {
  video_id: string;
  upload_url: string;
  object_key: string;
}

export interface PlaybackResponse {
  playback_url: string;
  content_type: string;
}

export interface ApiError {
  error?: string;
  errors?: Record<string, string>;
}
