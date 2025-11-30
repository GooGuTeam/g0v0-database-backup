package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	InitializeConfig()
	InitializeTracker()
	InitializeJobs()

	mux := http.NewServeMux()
	mux.HandleFunc("/full", HandleFullBackup)
	mux.HandleFunc("/incremental", HandleIncrementalBackup)
	mux.HandleFunc("/download", HandleDownloadBackup)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	address := ":" + strconv.Itoa(HttpPort)
	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	go func() {
		log.Println("HTTP server starting at " + address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown requested...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	} else {
		log.Println("HTTP server shutdown complete")
	}

	StopJobs()
	tracker.Close()
	log.Println("Application stopped")
}
