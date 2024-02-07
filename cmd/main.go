package main

import (
	"log"
	"os"

	"github.com/spf13/pflag"
	"github.com/waterfountain1996/tftp-go"
)

var (
	trace = pflag.Bool("trace", false, "enable packet tracing")
)

func main() {
	pflag.Parse()

	logFlags := log.LstdFlags
	for _, env := range os.Environ() {
		if env == "DEBUG=1" {
			logFlags |= log.Lshortfile
		}
	}

	logger := log.New(os.Stderr, "[tftp] ", logFlags)

	opts := []tftp.OptFunc{}

	if *trace {
		opts = append(opts, tftp.WithTracing)
	}

	server := tftp.NewServer(logger, opts...)
	if err := server.ListenAndServe(":6969"); err != nil {
		log.Fatal(err)
	}
}
