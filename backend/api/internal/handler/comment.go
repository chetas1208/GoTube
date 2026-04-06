package handler

import (
	"net/http"

	"github.com/chetasparekh/gotube-lite/api/internal/dto"
	"github.com/chetasparekh/gotube-lite/api/internal/middleware"
	"github.com/chetasparekh/gotube-lite/api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CommentHandler struct {
	comments *service.CommentService
}

func NewCommentHandler(comments *service.CommentService) *CommentHandler {
	return &CommentHandler{comments: comments}
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid video id")
		return
	}

	var req dto.CreateCommentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		writeValidationErrors(w, errs)
		return
	}

	resp, err := h.comments.Create(r.Context(), videoID, userID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create comment")
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *CommentHandler) List(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid video id")
		return
	}

	p := dto.PaginationParams{
		Page:    queryInt(r, "page", 1),
		PerPage: queryInt(r, "per_page", 20),
	}
	p.Normalize()

	resp, err := h.comments.ListByVideo(r.Context(), videoID, p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list comments")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
