import type {
  AuthResponse,
  VideoListResponse,
  Video,
  CommentListResponse,
  InitiateUploadResponse,
  PlaybackResponse,
  User,
  Comment,
} from "@/types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
const VIEWER_SESSION_STORAGE_KEY = "gotube:viewer-session-id";

let accessToken: string | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
}

export function getAccessToken(): string | null {
  return accessToken;
}

function getViewerSessionId(): string | null {
  if (typeof window === "undefined") {
    return null;
  }

  let sessionId = window.localStorage.getItem(VIEWER_SESSION_STORAGE_KEY);
  if (sessionId) {
    return sessionId;
  }

  sessionId = window.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`;
  window.localStorage.setItem(VIEWER_SESSION_STORAGE_KEY, sessionId);
  return sessionId;
}

async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`;
  }

  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiClientError(res.status, body.error || "Request failed", body.errors);
  }

  return res.json();
}

export class ApiClientError extends Error {
  constructor(
    public status: number,
    message: string,
    public fieldErrors?: Record<string, string>
  ) {
    super(message);
    this.name = "ApiClientError";
  }
}

// Auth
export const auth = {
  register: (data: { username: string; email: string; password: string }) =>
    apiFetch<AuthResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  login: (data: { email: string; password: string }) =>
    apiFetch<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  logout: () =>
    apiFetch<{ message: string }>("/auth/logout", { method: "POST" }),

  me: () => apiFetch<User>("/auth/me"),

  refresh: () => apiFetch<AuthResponse>("/auth/refresh", { method: "POST" }),
};

// Videos
export const videos = {
  list: (page = 1, perPage = 20) =>
    apiFetch<VideoListResponse>(`/videos?page=${page}&per_page=${perPage}`),

  get: (id: string) => apiFetch<Video>(`/videos/${id}`),

  my: (page = 1, perPage = 20) =>
    apiFetch<VideoListResponse>(`/videos/my?page=${page}&per_page=${perPage}`),

  initiateUpload: (data: {
    title: string;
    description: string;
    tags: string[];
    filename: string;
    content_type: string;
    file_size: number;
  }) =>
    apiFetch<InitiateUploadResponse>("/videos/initiate-upload", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  completeUpload: (videoId: string) =>
    apiFetch<{ message: string }>(`/videos/${videoId}/complete-upload`, {
      method: "POST",
    }),

  update: (id: string, data: { title?: string; description?: string; tags?: string[]; visibility?: string }) =>
    apiFetch<Video>(`/videos/${id}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    apiFetch<{ message: string }>(`/videos/${id}`, { method: "DELETE" }),

  playback: (id: string) => apiFetch<PlaybackResponse>(`/videos/${id}/playback`),

  recordView: (id: string) => {
    const sessionId = getViewerSessionId();
    const query = sessionId ? `?session_id=${encodeURIComponent(sessionId)}` : "";
    return apiFetch<{ message: string }>(`/videos/${id}/view${query}`, { method: "POST" });
  },

  like: (id: string) =>
    apiFetch<{ liked: boolean }>(`/videos/${id}/like`, { method: "POST" }),

  unlike: (id: string) =>
    apiFetch<{ message: string }>(`/videos/${id}/like`, { method: "DELETE" }),
};

// Comments
export const comments = {
  list: (videoId: string, page = 1, perPage = 20) =>
    apiFetch<CommentListResponse>(`/videos/${videoId}/comments?page=${page}&per_page=${perPage}`),

  create: (videoId: string, data: { body: string; parent_id?: string }) =>
    apiFetch<Comment>(`/videos/${videoId}/comments`, {
      method: "POST",
      body: JSON.stringify(data),
    }),
};

// Search & Trending
export const search = {
  query: (q: string, page = 1, perPage = 20, sortBy = "relevance") =>
    apiFetch<VideoListResponse>(`/search?q=${encodeURIComponent(q)}&page=${page}&per_page=${perPage}&sort_by=${sortBy}`),
};

export const trending = {
  get: (limit = 20) => apiFetch<Video[]>(`/trending?limit=${limit}`),
};

// Upload file directly to a presigned object storage URL.
export async function uploadToPresignedURL(
  url: string,
  file: File,
  onProgress?: (percent: number) => void
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("PUT", url);
    xhr.setRequestHeader("Content-Type", file.type);

    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable && onProgress) {
        onProgress(Math.round((e.loaded / e.total) * 100));
      }
    };

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve();
      } else {
        reject(new Error(`Upload failed with status ${xhr.status}`));
      }
    };

    xhr.onerror = () => reject(
      new Error("Upload failed before the file reached storage. Check the browser origin and object storage CORS settings.")
    );
    xhr.send(file);
  });
}
