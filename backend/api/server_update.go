package api

import (
	"net/http"
	"wireguard-ui/db"

	"github.com/gin-gonic/gin"
)

type UpdateServerReq struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	ListenPort int    `json:"listen_port"`
	Endpoint   string `json:"endpoint"`
	DNS        string `json:"dns"`
	MTU        int    `json:"mtu"`
}

func UpdateServer(c *gin.Context) {
	var req UpdateServerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	if req.Name != "" {
		server.Name = req.Name
	}
	if req.Address != "" {
		server.Address = req.Address
	}
	if req.ListenPort > 0 {
		server.ListenPort = req.ListenPort
	}
	if req.Endpoint != "" {
		server.Endpoint = req.Endpoint
	}
	if req.DNS != "" {
		server.DNS = req.DNS
	}
	if req.MTU > 0 {
		server.MTU = req.MTU
	}

	if err := db.UpdateServer(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, server)
}
