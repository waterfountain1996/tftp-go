package main

import (
	"log"

	"github.com/spf13/pflag"
	"github.com/waterfountain1996/tftp-go"
)

var (
	trace = pflag.Bool("trace", false, "enable packet tracing")
)

func main() {
	pflag.Parse()

	opts := []tftp.OptFunc{}

	if *trace {
		opts = append(opts, tftp.WithTracing)
	}

	server := tftp.NewServer(opts...)
	if err := server.ListenAndServe(":6969"); err != nil {
		log.Fatal(err)
	}
}
