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
	Name       string `json:"name"`
	Interface  string `json:"interface"`
	Address    string `json:"address" binding:"required"`
	ListenPort int    `json:"listen_port"`
	Endpoint   string `json:"endpoint" binding:"required"`
	DNS        string `json:"dns"`
	MTU        int    `json:"mtu"`
	FullTunnel bool   `json:"full_tunnel"`
	EnableNAT  bool   `json:"enable_nat"`
}

func CreateServer(c *gin.Context) {
	if n, err := db.ServerCount(); err == nil && n > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "服务器已配置，请使用更新或导入覆盖"})
		return
	}

	var req CreateServerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server := &model.Server{
		Name:       req.Name,
		Interface:  req.Interface,
		Address:    req.Address,
		ListenPort: req.ListenPort,
		Endpoint:   req.Endpoint,
		DNS:        req.DNS,
		MTU:        req.MTU,
		FullTunnel: req.FullTunnel,
		EnableNAT:  req.EnableNAT,
	}
	defaultsForServer(server)
	if err := validateServerFields(server.Interface, server.Address, server.Endpoint, server.ListenPort, server.MTU); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	privateKey, publicKey, err := wg.GenerateKeyPair()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	server.PrivateKey = privateKey
	server.PublicKey = publicKey

	if err := db.CreateServer(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, server)
}
