"use client";

import { useState, useRef } from "react";
import { useRouter } from "next/navigation";
import { videos as videosApi, uploadToPresignedURL } from "@/lib/api-client";
import { useToast } from "@/components/ui/Toast";

export function UploadForm() {
  const router = useRouter();
  const { addToast } = useToast();
  const fileRef = useRef<HTMLInputElement>(null);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [tags, setTags] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [progress, setProgress] = useState<number | null>(null);
  const [status, setStatus] = useState<"idle" | "uploading" | "processing" | "done" | "error">("idle");
  const [errorMsg, setErrorMsg] = useState("");

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selected = e.target.files?.[0];
    if (selected) {
      setFile(selected);
      if (!title) setTitle(selected.name.replace(/\.[^.]+$/, ""));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) return;

    setStatus("uploading");
    setErrorMsg("");

    try {
      const tagList = tags
        .split(",")
        .map((t) => t.trim())
        .filter(Boolean);

      const initResponse = await videosApi.initiateUpload({
        title,
        description,
        tags: tagList,
        filename: file.name,
        content_type: file.type,
        file_size: file.size,
      });

      await uploadToPresignedURL(initResponse.upload_url, file, setProgress);

      setStatus("processing");
      await videosApi.completeUpload(initResponse.video_id);

      setStatus("done");
      addToast("Upload complete! Video is being processed.", "success");
      router.push(`/studio?video=${initResponse.video_id}`);
    } catch (err) {
      setStatus("error");
      setErrorMsg(err instanceof Error ? err.message : "Upload failed");
      addToast("Upload failed. Please try again.", "error");
    }
  };

  return (
    <form onSubmit={handleSubmit} className="mx-auto max-w-2xl space-y-6">
      <div>
        <label className="block text-sm font-medium text-gray-700">Video File</label>
        <div
          onClick={() => fileRef.current?.click()}
          className="mt-1 flex cursor-pointer items-center justify-center rounded-lg border-2 border-dashed border-gray-300 p-8 transition hover:border-brand-500"
        >
          <input
            ref={fileRef}
            type="file"
            accept="video/*"
            onChange={handleFileChange}
            className="hidden"
          />
          {file ? (
            <div className="text-center">
              <p className="font-medium text-gray-900">{file.name}</p>
              <p className="mt-1 text-sm text-gray-500">
                {(file.size / (1024 * 1024)).toFixed(1)} MB
              </p>
            </div>
          ) : (
            <div className="text-center">
              <svg className="mx-auto h-10 w-10 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              <p className="mt-2 text-sm text-gray-600">Click to select a video file</p>
              <p className="mt-1 text-xs text-gray-400">MP4, MOV, WebM, AVI, MKV up to 500MB</p>
            </div>
          )}
        </div>
      </div>

      <div>
        <label htmlFor="title" className="block text-sm font-medium text-gray-700">Title</label>
        <input
          id="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
          maxLength={255}
          className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500"
        />
      </div>

      <div>
        <label htmlFor="description" className="block text-sm font-medium text-gray-700">Description</label>
        <textarea
          id="description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={4}
          maxLength={5000}
          className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500"
        />
      </div>

      <div>
        <label htmlFor="tags" className="block text-sm font-medium text-gray-700">Tags</label>
        <input
          id="tags"
          value={tags}
          onChange={(e) => setTags(e.target.value)}
          placeholder="tutorial, go, programming"
          className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500"
        />
        <p className="mt-1 text-xs text-gray-400">Comma-separated</p>
      </div>

      {progress !== null && status === "uploading" && (
        <div>
          <div className="h-2 overflow-hidden rounded-full bg-gray-200">
            <div
              className="h-full rounded-full bg-brand-600 transition-all"
              style={{ width: `${progress}%` }}
            />
          </div>
          <p className="mt-1 text-sm text-gray-500">Uploading... {progress}%</p>
        </div>
      )}

      {status === "processing" && (
        <div className="flex items-center gap-2 text-sm text-purple-600">
          <div className="h-4 w-4 animate-spin rounded-full border-2 border-purple-600 border-t-transparent" />
          Finalizing upload and starting processing...
        </div>
      )}

      {status === "error" && (
        <p className="text-sm text-red-600">{errorMsg}</p>
      )}

      <button
        type="submit"
        disabled={!file || !title || status === "uploading" || status === "processing"}
        className="w-full rounded-lg bg-brand-600 px-4 py-2.5 text-sm font-medium text-white hover:bg-brand-700 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {status === "uploading" ? "Uploading..." : status === "processing" ? "Processing..." : "Upload Video"}
      </button>
    </form>
  );
}
