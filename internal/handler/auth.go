package handler

import (
	"database-manager/internal/config"
	"database-manager/internal/logger"
	"database-manager/internal/middleware"
	"encoding/json"
	"net/http"
)

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	cfg := config.Get()
	if req.Username != cfg.User.Username {
		logger.Warn("Login failed: unknown user '%s' from %s", req.Username, r.RemoteAddr)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if !cfg.CheckPassword(req.Password) {
		logger.Warn("Login failed: wrong password for user '%s' from %s", req.Username, r.RemoteAddr)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	token, err := middleware.GenerateToken(cfg.JWTSecret, req.Username)
	if err != nil {
		logger.Error("Failed to generate token for user '%s': %v", req.Username, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	logger.Info("User '%s' logged in from %s", req.Username, r.RemoteAddr)
	writeJSON(w, http.StatusOK, map[string]string{
		"token":    token,
		"username": req.Username,
	})
}

func handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if len(req.NewPassword) < 4 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 4 characters"})
		return
	}

	cfg := config.Get()
	if !cfg.CheckPassword(req.OldPassword) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid old password"})
		return
	}

	if err := cfg.SetPassword(req.NewPassword); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save password"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password changed"})
}
