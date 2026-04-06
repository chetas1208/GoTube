"use client";

import { useAuth } from "@/lib/auth";
import { UploadForm } from "@/components/video/UploadForm";
import { redirect } from "next/navigation";

export default function UploadPage() {
  const { user, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-brand-600 border-t-transparent" />
      </div>
    );
  }

  if (!user) {
    redirect("/login");
  }

  return (
    <div>
      <h1 className="mb-8 text-2xl font-bold">Upload Video</h1>
      <UploadForm />
    </div>
  );
}
