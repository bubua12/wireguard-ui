package db

import (
	"fmt"
	"strings"
	"wireguard-ui/model"

	"golang.org/x/crypto/bcrypt"
)

const minPasswordLen = 8

// dummyHash is a valid bcrypt hash used to keep verify timing closer when the user is missing.
const dummyHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

func CreateUser(username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if len(password) < minPasswordLen {
		return fmt.Errorf("密码至少%d位", minPasswordLen)
	}
	if password == "admin" {
		return fmt.Errorf("不能使用默认弱密码")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = DB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, string(hash))
	return err
}

func GetUserByUsername(username string) (*model.User, error) {
	u := &model.User{}
	err := DB.QueryRow("SELECT id, username, password, created_at FROM users WHERE username = ?", username).
		Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func ValidatePassword(user *model.User, password string) bool {
	if user == nil {
		_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password))
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func GetUserCount() (int, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func UpdatePassword(username, newPassword string) error {
	if len(newPassword) < minPasswordLen {
		return fmt.Errorf("密码至少%d位", minPasswordLen)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = DB.Exec("UPDATE users SET password = ? WHERE username = ?", string(hash), username)
	return err
}
