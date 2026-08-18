package wg

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var configDir = "/etc/wireguard"

func SetConfigDir(dir string) {
	if strings.TrimSpace(dir) != "" {
		configDir = dir
	}
}

func ConfigDir() string {
	return configDir
}

func ConfigPath(name string) (string, error) {
	if err := ValidateInterfaceName(name); err != nil {
		return "", err
	}
	return filepath.Join(configDir, name+".conf"), nil
}

func runWG(args ...string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return out, fmt.Errorf("%s: %w", strings.Join(args, " "), err)
		}
		return out, fmt.Errorf("%s: %s", strings.Join(args, " "), msg)
	}
	return out, nil
}

func SaveServerConfig(name, content string) error {
	path, err := ConfigPath(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0600); err != nil {
		return fmt.Errorf("写入配置失败: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("保存配置失败: %w", err)
	}
	return nil
}

func InterfaceExists(name string) bool {
	if err := ValidateInterfaceName(name); err != nil {
		return false
	}
	iface, err := net.InterfaceByName(name)
	return err == nil && iface != nil
}

func InterfaceUp(name string) error {
	if err := ValidateInterfaceName(name); err != nil {
		return err
	}
	_, err := runWG("wg-quick", "up", name)
	return err
}

func InterfaceDown(name string) error {
	if err := ValidateInterfaceName(name); err != nil {
		return err
	}
	_, err := runWG("wg-quick", "down", name)
	return err
}

// SyncConfig writes are assumed done. If the interface is down, bring it up;
// otherwise apply the file with syncconf so existing sessions stay up.
func SyncConfig(name string) error {
	path, err := ConfigPath(name)
	if err != nil {
		return err
	}
	if !InterfaceExists(name) {
		return InterfaceUp(name)
	}
	_, err = runWG("wg", "syncconf", name, path)
	return err
}

func AddPeer(interfaceName, publicKey, presharedKey, allowedIPs string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	args := []string{"wg", "set", interfaceName, "peer", publicKey, "allowed-ips", allowedIPs}
	if presharedKey != "" {
		args = append(args, "preshared-key", "/dev/stdin")
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(presharedKey + "\n")
		out, err := cmd.CombinedOutput()
		if err != nil {
			msg := strings.TrimSpace(string(out))
			if msg == "" {
				return fmt.Errorf("wg set peer: %w", err)
			}
			return fmt.Errorf("wg set peer: %s", msg)
		}
		return nil
	}
	_, err := runWG(args...)
	return err
}

func RemovePeer(interfaceName, publicKey string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	_, err := runWG("wg", "set", interfaceName, "peer", publicKey, "remove")
	return err
}

func GetPeerHandshakes(interfaceName string) (map[string]int64, error) {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return nil, err
	}
	output, err := runWG("wg", "show", interfaceName, "latest-handshakes")
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 2 {
			timestamp, _ := strconv.ParseInt(parts[1], 10, 64)
			result[parts[0]] = timestamp
		}
	}
	return result, nil
}

func GetPeerTransfers(interfaceName string) (rx, tx int64, err error) {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return 0, 0, err
	}
	output, err := runWG("wg", "show", interfaceName, "transfer")
	if err != nil {
		return 0, 0, err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 3 {
			r, _ := strconv.ParseInt(parts[1], 10, 64)
			t, _ := strconv.ParseInt(parts[2], 10, 64)
			rx += r
			tx += t
		}
	}
	return rx, tx, nil
}

func IsPeerOnline(lastHandshake int64) bool {
	if lastHandshake == 0 {
		return false
	}
	return time.Now().Unix()-lastHandshake < 180
}
