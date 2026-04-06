"use client";

import { useState, useEffect, useCallback } from "react";
import { useAuth } from "@/lib/auth";
import { videos as videosApi } from "@/lib/api-client";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { Pagination } from "@/components/ui/Pagination";
import { formatViews, formatTimeAgo } from "@/lib/utils";
import { redirect } from "next/navigation";
import Link from "next/link";
import type { Video } from "@/types";

const inflightStatuses = new Set<Video["status"]>(["uploaded", "queued", "processing"]);

function isInflight(video: Video) {
  return inflightStatuses.has(video.status);
}

export default function StudioPage() {
  const { user, isLoading } = useAuth();
  const [videoList, setVideoList] = useState<Video[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editTitle, setEditTitle] = useState("");
  const [editDesc, setEditDesc] = useState("");
  const [focusedVideoId, setFocusedVideoId] = useState<string | null>(null);

  const loadVideos = useCallback(async (silent = false) => {
    if (!silent) {
      setLoading(true);
    }
    try {
      const data = await videosApi.my(page, 20);
      setVideoList(data.videos);
      setTotal(data.total_count);
    } catch {
      // handle
    } finally {
      if (!silent) {
        setLoading(false);
      }
    }
  }, [page]);

  useEffect(() => {
    if (user) {
      void loadVideos();
    }
  }, [user, loadVideos]);

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }
    setFocusedVideoId(new URLSearchParams(window.location.search).get("video"));
  }, []);

  const hasInflightVideos = videoList.some((video) => isInflight(video));

  useEffect(() => {
    if (!user || !hasInflightVideos) {
      return;
    }

    const intervalId = window.setInterval(() => {
      void loadVideos(true);
    }, 3000);

    return () => window.clearInterval(intervalId);
  }, [user, hasInflightVideos, loadVideos]);

  const handleSave = async (videoId: string) => {
    try {
      await videosApi.update(videoId, { title: editTitle, description: editDesc });
      setEditingId(null);
      await loadVideos(true);
    } catch {
      // handle
    }
  };

  const handleDelete = async (videoId: string) => {
    if (!confirm("Delete this video? This cannot be undone.")) return;
    try {
      await videosApi.delete(videoId);
      await loadVideos(true);
    } catch {
      // handle
    }
  };

  if (isLoading) {
    return <div className="flex items-center justify-center py-20"><div className="h-8 w-8 animate-spin rounded-full border-2 border-brand-600 border-t-transparent" /></div>;
  }

  if (!user) {
    redirect("/login");
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold">My Studio</h1>
        <Link href="/upload" className="rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white hover:bg-brand-700">
          Upload Video
        </Link>
      </div>

      {loading ? (
        <div className="space-y-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-20 animate-pulse rounded-lg bg-gray-200" />
          ))}
        </div>
      ) : videoList.length === 0 ? (
        <div className="py-20 text-center">
          <p className="text-lg text-gray-400">You haven&apos;t uploaded any videos yet.</p>
          <Link href="/upload" className="mt-4 inline-block rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white hover:bg-brand-700">
            Upload Your First Video
          </Link>
        </div>
      ) : (
        <>
          <div className="overflow-hidden rounded-lg border border-gray-200">
            <table className="w-full text-left text-sm">
              <thead className="bg-gray-50 text-xs uppercase text-gray-500">
                <tr>
                  <th className="px-4 py-3">Video</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Views</th>
                  <th className="px-4 py-3">Likes</th>
                  <th className="px-4 py-3">Date</th>
                  <th className="px-4 py-3">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {videoList.map((video) => (
                  <tr
                    key={video.id}
                    className={focusedVideoId === video.id ? "bg-brand-50 hover:bg-brand-50" : "hover:bg-gray-50"}
                  >
                    <td className="px-4 py-3">
                      <div className="flex items-start gap-4">
                        <div className="relative h-16 w-28 shrink-0 overflow-hidden rounded-md border border-gray-200 bg-gray-100">
                          {video.thumbnail_url ? (
                            <img
                              src={video.thumbnail_url}
                              alt=""
                              className="h-full w-full object-cover"
                            />
                          ) : isInflight(video) ? (
                            <div className="flex h-full items-center justify-center px-2 text-center text-[11px] font-medium text-gray-500">
                              {focusedVideoId === video.id ? "Generating preview..." : "Processing video"}
                            </div>
                          ) : (
                            <div className="flex h-full items-center justify-center px-2 text-center text-[11px] text-gray-400">
                              No thumbnail
                            </div>
                          )}
                        </div>

                        <div className="min-w-0 flex-1">
                          {editingId === video.id ? (
                            <div className="space-y-2">
                              <input
                                value={editTitle}
                                onChange={(e) => setEditTitle(e.target.value)}
                                className="w-full rounded border border-gray-300 px-2 py-1 text-sm"
                              />
                              <textarea
                                value={editDesc}
                                onChange={(e) => setEditDesc(e.target.value)}
                                rows={2}
                                className="w-full rounded border border-gray-300 px-2 py-1 text-sm"
                              />
                            </div>
                          ) : (
                            <div>
                              {video.status === "ready" ? (
                                <Link href={`/watch/${video.id}`} className="font-medium text-gray-900 hover:text-brand-600">
                                  {video.title}
                                </Link>
                              ) : (
                                <span className="font-medium text-gray-900">{video.title}</span>
                              )}
                              {video.description ? (
                                <p className="mt-1 max-w-xl text-xs text-gray-500">{video.description}</p>
                              ) : null}
                              {focusedVideoId === video.id && isInflight(video) && !video.thumbnail_url ? (
                                <p className="mt-2 text-xs font-medium text-brand-700">Generating preview...</p>
                              ) : null}
                            </div>
                          )}
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3"><StatusBadge status={video.status} /></td>
                    <td className="px-4 py-3 text-gray-500">{formatViews(video.views_count)}</td>
                    <td className="px-4 py-3 text-gray-500">{video.likes_count}</td>
                    <td className="px-4 py-3 text-gray-500">{formatTimeAgo(video.created_at)}</td>
                    <td className="px-4 py-3">
                      {editingId === video.id ? (
                        <div className="flex gap-1">
                          <button onClick={() => handleSave(video.id)} className="rounded bg-green-600 px-2 py-1 text-xs text-white hover:bg-green-700">Save</button>
                          <button onClick={() => setEditingId(null)} className="rounded bg-gray-200 px-2 py-1 text-xs hover:bg-gray-300">Cancel</button>
                        </div>
                      ) : (
                        <div className="flex gap-1">
                          <button
                            onClick={() => { setEditingId(video.id); setEditTitle(video.title); setEditDesc(video.description); }}
                            className="rounded bg-gray-100 px-2 py-1 text-xs hover:bg-gray-200"
                          >
                            Edit
                          </button>
                          <button onClick={() => handleDelete(video.id)} className="rounded bg-red-100 px-2 py-1 text-xs text-red-700 hover:bg-red-200">
                            Delete
                          </button>
                        </div>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <Pagination page={page} totalCount={total} perPage={20} onPageChange={setPage} />
        </>
      )}
    </div>
  );
}
