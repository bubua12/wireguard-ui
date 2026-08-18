package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultListen    = "127.0.0.1:8081"
	DefaultDBPath    = "wireguard.db"
	DefaultConfigDir = "/etc/wireguard"
	minSecretBytes   = 16
)

type Config struct {
	Listen      string
	DBPath      string
	JWTSecret   []byte
	TrustProxy  bool
	ConfigDir   string
	CORSOrigins []string
	Debug       bool
}

func Load() (*Config, error) {
	cfg := &Config{
		Listen:      envOr("WG_LISTEN", DefaultListen),
		DBPath:      envOr("WG_DB_PATH", DefaultDBPath),
		TrustProxy:  envBool("WG_TRUST_PROXY", false),
		ConfigDir:   envOr("WG_CONFIG_DIR", DefaultConfigDir),
		CORSOrigins: splitCSV(envOr("WG_CORS_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")),
		Debug:       envBool("WG_DEBUG", false),
	}

	secret, err := loadJWTSecret(cfg.DBPath)
	if err != nil {
		return nil, err
	}
	cfg.JWTSecret = secret
	return cfg, nil
}

func loadJWTSecret(dbPath string) ([]byte, error) {
	if s := strings.TrimSpace(os.Getenv("WG_JWT_SECRET")); s != "" {
		if len(s) < minSecretBytes {
			return nil, fmt.Errorf("WG_JWT_SECRET must be at least %d characters", minSecretBytes)
		}
		return []byte(s), nil
	}

	secretPath := os.Getenv("WG_JWT_SECRET_FILE")
	if secretPath == "" {
		secretPath = filepath.Join(filepath.Dir(absPath(dbPath)), ".jwt_secret")
	}

	if data, err := os.ReadFile(secretPath); err == nil {
		secret := strings.TrimSpace(string(data))
		if len(secret) >= minSecretBytes {
			return []byte(secret), nil
		}
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("generate jwt secret: %w", err)
	}
	secret := hex.EncodeToString(raw)
	if err := os.MkdirAll(filepath.Dir(secretPath), 0700); err != nil {
		return nil, fmt.Errorf("create jwt secret dir: %w", err)
	}
	if err := os.WriteFile(secretPath, []byte(secret+"\n"), 0600); err != nil {
		return nil, fmt.Errorf("write jwt secret: %w", err)
	}
	return []byte(secret), nil
}

func absPath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return abs
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
