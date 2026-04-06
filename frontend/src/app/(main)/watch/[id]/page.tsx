"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { videos as videosApi } from "@/lib/api-client";
import { useAuth } from "@/lib/auth";
import { VideoPlayer } from "@/components/video/VideoPlayer";
import { CommentSection } from "@/components/comments/CommentSection";
import { formatViews, formatTimeAgo } from "@/lib/utils";
import type { Video } from "@/types";

const VIEW_PING_COOLDOWN_MS = 15 * 1000;

function shouldRecordView(videoId: string) {
  if (typeof window === "undefined") {
    return false;
  }

  const key = `gotube:view-ping:${videoId}`;
  const now = Date.now();
  const previous = Number(window.sessionStorage.getItem(key) ?? "0");

  if (Number.isFinite(previous) && previous > 0 && now-previous < VIEW_PING_COOLDOWN_MS) {
    return false;
  }

  window.sessionStorage.setItem(key, String(now));
  return true;
}

export default function WatchPage() {
  const params = useParams();
  const videoId = params.id as string;
  const { user } = useAuth();
  const [video, setVideo] = useState<Video | null>(null);
  const [loading, setLoading] = useState(true);
  const [liked, setLiked] = useState(false);
  const [likesCount, setLikesCount] = useState(0);

  useEffect(() => {
    async function load() {
      try {
        const data = await videosApi.get(videoId);
        setVideo(data);
        setLiked(data.user_has_liked);
        setLikesCount(data.likes_count);

        if (shouldRecordView(videoId)) {
          videosApi.recordView(videoId).catch(() => {});
        }
      } catch {
        // Handle
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [videoId]);

  const handleLike = async () => {
    if (!user) return;
    try {
      const result = await videosApi.like(videoId);
      setLiked(result.liked);
      setLikesCount((prev) => (result.liked ? prev + 1 : prev - 1));
    } catch {
      // Handle silently
    }
  };

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="aspect-video animate-pulse rounded-lg bg-gray-200" />
        <div className="h-6 w-2/3 animate-pulse rounded bg-gray-200" />
        <div className="h-4 w-1/3 animate-pulse rounded bg-gray-200" />
      </div>
    );
  }

  if (!video) {
    return (
      <div className="py-20 text-center">
        <p className="text-lg text-gray-400">Video not found.</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-8 lg:grid-cols-3">
      <div className="lg:col-span-2">
        <VideoPlayer videoId={videoId} />

        <div className="mt-4 space-y-3">
          <h1 className="text-xl font-bold">{video.title}</h1>

          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-500">
              <span className="font-medium text-gray-700">{video.username}</span>
              <span className="mx-2">&middot;</span>
              {formatViews(video.views_count)}
              <span className="mx-2">&middot;</span>
              {formatTimeAgo(video.created_at)}
            </div>

            <button
              onClick={handleLike}
              disabled={!user}
              className={`flex items-center gap-1.5 rounded-full px-4 py-1.5 text-sm font-medium transition ${
                liked
                  ? "bg-brand-100 text-brand-700"
                  : "bg-gray-100 text-gray-700 hover:bg-gray-200"
              } disabled:cursor-not-allowed disabled:opacity-50`}
            >
              <svg className="h-4 w-4" fill={liked ? "currentColor" : "none"} viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5" />
              </svg>
              {likesCount}
            </button>
          </div>

          {video.description && (
            <div className="rounded-lg bg-gray-50 p-4">
              <p className="whitespace-pre-wrap text-sm text-gray-700">{video.description}</p>
            </div>
          )}

          {video.tags.length > 0 && (
            <div className="flex flex-wrap gap-1.5">
              {video.tags.map((tag) => (
                <span key={tag} className="rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-600">
                  #{tag}
                </span>
              ))}
            </div>
          )}
        </div>

        <div className="mt-8">
          <CommentSection videoId={videoId} />
        </div>
      </div>

      <div className="lg:col-span-1">
        <h3 className="mb-4 text-sm font-semibold text-gray-500 uppercase">More Videos</h3>
        <p className="text-sm text-gray-400">Recommendations coming soon.</p>
      </div>
    </div>
  );
}
