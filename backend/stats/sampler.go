package stats

import (
	"context"
	"log"
	"time"
	"wireguard-ui/db"
	"wireguard-ui/wg"
)

const sampleInterval = time.Minute

// Start records peer traffic every minute until ctx is cancelled. The returned
// channel closes after the final sample has been written, so shutdown can wait
// for it before the database goes away.
func Start(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(sampleInterval)
		defer ticker.Stop()
		Sample()
		for {
			select {
			case <-ctx.Done():
				Sample()
				return
			case <-ticker.C:
				Sample()
			}
		}
	}()
	return done
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
