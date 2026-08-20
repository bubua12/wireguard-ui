package wg

import (
	"bufio"
	"bytes"
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

// SyncConfig assumes the full wg-quick config is already on disk.
// If the interface is down, bring it up with wg-quick. If it is already
// running, strip wg-quick-only keys (Address, MTU, PostUp, ...) and apply
// the remainder with `wg syncconf` so existing sessions stay up.
// address is the server CIDR (e.g. 10.0.8.1/24) used to refuse creating a
// second interface that would collide with an existing one.
func SyncConfig(name, address string) error {
	if err := ValidateInterfaceName(name); err != nil {
		return err
	}
	if !InterfaceExists(name) {
		if owner, ok := interfaceHoldingAddress(address); ok && owner != name {
			return fmt.Errorf("内网地址 %s 已在接口 %s 上，但当前接口名是 %q（系统里没有这个网卡）。请把「接口名」改回 %s 后再同步；显示名称不要填到接口名里", hostIP(address), owner, name, owner)
		}
		return InterfaceUp(name)
	}

	stripped, err := stripQuickConfig(name)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "wg-sync-"+name+"-*.conf")
	if err != nil {
		return fmt.Errorf("创建临时配置失败: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(stripped); err != nil {
		tmp.Close()
		return fmt.Errorf("写入临时配置失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	_, err = runWG("wg", "syncconf", name, tmpName)
	return err
}

func hostIP(cidr string) string {
	ip, _, err := net.ParseCIDR(strings.TrimSpace(cidr))
	if err != nil {
		return strings.TrimSpace(cidr)
	}
	return ip.String()
}

func interfaceHoldingAddress(cidr string) (string, bool) {
	want := net.ParseIP(hostIP(cidr))
	if want == nil {
		return "", false
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", false
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet.IP == nil {
				continue
			}
			if ipnet.IP.Equal(want) {
				return iface.Name, true
			}
		}
	}
	return "", false
}

func stripQuickConfig(name string) ([]byte, error) {
	cmd := exec.Command("wg-quick", "strip", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return nil, fmt.Errorf("wg-quick strip %s: %w", name, err)
		}
		return nil, fmt.Errorf("wg-quick strip %s: %s", name, msg)
	}
	return out, nil
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

// Transfer is a peer's rx/tx byte counters from `wg show transfer`.
type Transfer struct {
	Rx int64
	Tx int64
}

func GetPeerTransferMap(interfaceName string) (map[string]Transfer, error) {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return nil, err
	}
	output, err := runWG("wg", "show", interfaceName, "transfer")
	if err != nil {
		return nil, err
	}
	return parseTransferOutput(string(output)), nil
}

func parseTransferOutput(output string) map[string]Transfer {
	result := make(map[string]Transfer)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 3 {
			continue
		}
		rx, err1 := strconv.ParseInt(parts[1], 10, 64)
		tx, err2 := strconv.ParseInt(parts[2], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		result[parts[0]] = Transfer{Rx: rx, Tx: tx}
	}
	return result
}

func GetPeerTransfers(interfaceName string) (rx, tx int64, err error) {
	m, err := GetPeerTransferMap(interfaceName)
	if err != nil {
		return 0, 0, err
	}
	for _, v := range m {
		rx += v.Rx
		tx += v.Tx
	}
	return rx, tx, nil
}

func IsPeerOnline(lastHandshake int64) bool {
	if lastHandshake == 0 {
		return false
	}
	return time.Now().Unix()-lastHandshake < 180
}
