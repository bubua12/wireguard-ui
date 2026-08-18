package model

import "time"

// Server WireGuard 服务器配置
type Server struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Interface  string    `json:"interface"`
	PrivateKey string    `json:"-"`
	PublicKey  string    `json:"public_key"`
	Address    string    `json:"address"`
	ListenPort int       `json:"listen_port"`
	Endpoint   string    `json:"endpoint"`
	DNS        string    `json:"dns"`
	MTU        int       `json:"mtu"`
	FullTunnel bool      `json:"full_tunnel"`
	EnableNAT  bool      `json:"enable_nat"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Peer 客户端配置
type Peer struct {
	ID                  int64     `json:"id"`
	ServerID            int64     `json:"server_id"`
	Name                string    `json:"name"`
	PrivateKey          string    `json:"-"`
	PublicKey           string    `json:"public_key"`
	PresharedKey        string    `json:"-"`
	AllowedIPs          string    `json:"allowed_ips"`
	PersistentKeepalive int       `json:"persistent_keepalive"`
	Enabled             bool      `json:"enabled"`
	HasPrivateKey       bool      `json:"has_private_key"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// User 管理员用户
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
