// Command evccmock serves a fake EVCC REST API backed by a YAML file.
//
//	evccmock [-config dev/evcc-mock.yaml] [-listen :7070]
//	evccmock generate [-out dev/evcc-mock.yaml] [-seed 1]
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/mafeko/ladegast/internal/evccmock"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "generate" {
		generate(os.Args[2:])
		return
	}
	serve(os.Args[1:])
}

func serve(args []string) {
	fs := flag.NewFlagSet("evccmock", flag.ExitOnError)
	path := fs.String("config", "dev/evcc-mock.yaml", "path to the YAML data file")
	listen := fs.String("listen", "", "bind address (overrides the file's listen)")
	_ = fs.Parse(args)

	cfg, err := evccmock.LoadConfig(*path)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if *listen != "" {
		cfg.Listen = *listen
	}
	srv, err := evccmock.NewServer(cfg, time.Now())
	if err != nil {
		log.Fatalf("server: %v", err)
	}
	log.Printf("evccmock listening on %s (%d users, data from %s)", cfg.Listen, len(cfg.Users), *path)
	if err := http.ListenAndServe(cfg.Listen, srv.Handler()); err != nil {
		log.Fatalf("http server: %v", err)
	}
}

func generate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	out := fs.String("out", "dev/evcc-mock.yaml", "output file")
	seed := fs.Int64("seed", 1, "random seed")
	_ = fs.Parse(args)

	f, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := evccmock.WriteYAML(f, evccmock.Generate(time.Now(), *seed)); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s", *out)
}
