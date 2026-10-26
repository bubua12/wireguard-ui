package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"wireguard-ui/db"
	"wireguard-ui/stats"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

// InterfaceControl starts, stops or restarts the WireGuard interface.
// Restart is the only way to apply Address, MTU and NAT changes to a running
// interface, but it drops every session for a moment.
func InterfaceControl(c *gin.Context) {
	action := c.Param("action")
	if action != "up" && action != "down" && action != "restart" {
		c.JSON(http.StatusNotFound, gin.H{"error": "未知操作"})
		return
	}

	mutateMu.Lock()
	defer mutateMu.Unlock()

	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not configured"})
		return
	}
	if err := wg.ValidateInterfaceName(server.Interface); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	running := wg.InterfaceExists(server.Interface)
	var message string
	var warning error
	switch action {
	case "up":
		if running {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("接口 %s 已在运行", server.Interface)})
			return
		}
		if err = persistPeers(server); err == nil {
			err = wg.SyncConfig(server.Interface, server.Address)
		}
		message = fmt.Sprintf("接口 %s 已启动", server.Interface)
	case "down":
		if !running {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("接口 %s 未在运行", server.Interface)})
			return
		}
		stats.Sample() // the kernel counters die with the interface
		err = wg.InterfaceDown(server.Interface)
		var hook *wg.HookError
		if errors.As(err, &hook) {
			warning, err = err, nil
		}
		message = fmt.Sprintf("接口 %s 已停止", server.Interface)
	case "restart":
		if running {
			stats.Sample()
		}
		warning, err = wg.RestartInterface(server.Interface, server.Address, func() error { return persistPeers(server) })
		message = fmt.Sprintf("接口 %s 已重启", server.Interface)
	}
	if err != nil {
		log.Printf("interface %s %s: %v", action, server.Interface, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if warning != nil {
		log.Printf("interface %s %s: hook failed: %v", action, server.Interface, warning)
	}

	applyWarning(c, gin.H{"message": message, "up": wg.InterfaceExists(server.Interface)}, warning)
}
