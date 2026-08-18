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

var jwtSecret []byte

func SetJWTSecret(secret []byte) {
	jwtSecret = append([]byte(nil), secret...)
}

const (
	maxLoginAttempts = 10
	lockDuration     = 30 * time.Minute
	attemptIdleFor   = 30 * time.Minute
)

type loginAttempt struct {
	count    int
	lockedAt time.Time
	lastAt   time.Time
}

var (
	loginAttempts = make(map[string]*loginAttempt)
	attemptsMutex sync.Mutex
)

func getClientIP(c *gin.Context) string {
	return c.ClientIP()
}

func cleanupAttemptsLocked(now time.Time) {
	for ip, attempt := range loginAttempts {
		if now.Sub(attempt.lastAt) > attemptIdleFor {
			delete(loginAttempts, ip)
			continue
		}
		if attempt.count >= maxLoginAttempts && now.Sub(attempt.lockedAt) > lockDuration {
			delete(loginAttempts, ip)
		}
	}
}

func isIPLocked(ip string) (bool, time.Duration) {
	attemptsMutex.Lock()
	defer attemptsMutex.Unlock()
	cleanupAttemptsLocked(time.Now())

	if attempt, exists := loginAttempts[ip]; exists {
		if attempt.count >= maxLoginAttempts {
			remaining := lockDuration - time.Since(attempt.lockedAt)
			if remaining > 0 {
				return true, remaining
			}
			delete(loginAttempts, ip)
		}
	}
	return false, 0
}

func recordFailedLogin(ip string) int {
	attemptsMutex.Lock()
	defer attemptsMutex.Unlock()
	now := time.Now()
	cleanupAttemptsLocked(now)

	attempt, exists := loginAttempts[ip]
	if !exists {
		attempt = &loginAttempt{}
		loginAttempts[ip] = attempt
	}
	if attempt.count >= maxLoginAttempts && now.Sub(attempt.lockedAt) > lockDuration {
		attempt.count = 0
	}
	attempt.count++
	attempt.lastAt = now
	if attempt.count >= maxLoginAttempts {
		attempt.lockedAt = now
	}
	return attempt.count
}

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
	Password string `json:"password" binding:"required,min=8"`
}

func Login(c *gin.Context) {
	ip := getClientIP(c)

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
		if err != nil {
			_ = db.ValidatePassword(nil, req.Password)
		}
		count := recordFailedLogin(ip)
		remaining := maxLoginAttempts - count
		if remaining <= 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":  "登录失败次数过多，IP已被锁定30分钟",
				"locked": true,
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":              "用户名或密码错误",
				"remaining_attempts": remaining,
			})
		}
		return
	}

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
	count, err := db.GetUserCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check users"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Registration disabled"})
		return
	}

	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名必填，密码至少8位"})
		return
	}

	if err := db.CreateUser(req.Username, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created"})
}

func CheckInit(c *gin.Context) {
	count, err := db.GetUserCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check users"})
		return
	}
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
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		if username, _ := claims["username"].(string); username != "" {
			c.Set("username", username)
		}
		c.Next()
	}
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func ChangePassword(c *gin.Context) {
	var req ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少8位"})
		return
	}

	usernameVal, ok := c.Get("username")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	username, _ := usernameVal.(string)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}
