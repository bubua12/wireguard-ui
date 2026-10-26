package api

import (
	"log"
	"net/http"
	"wireguard-ui/db"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

func SyncConfig(c *gin.Context) {
	mutateMu.Lock()
	defer mutateMu.Unlock()

	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not configured"})
		return
	}
	if err := wg.ValidateInterfaceName(server.Interface); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	peers, err := db.GetPeersByServer(server.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := wg.PersistServerConfig(server, peers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := wg.SyncConfig(server.Interface, server.Address); err != nil {
		log.Printf("sync %s: %v", server.Interface, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	up := wg.InterfaceExists(server.Interface)
	c.JSON(http.StatusOK, gin.H{
		"message":          "Config synced",
		"interface":        server.Interface,
		"up":               up,
		"restart_required": up && wg.NeedsRestart(server.Interface, server.Address, server.MTU),
	})
}
