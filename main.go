package main

import (
	"database-manager/internal/config"
	"database-manager/internal/handler"
	"database-manager/internal/logger"
	"database-manager/internal/service"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

//go:embed web/dist/*
var webFS embed.FS

func main() {
	// initialize config
	dataDir := "data"
	if dir := os.Getenv("DB_MANAGER_DATA"); dir != "" {
		dataDir = dir
	}

	cfg, err := config.Init(dataDir)
	if err != nil {
		logger.Error("Failed to initialize config: %v", err)
		os.Exit(1)
	}

	// get embedded web files
	webRoot, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		logger.Error("Failed to access web files: %v", err)
		os.Exit(1)
	}

	// create router
	mux := handler.NewRouter(webRoot, cfg)

	// get port
	port := cfg.Port
	if p := os.Getenv("DB_MANAGER_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	addr := fmt.Sprintf(":%d", port)
	logger.Info("Database Manager started at http://localhost%s", addr)
	logger.Info("Default login: admin / admin")

	// open browser
	go openBrowser(fmt.Sprintf("http://localhost%s", addr))

	// cleanup on exit
	defer service.GetDBManager().CloseAll()

	// start server
	logger.Info("HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("HTTP server error: %v", err)
		os.Exit(1)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}
