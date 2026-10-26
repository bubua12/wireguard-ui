package api

import (
	"net/http"
	"strings"
	"wireguard-ui/db"
	"wireguard-ui/model"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

// UpdateServerReq carries only the fields to change; omitted fields stay as is.
// DNS and MTU are pointers so they can be cleared: an empty DNS drops the line
// from client configs and MTU 0 leaves the choice to wg-quick.
type UpdateServerReq struct {
	Name       *string `json:"name"`
	Interface  *string `json:"interface"`
	Address    *string `json:"address"`
	ListenPort *int    `json:"listen_port"`
	Endpoint   *string `json:"endpoint"`
	DNS        *string `json:"dns"`
	MTU        *int    `json:"mtu"`
	FullTunnel *bool   `json:"full_tunnel"`
	EnableNAT  *bool   `json:"enable_nat"`
}

// serverResponse is the saved server plus whether the running interface needs a
// restart before it picks the change up.
type serverResponse struct {
	*model.Server
	RestartRequired bool `json:"restart_required"`
}

func UpdateServer(c *gin.Context) {
	var req UpdateServerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	mutateMu.Lock()
	defer mutateMu.Unlock()

	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}
	before := *server

	if req.Name != nil {
		server.Name = strings.TrimSpace(*req.Name)
	}
	if req.Interface != nil {
		server.Interface = strings.TrimSpace(*req.Interface)
	}
	if req.Address != nil {
		server.Address = strings.TrimSpace(*req.Address)
	}
	if req.ListenPort != nil {
		server.ListenPort = *req.ListenPort
	}
	if req.Endpoint != nil {
		server.Endpoint = strings.TrimSpace(*req.Endpoint)
	}
	if req.DNS != nil {
		server.DNS = *req.DNS
	}
	if req.MTU != nil {
		server.MTU = *req.MTU
	}
	if req.FullTunnel != nil {
		server.FullTunnel = *req.FullTunnel
	}
	if req.EnableNAT != nil {
		server.EnableNAT = *req.EnableNAT
	}
	if server.Name == "" {
		server.Name = server.Interface
	}

	dns, err := validateServerFields(server.Interface, server.Address, server.Endpoint, server.DNS, server.ListenPort, server.MTU)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	server.DNS = dns

	if err := db.UpdateServer(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// wg syncconf cannot change these on a running interface.
	liveSettingChanged := before.Address != server.Address || before.MTU != server.MTU || before.EnableNAT != server.EnableNAT
	restart := wg.InterfaceExists(before.Interface) &&
		(liveSettingChanged || wg.NeedsRestart(before.Interface, server.Address, server.MTU))

	c.JSON(http.StatusOK, serverResponse{Server: server, RestartRequired: restart})
}
