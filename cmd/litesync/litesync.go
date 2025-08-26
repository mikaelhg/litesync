package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mikaelhg/litesync/internal"
)

var (
	bindAddr = flag.String("bind", defaultBindAddr, "interface and port to bind the server to")
	dbPath   = flag.String("db", defaultDBPath, "database file path")
	showHelp = flag.Bool("help", false, "display usage information")
)

const (
	defaultBindAddr = ":8295"
	defaultDBPath   = "./litesync.sqlite"
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if *showHelp {
		usage()
		os.Exit(0)
	}

	if err := internal.StartServer(*bindAddr, *dbPath); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nBrowser startup example:\n")
	fmt.Fprintf(os.Stderr, "  brave-browser --sync-url=http://localhost:8295/litesync\n")
}
