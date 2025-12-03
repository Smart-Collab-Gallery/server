package pkg

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT Claims
type Claims struct {
	UserID      int64  `json:"user_id"`
	UserAccount string `json:"user_account"`
	UserRole    string `json:"user_role"`
	jwt.RegisteredClaims
}

// JWTManager JWT 管理器
type JWTManager struct {
	secret string
	expire time.Duration
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(secret string, expire time.Duration) *JWTManager {
	return &JWTManager{
		secret: secret,
		expire: expire,
	}
}

// GenerateToken 生成 JWT Token
func (m *JWTManager) GenerateToken(userID int64, userAccount string, userRole string) (string, error) {
	claims := Claims{
		UserID:      userID,
		UserAccount: userAccount,
		UserRole:    userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "smart-collab-gallery",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

// ParseToken 解析 JWT Token
func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}
