package api

import (
	"net/http"
	"strconv"
	"wireguard-ui/db"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

type TogglePeerReq struct {
	Enabled bool `json:"enabled"`
}

func TogglePeer(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req TogglePeerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取 peer 信息
	peer, err := db.GetPeer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}

	// 获取 server 信息
	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not configured"})
		return
	}

	// 更新数据库
	if err := db.TogglePeer(id, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新配置文件
	peers, _ := db.GetPeersByServer(server.ID)
	config := wg.GenerateServerConfig(server, peers)
	wg.SaveServerConfig(server.Name, config)

	// 动态更新 WireGuard
	if req.Enabled {
		wg.AddPeer(server.Name, peer.PublicKey, peer.PresharedKey, peer.AllowedIPs)
	} else {
		wg.RemovePeer(server.Name, peer.PublicKey)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Toggled", "enabled": req.Enabled})
}
