package handler

import (
	"database-manager/internal/config"
	"database-manager/internal/middleware"
	"io/fs"
	"net/http"
	"strings"
)

func NewRouter(webFS fs.FS, cfg *config.Config) *http.ServeMux {
	mux := http.NewServeMux()
	auth := middleware.AuthMiddleware(cfg.JWTSecret)

	// auth routes (no auth required)
	mux.HandleFunc("/api/auth/login", handleLogin)

	// routes that require auth - wrap with auth middleware
	mux.Handle("/api/auth/change-password", auth(http.HandlerFunc(handleChangePassword)))
	mux.Handle("/api/connections", auth(http.HandlerFunc(handleConnections)))
	mux.Handle("/api/connections/", auth(http.HandlerFunc(handleConnectionByID)))
	mux.Handle("/api/conn/", auth(http.HandlerFunc(handleDBRoutes)))

	// SPA fallback handler
	fileServer := http.FileServer(http.FS(webFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If it's an API route, should not reach here
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve static file
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if file exists in the embedded filesystem
		if _, err := fs.Stat(webFS, path); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// For SPA routes, serve index.html
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	return mux
}
