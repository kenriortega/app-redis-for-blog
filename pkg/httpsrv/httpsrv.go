package httpsrv

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type server struct {
	*http.Server
}

func NewServer(host string, port int, mux http.Handler) *server {

	s := &http.Server{
		Handler: mux,
		Addr:    fmt.Sprintf("%s:%d", host, port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	return &server{s}
}

// Start runs ListenAndServe on the http.Server with graceful shutdown
func (srv *server) Start() {
	log.Println("server: starting server...")

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not listen on %s due to %s", srv.Addr, err)
		}
	}()
	log.Println(fmt.Sprintf("server: server is ready to handle requests %s", srv.Addr))
	srv.gracefulShutdown()
}

func (srv *server) gracefulShutdown() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)
	sig := <-quit
	log.Println(fmt.Sprintf("server: server is shutting down %s", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	srv.SetKeepAlivesEnabled(false)
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("could not gracefully shutdown the server %s", err)

	}
	log.Println("server: server stopped")
}
