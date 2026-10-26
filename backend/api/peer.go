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

// GetNextPeerIP previews the address a new peer would get when none is given.
func GetNextPeerIP(c *gin.Context) {
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

	used := make([]string, 0, len(peers))
	for _, p := range peers {
		used = append(used, p.AllowedIPs)
	}
	ip, err := wg.NextAvailableIP(server.Address, used)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ip": "", "reason": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ip": ip})
}

type CreatePeerReq struct {
	Name       string `json:"name" binding:"required"`
	AllowedIPs string `json:"allowed_ips"`
}

func CreatePeer(c *gin.Context) {
	var req CreatePeerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称不能为空"})
		return
	}
	name, err := validatePeerName(req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mutateMu.Lock()
	defer mutateMu.Unlock()

	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not configured"})
		return
	}

	requestedIP := strings.TrimSpace(req.AllowedIPs)
	if requestedIP != "" {
		if err := wg.ValidateClientIP(server.Address, requestedIP); err != nil {
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
		Name:                name,
		PrivateKey:          privateKey,
		PublicKey:           publicKey,
		PresharedKey:        psk,
		PersistentKeepalive: 25,
		Enabled:             true,
	}

	if err := db.AllocateAndCreatePeer(server, peer, requestedIP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	warning := applyChange(server, func() error { return wg.ApplyPeerAdd(server, peer) })
	c.JSON(http.StatusOK, newPeerResponse(peer, warning))
}
