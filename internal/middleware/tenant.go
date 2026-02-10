package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	TenantIDContextKey contextKey = "tenant_id"
	ClaimsContextKey   contextKey = "claims"
)

// TenantExtractor extracts tenant ID from X-Tenant-ID header or subdomain
func TenantExtractor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")

		// If not in header, try to extract from subdomain
		if tenantID == "" {
			host := r.Host
			parts := strings.Split(host, ".")
			if len(parts) > 2 {
				// Assume first part is tenant slug
				tenantID = parts[0]
			}
		}

		if tenantID != "" {
			ctx := context.WithValue(r.Context(), TenantIDContextKey, tenantID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDContextKey).(string); ok {
		return tenantID
	}
	return ""
}
