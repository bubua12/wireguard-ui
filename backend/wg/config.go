package wg

import (
	"fmt"
	"wireguard-ui/model"
)

func GenerateServerConfig(server *model.Server, peers []model.Peer) string {
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
ListenPort = %d
`, server.PrivateKey, server.Address, server.ListenPort)

	if server.MTU > 0 {
		config += fmt.Sprintf("MTU = %d\n", server.MTU)
	}

	if server.EnableNAT {
		subnet := GetNetworkCIDR(server.Address)
		config += fmt.Sprintf(`PostUp = iptables -A FORWARD -i %%i -j ACCEPT; iptables -A FORWARD -o %%i -j ACCEPT; iptables -t nat -A POSTROUTING -s %s ! -o %%i -j MASQUERADE
PostDown = iptables -D FORWARD -i %%i -j ACCEPT; iptables -D FORWARD -o %%i -j ACCEPT; iptables -t nat -D POSTROUTING -s %s ! -o %%i -j MASQUERADE
`, subnet, subnet)
	}

	for _, peer := range peers {
		if !peer.Enabled || peer.PublicKey == "" {
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
