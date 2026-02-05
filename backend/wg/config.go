package wg

import (
	"fmt"
	"wireguard-ui/model"
)

// GenerateServerConfig 生成服务器配置文件内容
func GenerateServerConfig(server *model.Server, peers []model.Peer) string {
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
ListenPort = %d
`, server.PrivateKey, server.Address, server.ListenPort)

	for _, peer := range peers {
		if !peer.Enabled {
			continue
		}
		config += fmt.Sprintf(`
[Peer]
PublicKey = %s
AllowedIPs = %s
`, peer.PublicKey, peer.AllowedIPs)
		if peer.PresharedKey != "" {
			config += fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey)
		}
	}

	return config
}
