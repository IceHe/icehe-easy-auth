package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestExtendWeekDatabase(t *testing.T) {
	url := os.Getenv("AUTH_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("AUTH_TEST_DATABASE_URL is not set")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx := context.Background()
	_, err = db.ExecContext(ctx, `CREATE TEMP TABLE users (LIKE public.users INCLUDING DEFAULTS);
		INSERT INTO users (id, name, token, expires_at, permissions, domains, created_at, updated_at, is_admin)
		VALUES (1, 'admin', 'admin-test', '2099-01-01T00:00:00Z', 'manage', '*', '', '', true),
		       (2, 'user', 'user-test', '2020-01-01T00:00:00Z', 'view', '*', '', '', false)`)
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{db: db, cfg: Config{SecretKey: "test-secret"}, tokenCache: NewTokenCache(true, time.Minute, 10)}
	handler := s.routes()
	request := func(cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/admin/users/2/extend-week", nil)
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	if w := request(nil); w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/admin/login" {
		t.Fatalf("unauthenticated response: %d %s", w.Code, w.Header().Get("Location"))
	}
	w := httptest.NewRecorder()
	s.setSessionCookie(w, 1)
	cookie := w.Result().Cookies()[0]
	for i := 0; i < 2; i++ {
		if w := request(cookie); w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/admin" {
			t.Fatalf("extension response: %d %s", w.Code, w.Body.String())
		}
	}
	var expiry string
	if err := db.QueryRowContext(ctx, `SELECT expires_at FROM users WHERE id = 2`).Scan(&expiry); err != nil {
		t.Fatal(err)
	}
	if expiry != "2020-01-15T00:00:00Z" {
		t.Fatalf("expiry = %q, want 2020-01-15T00:00:00Z", expiry)
	}
}
