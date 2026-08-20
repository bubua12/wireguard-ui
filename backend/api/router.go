package api

import (
	"io/fs"
	"net/http"
	"strings"
	"wireguard-ui/config"
	"wireguard-ui/web"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	if cfg.Debug {
		r.Use(gin.Logger())
	}

	if cfg.TrustProxy {
		_ = r.SetTrustedProxies([]string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})
	} else {
		_ = r.SetTrustedProxies(nil)
	}

	corsCfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
	}
	if len(cfg.CORSOrigins) == 1 && cfg.CORSOrigins[0] == "*" {
		corsCfg.AllowAllOrigins = true
	} else {
		corsCfg.AllowOrigins = cfg.CORSOrigins
	}
	r.Use(cors.New(corsCfg))

	SetJWTSecret(cfg.JWTSecret)

	r.POST("/api/login", Login)
	r.POST("/api/register", Register)
	r.GET("/api/init", CheckInit)
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	api := r.Group("/api")
	api.Use(AuthMiddleware())
	{
		api.GET("/server", GetServer)
		api.POST("/server", CreateServer)
		api.PUT("/server", UpdateServer)
		api.GET("/status", GetStatus)

		api.GET("/peers", GetPeers)
		api.GET("/peers/status", GetPeersStatus)
		api.POST("/peers", CreatePeer)
		api.PUT("/peers/:id", UpdatePeer)
		api.DELETE("/peers/:id", DeletePeer)
		api.POST("/peers/:id/toggle", TogglePeer)

		api.GET("/peers/:id/config", GetPeerConfig)
		api.GET("/peers/:id/qrcode", GetPeerQRCode)
		api.GET("/peers/:id/traffic", GetPeerTraffic)

		api.POST("/sync", SyncConfig)
		api.POST("/import", ImportConfig)
		api.POST("/change-password", ChangePassword)
	}

	serveFrontend(r)
	return r
}

func serveFrontend(r *gin.Engine) {
	sub, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		return
	}
	fileServer := http.FileServer(http.FS(sub))
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		if path != "" {
			if f, err := sub.Open(path); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
