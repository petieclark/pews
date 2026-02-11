package middleware

import (
	"context"
	"fmt"
	"net/http"
)

// TenantRLS middleware ensures tenant ID from JWT claims is stored in request context
// The database pool's AfterConnect hook will automatically set app.current_tenant_id
func TenantRLS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract claims from context (set by auth middleware)
		claims, ok := GetClaims(r.Context())
		if !ok {
			// No claims means this is a public route - pass through
			next.ServeHTTP(w, r)
			return
		}

		// Get tenant ID from claims
		tenantID := claims.TenantID
		if tenantID == "" {
			http.Error(w, "Missing tenant_id in token claims", http.StatusUnauthorized)
			return
		}

		// Store tenant ID in context for the database hook to use
		ctx := context.WithValue(r.Context(), TenantIDContextKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTenantIDFromClaims extracts tenant ID from JWT claims in context
func GetTenantIDFromClaims(ctx context.Context) (string, error) {
	claims, ok := GetClaims(ctx)
	if !ok {
		return "", fmt.Errorf("no claims in context")
	}
	if claims.TenantID == "" {
		return "", fmt.Errorf("no tenant_id in claims")
	}
	return claims.TenantID, nil
}
