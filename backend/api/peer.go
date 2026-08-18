package api

import (
	"net/http"
	"strings"
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
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称不能为空"})
		return
	}

	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not configured"})
		return
	}

	if ip := strings.TrimSpace(req.AllowedIPs); ip != "" {
		if err := wg.ValidateClientIP(server.Address, ip); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	privateKey, publicKey, err := wg.GenerateKeyPair()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	psk, err := wg.GeneratePresharedKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	peer := &model.Peer{
		ServerID:            server.ID,
		Name:                req.Name,
		PrivateKey:          privateKey,
		PublicKey:           publicKey,
		PresharedKey:        psk,
		PersistentKeepalive: 25,
		Enabled:             true,
	}

	if err := db.AllocateAndCreatePeer(server, peer, strings.TrimSpace(req.AllowedIPs)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var warning error
	if err := persistPeers(server); err != nil {
		warning = err
	} else if err := wg.ApplyPeerAdd(server, peer); err != nil {
		warning = err
	}

	payload := gin.H{
		"id":                   peer.ID,
		"server_id":            peer.ServerID,
		"name":                 peer.Name,
		"public_key":           peer.PublicKey,
		"allowed_ips":          peer.AllowedIPs,
		"persistent_keepalive": peer.PersistentKeepalive,
		"enabled":              peer.Enabled,
		"has_private_key":      peer.HasPrivateKey,
		"created_at":           peer.CreatedAt,
		"updated_at":           peer.UpdatedAt,
	}
	if warning != nil {
		payload["warning"] = warning.Error()
	}
	c.JSON(http.StatusOK, payload)
}
