package wg

import (
	"fmt"
	"wireguard-ui/model"
)

// GenerateClientConfig 生成客户端配置文件内容
func GenerateClientConfig(server *model.Server, peer *model.Peer) string {
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = %s
`, peer.PrivateKey, peer.AllowedIPs, server.DNS)

	if server.MTU > 0 {
		config += fmt.Sprintf("MTU = %d\n", server.MTU)
	}

	config += fmt.Sprintf(`
[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = %d
`, server.PublicKey, server.Endpoint, peer.PersistentKeepalive)

	if peer.PresharedKey != "" {
		config += fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey)
	}

	return config
}
