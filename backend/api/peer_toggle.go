package api

import (
	"net/http"
	"wireguard-ui/db"
	"wireguard-ui/stats"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

type TogglePeerReq struct {
	Enabled bool `json:"enabled"`
}

func TogglePeer(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req TogglePeerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mutateMu.Lock()
	defer mutateMu.Unlock()

	peer, err := db.GetPeer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}

	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not configured"})
		return
	}

	if !req.Enabled && peer.Enabled {
		stats.Sample() // record what the peer moved before its counters vanish
	}
	if err := db.TogglePeer(id, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	warning := applyChange(server, func() error {
		if req.Enabled {
			return wg.ApplyPeerAdd(server, peer)
		}
		return wg.ApplyPeerRemove(server, peer.PublicKey)
	})
	applyWarning(c, gin.H{"message": "Toggled", "enabled": req.Enabled}, warning)
}
