"use client";

import Link from "next/link";
import { redirect } from "next/navigation";
import { Sparkles, Upload, UserRound } from "lucide-react";
import { useAuth } from "@/lib/auth";

function ProfileAvatar({ username, avatarUrl }: { username: string; avatarUrl?: string }) {
  const initial = username.slice(0, 1).toUpperCase();

  return (
    <div className="relative h-20 w-20 overflow-hidden rounded-full border border-white/70 bg-gradient-to-br from-sky-500 via-blue-500 to-indigo-500 shadow-[inset_0_1px_0_rgba(255,255,255,0.35)]">
      {avatarUrl ? (
        <img src={avatarUrl} alt={`${username} avatar`} className="h-full w-full object-cover" />
      ) : (
        <div className="flex h-full w-full items-center justify-center font-[family:var(--font-display)] text-3xl font-semibold text-white">
          {initial}
        </div>
      )}
      <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(135deg,rgba(255,255,255,0.32),transparent_58%)]" />
    </div>
  );
}

export default function ProfilePage() {
  const { user, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-24">
        <div className="h-9 w-9 animate-spin rounded-full border-2 border-sky-500 border-t-transparent" />
      </div>
    );
  }

  if (!user) {
    redirect("/login");
  }

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <section className="relative overflow-hidden rounded-[32px] border border-white/70 bg-white/80 p-6 shadow-[0_28px_90px_rgba(15,23,42,0.12)] backdrop-blur-2xl sm:p-8">
        <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(56,189,248,0.16),_transparent_35%),linear-gradient(135deg,_rgba(255,255,255,0.4),_rgba(255,255,255,0.12)_48%,_rgba(99,102,241,0.08))]" />
        <div className="relative flex flex-col gap-6 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-4">
            <ProfileAvatar username={user.username} avatarUrl={user.avatar_url} />
            <div className="min-w-0">
              <div className="inline-flex items-center gap-2 rounded-full border border-sky-200/80 bg-sky-50/80 px-3 py-1 text-xs font-semibold uppercase tracking-[0.28em] text-sky-700">
                <UserRound className="h-3.5 w-3.5" />
                Profile
              </div>
              <h1 className="mt-3 font-[family:var(--font-display)] text-3xl font-semibold tracking-tight text-slate-950">
                {user.username}
              </h1>
              <p className="mt-1 text-sm text-slate-600">{user.email}</p>
            </div>
          </div>

          <div className="flex flex-wrap gap-3">
            <Link
              href="/studio"
              className="inline-flex items-center gap-2 rounded-full border border-white/80 bg-white/85 px-4 py-2 text-sm font-medium text-slate-700 shadow-[inset_0_1px_0_rgba(255,255,255,0.74)] transition hover:-translate-y-0.5 hover:bg-white"
            >
              <Sparkles className="h-4 w-4 text-indigo-600" />
              My Studio
            </Link>
            <Link
              href="/upload"
              className="inline-flex items-center gap-2 rounded-full bg-slate-950 px-4 py-2 text-sm font-medium text-white shadow-[0_16px_32px_rgba(15,23,42,0.18)] transition hover:-translate-y-0.5 hover:bg-slate-800"
            >
              <Upload className="h-4 w-4" />
              Upload Video
            </Link>
          </div>
        </div>
      </section>

      <section className="grid gap-4 md:grid-cols-2">
        <div className="rounded-[28px] border border-white/70 bg-white/78 p-6 shadow-[0_20px_70px_rgba(15,23,42,0.08)] backdrop-blur-xl">
          <p className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Account</p>
          <dl className="mt-4 space-y-4">
            <div>
              <dt className="text-sm font-medium text-slate-500">Username</dt>
              <dd className="mt-1 text-base font-medium text-slate-900">{user.username}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-slate-500">Email</dt>
              <dd className="mt-1 break-all text-base font-medium text-slate-900">{user.email}</dd>
            </div>
          </dl>
        </div>

        <div className="rounded-[28px] border border-white/70 bg-white/78 p-6 shadow-[0_20px_70px_rgba(15,23,42,0.08)] backdrop-blur-xl">
          <p className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">Quick Actions</p>
          <div className="mt-4 space-y-3">
            <Link
              href="/studio"
              className="flex items-center justify-between rounded-2xl bg-slate-50 px-4 py-3 text-sm font-medium text-slate-700 transition hover:bg-slate-100"
            >
              Go to studio
              <Sparkles className="h-4 w-4 text-indigo-600" />
            </Link>
            <Link
              href="/upload"
              className="flex items-center justify-between rounded-2xl bg-slate-950 px-4 py-3 text-sm font-medium text-white transition hover:bg-slate-800"
            >
              Upload a new video
              <Upload className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </section>
    </div>
  );
}
