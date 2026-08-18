package api

import (
	"net/http"
	"wireguard-ui/db"
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

	if err := db.TogglePeer(id, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var warning error
	if err := persistPeers(server); err != nil {
		warning = err
	} else if req.Enabled {
		if err := wg.ApplyPeerAdd(server, peer); err != nil {
			warning = err
		}
	} else if err := wg.ApplyPeerRemove(server, peer.PublicKey); err != nil {
		warning = err
	}

	applyWarning(c, gin.H{"message": "Toggled", "enabled": req.Enabled}, warning)
}
