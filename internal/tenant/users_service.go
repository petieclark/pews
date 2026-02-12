package tenant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) ListUsers(ctx context.Context, tenantID string) ([]UserResponse, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, email, role 
		 FROM users WHERE tenant_id = $1 ORDER BY created_at`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := []UserResponse{}
	for rows.Next() {
		var u UserResponse
		if err := rows.Scan(&u.ID, &u.Email, &u.Role); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		u.Name = u.Email // Use email as display name since no name column exists
		users = append(users, u)
	}

	if users == nil {
		users = []UserResponse{}
	}

	return users, nil
}

func (s *Service) InviteUser(ctx context.Context, tenantID, email, role string) (*UserResponse, error) {
	// Check if user already exists in this tenant
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE tenant_id = $1 AND email = $2)`,
		tenantID, email,
	).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// Create user with a temporary password (they'll need to reset)
	tempPassword := uuid.New().String()[:12]
	hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	id := uuid.New().String()
	_, err = s.db.Exec(ctx,
		`INSERT INTO users (id, tenant_id, email, password_hash, role, verified)
		 VALUES ($1, $2, $3, $4, $5, false)`,
		id, tenantID, email, string(hash), role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &UserResponse{
		ID:    id,
		Email: email,
		Role:  role,
	}, nil
}

func (s *Service) UpdateUserRole(ctx context.Context, tenantID, userID, role string) error {
	result, err := s.db.Exec(ctx,
		`UPDATE users SET role = $1 WHERE id = $2 AND tenant_id = $3`,
		role, userID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (s *Service) RemoveUser(ctx context.Context, tenantID, userID string) error {
	result, err := s.db.Exec(ctx,
		`DELETE FROM users WHERE id = $1 AND tenant_id = $2`,
		userID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
