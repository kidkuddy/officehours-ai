// Package auth implements JWT login and request authentication middleware.
// See BUILD_SPEC §4: JWT in Authorization: Bearer (or httpOnly cookie),
// no public register endpoint (accounts come from `ohctl seed user`).
package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ctxKey is the private context key type for the authenticated user id.
type ctxKey string

const userIDKey ctxKey = "auth.user_id"

// CookieName is the httpOnly cookie that may carry the JWT.
const CookieName = "oh_token"

// TokenTTL is the lifetime of an issued token.
const TokenTTL = 7 * 24 * time.Hour

// HashPassword returns a bcrypt hash of the plaintext password.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword compares a bcrypt hash with a plaintext password.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// Claims is the JWT claim set.
type Claims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// IssueToken signs a JWT for the given user.
func IssueToken(secret, userID, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenTTL)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

// ParseToken validates a signed token and returns its claims.
func ParseToken(secret, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// Middleware authenticates requests via Bearer header or the httpOnly cookie.
// On success the user id is stored in the request context (use UserID to read).
func Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractToken(r)
			if tokenStr == "" {
				http.Error(w, `{"error":"missing token"}`, http.StatusUnauthorized)
				return
			}
			claims, err := ParseToken(secret, tokenStr)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Require wraps a handler with the auth middleware for the given secret.
func Require(secret string, h http.HandlerFunc) http.HandlerFunc {
	mw := Middleware(secret)
	wrapped := mw(h)
	return wrapped.ServeHTTP
}

func extractToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	if c, err := r.Cookie(CookieName); err == nil {
		return c.Value
	}
	return ""
}

// UserID returns the authenticated user id from the request context.
func UserID(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}
