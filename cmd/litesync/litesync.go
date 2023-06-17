package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mikaelhg/litesync/internal"
)

func main() {
	bind := flag.String("bind", ":8295", "interface and port to bind the server to (default :8295)")
	dbFile := flag.String("db", "./litesync.sqlite", "database file (default ./litesync.sqlite)")
	help := flag.Bool("help", false, "usage")
	flag.Parse()
	if *help {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nStart your browser with the command:\n")
		fmt.Fprintf(os.Stderr, "  brave-browser --sync-url=http://localhost:8295/litesync")
	} else {
		internal.StartServer(*bind, *dbFile)
	}
}
