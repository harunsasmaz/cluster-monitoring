package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/harunsasmaz/cluster-monitoring/internal/server"
	"github.com/harunsasmaz/cluster-monitoring/pkg/cluster"
	"k8s.io/client-go/rest"
)

func main() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("failed to load in-cluster config: %v\n", err)
	}

	client, err := cluster.NewClient(config)
	if err != nil {
		log.Fatalf("failed to create cluster client: %v\n", err)
	}

	srv := server.New(client)
	errCh := make(chan error)

	go func() {
		if err := srv.Serve(); err != nil {
			errCh <- err
		}
		close(errCh)
	}()
	log.Printf("Starting server at %s\n", srv.Addr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
	case err, ok := <-errCh:
		if ok {
			log.Printf("server error: %v\n", err)
		}

	case sig := <-sigCh:
		log.Printf("Signal %s received\n", sig)
		if err := srv.Shutdown(); err != nil {
			log.Printf("failed to shutdown server: %v\n", err)
		}
		log.Println("server shutdown")
	}
}
