package main

import (
	"log"

	"github.com/waterfountain1996/tftp-go"
)

func main() {
	server := tftp.NewServer()
	if err := server.ListenAndServe(":6969"); err != nil {
		log.Fatal(err)
	}
}
