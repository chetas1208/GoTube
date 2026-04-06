import type { VideoStatus } from "@/types";
import { cn } from "@/lib/utils";

const statusConfig: Record<VideoStatus, { label: string; className: string }> = {
  uploaded: { label: "Uploaded", className: "bg-yellow-100 text-yellow-800" },
  queued: { label: "Queued", className: "bg-blue-100 text-blue-800" },
  processing: { label: "Processing", className: "bg-purple-100 text-purple-800" },
  ready: { label: "Ready", className: "bg-green-100 text-green-800" },
  failed: { label: "Failed", className: "bg-red-100 text-red-800" },
};

export function StatusBadge({ status }: { status: VideoStatus }) {
  const config = statusConfig[status] || { label: status, className: "bg-gray-100 text-gray-800" };
  return (
    <span className={cn("inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium", config.className)}>
      {config.label}
    </span>
  );
}
