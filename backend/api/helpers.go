package api

import (
	"fmt"
	"net/http"
	"strconv"
	"wireguard-ui/db"
	"wireguard-ui/model"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return 0, false
	}
	return id, true
}

func persistPeers(server *model.Server) error {
	peers, err := db.GetPeersByServer(server.ID)
	if err != nil {
		return err
	}
	return wg.PersistServerConfig(server, peers)
}

func applyWarning(c *gin.Context, payload gin.H, err error) {
	if err == nil {
		c.JSON(http.StatusOK, payload)
		return
	}
	payload["warning"] = err.Error()
	c.JSON(http.StatusOK, payload)
}

func defaultsForServer(s *model.Server) {
	if s.Interface == "" {
		s.Interface = "wg0"
	}
	if s.ListenPort == 0 {
		s.ListenPort = 51820
	}
	if s.DNS == "" {
		s.DNS = "8.8.8.8"
	}
	if s.MTU == 0 {
		s.MTU = 1420
	}
	if s.Name == "" {
		s.Name = s.Interface
	}
}

func validateServerFields(iface, address, endpoint string, port, mtu int) error {
	if err := wg.ValidateInterfaceName(iface); err != nil {
		return err
	}
	if err := wg.ValidateServerAddress(address); err != nil {
		return err
	}
	if err := wg.ValidateEndpoint(endpoint); err != nil {
		return err
	}
	if err := wg.ValidateListenPort(port); err != nil {
		return err
	}
	if mtu < 0 || mtu > 9000 {
		return fmt.Errorf("MTU 无效")
	}
	return nil
}
