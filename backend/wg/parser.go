package wg

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParsedConfig 解析后的配置
type ParsedConfig struct {
	Interface InterfaceConfig
	Peers     []PeerConfig
}

// InterfaceConfig 服务器接口配置
type InterfaceConfig struct {
	PrivateKey string
	Address    string
	ListenPort int
	DNS        string
	MTU        int
}

// PeerConfig 客户端配置
type PeerConfig struct {
	PublicKey           string
	PresharedKey        string
	AllowedIPs          string
	PersistentKeepalive int
}

// ParseConfigFile 解析 WireGuard 配置文件
func ParseConfigFile(path string) (*ParsedConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("无法打开配置文件: %v", err)
	}
	defer file.Close()

	config := &ParsedConfig{}
	var currentSection string
	var currentPeer *PeerConfig

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 检查段落标记
		if line == "[Interface]" {
			currentSection = "interface"
			continue
		}
		if line == "[Peer]" {
			currentSection = "peer"
			if currentPeer != nil {
				config.Peers = append(config.Peers, *currentPeer)
			}
			currentPeer = &PeerConfig{}
			continue
		}

		// 解析键值对
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch currentSection {
		case "interface":
			parseInterfaceLine(&config.Interface, key, value)
		case "peer":
			if currentPeer != nil {
				parsePeerLine(currentPeer, key, value)
			}
		}
	}

	// 添加最后一个 peer
	if currentPeer != nil {
		config.Peers = append(config.Peers, *currentPeer)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取配置文件错误: %v", err)
	}

	return config, nil
}

func parseInterfaceLine(iface *InterfaceConfig, key, value string) {
	switch key {
	case "PrivateKey":
		iface.PrivateKey = value
	case "Address":
		iface.Address = value
	case "ListenPort":
		iface.ListenPort, _ = strconv.Atoi(value)
	case "DNS":
		iface.DNS = value
	case "MTU":
		iface.MTU, _ = strconv.Atoi(value)
	}
}

func parsePeerLine(peer *PeerConfig, key, value string) {
	switch key {
	case "PublicKey":
		peer.PublicKey = value
	case "PresharedKey":
		peer.PresharedKey = value
	case "AllowedIPs":
		peer.AllowedIPs = value
	case "PersistentKeepalive":
		peer.PersistentKeepalive, _ = strconv.Atoi(value)
	}
}
