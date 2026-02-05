package api

import (
	"net/http"
	"strconv"
	"wireguard-ui/db"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"
)

func GetPeerConfig(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	peer, err := db.GetPeer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}

	server, err := db.GetServer(peer.ServerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	config := wg.GenerateClientConfig(server, peer)

	c.Header("Content-Disposition", "attachment; filename="+peer.Name+".conf")
	c.Data(http.StatusOK, "text/plain", []byte(config))
}

func GetPeerQRCode(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	peer, err := db.GetPeer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}

	server, err := db.GetServer(peer.ServerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	config := wg.GenerateClientConfig(server, peer)
	png, err := qrcode.Encode(config, qrcode.Medium, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
		return
	}

	c.Data(http.StatusOK, "image/png", png)
}
