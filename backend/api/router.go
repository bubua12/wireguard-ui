package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// 公开路由
	r.POST("/api/login", Login)
	r.POST("/api/register", Register)
	r.GET("/api/init", CheckInit)

	// 需要认证的 API 路由组
	api := r.Group("/api")
	api.Use(AuthMiddleware())
	{
		// 服务器配置
		api.GET("/server", GetServer)
		api.POST("/server", CreateServer)
		api.PUT("/server", UpdateServer)

		// 客户端管理
		api.GET("/peers", GetPeers)
		api.GET("/peers/status", GetPeersStatus)
		api.POST("/peers", CreatePeer)
		api.PUT("/peers/:id", UpdatePeer)
		api.DELETE("/peers/:id", DeletePeer)
		api.POST("/peers/:id/toggle", TogglePeer)

		// 配置下载
		api.GET("/peers/:id/config", GetPeerConfig)
		api.GET("/peers/:id/qrcode", GetPeerQRCode)

		// 同步配置到系统
		api.POST("/sync", SyncConfig)

		// 导入现有配置
		api.POST("/import", ImportConfig)

		// 用户管理
		api.POST("/change-password", ChangePassword)
	}

	return r
}
