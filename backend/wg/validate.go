package wg

import (
	"encoding/base64"
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
	if i := strings.LastIndex(name, "."); i >= 0 {
		suffix := name[i+1:]
		if len(suffix) >= 2 && isAlpha(suffix) {
			return fmt.Errorf("接口名不能是域名（%s），请填写系统网卡名，一般是 wg0", name)
		}
	}
	return nil
}

func isAlpha(s string) bool {
	for _, r := range s {
		if r < 'A' || (r > 'Z' && r < 'a') || r > 'z' {
			return false
		}
	}
	return s != ""
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

// hostRe matches a hostname (labels of letters, digits, '-' and '_').
var hostRe = regexp.MustCompile(`^[A-Za-z0-9_]([A-Za-z0-9_-]{0,61}[A-Za-z0-9_])?(\.[A-Za-z0-9_]([A-Za-z0-9_-]{0,61}[A-Za-z0-9_])?)*$`)

func isHostOrIP(s string) bool {
	return net.ParseIP(s) != nil || (len(s) <= 253 && hostRe.MatchString(s))
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
	if !isHostOrIP(host) {
		return fmt.Errorf("公网地址的主机名 %q 无效", host)
	}
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("公网地址端口无效，应在 1-65535 之间")
	}
	return nil
}

// ValidateDNS accepts an empty value or a comma separated list of IP addresses
// and search domains. Anything else could smuggle extra lines into the
// downloaded client config.
func ValidateDNS(dns string) error {
	dns = strings.TrimSpace(dns)
	if dns == "" {
		return nil
	}
	for _, item := range strings.Split(dns, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			return fmt.Errorf("DNS 格式无效，多个地址请用逗号分隔，例如 1.1.1.1, 8.8.8.8")
		}
		if !isHostOrIP(item) {
			return fmt.Errorf("DNS %q 无效，应为 IP 地址或域名", item)
		}
	}
	return nil
}

// NormalizeDNS trims and joins a validated DNS list as "a, b".
func NormalizeDNS(dns string) string {
	parts := strings.Split(dns, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, ", ")
}

func ValidateMTU(mtu int) error {
	if mtu != 0 && (mtu < 576 || mtu > 9000) {
		return fmt.Errorf("MTU 应为 0（自动）或 576-9000 之间的数值")
	}
	return nil
}

// ValidateKey checks a WireGuard key: 32 bytes of standard base64.
func ValidateKey(key string) error {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(key))
	if err != nil || len(raw) != 32 {
		return fmt.Errorf("密钥格式无效")
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
