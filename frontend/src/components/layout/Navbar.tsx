"use client";

import Link from "next/link";
import { useAuth } from "@/lib/auth";
import { BellRing, LogOut, Radio, Sparkles, UserRound } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { SearchBar } from "@/components/ui/SearchBar";
import { cn } from "@/lib/utils";

function formatCountdown(totalSeconds: number | null) {
  if (totalSeconds === null) {
    return null;
  }
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

function AvatarBadge({
  username,
  avatarUrl,
  sizeClassName,
  textClassName,
}: {
  username: string;
  avatarUrl?: string;
  sizeClassName: string;
  textClassName: string;
}) {
  const [imageFailed, setImageFailed] = useState(false);
  const initial = username.slice(0, 1).toUpperCase();

  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-full border border-white/60 bg-gradient-to-br from-sky-500 via-blue-500 to-indigo-500 shadow-[inset_0_1px_0_rgba(255,255,255,0.4)]",
        sizeClassName
      )}
    >
      {avatarUrl && !imageFailed ? (
        <img
          src={avatarUrl}
          alt={`${username} avatar`}
          className="h-full w-full object-cover"
          onError={() => setImageFailed(true)}
        />
      ) : (
        <div className={cn("flex h-full w-full items-center justify-center font-semibold text-white", textClassName)}>
          {initial}
        </div>
      )}
      <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(135deg,rgba(255,255,255,0.35),transparent_55%)]" />
    </div>
  );
}

