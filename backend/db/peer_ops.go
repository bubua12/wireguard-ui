package db

import "time"

func UpdatePeer(id int64, name string, enabled bool) error {
	_, err := DB.Exec(`UPDATE peers SET name=?, enabled=?, updated_at=? WHERE id=?`,
		name, enabled, time.Now(), id)
	return err
}

func DeletePeer(id int64) error {
	_, err := DB.Exec(`DELETE FROM peers WHERE id = ?`, id)
	return err
}

func TogglePeer(id int64, enabled bool) error {
	_, err := DB.Exec(`UPDATE peers SET enabled=?, updated_at=? WHERE id=?`,
		enabled, time.Now(), id)
	return err
}

func GetNextAvailableIP(serverID int64, subnet string) (string, error) {
	// 简单实现：获取最大IP+1
	var maxIP string
	err := DB.QueryRow(`SELECT allowed_ips FROM peers WHERE server_id = ? ORDER BY id DESC LIMIT 1`, serverID).Scan(&maxIP)
	if err != nil {
		// 没有peer，返回第一个可用IP（.2，因为.1是服务器）
		return incrementIP(subnet, 2), nil
	}
	return incrementIP(maxIP, 1), nil
}

func incrementIP(ip string, inc int) string {
	// 简化实现：假设是 10.0.0.x/32 格式
	// 实际项目中应该用 net 包处理
	var a, b, c, d int
	var mask string
	n, _ := parseIP(ip, &a, &b, &c, &d, &mask)
	if n >= 4 {
		d += inc
		if mask != "" {
			return formatIP(a, b, c, d, mask)
		}
		return formatIPNoMask(a, b, c, d)
	}
	return ip
}

func parseIP(ip string, a, b, c, d *int, mask *string) (int, error) {
	// 尝试解析带掩码的IP
	n, err := scanIP(ip, a, b, c, d, mask)
	return n, err
}

func scanIP(ip string, a, b, c, d *int, mask *string) (int, error) {
	var m string
	n, _ := scanIPWithMask(ip, a, b, c, d, &m)
	*mask = m
	return n, nil
}

func scanIPWithMask(ip string, a, b, c, d *int, mask *string) (int, error) {
	// 简单解析
	for i, ch := range ip {
		if ch == '/' {
			*mask = ip[i:]
			ip = ip[:i]
			break
		}
	}
	n := 0
	parts := splitIP(ip)
	if len(parts) >= 1 { *a = atoi(parts[0]); n++ }
	if len(parts) >= 2 { *b = atoi(parts[1]); n++ }
	if len(parts) >= 3 { *c = atoi(parts[2]); n++ }
	if len(parts) >= 4 { *d = atoi(parts[3]); n++ }
	return n, nil
}

func splitIP(ip string) []string {
	var parts []string
	var current string
	for _, ch := range ip {
		if ch == '.' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func atoi(s string) int {
	n := 0
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		}
	}
	return n
}

func formatIP(a, b, c, d int, mask string) string {
	return formatIPNoMask(a, b, c, d) + mask
}

func formatIPNoMask(a, b, c, d int) string {
	return itoa(a) + "." + itoa(b) + "." + itoa(c) + "." + itoa(d)
}

func itoa(n int) string {
	if n == 0 { return "0" }
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
