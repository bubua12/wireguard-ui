package model

import "time"

// Server WireGuard 服务器配置
type Server struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	PrivateKey string    `json:"-"`
	PublicKey  string    `json:"public_key"`
	Address    string    `json:"address"`    // 例如: 10.0.0.1/24
	ListenPort int       `json:"listen_port"`
	Endpoint   string    `json:"endpoint"`   // 公网地址:端口
	DNS        string    `json:"dns"`
	MTU        int       `json:"mtu"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Peer 客户端配置
type Peer struct {
	ID              int64     `json:"id"`
	ServerID        int64     `json:"server_id"`
	Name            string    `json:"name"`
	PrivateKey      string    `json:"-"`
	PublicKey       string    `json:"public_key"`
	PresharedKey    string    `json:"-"`
	AllowedIPs      string    `json:"allowed_ips"`      // 分配的IP
	PersistentKeepalive int   `json:"persistent_keepalive"`
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// User 管理员用户
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
