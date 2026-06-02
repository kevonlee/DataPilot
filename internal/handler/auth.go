package handler

import (
	"database-manager/internal/config"
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
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if !cfg.CheckPassword(req.Password) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	token, err := middleware.GenerateToken(cfg.JWTSecret, req.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

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
