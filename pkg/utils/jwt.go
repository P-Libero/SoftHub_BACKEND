package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 是 Token 的载荷（payload）：自定义字段 + 标准字段（过期时间等）
type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

const tokenTTL = 30 * time.Minute

func getSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET 未配置")
	}
	return []byte(secret), nil
}

// 签发Token

func GenerateToken(userID uint64, username string, role string) (string, error) {
	secret, err := getSecret()
	if err != nil {
		return "", err
	}
	claims := Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "SoftHub",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) //这里是创建jwt对象
	return token.SignedString(secret)                          //这里真正进行签发Token
}

func ParseToken(tokenString string) (*Claims, error) {
	secret, err := getSecret()
	if err != nil {
		return nil, err
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	// Valid表示这个token是否通过校验
	if !token.Valid {
		return nil, errors.New("token无效")
	}
	return claims, nil
}
