package api

import (
	"net/http"
	"wireguard-ui/db"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

type UpdatePeerReq struct {
	Name    string `json:"name"`
	Enabled *bool  `json:"enabled"`
}

func UpdatePeer(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req UpdatePeerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	peer, err := db.GetPeer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}

	name := peer.Name
	enabled := peer.Enabled
	if req.Name != "" {
		name = req.Name
	}
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	if err := db.UpdatePeer(id, name, enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}

func DeletePeer(c *gin.Context) {
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

	if err := db.DeletePeer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = db.DeletePeerTransfer(peer.PublicKey)

	var warning error
	if err := persistPeers(server); err != nil {
		warning = err
	} else if err := wg.ApplyPeerRemove(server, peer.PublicKey); err != nil {
		warning = err
	}

	applyWarning(c, gin.H{"message": "Deleted"}, warning)
}
