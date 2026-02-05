package api

import (
	"fmt"
	"net/http"
	"wireguard-ui/db"
	"wireguard-ui/model"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

type ImportRequest struct {
	ConfigPath string `json:"config_path"`
	Endpoint   string `json:"endpoint"`
	DNS        string `json:"dns"`
}

type ImportResult struct {
	Server      *model.Server `json:"server"`
	PeersCount  int           `json:"peers_count"`
	Message     string        `json:"message"`
}

func ImportConfig(c *gin.Context) {
	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if req.ConfigPath == "" {
		req.ConfigPath = "/etc/wireguard/wg0.conf"
	}

	// 解析配置文件
	parsed, err := wg.ParseConfigFile(req.ConfigPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从私钥生成公钥
	publicKey, err := wg.GeneratePublicKey(parsed.Interface.PrivateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法生成公钥"})
		return
	}

	// 创建服务器配置
	server := &model.Server{
		Name:       "wg0",
		PrivateKey: parsed.Interface.PrivateKey,
		PublicKey:  publicKey,
		Address:    parsed.Interface.Address,
		ListenPort: parsed.Interface.ListenPort,
		Endpoint:   req.Endpoint,
		DNS:        req.DNS,
		MTU:        parsed.Interface.MTU,
	}

	// 设置默认值
	if server.ListenPort == 0 {
		server.ListenPort = 51820
	}
	if server.DNS == "" {
		server.DNS = parsed.Interface.DNS
	}
	if server.DNS == "" {
		server.DNS = "8.8.8.8"
	}
	if server.MTU == 0 {
		server.MTU = 1420
	}

	// 保存服务器
	if err := db.CreateServer(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存服务器配置失败: " + err.Error()})
		return
	}

	// 导入客户端
	importedCount := 0
	for i, p := range parsed.Peers {
		peer := &model.Peer{
			ServerID:            server.ID,
			Name:                fmt.Sprintf("peer-%d", i+1),
			PrivateKey:          "",  // 服务端没有客户端私钥
			PublicKey:           p.PublicKey,
			PresharedKey:        p.PresharedKey,
			AllowedIPs:          p.AllowedIPs,
			PersistentKeepalive: p.PersistentKeepalive,
			Enabled:             true,
		}
		if err := db.CreatePeer(peer); err == nil {
			importedCount++
		}
	}

	c.JSON(http.StatusOK, ImportResult{
		Server:     server,
		PeersCount: importedCount,
		Message:    fmt.Sprintf("成功导入服务器配置和 %d 个客户端", importedCount),
	})
}
