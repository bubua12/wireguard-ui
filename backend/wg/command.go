package wg

import (
	"fmt"
	"os"
	"os/exec"
)

const configDir = "/etc/wireguard"

// SaveServerConfig 保存服务器配置到文件
func SaveServerConfig(name, content string) error {
	path := fmt.Sprintf("%s/%s.conf", configDir, name)
	return os.WriteFile(path, []byte(content), 0600)
}

// InterfaceUp 启动 WireGuard 接口
func InterfaceUp(name string) error {
	cmd := exec.Command("wg-quick", "up", name)
	return cmd.Run()
}

// InterfaceDown 关闭 WireGuard 接口
func InterfaceDown(name string) error {
	cmd := exec.Command("wg-quick", "down", name)
	return cmd.Run()
}

// SyncConfig 同步配置（不中断连接）
func SyncConfig(name string) error {
	path := fmt.Sprintf("%s/%s.conf", configDir, name)
	cmd := exec.Command("wg", "syncconf", name, path)
	return cmd.Run()
}
