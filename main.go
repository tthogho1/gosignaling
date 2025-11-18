package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var (
		addrFlag = flag.String("addr", "0.0.0.0", "server address")
		portFlag = flag.Int("port", 5000, "server port")
	)
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *addrFlag, *portFlag)
	log.Printf("Starting WebRTC signaling server on %s", addr)

	return serve(addr)
}
