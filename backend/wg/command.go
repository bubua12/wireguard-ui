package wg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// Plain wg calls return immediately; wg-quick may resolve endpoints and run
	// PostUp/PostDown hooks, so it gets a longer leash.
	cmdTimeout   = 15 * time.Second
	quickTimeout = 60 * time.Second
	// A killed wg-quick can leave children holding its pipes open; stop waiting
	// for them shortly after the kill instead of hanging the request.
	cmdWaitDelay = 2 * time.Second
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

// runCmd runs args[0] with a deadline and returns its combined output. stdin is
// only attached when non-empty.
func runCmd(timeout time.Duration, stdin string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.WaitDelay = cmdWaitDelay
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	if err == nil {
		return out, nil
	}

	line := strings.Join(args, " ")
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		return out, fmt.Errorf("%s: 执行超时（%s）", line, timeout)
	case errors.Is(err, exec.ErrNotFound):
		return out, fmt.Errorf("未找到 %s 命令，请先安装 wireguard-tools", args[0])
	}
	if msg := strings.TrimSpace(string(out)); msg != "" {
		return out, fmt.Errorf("%s: %s", line, msg)
	}
	return out, fmt.Errorf("%s: %w", line, err)
}

func runWG(args ...string) ([]byte, error) {
	return runCmd(cmdTimeout, "", args...)
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
	_, err := runCmd(quickTimeout, "", "wg-quick", "up", name)
	return err
}

// HookError means wg-quick took the interface down but a PostDown hook failed.
// That happens when the config on disk no longer matches what was applied at
// startup (for example NAT was switched on after the interface came up, so the
// iptables -D rules have nothing to delete). The interface is down regardless,
// so callers may carry on and surface it as a warning.
type HookError struct{ Err error }

func (e *HookError) Error() string { return e.Err.Error() }
func (e *HookError) Unwrap() error { return e.Err }

func InterfaceDown(name string) error {
	if err := ValidateInterfaceName(name); err != nil {
		return err
	}
	_, err := runCmd(quickTimeout, "", "wg-quick", "down", name)
	if err != nil && !InterfaceExists(name) {
		return &HookError{Err: err}
	}
	return err
}

// RestartInterface takes the interface down (when it is up) and brings it up
// again from the config on disk. Unlike SyncConfig this also applies Address,
// MTU and PostUp/PostDown changes, at the price of dropping every session.
//
// rewrite regenerates the config file. It runs between stop and start so that
// the PostDown hooks of the old interface execute against the config it was
// started with, and the new PostUp hooks against the new one; with NAT toggled
// the other order leaves iptables rules behind or fails to delete them.
// A failing PostDown hook does not abort the restart; it is returned as warning.
func RestartInterface(name, address string, rewrite func() error) (warning error, err error) {
	wasUp := InterfaceExists(name)
	if wasUp {
		if err := InterfaceDown(name); err != nil {
			var hook *HookError
			if !errors.As(err, &hook) {
				return nil, err
			}
			warning = err
		}
	}
	err = rewrite()
	if err == nil {
		err = checkAddressFree(name, address)
	}
	if err == nil {
		err = InterfaceUp(name)
	}
	if err != nil && wasUp {
		return warning, fmt.Errorf("接口已停止，但重新启动失败: %w", err)
	}
	return warning, err
}

// checkAddressFree refuses to bring up name when another interface already
// carries the server address, which would silently create a second interface
// with a colliding address.
func checkAddressFree(name, address string) error {
	if owner, ok := interfaceHoldingAddress(address); ok && owner != name {
		return fmt.Errorf("内网地址 %s 已在接口 %s 上，但当前接口名是 %q（系统里没有这个网卡）。请把「接口名」改回 %s 后再同步；显示名称不要填到接口名里", hostIP(address), owner, name, owner)
	}
	return nil
}

