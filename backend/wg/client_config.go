package wg

import (
	"fmt"
	"net"
	"wireguard-ui/model"
)

// GetNetworkCIDR 从服务器地址提取网段 CIDR（例如 10.0.8.1/24 -> 10.0.8.0/24）
func GetNetworkCIDR(address string) string {
	_, ipNet, err := net.ParseCIDR(address)
	if err != nil {
		return "10.0.8.0/24" // 默认值
	}
	return ipNet.String()
}

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

	// 使用服务器网段作为 AllowedIPs，而不是全流量转发
	allowedIPs := GetNetworkCIDR(server.Address)

	config += fmt.Sprintf(`
[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = %s
PersistentKeepalive = %d
`, server.PublicKey, server.Endpoint, allowedIPs, peer.PersistentKeepalive)

	if peer.PresharedKey != "" {
		config += fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey)
	}

	return config
}
