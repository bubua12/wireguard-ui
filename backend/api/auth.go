package api

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	"wireguard-ui/db"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("wireguard-ui-secret-key-change-in-production")

// IP 登录限制
const (
	maxLoginAttempts = 10              // 最大失败次数
	lockDuration     = 30 * time.Minute // 锁定时长
)

type loginAttempt struct {
	count    int
	lockedAt time.Time
}

var (
	loginAttempts = make(map[string]*loginAttempt)
	attemptsMutex sync.RWMutex
)

// 获取客户端真实IP
func getClientIP(c *gin.Context) string {
	// 优先从 X-Real-IP 获取（nginx 代理）
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	// 其次从 X-Forwarded-For 获取
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return c.ClientIP()
}

// 检查IP是否被锁定
func isIPLocked(ip string) (bool, time.Duration) {
	attemptsMutex.RLock()
	defer attemptsMutex.RUnlock()

	if attempt, exists := loginAttempts[ip]; exists {
		if attempt.count >= maxLoginAttempts {
			remaining := lockDuration - time.Since(attempt.lockedAt)
			if remaining > 0 {
				return true, remaining
			}
		}
	}
	return false, 0
}

// 记录登录失败
func recordFailedLogin(ip string) int {
	attemptsMutex.Lock()
	defer attemptsMutex.Unlock()

	if _, exists := loginAttempts[ip]; !exists {
		loginAttempts[ip] = &loginAttempt{}
	}

	attempt := loginAttempts[ip]
	// 如果锁定已过期，重置计数
	if attempt.count >= maxLoginAttempts && time.Since(attempt.lockedAt) > lockDuration {
		attempt.count = 0
	}

	attempt.count++
	if attempt.count >= maxLoginAttempts {
		attempt.lockedAt = time.Now()
	}
	return attempt.count
}

// 清除登录失败记录
func clearFailedLogin(ip string) {
	attemptsMutex.Lock()
	defer attemptsMutex.Unlock()
	delete(loginAttempts, ip)
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func Login(c *gin.Context) {
	ip := getClientIP(c)

	// 检查IP是否被锁定
	if locked, remaining := isIPLocked(ip); locked {
		minutes := int(remaining.Minutes()) + 1
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":             fmt.Sprintf("登录失败次数过多，请%d分钟后再试", minutes),
			"locked":            true,
			"remaining_minutes": minutes,
		})
		return
	}

	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := db.GetUserByUsername(req.Username)
	if err != nil || !db.ValidatePassword(user, req.Password) {
		count := recordFailedLogin(ip)
		remaining := maxLoginAttempts - count
		if remaining <= 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "登录失败次数过多，IP已被锁定30分钟",
				"locked": true,
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户名或密码错误",
				"remaining_attempts": remaining,
			})
		}
		return
	}

	// 登录成功，清除失败记录
	clearFailedLogin(ip)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString, "username": user.Username})
}

func Register(c *gin.Context) {
	count, _ := db.GetUserCount()
	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Registration disabled"})
		return
	}

	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := db.CreateUser(req.Username, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created"})
}

func CheckInit(c *gin.Context) {
	count, _ := db.GetUserCount()
	c.JSON(http.StatusOK, gin.H{"initialized": count > 0})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func ChangePassword(c *gin.Context) {
	var req ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少6位"})
		return
	}

	auth := c.GetHeader("Authorization")
	tokenString := strings.TrimPrefix(auth, "Bearer ")
	token, _ := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	claims := token.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	user, err := db.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户不存在"})
		return
	}

	if !db.ValidatePassword(user, req.OldPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "原密码错误"})
		return
	}

	if err := db.UpdatePassword(username, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "修改失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}
