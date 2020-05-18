package main

import (
	"context"
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bugimetal/facedetection/handler"
	"github.com/bugimetal/facedetection/service"

	"github.com/sirupsen/logrus"
)

var (
	bind = flag.String("bind", ":8080", "The socket to bind the HTTP server")
)

func main() {
	flag.Parse()

	// Service covers the high-level business logic.
	services, err := service.New()
	if err != nil {
		fmt.Errorf("can't initialize services %v", err)
		return
	}

	// Handler handles the incoming requests using services to perform business logic operations
	h := handler.New(handler.Services{
		FaceDetection: services.FaceDetection,
		ImageFetcher:  services.ImageFetcher,
	})

	httpServer := &http.Server{
		Addr:    *bind,
		Handler: h,
	}

	// Start the HTTP server.
	httpServerErrorChan := make(chan error)
	go func() {
		fmt.Printf("HTTP server listening on %s\n", *bind)
		httpServerErrorChan <- httpServer.ListenAndServe()
	}()

	// Set up the signal channel.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	// If the HTTP server returned an error, exit here.
	case err := <-httpServerErrorChan:
		logrus.Printf("HTTP server error: %s", err)
	// If a termination signal was received, shutdown the server.
	case sig := <-signalChan:
		logrus.Printf("Signal received: %s", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logrus.Fatalf("HTTP Server graceful shutdown failed with an error: %s\n", err)
	}
}
