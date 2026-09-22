package middleware

import (
	"strings"
	"time"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// JWTOrDemoAuth validates a supplied bearer token. When no token is supplied,
// it assigns the documented demo identity so the catalog demo remains usable.
func JWTOrDemoAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(constants.AuthorizationHeader)
		if header == "" {
			setIdentity(c, constants.DemoUserID, constants.RoleUser)
			c.Next()
			return
		}
		if !strings.HasPrefix(header, constants.BearerPrefix) {
			rejectUnauthorized(c)
			return
		}
		tokenString := strings.TrimPrefix(header, constants.BearerPrefix)
		parsed, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperrors.ErrUnauthorized
			}
			return []byte(secret), nil
		})
		if err != nil || !parsed.Valid {
			rejectUnauthorized(c)
			return
		}
		payload, ok := parsed.Claims.(*claims)
		if !ok || payload.Subject == "" || !validRole(payload.Role) {
			rejectUnauthorized(c)
			return
		}
		setIdentity(c, payload.Subject, payload.Role)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(constants.RoleContextKey)
		for _, allowed := range roles {
			if role == allowed {
				c.Next()
				return
			}
		}
		rejectUnauthorized(c)
	}
}

func setIdentity(c *gin.Context, userID, role string) {
	c.Set(constants.UserIDContextKey, userID)
	c.Set(constants.RoleContextKey, role)
}

func validRole(role string) bool {
	return role == constants.RoleAdmin || role == constants.RoleSupplier || role == constants.RoleUser
}

func rejectUnauthorized(c *gin.Context) {
	_ = c.Error(apperrors.ErrUnauthorized)
	c.Abort()
}

// NewDemoToken is intentionally small and supports local RBAC smoke tests.
func NewDemoToken(secret, subject, role string) (string, error) {
	payload := claims{Role: role, RegisteredClaims: jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(constants.DemoTokenLifetime)),
	}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString([]byte(secret))
}
