package shared

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// RunWithGracefulShutdown menjalankan server HTTP dan menangani sinyal matikan (SIGINT/SIGTERM)
// parameter 'cleanup' adalah fungsi yang akan dijalankan saat shutdown (misal: tutup DB)
func RunWithGracefulShutdown(srv *http.Server, timeout time.Duration, cleanup func() error) error {
	// Channel untuk menangkap error saat startup server
	serverErrors := make(chan error, 1)

	// Jalankan server di goroutine
	go func() {
		log.Printf("Server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Channel untuk menangkap sinyal OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking: Tunggu sampai ada error server ATAU sinyal shutdown
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server startup error: %w", err)

	case <-shutdown:
		log.Println("Shutdown signal received, stopping server...")

		// Buat context timeout untuk shutdown
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// 1. Coba hentikan terima request baru (Graceful Shutdown HTTP)
		if err := srv.Shutdown(ctx); err != nil {
			// Jika gagal graceful, paksa close
			srv.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}

		// 2. Jalankan fungsi cleanup (misal: tutup DB)
		if cleanup != nil {
			log.Println("Cleaning up resources (DB, Redis, etc)...")
			if err := cleanup(); err != nil {
				return fmt.Errorf("cleanup failed: %w", err)
			}
		}

		log.Println("Server stopped gracefully")
		return nil
	}
}
