package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/btouchard/ackify-ce/internal/presentation/admin"
	"github.com/btouchard/ackify-ce/pkg/web"
)

func main() {
	ctx := context.Background()

	server, err := web.NewServer(ctx)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	server.RegisterRoutes(admin.RegisterAdminRoutes(server.GetBaseURL(), server.GetTemplates(), server.GetDB()))

	go func() {
		log.Printf("Community Edition server starting on %s", server.GetAddr())
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Community Edition server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Community Edition server exited")
}
