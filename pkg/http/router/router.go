package router

import (
	"smart-slowquery/pkg/log"

	"context"
	"fmt"
	"net/http"
	"time"
)

type Router struct {
	port, shutDownWaitTimeout int
	Srv                       *http.Server
}

func InitRouter(port, shutDownWaitTimeout int) (router *Router) {
	return &Router{
		port:                port,
		shutDownWaitTimeout: shutDownWaitTimeout,
	}
}

func (router *Router) Run(engine http.Handler) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", router.port),
		Handler: engine,
	}
	router.Srv = srv
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Info(fmt.Sprintf("Listen_err: %v", err))
		}
	}()
}

// Shutdown router
func (router *Router) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(router.shutDownWaitTimeout)*time.Second)
	defer cancel()

	select {
	// when timeout or cancel, Done() will be captured
	case <-ctx.Done():
		log.Errorf("Server_shutdown_timeout")
	default:
		log.Info("Server_exiting")
	}
	log.Info("Server_exited")

	return router.Srv.Shutdown(ctx)
}
