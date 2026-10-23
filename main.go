package main

import (
	"log"
	"net/http"

	"github.com/mafeko/ladegast/internal/config"
	"github.com/mafeko/ladegast/internal/evcc"
	"github.com/mafeko/ladegast/internal/httpserver"
	"github.com/mafeko/ladegast/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	st, err := store.Open(cfg.DataDir)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer st.Close()

	if err := httpserver.EnsureAdminBootstrap(st, cfg); err != nil {
		log.Fatalf("admin bootstrap: %v", err)
	}

	evccClient := evcc.NewClient(cfg.EVCCBaseURL, cfg.EVCCAPIKey)
	srv := httpserver.New(cfg, st, evccClient)

	log.Printf("ladegast listening on %s (public=%s admin=%s)", cfg.ListenAddr, cfg.PublicPathPrefix, cfg.AdminPathPrefix)
	if err := http.ListenAndServe(cfg.ListenAddr, srv.Handler()); err != nil {
		log.Fatalf("http server: %v", err)
	}
}
