package wg

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Linux IFNAMSIZ is 16 including the trailing NUL.
var ifaceNameRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_=+.-]{0,14}$`)

func ValidateInterfaceName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("接口名不能为空")
	}
	if !ifaceNameRe.MatchString(name) {
		return fmt.Errorf("接口名 %q 无效，需为 1-15 位、以字母开头，仅含字母数字和 _ = + . -", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("接口名无效")
	}
	return nil
}

func SanitizeDownloadName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "client"
	}
	var b strings.Builder
	for _, r := range name {
		if r == '.' || r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	out := strings.Trim(b.String(), "._")
	if out == "" {
		return "client"
	}
	if len(out) > 64 {
		out = out[:64]
	}
	return out
}

func ValidateImportPath(configDir, p string) (string, error) {
	configDir = filepath.Clean(configDir)
	if p == "" {
		p = filepath.Join(configDir, "wg0.conf")
	}
	p = strings.TrimSpace(p)
	if !filepath.IsAbs(p) {
		p = filepath.Join(configDir, p)
	}
	p = filepath.Clean(p)

	rel, err := filepath.Rel(configDir, p)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("配置文件必须位于 %s 目录内", configDir)
	}
	if filepath.Ext(p) != ".conf" {
		return "", fmt.Errorf("配置文件必须是 .conf")
	}
	base := strings.TrimSuffix(filepath.Base(p), ".conf")
	if err := ValidateInterfaceName(base); err != nil {
		return "", err
	}
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("无法读取配置文件: %w", err)
	}
	return p, nil
}

func InterfaceNameFromPath(p string) string {
	return strings.TrimSuffix(filepath.Base(p), ".conf")
}

func ValidateListenPort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("监听端口必须在 1-65535 之间")
	}
	return nil
}

func ValidateEndpoint(endpoint string) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return fmt.Errorf("公网地址不能为空")
	}
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return fmt.Errorf("公网地址格式应为 host:port，例如 vpn.example.com:51820")
	}
	if host == "" {
		return fmt.Errorf("公网地址缺少主机名")
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("公网地址端口无效")
	}
	return nil
}

func ValidateServerAddress(addr string) error {
	ip, ipNet, err := net.ParseCIDR(strings.TrimSpace(addr))
	if err != nil {
		return fmt.Errorf("内网地址必须是 CIDR，例如 10.0.0.1/24")
	}
	if ip.To4() == nil {
		return fmt.Errorf("暂不支持 IPv6 服务器地址")
	}
	ones, bits := ipNet.Mask.Size()
	if bits != 32 || ones > 30 {
		return fmt.Errorf("内网网段过小，至少需要 /30")
	}
	return nil
}
