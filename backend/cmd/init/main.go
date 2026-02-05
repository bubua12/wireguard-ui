package main

import (
	"fmt"
	"wireguard-ui/db"
)

func main() {
	if err := db.Init("wireguard.db"); err != nil {
		fmt.Println("Failed to init db:", err)
		return
	}

	if err := db.CreateUser("admin", "admin"); err != nil {
		fmt.Println("Failed to create user:", err)
		return
	}

	fmt.Println("Admin user created: admin/admin")
}
