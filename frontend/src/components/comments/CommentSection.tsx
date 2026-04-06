"use client";

import { useState, useEffect, useCallback } from "react";
import { comments as commentsApi } from "@/lib/api-client";
import { useAuth } from "@/lib/auth";
import { useToast } from "@/components/ui/Toast";
import { formatTimeAgo } from "@/lib/utils";
import { Pagination } from "@/components/ui/Pagination";
import type { Comment } from "@/types";

export function CommentSection({ videoId }: { videoId: string }) {
  const { user } = useAuth();
  const { addToast } = useToast();
  const [commentList, setCommentList] = useState<Comment[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [body, setBody] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const loadComments = useCallback(async () => {
    try {
      const data = await commentsApi.list(videoId, page, 20);
      setCommentList(data.comments);
      setTotal(data.total_count);
    } catch {
      // Silently handle
    }
  }, [videoId, page]);

  useEffect(() => {
    loadComments();
  }, [loadComments]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!body.trim() || !user) return;
    setSubmitting(true);
    try {
      await commentsApi.create(videoId, { body: body.trim() });
      setBody("");
      loadComments();
      addToast("Comment posted!", "success");
    } catch {
      addToast("Failed to post comment.", "error");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold">{total} Comments</h3>

      {user && (
        <form onSubmit={handleSubmit} className="flex gap-2">
          <input
            value={body}
            onChange={(e) => setBody(e.target.value)}
            placeholder="Add a comment..."
            maxLength={2000}
            className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500"
          />
          <button
            type="submit"
            disabled={!body.trim() || submitting}
            className="rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white hover:bg-brand-700 disabled:opacity-50"
          >
            Post
          </button>
        </form>
      )}

      <div className="space-y-3">
        {commentList.length === 0 && (
          <p className="py-8 text-center text-sm text-gray-400">No comments yet. Be the first!</p>
        )}
        {commentList.map((comment) => (
          <div key={comment.id} className="flex gap-3 rounded-lg bg-white p-3 shadow-sm">
            <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-gray-200 text-xs font-bold text-gray-600">
              {comment.username[0].toUpperCase()}
            </div>
            <div className="flex-1">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium">{comment.username}</span>
                <span className="text-xs text-gray-400">{formatTimeAgo(comment.created_at)}</span>
              </div>
              <p className="mt-0.5 text-sm text-gray-700">{comment.body}</p>
            </div>
          </div>
        ))}
      </div>

      <Pagination page={page} totalCount={total} perPage={20} onPageChange={setPage} />
    </div>
  );
}
