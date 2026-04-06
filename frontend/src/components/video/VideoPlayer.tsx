"use client";

import { useEffect, useRef, useState } from "react";
import { videos as videosApi } from "@/lib/api-client";

interface VideoPlayerProps {
  videoId: string;
}

export function VideoPlayer({ videoId }: VideoPlayerProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [playbackUrl, setPlaybackUrl] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchPlayback() {
      try {
        const data = await videosApi.playback(videoId);
        setPlaybackUrl(data.playback_url);
      } catch {
        setError("Video is not available for playback.");
      } finally {
        setLoading(false);
      }
    }
    fetchPlayback();
  }, [videoId]);

  if (loading) {
    return (
      <div className="flex aspect-video items-center justify-center rounded-lg bg-black">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-white border-t-transparent" />
      </div>
    );
  }

  if (error || !playbackUrl) {
    return (
      <div className="flex aspect-video items-center justify-center rounded-lg bg-gray-900">
        <p className="text-sm text-gray-400">{error || "Video unavailable"}</p>
      </div>
    );
  }

  return (
    <video
      ref={videoRef}
      src={playbackUrl}
      controls
      className="aspect-video w-full rounded-lg bg-black"
      preload="metadata"
    />
  );
}
