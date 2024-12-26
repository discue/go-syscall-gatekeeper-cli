package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from Go server")
}

func main() {
	server := &http.Server{Addr: ":8081", Handler: nil} // Create server instance
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Run server in a separate goroutine
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Println("Server error:", err) // Handle unexpected server errors
		}
	}()

	time.Sleep(5 * time.Second)
	server.Shutdown(ctx)

	// sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// <-sigChan // Wait for a signal
	<-ctx.Done() // Wait for the context to be canceled

	fmt.Println("Shutting down server...")

	fmt.Println("Server stopped")

}
