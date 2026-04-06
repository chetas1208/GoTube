"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";

interface SearchBarProps {
  className?: string;
  inputClassName?: string;
  buttonClassName?: string;
}

export function SearchBar({ className, inputClassName, buttonClassName }: SearchBarProps) {
  const [query, setQuery] = useState("");
  const router = useRouter();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (query.trim()) {
      router.push(`/search?q=${encodeURIComponent(query.trim())}`);
    }
  };

  return (
    <form onSubmit={handleSubmit} className={cn("flex w-full max-w-md", className)}>
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search videos..."
        className={cn(
          "flex-1 rounded-l-lg border border-gray-300 px-3 py-1.5 text-sm outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500",
          inputClassName
        )}
      />
      <button
        type="submit"
        className={cn(
          "rounded-r-lg border border-l-0 border-gray-300 bg-gray-50 px-4 py-1.5 text-sm text-gray-600 hover:bg-gray-100",
          buttonClassName
        )}
      >
        Search
      </button>
    </form>
  );
}
