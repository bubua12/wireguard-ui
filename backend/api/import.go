package api

import (
	"fmt"
	"net/http"
	"strings"
	"wireguard-ui/db"
	"wireguard-ui/model"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

type ImportRequest struct {
	ConfigPath string `json:"config_path"`
	Endpoint   string `json:"endpoint"`
	DNS        string `json:"dns"`
	FullTunnel bool   `json:"full_tunnel"`
	EnableNAT  bool   `json:"enable_nat"`
}

type ImportResult struct {
	Server     *model.Server `json:"server"`
	PeersCount int           `json:"peers_count"`
	Message    string        `json:"message"`
}

func ImportConfig(c *gin.Context) {
	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}
	if err := wg.ValidateEndpoint(req.Endpoint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	path, err := wg.ValidateImportPath(wg.ConfigDir(), req.ConfigPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsed, err := wg.ParseConfigFile(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if parsed.Interface.PrivateKey == "" || parsed.Interface.Address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置文件缺少 Interface 的 PrivateKey 或 Address"})
		return
	}

	publicKey, err := wg.GeneratePublicKey(parsed.Interface.PrivateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法生成公钥: " + err.Error()})
		return
	}

	iface := wg.InterfaceNameFromPath(path)
	server := &model.Server{
		Name:       iface,
		Interface:  iface,
		PrivateKey: parsed.Interface.PrivateKey,
		PublicKey:  publicKey,
		Address:    parsed.Interface.Address,
		ListenPort: parsed.Interface.ListenPort,
		Endpoint:   req.Endpoint,
		DNS:        req.DNS,
		MTU:        parsed.Interface.MTU,
		FullTunnel: req.FullTunnel,
		EnableNAT:  req.EnableNAT,
	}
	defaultsForServer(server)
	if parsed.Interface.DNS != "" && req.DNS == "" {
		server.DNS = parsed.Interface.DNS
	}
	if err := validateServerFields(server.Interface, server.Address, server.Endpoint, server.ListenPort, server.MTU); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	peers := make([]model.Peer, 0, len(parsed.Peers))
	for i, p := range parsed.Peers {
		if p.PublicKey == "" || strings.TrimSpace(p.AllowedIPs) == "" {
			continue
		}
		name := fmt.Sprintf("peer-%d", i+1)
		keepalive := p.PersistentKeepalive
		if keepalive == 0 {
			keepalive = 25
		}
		peers = append(peers, model.Peer{
			Name:                name,
			PrivateKey:          "",
			PublicKey:           p.PublicKey,
			PresharedKey:        p.PresharedKey,
			AllowedIPs:          strings.TrimSpace(p.AllowedIPs),
			PersistentKeepalive: keepalive,
			Enabled:             true,
			HasPrivateKey:       false,
		})
	}

	if err := db.ReplaceAllConfig(server, peers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导入失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, ImportResult{
		Server:     server,
		PeersCount: len(peers),
		Message:    fmt.Sprintf("已覆盖导入服务器 %s 和 %d 个客户端（导入的客户端没有私钥，无法下载配置）", server.Interface, len(peers)),
	})
}
