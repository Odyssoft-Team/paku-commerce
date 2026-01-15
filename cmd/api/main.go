package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "paku-commerce/docs" // ⬅️ CRÍTICO: debe estar presente
	"paku-commerce/pkg/server"
)

// @title           Paku Commerce API
// @version         1.0
// @description     API de comercio para servicios y productos de Paku
// @termsOfService  http://paku.pe/terms/

/// @contact.name   Paku Support
// @contact.url    http://paku.pe/support
// @contact.email  support@paku.pe

// @license.name  Proprietary
// @license.url   http://paku.pe/license

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey UserID
// @in header
// @name X-User-ID
// @description User ID header for authentication (dev only)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           server.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server stopped")
}
