package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Farengier/smart-home/internal/signal"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type Config interface {
	Addr() string
	WriteTimeout() time.Duration
	ReadTimeout() time.Duration
}

func Start(cfg Config) {
	log.Info("[Web] Starting server")

	r := mux.NewRouter()
	r.HandleFunc("/", testHandler)
	r.HandleFunc("/products", testHandler)
	r.HandleFunc("/articles", testHandler)

	bctx, cncl := context.WithCancel(context.Background())
	srv := &http.Server{
		Handler:      r,
		Addr:         cfg.Addr(),
		WriteTimeout: cfg.WriteTimeout(),
		ReadTimeout:  cfg.ReadTimeout(),
		BaseContext: func(_ net.Listener) context.Context {
			return bctx
		},
	}

	signal.OnShutdown(func() error {
		log.Info("[Web] Shutdown server")
		cncl()
		err := srv.Shutdown(context.Background())
		if err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		return nil
	})
	_ = srv.ListenAndServe()
}

func testHandler(rw http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(rw, "test handler ok")
}
