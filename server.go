package main

import (
	"log"
	"net/http"

	"gosignaling/handler"
	"gosignaling/manager"
	"gosignaling/repository/mem"
)

func serve(addr string) error {
	roomRepo := mem.NewRoomRepository()
	roomManager := manager.NewRoomManager(roomRepo)
	h := handler.NewHandler(roomManager)

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
