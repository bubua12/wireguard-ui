package wg

import (
	"fmt"
	"net"
	"wireguard-ui/model"
)

func GetNetworkCIDR(address string) string {
	_, ipNet, err := net.ParseCIDR(address)
	if err != nil {
		return address
	}
	return ipNet.String()
}

func ClientAllowedIPs(server *model.Server) string {
	if server.FullTunnel {
		return "0.0.0.0/0, ::/0"
	}
	return GetNetworkCIDR(server.Address)
}

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
AllowedIPs = %s
PersistentKeepalive = %d
`, server.PublicKey, server.Endpoint, ClientAllowedIPs(server), peer.PersistentKeepalive)

	if peer.PresharedKey != "" {
		config += fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey)
	}

	return config
}
