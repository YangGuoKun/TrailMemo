package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/trailmemo/internal/config"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"github.com/trailmemo/pkg/response"
	"go.uber.org/zap"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

var jwtSecret []byte

// InitJWT 初始化JWT，验证Secret长度（至少32位）
func InitJWT() {
	secret := config.Get().JWT.Secret
	if len(secret) < 32 {
		panic("JWT secret must be at least 32 characters")
	}
	jwtSecret = []byte(secret)
}

// GenerateToken 生成JWT
func GenerateToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.Get().JWT.ExpireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // 生成JWT
	return token.SignedString(jwtSecret)
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (*Claims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ") // 移除Bearer前缀
	if tokenString == "" {
		return nil, errors.New("token is empty")
	}
	// 解析JWT
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid { // 验证JWT是否有效
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		//tokenString := c.GetHeader("Authorization")

		// 1. 检查是否是忽略路径
		path := c.Request.URL.Path
		for _, ignorePath := range config.Get().JWT.IgnorePaths {
			if path == ignorePath || strings.HasPrefix(path, ignorePath+"/") {
				c.Next() // 继续执行后续中间件
				return
			}
		}
		var token string
		queryToken := c.Query("token")
		authHeader := c.GetHeader("Authorization")

		if queryToken != "" {
			// 从查询参数获取 token（用于 WebSocket）
			token = queryToken
		} else if authHeader != "" {
			// 从请求头获取 token
			parts := strings.SplitN(authHeader, " ", 2)
			if !(len(parts) == 2 && parts[0] == "Bearer") {
				platformlogger.FromGinContext(c).Warn("auth_failed",
					zap.String("event", "auth_failed"),
					zap.String("module", "auth"),
					zap.String("error_kind", "auth"),
					zap.String("reason", "invalid_format"),
				)
				response.Unauthorized(c, "Authorization header format must be Bearer {token}")
				c.Abort()
				return
			}
			token = parts[1]
		} else {
			platformlogger.FromGinContext(c).Warn("auth_failed",
				zap.String("event", "auth_failed"),
				zap.String("module", "auth"),
				zap.String("error_kind", "auth"),
				zap.String("reason", "missing_token"),
			)
			response.Unauthorized(c, "authorization token is required")
			c.Abort()
			return
		}

		claims, err := ParseToken(token) // 解析JWT
		if err != nil {
			// 隐藏具体错误信息，防止泄露签名细节
			platformlogger.FromGinContext(c).Warn("auth_failed",
				zap.String("event", "auth_failed"),
				zap.String("module", "auth"),
				zap.String("error_kind", "auth"),
				zap.String("reason", "invalid_or_expired"),
			)
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID) // 存储用户ID到上下文
		platformlogger.WithGinFields(c, zap.String("user_id", claims.UserID))
		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return userID.(string)
}
