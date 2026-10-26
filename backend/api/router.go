package api

import (
	"io/fs"
	"net/http"
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
	r.Use(securityHeaders(), limitBody(maxBodyBytes))

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
		api.GET("/traffic", GetTrafficHistory)

		api.GET("/peers", GetPeers)
		api.GET("/peers/status", GetPeersStatus)
		api.GET("/peers/next-ip", GetNextPeerIP)
		api.POST("/peers", CreatePeer)
		api.PUT("/peers/:id", UpdatePeer)
		api.DELETE("/peers/:id", DeletePeer)
		api.POST("/peers/:id/toggle", TogglePeer)

		api.GET("/peers/:id/config", GetPeerConfig)
		api.GET("/peers/:id/qrcode", GetPeerQRCode)
		api.GET("/peers/:id/traffic", GetPeerTraffic)

		api.POST("/sync", SyncConfig)
		api.POST("/interface/:action", InterfaceControl)
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
	r.NoRoute(newStaticSite(sub).handle)
}

const maxBodyBytes = 1 << 20

// securityHeaders applies conservative browser protections. The app is a
// same-origin SPA, so everything may only load from 'self'; blob: and data:
// images cover the QR code, and inline styles are what Vue bindings produce.
func securityHeaders() gin.HandlerFunc {
	const csp = "default-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; " +
		"object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Content-Security-Policy", csp)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "same-origin")
		c.Next()
	}
}

func limitBody(max int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max)
		c.Next()
	}
}
