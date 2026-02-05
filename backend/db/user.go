package db

import (
	"wireguard-ui/model"

	"golang.org/x/crypto/bcrypt"
)

func CreateUser(username, password string) error {
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
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func GetUserCount() (int, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func UpdatePassword(username, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = DB.Exec("UPDATE users SET password = ? WHERE username = ?", string(hash), username)
	return err
}
