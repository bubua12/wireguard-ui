package api

import (
	"net/http"
	"strconv"
	"wireguard-ui/db"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

type UpdatePeerReq struct {
	Name    string `json:"name"`
	Enabled *bool  `json:"enabled"`
}

func UpdatePeer(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req UpdatePeerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	peer, err := db.GetPeer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}

	name := peer.Name
	enabled := peer.Enabled
	if req.Name != "" {
		name = req.Name
	}
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	if err := db.UpdatePeer(id, name, enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}

func DeletePeer(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	// 先获取 peer 信息
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

	// 从数据库删除
	if err := db.DeletePeer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新配置文件
	peers, _ := db.GetPeersByServer(server.ID)
	config := wg.GenerateServerConfig(server, peers)
	wg.SaveServerConfig(server.Name, config)

	// 动态移除 peer
	wg.RemovePeer(server.Name, peer.PublicKey)

	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}
