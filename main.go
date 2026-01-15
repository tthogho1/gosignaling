package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"gosignaling/config"
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

	// Initialize environment variables and Redis
	config.InitEnv()
	config.InitRedis()

	// Prioritize Fly.io PORT environment variable
	port := *portFlag
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	addr := fmt.Sprintf("%s:%d", *addrFlag, port)
	log.Printf("Starting WebRTC signaling server on %s", addr)

	return serve(addr)
}
