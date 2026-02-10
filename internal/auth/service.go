package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petieclark/pews/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db        *pgxpool.Pool
	jwtSecret []byte
}

type jwtClaims struct {
	middleware.Claims
	jwt.RegisteredClaims
}

type User struct {
	ID           string
	TenantID     string
	Email        string
	PasswordHash string
	Role         string
	Verified     bool
	CreatedAt    time.Time
}

func NewService(db *pgxpool.Pool, jwtSecret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

func (s *Service) VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s *Service) GenerateToken(userID, tenantID, email, role string) (string, error) {
	claims := &jwtClaims{
		Claims: middleware.Claims{
			UserID:   userID,
			TenantID: tenantID,
			Email:    email,
			Role:     role,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *Service) ValidateToken(tokenString string) (*middleware.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return &claims.Claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (s *Service) CreateUser(ctx context.Context, tenantID, email, password, role string) (*User, error) {
	hash, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		Verified:     false,
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO users (id, tenant_id, email, password_hash, role, verified) 
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.TenantID, user.Email, user.PasswordHash, user.Role, user.Verified,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, tenantID, email string) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(ctx,
		`SELECT id, tenant_id, email, password_hash, role, verified, created_at 
		 FROM users WHERE tenant_id = $1 AND email = $2`,
		tenantID, email,
	).Scan(&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &user.Role, &user.Verified, &user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
