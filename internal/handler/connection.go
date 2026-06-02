package handler

import (
	"database-manager/internal/config"
	"database-manager/internal/model"
	"database-manager/internal/service"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, cfg.GetConnections())
	case http.MethodPost:
		var conn model.Connection
		if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		conn.ID = generateID()
		conn.CreatedAt = time.Now()
		conn.UpdatedAt = time.Now()
		if err := cfg.AddConnection(conn); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, conn)
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func handleConnectionByID(w http.ResponseWriter, r *http.Request) {
	// path: /api/connections/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/connections/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, `{"error":"missing connection id"}`, http.StatusBadRequest)
		return
	}

	id := parts[0]

	// check for test action
	if len(parts) > 1 && parts[1] == "test" {
		handleTestConnection(w, r, id)
		return
	}

	cfg := config.Get()
	switch r.Method {
	case http.MethodGet:
		conn := cfg.GetConnection(id)
		if conn == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "connection not found"})
			return
		}
		writeJSON(w, http.StatusOK, conn)
	case http.MethodPut:
		var conn model.Connection
		if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		conn.ID = id
		conn.UpdatedAt = time.Now()
		if err := cfg.UpdateConnection(conn); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, conn)
	case http.MethodDelete:
		service.GetDBManager().Close(id)
		if err := cfg.DeleteConnection(id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func handleTestConnection(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	cfg := config.Get()
	conn := cfg.GetConnection(id)
	if conn == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "connection not found"})
		return
	}

	err := service.GetDBManager().Test(conn)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "connection successful",
	})
}

func generateID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}
