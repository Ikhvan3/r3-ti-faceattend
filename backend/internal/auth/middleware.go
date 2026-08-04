package auth

import (
	"context"
	"net/http"
	"strings"

	"r3-ti-faceattend/backend/internal/user"
)

type contextKey string

const claimsContextKey contextKey = "auth_claims"

type TokenVerifier interface {
	VerifyAccessToken(token string) (Claims, error)
}

func Authenticate(verifier TokenVerifier, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		if header == "" {
			writeError(w, http.StatusUnauthorized, "token tidak valid")
			return
		}

		parts := strings.Split(header, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			writeError(w, http.StatusUnauthorized, "token tidak valid")
			return
		}

		claims, err := verifier.VerifyAccessToken(parts[1])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "token tidak valid")
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(role user.Role, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok || claims.Role != role {
			writeError(w, http.StatusForbidden, "akses tidak diizinkan")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(Claims)
	return claims, ok
}
