package api

import (
	"net/http"
	"time"
	"wireguard-ui/db"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

func GetPeerTraffic(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
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

	var live wg.Transfer
	if m, err := wg.GetPeerTransferMap(server.Interface); err == nil {
		live = m[peer.PublicKey]
	}

	report, err := db.GetPeerTrafficReport(peer.PublicKey, live, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}
