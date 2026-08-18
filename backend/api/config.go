package api

import (
	"fmt"
	"net/http"
	"wireguard-ui/db"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"
)

func GetPeerConfig(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	peer, err := db.GetPeer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}
	if peer.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该客户端由服务端配置导入，没有私钥，无法生成客户端配置"})
		return
	}

	server, err := db.GetServer(peer.ServerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	config := wg.GenerateClientConfig(server, peer)
	filename := wg.SanitizeDownloadName(peer.Name) + ".conf"
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(config))
}

func GetPeerQRCode(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	peer, err := db.GetPeer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}
	if peer.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该客户端由服务端配置导入，没有私钥，无法生成二维码"})
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
