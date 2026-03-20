package security

import (
	"MuXi/2026-MuxiShooter-Backend/models"
	"MuXi/2026-MuxiShooter-Backend/utils"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var ErrJWTWrongSigningMethod = errors.New("无效的签名算法")

type BcryptPasswordHasher struct{}

func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{}
}

func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	return utils.Hashtool(password)
}

func (h *BcryptPasswordHasher) Compare(hashedPassword, password string) error {
	return utils.ComparePassword(hashedPassword, password)
}

type JWTTokenService struct {
	secret []byte
}

func NewJWTTokenService(secret []byte) *JWTTokenService {
	return &JWTTokenService{secret: secret}
}

func (s *JWTTokenService) GenerateToken(user models.User) (string, time.Time, error) {
	return utils.GenerateToken(user, s.secret)
}

func (s *JWTTokenService) ParseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrJWTWrongSigningMethod
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("无效的token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("无效的token声明")
	}
	return claims, nil
}
