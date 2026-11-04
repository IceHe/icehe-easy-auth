package main

import (
	"strings"
	"testing"
	"time"
)

func TestAdminPageExtendWeekButton(t *testing.T) {
	page := renderAdminHTML([]User{{ID: 42, Name: "test", ExpiresAt: "2027-01-01T00:00:00Z", Permissions: "view", Domains: "*"}}, User{Name: "admin"})
	if !strings.Contains(page, `action="/admin/users/42/extend-week"`) || !strings.Contains(page, `>延期一周</button>`) {
		t.Fatal("admin page is missing the user's extend-week button")
	}
}

func TestNormalizeAdminFormExpiresAtAcceptsDateTimeLocal(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("time.LoadLocation returned error: %v", err)
	}
	previousLocal := time.Local
	time.Local = location
	t.Cleanup(func() {
		time.Local = previousLocal
	})

	got, err := normalizeAdminFormExpiresAt("2026-04-15T08:30:45")
	if err != nil {
		t.Fatalf("normalizeAdminFormExpiresAt returned error: %v", err)
	}

	want := time.Date(2026, 4, 15, 0, 30, 45, 0, time.UTC).Format(time.RFC3339Nano)
	if got != want {
		t.Fatalf("normalizeAdminFormExpiresAt() = %q, want %q", got, want)
	}
}

func TestNormalizeAdminFormExpiresAtAcceptsISO(t *testing.T) {
	input := generateQuickExpiresAt()

	got, err := normalizeAdminFormExpiresAt(input)
	if err != nil {
		t.Fatalf("normalizeAdminFormExpiresAt returned error: %v", err)
	}

	if _, err := parseISO(got); err != nil {
		t.Fatalf("normalized value %q is not valid ISO: %v", got, err)
	}

	inputParsed, err := parseISO(input)
	if err != nil {
		t.Fatalf("parseISO(input) returned error: %v", err)
	}
	gotParsed, err := parseISO(got)
	if err != nil {
		t.Fatalf("parseISO(got) returned error: %v", err)
	}
	if !gotParsed.Equal(inputParsed) {
		t.Fatalf("normalizeAdminFormExpiresAt() changed instant: got %q, want %q", got, input)
	}
}
