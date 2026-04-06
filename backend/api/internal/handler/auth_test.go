package handler

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestSetRefreshCookieUsesConfiguredTTL(t *testing.T) {
	rec := httptest.NewRecorder()

	setRefreshCookie(rec, "refresh-token", 36*time.Hour)

	res := rec.Result()
	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected exactly one cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "refresh_token" {
		t.Fatalf("unexpected cookie name: %s", cookies[0].Name)
	}
	if cookies[0].MaxAge != int((36*time.Hour)/time.Second) {
		t.Fatalf("unexpected cookie max age: got %d", cookies[0].MaxAge)
	}
}
