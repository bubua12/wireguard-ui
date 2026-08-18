package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"wireguard-ui/config"
	"wireguard-ui/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(1)
	}
	if err := db.Init(cfg.DBPath); err != nil {
		fmt.Println("Failed to init db:", err)
		os.Exit(1)
	}

	count, err := db.GetUserCount()
	if err != nil {
		fmt.Println("Failed to check users:", err)
		os.Exit(1)
	}
	if count > 0 {
		fmt.Println("Admin user already exists, nothing to do.")
		os.Exit(0)
	}

	username := strings.TrimSpace(os.Getenv("WG_ADMIN_USER"))
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("WG_ADMIN_PASSWORD")
	if password == "" {
		password, err = randomPassword(16)
		if err != nil {
			fmt.Println("Failed to generate password:", err)
			os.Exit(1)
		}
	}

	if err := db.CreateUser(username, password); err != nil {
		fmt.Println("Failed to create user:", err)
		os.Exit(1)
	}

	fmt.Println("Admin user created.")
	fmt.Println("Username:", username)
	fmt.Println("Password:", password)
	fmt.Println("Store this password now; it will not be shown again.")
}

func randomPassword(n int) (string, error) {
	const chars = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = chars[int(buf[i])%len(chars)]
	}
	return string(buf), nil
}
