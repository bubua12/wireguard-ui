package wg

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
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

// AddPeer 动态添加 peer（不重启接口）
func AddPeer(interfaceName, publicKey, presharedKey, allowedIPs string) error {
	args := []string{"set", interfaceName, "peer", publicKey, "allowed-ips", allowedIPs}
	if presharedKey != "" {
		args = append(args, "preshared-key", "/dev/stdin")
		cmd := exec.Command("wg", args...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return err
		}
		stdin.Write([]byte(presharedKey))
		stdin.Close()
		return cmd.Wait()
	}
	cmd := exec.Command("wg", args...)
	return cmd.Run()
}

// RemovePeer 动态移除 peer（不重启接口）
func RemovePeer(interfaceName, publicKey string) error {
	cmd := exec.Command("wg", "set", interfaceName, "peer", publicKey, "remove")
	return cmd.Run()
}

// GetPeerHandshakes 获取所有 peer 的最后握手时间
func GetPeerHandshakes(interfaceName string) (map[string]int64, error) {
	cmd := exec.Command("wg", "show", interfaceName, "latest-handshakes")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			publicKey := parts[0]
			timestamp, _ := strconv.ParseInt(parts[1], 10, 64)
			result[publicKey] = timestamp
		}
	}
	return result, nil
}

// IsPeerOnline 判断 peer 是否在线（3分钟内有握手）
func IsPeerOnline(lastHandshake int64) bool {
	if lastHandshake == 0 {
		return false
	}
	return time.Now().Unix()-lastHandshake < 180
}
