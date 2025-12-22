package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	server := &http.Server{Addr: ":8082", Handler: nil} // Create server instance
	defer server.Close()

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}

	// Run server in a separate goroutine
	go func() {
		if err := server.Serve(ln); err != http.ErrServerClosed {
			fmt.Println("Server error:", err) // Handle unexpected server errors
			os.Exit(1)
		}
	}()

	fmt.Println("Starting server")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	fmt.Println("Server stopped")
}
