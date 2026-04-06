import Link from "next/link";
import type { Video } from "@/types";
import { formatViews, formatTimeAgo, formatDuration } from "@/lib/utils";

export function VideoCard({ video }: { video: Video }) {
  return (
    <Link href={`/watch/${video.id}`} className="group block">
      <div className="relative aspect-video overflow-hidden rounded-lg bg-gray-200">
        {video.thumbnail_url ? (
          <img
            src={video.thumbnail_url}
            alt={video.title}
            className="h-full w-full object-cover transition-transform group-hover:scale-105"
          />
        ) : (
          <div className="flex h-full items-center justify-center bg-gradient-to-br from-gray-300 to-gray-400">
            <svg className="h-12 w-12 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
        )}
        {video.duration_seconds && (
          <span className="absolute bottom-1.5 right-1.5 rounded bg-black/80 px-1.5 py-0.5 text-xs font-medium text-white">
            {formatDuration(video.duration_seconds)}
          </span>
        )}
      </div>
      <div className="mt-2 space-y-1">
        <h3 className="line-clamp-2 text-sm font-medium leading-snug text-gray-900 group-hover:text-brand-600">
          {video.title}
        </h3>
        <p className="text-xs text-gray-500">{video.username}</p>
        <p className="text-xs text-gray-500">
          {formatViews(video.views_count)} &middot; {formatTimeAgo(video.created_at)}
        </p>
      </div>
    </Link>
  );
}
