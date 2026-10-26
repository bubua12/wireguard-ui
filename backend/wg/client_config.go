package wg

import (
	"fmt"
	"net"
	"strings"
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
	var b strings.Builder

	fmt.Fprintf(&b, "[Interface]\nPrivateKey = %s\nAddress = %s\n", peer.PrivateKey, peer.AllowedIPs)
	if dns := strings.TrimSpace(server.DNS); dns != "" {
		fmt.Fprintf(&b, "DNS = %s\n", dns)
	}
	if server.MTU > 0 {
		fmt.Fprintf(&b, "MTU = %d\n", server.MTU)
	}

	fmt.Fprintf(&b, "\n[Peer]\nPublicKey = %s\nEndpoint = %s\nAllowedIPs = %s\n",
		server.PublicKey, server.Endpoint, ClientAllowedIPs(server))
	if peer.PersistentKeepalive > 0 {
		fmt.Fprintf(&b, "PersistentKeepalive = %d\n", peer.PersistentKeepalive)
	}
	if peer.PresharedKey != "" {
		fmt.Fprintf(&b, "PresharedKey = %s\n", peer.PresharedKey)
	}

	return b.String()
}
