package jwt

import (
	"errors"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	JTI    string `json:"jti"`
	gojwt.RegisteredClaims
}

type JWTManager struct {
	secretKey        []byte
	accessTokenExpy  time.Duration
	refreshTokenExpy time.Duration
}

func NewJWTManager(secretKey string, accessTokenExpy, refreshTokenExpy time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:        []byte(secretKey),
		accessTokenExpy:  accessTokenExpy,
		refreshTokenExpy: refreshTokenExpy,
	}
}

// GenerateAccessToken creates a new access token and returns (token, jti, error)
func (j *JWTManager) GenerateAccessToken(userID uuid.UUID, role, email string) (string, string, error) {
	tokenJTI := uuid.New().String()

	claims := &Claims{
		UserID: userID.String(),
		Role:   role,
		Email:  email,
		JTI:    tokenJTI,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(j.accessTokenExpy)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
			NotBefore: gojwt.NewNumericDate(time.Now()),
			Issuer:    "simtas-filkom",
		},
	}

	tokenObj := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	tokenString, err := tokenObj.SignedString(j.secretKey)
	if err != nil {
		return "", "", err
	}

	return tokenString, tokenJTI, nil
}

// GenerateRefreshToken creates a long-lived refresh token
func (j *JWTManager) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	tokenJTI := uuid.New().String()

	claims := &Claims{
		UserID: userID.String(),
		JTI:    tokenJTI,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(j.refreshTokenExpy)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
			NotBefore: gojwt.NewNumericDate(time.Now()),
			Issuer:    "simtas-filkom-refresh",
		},
	}

	tokenObj := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return tokenObj.SignedString(j.secretKey)
}

// ValidateToken parses and validates a JWT string
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenString, &Claims{}, func(token *gojwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return j.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, gojwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// AccessTokenExpiry returns the configured access token duration
func (j *JWTManager) AccessTokenExpiry() time.Duration {
	return j.accessTokenExpy
}
