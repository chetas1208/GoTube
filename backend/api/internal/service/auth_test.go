package service

import (
	"testing"
	"time"
)

func TestIsRefreshTokenIdleExpired(t *testing.T) {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)

	if isRefreshTokenIdleExpired(now.Add(-10*time.Minute), now, 15*time.Minute) {
		t.Fatalf("expected token used 10 minutes ago to stay active")
	}
	if !isRefreshTokenIdleExpired(now.Add(-16*time.Minute), now, 15*time.Minute) {
		t.Fatalf("expected token used 16 minutes ago to be idle expired")
	}
}
