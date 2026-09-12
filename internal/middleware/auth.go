package middleware

import (
	"net/http"
	"strings"

	pkgJwt "github.com/doanvuduy170921/quick-link/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const AuthorizationHeaderKey = "authorization"
const AuthorizationTypeBearer = "bearer"
const AuthorizationPayloadKey = "user_id"

func AuthMiddleware(jwtManager *pkgJwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		if len(authHeader) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is missing"})
			c.Abort()
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != AuthorizationTypeBearer {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unsupported authorization type"})
			c.Abort()
			return
		}

		accessToken := fields[1]
		claims, err := jwtManager.ValidateToken(accessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Lưu user_id vào context để các handler phía sau lấy dùng
		c.Set(AuthorizationPayloadKey, claims.UserID)
		c.Next()
	}
}

func OptionalAuthMiddleware(jwtManager *pkgJwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		if len(authHeader) == 0 {
			c.Next()
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 || strings.ToLower(fields[0]) != AuthorizationTypeBearer {
			c.Next()
			return
		}

		claims, err := jwtManager.ValidateToken(fields[1])
		if err != nil {
			c.Next()
			return
		}

		c.Set(AuthorizationPayloadKey, claims.UserID)
		c.Next()
	}
}
