package web

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
)

//go:embed static/*
var staticFS embed.FS

// handleSPA serves the embedded SPA. Any non-API path that doesn't match
// a static file falls back to index.html (client-side routing).
func (s *Server) handleSPA(w http.ResponseWriter, r *http.Request) {
	// If auth is configured and this is a login POST, handle it.
	if s.authToken != "" && r.Method == "POST" && r.URL.Path == "/login" {
		s.handleLoginSubmit(w, r)
		return
	}

	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// Try to serve from embedded static files.
	data, err := staticFS.ReadFile("static" + path)
	if err == nil {
		serveStatic(w, path, data)
		return
	}

	// Fallback to index.html for client-side routing.
	data, err = staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	serveStatic(w, "index.html", data)
}

func serveStatic(w http.ResponseWriter, path string, data []byte) {
	switch {
	case strings.HasSuffix(path, ".html"):
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case strings.HasSuffix(path, ".svg"):
		w.Header().Set("Content-Type", "image/svg+xml")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Write(data)
}

// serveLogin serves the login page when auth is required.
func (s *Server) serveLogin(w http.ResponseWriter, r *http.Request) {
	loginHTML := `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Pressie — Login</title>
<link rel="stylesheet" href="/style.css">
</head>
<body>
<div class="login-container">
<h1>Pressie</h1>
<p>Enter your access token to continue.</p>
<form method="POST" action="/login">
<input type="password" name="token" placeholder="Access token" autocomplete="current-password" autofocus>
<button type="submit">Sign in</button>
</form>
</div>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(loginHTML))
}

// handleLoginSubmit processes the login form. On success, sets a cookie
// and redirects to /.
func (s *Server) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	token := r.FormValue("token")
	if token == s.authToken {
		http.SetCookie(w, &http.Cookie{
			Name:     "pressie_auth",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   86400 * 30, // 30 days
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Pressie — Login</title><link rel="stylesheet" href="/style.css"></head><body><div class="login-container"><h1>Pressie</h1><p style="color:#c00">Invalid token. Try again.</p><form method="POST" action="/login"><input type="password" name="token" placeholder="Access token" autofocus><button type="submit">Sign in</button></form></div></body></html>`)
}