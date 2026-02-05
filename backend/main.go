package main

import (
	"log"
	"wireguard-ui/api"
	"wireguard-ui/db"
)

func main() {
	// 初始化数据库
	if err := db.Init("wireguard.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 启动 API 服务
	r := api.SetupRouter()
	log.Println("WireGuard UI starting on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
