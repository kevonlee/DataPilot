package main

import (
	"database-manager/internal/config"
	"database-manager/internal/handler"
	"database-manager/internal/service"
	"embed"
	"fmt"
	"io/fs"
	"log"
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
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// get embedded web files
	webRoot, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		log.Fatalf("Failed to access web files: %v", err)
	}

	// create router
	mux := handler.NewRouter(webRoot, cfg)

	// get port
	port := cfg.Port
	if p := os.Getenv("DB_MANAGER_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Database Manager started at http://localhost%s\n", addr)
	fmt.Println("Default login: admin / admin")

	// open browser
	go openBrowser(fmt.Sprintf("http://localhost%s", addr))

	// cleanup on exit
	defer service.GetDBManager().CloseAll()

	// start server
	log.Fatal(http.ListenAndServe(addr, mux))
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
