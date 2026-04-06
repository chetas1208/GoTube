package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/chetasparekh/gotube-lite/api/internal/dto"
	"github.com/chetasparekh/gotube-lite/api/internal/middleware"
	"github.com/chetasparekh/gotube-lite/api/internal/service"
)

type AuthHandler struct {
	auth       *service.AuthService
	refreshTTL time.Duration
}

func NewAuthHandler(auth *service.AuthService, refreshTTL time.Duration) *AuthHandler {
	return &AuthHandler{auth: auth, refreshTTL: refreshTTL}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		writeValidationErrors(w, errs)
		return
	}

	resp, refreshToken, err := h.auth.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			writeError(w, http.StatusConflict, "user already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "registration failed")
		return
	}

	setRefreshCookie(w, refreshToken, h.refreshTTL)
	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		writeValidationErrors(w, errs)
		return
	}

	resp, refreshToken, err := h.auth.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}

	setRefreshCookie(w, refreshToken, h.refreshTTL)
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	_ = h.auth.Logout(r.Context(), userID)
	clearRefreshCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.auth.GetCurrentUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "no refresh token")
		return
	}

	resp, newRefresh, err := h.auth.RefreshAccessToken(r.Context(), cookie.Value)
	if err != nil {
		clearRefreshCookie(w)
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	setRefreshCookie(w, newRefresh, h.refreshTTL)
	writeJSON(w, http.StatusOK, resp)
}

func setRefreshCookie(w http.ResponseWriter, token string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   false, // Set true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   refreshCookieMaxAgeSeconds(ttl),
	})
}

func refreshCookieMaxAgeSeconds(ttl time.Duration) int {
	if ttl <= 0 {
		return 0
	}
	return int(ttl / time.Second)
}

func clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		MaxAge:   -1,
	})
}
