package main

import (
	"log"
	"net/http"

	"gosignaling/config"
	"gosignaling/handler"
	"gosignaling/manager"
	"gosignaling/repository/mem"
	"gosignaling/services"
)

func serve(addr string) error {
	roomRepo := mem.NewRoomRepository()
	roomManager := manager.NewRoomManager(roomRepo)
	h := handler.NewHandler(roomManager)

	// Initialize clustering service for multi-pod support (if Redis is available)
	if config.Rdb != nil {
		clusteringService := services.NewClusteringService(roomManager)
		clusteringService.InitializeRedisSubscriptions()
		log.Println("✅ Redis Pub/Sub clustering initialized for WebRTC signaling")
	} else {
		log.Println("ℹ️ Running in standalone mode (no Redis clustering)")
	}

	// Health check endpoint for Fly.io
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		h.CreateConnection(w, r)
	})

	// Add /ws endpoint as alias for /connect
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		h.CreateConnection(w, r)
	})

	log.Printf("WebRTC signaling server listening on %s", addr)
	return http.ListenAndServe(addr, nil)
}
