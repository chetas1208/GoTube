"use client";

import { Loader2, TimerReset } from "lucide-react";

function formatCountdown(totalSeconds: number) {
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

interface SessionTimeoutDialogProps {
  isOpen: boolean;
  secondsRemaining: number | null;
  isRefreshing: boolean;
  onStaySignedIn: () => void | Promise<void>;
}

export function SessionTimeoutDialog({
  isOpen,
  secondsRemaining,
  isRefreshing,
  onStaySignedIn,
}: SessionTimeoutDialogProps) {
  if (!isOpen || secondsRemaining === null) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-[70] flex items-end justify-center bg-slate-950/20 px-4 pb-6 pt-24 backdrop-blur-md sm:items-center">
      <div className="relative w-full max-w-md overflow-hidden rounded-[30px] border border-white/70 bg-white/80 p-6 shadow-[0_28px_120px_rgba(15,23,42,0.22)] backdrop-blur-2xl">
        <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top,_rgba(56,189,248,0.18),_transparent_42%),linear-gradient(135deg,_rgba(14,165,233,0.08),_rgba(59,130,246,0.03)_55%,_rgba(249,115,22,0.08))]" />
        <div className="pointer-events-none absolute -right-10 top-0 h-28 w-28 rounded-full bg-sky-300/35 blur-3xl" />

        <div className="relative space-y-5">
          <div className="inline-flex items-center gap-2 rounded-full border border-sky-200/80 bg-sky-50/80 px-3 py-1 text-xs font-semibold uppercase tracking-[0.28em] text-sky-700">
            <TimerReset className="h-3.5 w-3.5" />
            Session warning
          </div>

          <div className="space-y-2">
            <h2 className="font-[family:var(--font-display)] text-3xl font-semibold tracking-tight text-slate-950">
              You&apos;re about to be signed out
            </h2>
            <p className="max-w-sm text-sm leading-6 text-slate-600">
              Your session ends after 15 minutes without on-screen activity. Stay active to keep working.
            </p>
          </div>

          <div className="rounded-[26px] border border-white/70 bg-white/75 p-4 shadow-[inset_0_1px_0_rgba(255,255,255,0.7)]">
            <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">
              Auto sign-out in
            </div>
            <div className="mt-2 font-[family:var(--font-display)] text-5xl font-semibold tracking-tight text-slate-950">
              {formatCountdown(secondsRemaining)}
            </div>
          </div>

          <button
            type="button"
            onClick={() => void onStaySignedIn()}
            disabled={isRefreshing}
            className="inline-flex w-full items-center justify-center gap-2 rounded-full bg-slate-950 px-5 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-wait disabled:opacity-70"
          >
            {isRefreshing ? <Loader2 className="h-4 w-4 animate-spin" /> : <TimerReset className="h-4 w-4" />}
            {isRefreshing ? "Refreshing session..." : "Stay signed in"}
          </button>
        </div>
      </div>
    </div>
  );
}
