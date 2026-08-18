package wg

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

// NextAvailableIP returns the first free host address in serverCIDR as addr/32.
// used may be CIDR or bare IP strings. The server address itself is reserved.
func NextAvailableIP(serverCIDR string, used []string) (string, error) {
	serverIP, ipNet, err := net.ParseCIDR(strings.TrimSpace(serverCIDR))
	if err != nil {
		return "", fmt.Errorf("无效的服务器地址 %q", serverCIDR)
	}
	if serverIP.To4() == nil {
		return "", fmt.Errorf("暂不支持自动分配 IPv6，请手动指定客户端 IP")
	}

	taken := make(map[uint32]struct{}, len(used)+1)
	taken[ipToUint(serverIP.To4())] = struct{}{}
	for _, raw := range used {
		ip, err := parseHostIP(raw)
		if err != nil {
			continue
		}
		if ip4 := ip.To4(); ip4 != nil {
			taken[ipToUint(ip4)] = struct{}{}
		}
	}

	start := ipToUint(ipNet.IP.To4())
	ones, bits := ipNet.Mask.Size()
	if bits != 32 || ones >= 31 {
		return "", fmt.Errorf("网段 %s 没有可用的客户端地址", serverCIDR)
	}
	hostBits := uint(bits - ones)
	count := uint32(1) << hostBits
	// skip network (.0) and broadcast
	for i := uint32(1); i < count-1; i++ {
		candidate := start + i
		if _, ok := taken[candidate]; ok {
			continue
		}
		return uintToIP(candidate).String() + "/32", nil
	}
	return "", fmt.Errorf("网段 %s 已无可用 IP", ipNet.String())
}

// ValidateClientIP checks that clientCIDR is a single host (/32) inside the
// server subnet, and is not the network, broadcast, or server address.
func ValidateClientIP(serverCIDR, clientCIDR string) error {
	serverIP, serverNet, err := net.ParseCIDR(strings.TrimSpace(serverCIDR))
	if err != nil {
		return fmt.Errorf("无效的服务器地址 %q", serverCIDR)
	}
	clientCIDR = strings.TrimSpace(clientCIDR)
	clientIP, clientNet, err := net.ParseCIDR(clientCIDR)
	if err != nil {
		return fmt.Errorf("IP 地址格式无效，请使用 CIDR，例如 10.0.0.2/32")
	}
	ones, bits := clientNet.Mask.Size()
	if bits != 32 || ones != 32 {
		return fmt.Errorf("客户端 IP 必须以 /32 结尾，例如：10.0.0.2/32")
	}
	if clientIP.To4() == nil {
		return fmt.Errorf("暂不支持 IPv6 客户端地址")
	}
	if !serverNet.Contains(clientIP) {
		return fmt.Errorf("IP %s 不在服务器网段 %s 内", clientIP, serverNet)
	}
	if clientIP.Equal(serverIP) {
		return fmt.Errorf("IP %s 是服务器地址，不能分配给客户端", clientIP)
	}
	if isNetworkOrBroadcast(clientIP, serverNet) {
		return fmt.Errorf("IP %s 是网段的网络地址或广播地址", clientIP)
	}
	return nil
}

func parseHostIP(raw string) (net.IP, error) {
	raw = strings.TrimSpace(strings.Split(raw, ",")[0])
	if raw == "" {
		return nil, fmt.Errorf("empty")
	}
	if strings.Contains(raw, "/") {
		ip, _, err := net.ParseCIDR(raw)
		return ip, err
	}
	ip := net.ParseIP(raw)
	if ip == nil {
		return nil, fmt.Errorf("invalid ip")
	}
	return ip, nil
}

func isNetworkOrBroadcast(ip net.IP, n *net.IPNet) bool {
	ip4 := ip.To4()
	net4 := n.IP.To4()
	if ip4 == nil || net4 == nil {
		return false
	}
	ones, bits := n.Mask.Size()
	if bits != 32 || ones >= 31 {
		return false
	}
	start := ipToUint(net4)
	hostBits := uint(bits - ones)
	count := uint32(1) << hostBits
	v := ipToUint(ip4)
	return v == start || v == start+count-1
}

func ipToUint(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

func uintToIP(v uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, v)
	return ip
}

// CanonicalClientIP normalizes a validated client CIDR to dotted-quad/32.
func CanonicalClientIP(clientCIDR string) (string, error) {
	ip, _, err := net.ParseCIDR(strings.TrimSpace(clientCIDR))
	if err != nil {
		return "", err
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return "", fmt.Errorf("not ipv4")
	}
	return ip4.String() + "/32", nil
}

// IPLess reports whether a should sort before b when both are CIDR/IP strings.
func IPLess(a, b string) bool {
	ia, errA := parseHostIP(a)
	ib, errB := parseHostIP(b)
	if errA != nil || errB != nil {
		return a < b
	}
	a4, b4 := ia.To4(), ib.To4()
	if a4 == nil || b4 == nil {
		return a < b
	}
	return ipToUint(a4) < ipToUint(b4)
}
