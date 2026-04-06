"use client";

import { useState, useEffect } from "react";
import { videos as videosApi } from "@/lib/api-client";
import { VideoCard } from "@/components/video/VideoCard";
import { VideoCardSkeleton } from "@/components/ui/Skeleton";
import { Pagination } from "@/components/ui/Pagination";
import type { Video } from "@/types";

export default function HomePage() {
  const [videoList, setVideoList] = useState<Video[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      setLoading(true);
      try {
        const data = await videosApi.list(page, 20);
        setVideoList(data.videos);
        setTotal(data.total_count);
      } catch {
        // Handle silently
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [page]);

  return (
    <div>
      <h1 className="mb-6 text-2xl font-bold">Recent Videos</h1>

      {loading ? (
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <VideoCardSkeleton key={i} />
          ))}
        </div>
      ) : videoList.length === 0 ? (
        <div className="py-20 text-center">
          <p className="text-lg text-gray-400">No videos yet. Be the first to upload!</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {videoList.map((video) => (
              <VideoCard key={video.id} video={video} />
            ))}
          </div>
          <Pagination page={page} totalCount={total} perPage={20} onPageChange={setPage} />
        </>
      )}
    </div>
  );
}
