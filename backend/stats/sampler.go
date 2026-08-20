package stats

import (
	"log"
	"time"
	"wireguard-ui/db"
	"wireguard-ui/wg"
)

const sampleInterval = time.Minute

func Start() {
	go run()
}

func run() {
	ticker := time.NewTicker(sampleInterval)
	defer ticker.Stop()
	Sample()
	for range ticker.C {
		Sample()
	}
}

func Sample() {
	server, err := db.GetFirstServer()
	if err != nil {
		return
	}
	if !wg.InterfaceExists(server.Interface) {
		return
	}
	live, err := wg.GetPeerTransferMap(server.Interface)
	if err != nil {
		return
	}
	if err := db.ApplyTransferSample(live, time.Now()); err != nil {
		log.Printf("traffic sample: %v", err)
	}
}
