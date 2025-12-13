// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-sis-be/internal/configs"
	"go-sis-be/routes"
)

func main() {
	configs.ConnectDB()
	configs.SeedDatabase()
	r := routes.InitRouter()

	host := "localhost"
	if envHost := os.Getenv("HOST"); envHost != "" {
		host = envHost
	}

	port := ":8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Print("\033[H\033[2J")
		fmt.Println("=================================================")
		fmt.Println("STATUS   : SERVICE RUNNING")
		fmt.Printf("PORT     : %s\n", port)
		fmt.Printf("HOST     : %s\n", host)
		fmt.Println("=================================================")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("\nServer is shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown: %v", err)
	}
	configs.CloseDB()
}