// NeedsRestart reports whether the running interface still carries an older
// Address or MTU than the saved settings. wg syncconf cannot change either, so
// only a restart applies them. It is false when the interface is not running.
func NeedsRestart(name, address string, mtu int) bool {
	if ValidateInterfaceName(name) != nil {
		return false
	}
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return false
	}
	if mtu > 0 && iface.MTU != mtu {
		return true
	}

	want, wantNet, err := net.ParseCIDR(strings.TrimSpace(address))
	if err != nil {
		return false
	}
	wantOnes, _ := wantNet.Mask.Size()
	addrs, err := iface.Addrs()
	if err != nil {
		return false
	}
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		if ones, _ := ipnet.Mask.Size(); ipnet.IP.Equal(want) && ones == wantOnes {
			return false
		}
	}
	return true
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
		if err := checkAddressFree(name, address); err != nil {
			return err
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
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wg-quick", "strip", name)
	cmd.WaitDelay = cmdWaitDelay
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("未找到 wg-quick 命令，请先安装 wireguard-tools")
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return nil, fmt.Errorf("wg-quick strip %s: %w", name, err)
		}
		return nil, fmt.Errorf("wg-quick strip %s: %s", name, msg)
	}
	return out, nil
}

// AddPeer adds the peer to the live interface, or replaces its allowed IPs and
// preshared key when it is already there.
func AddPeer(interfaceName, publicKey, presharedKey, allowedIPs string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	args := []string{"wg", "set", interfaceName, "peer", publicKey, "allowed-ips", allowedIPs}
	stdin := ""
	if presharedKey != "" {
		args = append(args, "preshared-key", "/dev/stdin")
		stdin = presharedKey + "\n"
	}
	_, err := runCmd(cmdTimeout, stdin, args...)
	return err
}

func RemovePeer(interfaceName, publicKey string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	_, err := runWG("wg", "set", interfaceName, "peer", publicKey, "remove")
	return err
}

// Transfer is a peer's rx/tx byte counters. Both are from the server's point of
// view: rx was received from the peer (its upload), tx was sent to it (its
// download).
type Transfer struct {
	Rx int64
	Tx int64
}

// PeerState is what the kernel currently knows about a peer.
type PeerState struct {
	Endpoint        string // remote ip:port of the last packet, empty if never seen
	LatestHandshake int64  // unix seconds, 0 if never
	Transfer        Transfer
}

// GetPeerStates reads every peer's endpoint, handshake and counters with a
// single `wg show <iface> dump`.
func GetPeerStates(interfaceName string) (map[string]PeerState, error) {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return nil, err
	}
	output, err := runWG("wg", "show", interfaceName, "dump")
	if err != nil {
		return nil, err
	}
	return parseDump(string(output)), nil
}

// parseDump reads `wg show <iface> dump`. The first line describes the
// interface; every other line is tab separated:
// public-key preshared-key endpoint allowed-ips latest-handshake rx tx keepalive
func parseDump(output string) map[string]PeerState {
	result := make(map[string]PeerState)
	for i, line := range strings.Split(strings.TrimRight(output, "\r\n"), "\n") {
		if i == 0 {
			continue
		}
		f := strings.Split(strings.TrimRight(line, "\r"), "\t")
		if len(f) < 8 {
			continue
		}
		handshake, err1 := strconv.ParseInt(f[4], 10, 64)
		rx, err2 := strconv.ParseInt(f[5], 10, 64)
		tx, err3 := strconv.ParseInt(f[6], 10, 64)
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}
		endpoint := f[2]
		if endpoint == "(none)" {
			endpoint = ""
		}
		result[f[0]] = PeerState{
			Endpoint:        endpoint,
			LatestHandshake: handshake,
			Transfer:        Transfer{Rx: rx, Tx: tx},
		}
	}
	return result
}

// Transfers keeps only the byte counters of states.
func Transfers(states map[string]PeerState) map[string]Transfer {
	out := make(map[string]Transfer, len(states))
	for key, st := range states {
		out[key] = st.Transfer
	}
	return out
}

func GetPeerTransferMap(interfaceName string) (map[string]Transfer, error) {
	states, err := GetPeerStates(interfaceName)
	if err != nil {
		return nil, err
	}
	return Transfers(states), nil
}

func IsPeerOnline(lastHandshake int64) bool {
	if lastHandshake == 0 {
		return false
	}
	return time.Now().Unix()-lastHandshake < 180
}

// HandshakeAge returns the seconds since lastHandshake, or -1 if the peer never
// completed one.
func HandshakeAge(lastHandshake int64) int64 {
	if lastHandshake <= 0 {
		return -1
	}
	if age := time.Now().Unix() - lastHandshake; age > 0 {
		return age
	}
	return 0
}
