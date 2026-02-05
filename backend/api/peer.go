package api

import (
	"net/http"
	"wireguard-ui/db"
	"wireguard-ui/model"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

func GetPeers(c *gin.Context) {
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

	c.JSON(http.StatusOK, peers)
}

type CreatePeerReq struct {
	Name       string `json:"name" binding:"required"`
	AllowedIPs string `json:"allowed_ips"`
}

func CreatePeer(c *gin.Context) {
	var req CreatePeerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not configured"})
		return
	}

	// 生成密钥
	privateKey, publicKey, err := wg.GenerateKeyPair()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate keys"})
		return
	}

	psk, _ := wg.GeneratePresharedKey()

	// 分配IP：如果指定了IP则使用指定的，否则自动分配
	ip := req.AllowedIPs
	if ip == "" {
		ip, err = db.GetNextAvailableIP(server.ID, server.Address)
		if err != nil {
			ip = "10.0.0.2/32"
		}
	}

	peer := &model.Peer{
		ServerID:            server.ID,
		Name:                req.Name,
		PrivateKey:          privateKey,
		PublicKey:           publicKey,
		PresharedKey:        psk,
		AllowedIPs:          ip,
		PersistentKeepalive: 25,
		Enabled:             true,
	}

	if err := db.CreatePeer(peer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新服务端配置文件
	peers, _ := db.GetPeersByServer(server.ID)
	config := wg.GenerateServerConfig(server, peers)
	wg.SaveServerConfig(server.Name, config)

	// 动态添加 peer，不需要重启接口
	wg.AddPeer(server.Name, peer.PublicKey, peer.PresharedKey, peer.AllowedIPs)

	c.JSON(http.StatusOK, peer)
}
