package api

import (
	"net/http"
	"wireguard-ui/db"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

func SyncConfig(c *gin.Context) {
	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not configured"})
		return
	}

	peers, err := db.GetPeersByServer(server.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	config := wg.GenerateServerConfig(server, peers)

	if err := wg.SaveServerConfig(server.Name, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config"})
		return
	}

	if err := wg.SyncConfig(server.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config synced"})
}
