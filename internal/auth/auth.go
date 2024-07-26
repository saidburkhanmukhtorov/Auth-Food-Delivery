package auth

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"github.com/time_capsule/Auth-Servic-Timecapsule/config"
	"github.com/time_capsule/Auth-Servic-Timecapsule/internal/models"
)

// JWTManager manages JWT tokens.
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTManager creates a new JWTManager.
func NewJWTManager(cfg *config.Config) *JWTManager {
	return &JWTManager{
		secretKey:     cfg.JWTSecretKey,
		tokenDuration: time.Duration(cfg.JWTExpiry) * time.Minute,
	}
}

// Generate generates and signs a new JWT token for the given user.
func (manager *JWTManager) Generate(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"id":   user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(manager.tokenDuration).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))
}

// Verify verifies the signature of the given JWT token and returns the user claims if valid.
func (manager *JWTManager) Verify(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// UserClaims represents the claims embedded in a JWT token.
type UserClaims struct {
	jwt.StandardClaims
	ID   string `json:"id"`
	Role string `json:"role"`
	Iat  int64  `json:"iat"`
}

// GetUserID returns the user ID from the token claims.
func (c *UserClaims) GetUserID() string {
	return c.ID
}

// GetUserRole returns the user role from the token claims.
func (c *UserClaims) GetUserRole() string {
	return c.Role
}

// GetUserRole returns the user role from the token claims.
func (c *UserClaims) GetIat() int64 {
	return c.Iat
}

// HashPassword hashes the given password using bcrypt.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPasswordHash compares a plain text password with a bcrypt hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateOTP generates a 6-digit numeric OTP.
func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
