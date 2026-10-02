package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"eskhan/internal/api"
	"eskhan/internal/config"
	"eskhan/internal/storage"
	"eskhan/web"
)

var (
	version = "1.0.0"
)

func main() {
	portFlag := flag.Int("port", 8989, "HTTP port to listen on")
	hostFlag := flag.String("host", "localhost", "Host interface to bind")
	noBrowserFlag := flag.Bool("no-browser", false, "Do not auto-open the default browser on launch")
	configFlag := flag.String("config", "", "Custom path to config.json")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("ESKhan v%s (Elasticsearch Query IDE)\n", version)
		os.Exit(0)
	}

	// 1. Initialize configuration manager
	cfgMgr, err := config.NewManager(*configFlag)
	if err != nil {
		log.Fatalf("❌ Failed to initialize configuration: %v", err)
	}

	// 2. Initialize history & snippet store
	store, err := storage.NewStore("", cfgMgr.GetConfig().MaxHistoryItems)
	if err != nil {
		log.Fatalf("❌ Failed to initialize storage: %v", err)
	}

	// 3. Mount embedded static web assets
	staticFS, err := web.GetFS()
	if err != nil {
		log.Fatalf("❌ Failed to load embedded assets: %v", err)
	}

	// 4. Initialize REST API server
	server, err := api.NewServer(cfgMgr, store, staticFS)
	if err != nil {
		log.Fatalf("❌ Failed to initialize server: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", *hostFlag, *portFlag)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server.Router(),
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	targetURL := fmt.Sprintf("http://localhost:%d", *portFlag)

	// Graceful shutdown channel
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Start listener
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("❌ Failed to bind to %s: %v", addr, err)
	}

	go func() {
		log.Printf("═══════════════════════════════════════════════════════")
		log.Printf(" ⚡ ESKhan - Elasticsearch Query IDE v%s", version)
		log.Printf(" 🚀 Running at: %s", targetURL)
		log.Printf(" ⌨️  Press Ctrl+C to stop")
		log.Printf("═══════════════════════════════════════════════════════")

		if !*noBrowserFlag {
			go openBrowser(targetURL)
		}

		if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stopChan
	log.Println("\n🛑 Shutting down ESKhan gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("👋 ESKhan stopped. Goodbye!")
}

// openBrowser opens the specified URL in the user's default browser.
func openBrowser(url string) {
	time.Sleep(350 * time.Millisecond) // Give server a moment to start
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}

	if cmd != nil {
		_ = cmd.Start()
	}
}