export function Navbar() {
  const { user, logout, isLoading, isSessionExpiring, secondsUntilLogout } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement | null>(null);
  const buttonRef = useRef<HTMLButtonElement | null>(null);
  const countdown = formatCountdown(secondsUntilLogout);

  useEffect(() => {
    if (!menuOpen) {
      return;
    }

    const handlePointerDown = (event: MouseEvent) => {
      const target = event.target as Node;
      if (menuRef.current?.contains(target) || buttonRef.current?.contains(target)) {
        return;
      }
      setMenuOpen(false);
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setMenuOpen(false);
      }
    };

    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [menuOpen]);

  return (
    <nav className="sticky top-0 z-50 px-3 pt-3 sm:px-4">
      <div className="mx-auto max-w-7xl">
        <div className="relative rounded-[30px] border border-white/70 bg-white/72 shadow-[0_24px_90px_rgba(15,23,42,0.12)] backdrop-blur-2xl">
          <div className="pointer-events-none absolute inset-0 overflow-hidden rounded-[30px]">
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(59,130,246,0.18),_transparent_34%),radial-gradient(circle_at_top_right,_rgba(34,197,94,0.14),_transparent_28%),linear-gradient(135deg,_rgba(255,255,255,0.42),_rgba(255,255,255,0.16)_40%,_rgba(59,130,246,0.06))]" />
            <div className="absolute -left-16 top-0 h-36 w-36 rounded-full bg-sky-300/30 blur-3xl animate-float-orb" />
            <div className="absolute -right-10 top-2 h-28 w-28 rounded-full bg-cyan-200/40 blur-3xl animate-float-orb-delayed" />
            <div className="absolute inset-y-0 left-[-25%] w-1/3 bg-[linear-gradient(90deg,transparent,rgba(255,255,255,0.34),transparent)] animate-glass-sweep" />
          </div>

          <div className="relative mx-auto flex h-16 items-center justify-between gap-4 px-4 sm:px-5">
            <div className="flex min-w-0 items-center gap-3 sm:gap-5">
              <Link href="/" className="group flex min-w-0 items-center gap-3">
                <div className="relative flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl bg-slate-950 text-white shadow-[0_16px_36px_rgba(15,23,42,0.22)]">
                  <div className="absolute inset-0 bg-[linear-gradient(135deg,rgba(56,189,248,0.55),transparent_55%,rgba(96,165,250,0.7))]" />
                  <svg viewBox="0 0 24 24" className="relative h-6 w-6 fill-current" aria-hidden="true">
                    <path d="M19.615 3.184c-3.604-.246-11.631-.245-15.23 0C.488 3.45.029 5.804 0 12c.029 6.185.484 8.549 4.385 8.816 3.6.245 11.626.246 15.23 0C23.512 20.55 23.971 18.196 24 12c-.029-6.185-.484-8.549-4.385-8.816zM9 16V8l8 4-8 4z" />
                  </svg>
                </div>
                <div className="min-w-0">
                  <div className="font-[family:var(--font-display)] text-xl font-semibold tracking-tight text-slate-950">
                    GoTube
                  </div>
                  <div className="hidden text-[11px] uppercase tracking-[0.32em] text-slate-500 sm:block">
                    Motion studio
                  </div>
                </div>
              </Link>

              <div className="hidden xl:block">
                <SearchBar
                  className="max-w-xl"
                  inputClassName="border-white/70 bg-white/75 text-slate-700 placeholder:text-slate-400 shadow-[inset_0_1px_0_rgba(255,255,255,0.7)] focus:border-sky-400 focus:ring-sky-300/70"
                  buttonClassName="border-white/70 bg-slate-950 text-white hover:bg-slate-800"
                />
              </div>
            </div>

            <div className="flex items-center gap-2 sm:gap-3">
              <Link
                href="/trending"
                className="hidden rounded-full border border-white/70 bg-white/70 px-4 py-2 text-sm font-medium text-slate-700 shadow-[inset_0_1px_0_rgba(255,255,255,0.7)] transition hover:-translate-y-0.5 hover:bg-white sm:inline-flex"
              >
                Trending
              </Link>

              {isLoading ? (
                <div className="h-11 w-11 animate-pulse rounded-full bg-slate-200/80" />
              ) : user ? (
                <div className="relative">
                  <button
                    ref={buttonRef}
                    type="button"
                    onClick={() => setMenuOpen((open) => !open)}
                    aria-expanded={menuOpen}
                    aria-haspopup="menu"
                    aria-label="Open account menu"
                    className={cn(
                      "group relative flex h-12 w-12 items-center justify-center rounded-full border border-white/80 bg-white/80 shadow-[0_18px_42px_rgba(15,23,42,0.16),inset_0_1px_0_rgba(255,255,255,0.72)] transition duration-300 hover:-translate-y-0.5 hover:shadow-[0_24px_50px_rgba(15,23,42,0.2),inset_0_1px_0_rgba(255,255,255,0.78)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-400/70",
                      (menuOpen || isSessionExpiring) && "ring-2 ring-sky-300/60",
                      isSessionExpiring && "shadow-[0_0_0_7px_rgba(251,191,36,0.18),0_24px_52px_rgba(245,158,11,0.18)]"
                    )}
                  >
                    <span className="pointer-events-none absolute inset-0 rounded-full bg-[linear-gradient(135deg,rgba(56,189,248,0.22),transparent_48%,rgba(59,130,246,0.12))]" />
                    <AvatarBadge
                      username={user.username}
                      avatarUrl={user.avatar_url}
                      sizeClassName="h-10 w-10"
                      textClassName="text-sm"
                    />
                    <span
                      className={cn(
                        "absolute -bottom-0.5 -right-0.5 flex h-4 w-4 items-center justify-center rounded-full border border-white/80 bg-emerald-400 text-[9px] text-white shadow-sm transition",
                        isSessionExpiring && "bg-amber-400 animate-pulse"
                      )}
                    >
                      <Radio className="h-2.5 w-2.5" />
                    </span>
                  </button>

                  {menuOpen && (
                    <div
                      ref={menuRef}
                      role="menu"
                      className="absolute right-0 mt-3 w-72 overflow-hidden rounded-[28px] border border-white/75 bg-white/82 p-2 shadow-[0_30px_80px_rgba(15,23,42,0.22)] backdrop-blur-2xl animate-dropdown-in"
                    >
                      <div className="relative overflow-hidden rounded-[22px] border border-white/70 bg-slate-950 px-4 py-4 text-white">
                        <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(56,189,248,0.3),_transparent_35%),linear-gradient(135deg,_rgba(255,255,255,0.18),transparent_55%)]" />
                        <div className="relative flex items-center gap-3">
                          <AvatarBadge
                            username={user.username}
                            avatarUrl={user.avatar_url}
                            sizeClassName="h-14 w-14 border-white/25"
                            textClassName="text-lg"
                          />
                          <div className="min-w-0">
                            <p className="font-[family:var(--font-display)] text-lg font-semibold tracking-tight">
                              {user.username}
                            </p>
                            <p className="truncate text-sm text-white/75">{user.email}</p>
                          </div>
                        </div>
                        {isSessionExpiring && countdown && (
                          <div className="relative mt-4 flex items-center gap-2 rounded-full bg-white/10 px-3 py-2 text-xs font-medium text-white/90">
                            <BellRing className="h-3.5 w-3.5 text-amber-300" />
                            Session ends in {countdown}
                          </div>
                        )}
                      </div>

                      <div className="mt-2 space-y-1">
                        <Link
                          href="/profile"
                          className="flex items-center gap-3 rounded-2xl px-3 py-3 text-sm font-medium text-slate-700 transition hover:bg-slate-100/80"
                          onClick={() => setMenuOpen(false)}
                        >
                          <UserRound className="h-4 w-4 text-sky-600" />
                          Profile
                        </Link>
                        <Link
                          href="/studio"
                          className="flex items-center gap-3 rounded-2xl px-3 py-3 text-sm font-medium text-slate-700 transition hover:bg-slate-100/80"
                          onClick={() => setMenuOpen(false)}
                        >
                          <Sparkles className="h-4 w-4 text-indigo-600" />
                          My Studio
                        </Link>
                        <button
                          type="button"
                          onClick={() => {
                            void logout();
                            setMenuOpen(false);
                          }}
                          className="flex w-full items-center gap-3 rounded-2xl px-3 py-3 text-left text-sm font-medium text-slate-700 transition hover:bg-rose-50 hover:text-rose-700"
                        >
                          <LogOut className="h-4 w-4 text-rose-500" />
                          Sign Out
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <Link
                    href="/login"
                    className="rounded-full border border-white/80 bg-white/80 px-4 py-2 text-sm font-medium text-slate-700 shadow-[inset_0_1px_0_rgba(255,255,255,0.74)] transition hover:-translate-y-0.5 hover:bg-white"
                  >
                    Sign In
                  </Link>
                  <Link
                    href="/register"
                    className="rounded-full bg-slate-950 px-4 py-2 text-sm font-medium text-white shadow-[0_16px_32px_rgba(15,23,42,0.18)] transition hover:-translate-y-0.5 hover:bg-slate-800"
                  >
                    Sign Up
                  </Link>
                </div>
              )}
            </div>
          </div>

          <div className="border-t border-white/60 px-4 py-3 xl:hidden">
            <SearchBar
              className="max-w-none"
              inputClassName="border-white/70 bg-white/78 text-slate-700 placeholder:text-slate-400 focus:border-sky-400 focus:ring-sky-300/70"
              buttonClassName="border-white/70 bg-slate-950 text-white hover:bg-slate-800"
            />
          </div>
        </div>
      </div>
    </nav>
  );
}
