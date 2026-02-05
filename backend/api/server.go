package api

import (
	"net/http"
	"wireguard-ui/db"
	"wireguard-ui/model"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

func GetServer(c *gin.Context) {
	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not configured"})
		return
	}
	c.JSON(http.StatusOK, server)
}

type CreateServerReq struct {
	Name       string `json:"name" binding:"required"`
	Address    string `json:"address" binding:"required"`
	ListenPort int    `json:"listen_port"`
	Endpoint   string `json:"endpoint" binding:"required"`
	DNS        string `json:"dns"`
	MTU        int    `json:"mtu"`
}

func CreateServer(c *gin.Context) {
	var req CreateServerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成密钥对
	privateKey, publicKey, err := wg.GenerateKeyPair()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate keys"})
		return
	}

	server := &model.Server{
		Name:       req.Name,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    req.Address,
		ListenPort: req.ListenPort,
		Endpoint:   req.Endpoint,
		DNS:        req.DNS,
		MTU:        req.MTU,
	}

	if server.ListenPort == 0 {
		server.ListenPort = 51820
	}
	if server.DNS == "" {
		server.DNS = "8.8.8.8"
	}
	if server.MTU == 0 {
		server.MTU = 1420
	}

	if err := db.CreateServer(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, server)
}
