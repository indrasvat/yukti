package views

import (
	"testing"
	"time"
)

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		contains string // Expected substring in result
	}{
		{
			name:     "zero time",
			time:     time.Time{},
			contains: "unknown",
		},
		{
			name:     "just now",
			time:     now.Add(-30 * time.Second),
			contains: "just now",
		},
		{
			name:     "1 minute ago",
			time:     now.Add(-1 * time.Minute),
			contains: "1 minute ago",
		},
		{
			name:     "5 minutes ago",
			time:     now.Add(-5 * time.Minute),
			contains: "5 minutes ago",
		},
		{
			name:     "1 hour ago",
			time:     now.Add(-1 * time.Hour),
			contains: "1 hour ago",
		},
		{
			name:     "3 hours ago",
			time:     now.Add(-3 * time.Hour),
			contains: "3 hours ago",
		},
		{
			name:     "yesterday",
			time:     now.Add(-24 * time.Hour),
			contains: "yesterday",
		},
		{
			name:     "3 days ago",
			time:     now.Add(-3 * 24 * time.Hour),
			contains: "3 days ago",
		},
		{
			name:     "1 week ago",
			time:     now.Add(-7 * 24 * time.Hour),
			contains: "1 week ago",
		},
		{
			name:     "2 weeks ago",
			time:     now.Add(-14 * 24 * time.Hour),
			contains: "2 weeks ago",
		},
		{
			name:     "1 month ago",
			time:     now.Add(-35 * 24 * time.Hour),
			contains: "1 month ago",
		},
		{
			name:     "3 months ago",
			time:     now.Add(-100 * 24 * time.Hour),
			contains: "3 months ago",
		},
		{
			name:     "1 year ago",
			time:     now.Add(-400 * 24 * time.Hour),
			contains: "1 year ago",
		},
		{
			name:     "2 years ago",
			time:     now.Add(-800 * 24 * time.Hour),
			contains: "2 years ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimeAgo(tt.time)

			if !containsStr(result, tt.contains) {
				t.Errorf("formatTimeAgo() = %q, want to contain %q", result, tt.contains)
			}
		})
	}
}

func TestFormatTimeAgo_EdgeCases(t *testing.T) {
	now := time.Now()

	// Test boundary between units
	tests := []struct {
		name       string
		duration   time.Duration
		shouldNot  string // Should NOT contain this
		shouldHave string // Should contain this
	}{
		{
			name:       "59 seconds is just now",
			duration:   59 * time.Second,
			shouldHave: "just now",
		},
		{
			name:       "61 seconds is minutes",
			duration:   61 * time.Second,
			shouldHave: "minute",
		},
		{
			name:       "59 minutes is minutes",
			duration:   59 * time.Minute,
			shouldHave: "minutes",
		},
		{
			name:       "61 minutes is hours",
			duration:   61 * time.Minute,
			shouldHave: "hour",
		},
		{
			name:       "23 hours is hours",
			duration:   23 * time.Hour,
			shouldHave: "hours",
		},
		{
			name:       "25 hours is yesterday",
			duration:   25 * time.Hour,
			shouldHave: "yesterday",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimeAgo(now.Add(-tt.duration))

			if tt.shouldHave != "" && !containsStr(result, tt.shouldHave) {
				t.Errorf("formatTimeAgo(-%v) = %q, want to contain %q",
					tt.duration, result, tt.shouldHave)
			}
			if tt.shouldNot != "" && containsStr(result, tt.shouldNot) {
				t.Errorf("formatTimeAgo(-%v) = %q, should NOT contain %q",
					tt.duration, result, tt.shouldNot)
			}
		})
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
