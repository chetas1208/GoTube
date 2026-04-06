package handler

import (
	"errors"
	"net/http"

	"github.com/chetasparekh/gotube-lite/api/internal/dto"
	"github.com/chetasparekh/gotube-lite/api/internal/middleware"
	"github.com/chetasparekh/gotube-lite/api/internal/repository"
	"github.com/chetasparekh/gotube-lite/api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type VideoHandler struct {
	videos *service.VideoService
}

func NewVideoHandler(videos *service.VideoService) *VideoHandler {
	return &VideoHandler{videos: videos}
}

func (h *VideoHandler) InitiateUpload(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var req dto.InitiateUploadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		writeValidationErrors(w, errs)
		return
	}

	resp, err := h.videos.InitiateUpload(r.Context(), userID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *VideoHandler) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid video id")
		return
	}

	if err := h.videos.CompleteUpload(r.Context(), userID, videoID); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			writeError(w, http.StatusForbidden, "not your video")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "upload completed, processing started"})
}

func (h *VideoHandler) Get(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid video id")
		return
	}

	var viewerID *uuid.UUID
	if uid, ok := middleware.GetUserID(r.Context()); ok {
		viewerID = &uid
	}

	resp, err := h.videos.GetVideo(r.Context(), videoID, viewerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "video not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get video")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *VideoHandler) ListRecent(w http.ResponseWriter, r *http.Request) {
	p := dto.PaginationParams{
		Page:    queryInt(r, "page", 1),
		PerPage: queryInt(r, "per_page", 20),
	}
	p.Normalize()

	var viewerID *uuid.UUID
	if uid, ok := middleware.GetUserID(r.Context()); ok {
		viewerID = &uid
	}

	resp, err := h.videos.ListRecent(r.Context(), p, viewerID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list recent videos")
		writeError(w, http.StatusInternalServerError, "failed to list videos")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *VideoHandler) ListMyVideos(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	p := dto.PaginationParams{
		Page:    queryInt(r, "page", 1),
		PerPage: queryInt(r, "per_page", 20),
	}
	p.Normalize()

	resp, err := h.videos.ListByUser(r.Context(), userID, p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list videos")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *VideoHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid video id")
		return
	}

	var req dto.UpdateVideoRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		writeValidationErrors(w, errs)
		return
	}

	resp, err := h.videos.UpdateVideo(r.Context(), userID, videoID, req)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			writeError(w, http.StatusForbidden, "not your video")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update video")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *VideoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid video id")
		return
	}

	if err := h.videos.DeleteVideo(r.Context(), userID, videoID); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			writeError(w, http.StatusForbidden, "not your video")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete video")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "video deleted"})
}

func (h *VideoHandler) RecordView(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid video id")
		return
	}

	var userID *uuid.UUID
	if uid, ok := middleware.GetUserID(r.Context()); ok {
		userID = &uid
	}

	sessionID := r.URL.Query().Get("session_id")
	var sid *string
	if sessionID != "" {
		sid = &sessionID
	}

	if err := h.videos.RecordView(r.Context(), videoID, userID, sid); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to record view")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "view recorded"})
}

func (h *VideoHandler) Like(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid video id")
		return
	}

	liked, err := h.videos.ToggleLike(r.Context(), videoID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to toggle like")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"liked": liked})
}

func (h *VideoHandler) Unlike(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid video id")
		return
	}

	if err := h.videos.RemoveLike(r.Context(), userID, videoID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove like")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "like removed"})
}

func (h *VideoHandler) Playback(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid video id")
		return
	}

	resp, err := h.videos.GetPlaybackURL(r.Context(), videoID)
	if err != nil {
		if errors.Is(err, service.ErrVideoNotReady) {
			writeError(w, http.StatusNotFound, "video not ready for playback")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get playback URL")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *VideoHandler) Search(w http.ResponseWriter, r *http.Request) {
	params := dto.SearchParams{
		Query:   r.URL.Query().Get("q"),
		Page:    queryInt(r, "page", 1),
		PerPage: queryInt(r, "per_page", 20),
		SortBy:  r.URL.Query().Get("sort_by"),
	}
	params.Normalize()

	if params.Query == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	var viewerID *uuid.UUID
	if uid, ok := middleware.GetUserID(r.Context()); ok {
		viewerID = &uid
	}

	resp, err := h.videos.Search(r.Context(), params, viewerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *VideoHandler) Trending(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	if limit > 50 {
		limit = 50
	}

	var viewerID *uuid.UUID
	if uid, ok := middleware.GetUserID(r.Context()); ok {
		viewerID = &uid
	}

	resp, err := h.videos.Trending(r.Context(), limit, viewerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "trending failed")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
