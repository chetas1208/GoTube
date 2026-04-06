"use client";

import { Suspense, useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { search as searchApi } from "@/lib/api-client";
import { VideoCard } from "@/components/video/VideoCard";
import { VideoCardSkeleton } from "@/components/ui/Skeleton";
import { Pagination } from "@/components/ui/Pagination";
import type { Video } from "@/types";

function SearchPageContent() {
  const searchParams = useSearchParams();
  const q = searchParams.get("q") || "";
  const [videoList, setVideoList] = useState<Video[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [sortBy, setSortBy] = useState("relevance");

  useEffect(() => {
    if (!q) return;
    async function load() {
      setLoading(true);
      try {
        const data = await searchApi.query(q, page, 20, sortBy);
        setVideoList(data.videos);
        setTotal(data.total_count);
      } catch {
        // Handle silently
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [q, page, sortBy]);

  if (!q) {
    return (
      <div className="py-20 text-center">
        <p className="text-lg text-gray-400">Enter a search query to find videos.</p>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold">
          Results for &ldquo;{q}&rdquo;
          {!loading && <span className="ml-2 text-base font-normal text-gray-400">({total} results)</span>}
        </h1>
        <select
          value={sortBy}
          onChange={(e) => { setSortBy(e.target.value); setPage(1); }}
          className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm"
        >
          <option value="relevance">Relevance</option>
          <option value="recent">Most Recent</option>
          <option value="views">Most Views</option>
        </select>
      </div>

      {loading ? (
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <VideoCardSkeleton key={i} />
          ))}
        </div>
      ) : videoList.length === 0 ? (
        <div className="py-20 text-center">
          <p className="text-lg text-gray-400">No videos found for &ldquo;{q}&rdquo;</p>
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

export default function SearchPage() {
  return (
    <Suspense fallback={<div className="py-20 text-center"><p className="text-lg text-gray-400">Loading search...</p></div>}>
      <SearchPageContent />
    </Suspense>
  );
}
