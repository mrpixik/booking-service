package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/auth"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/server"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
)

type jwtManager interface {
	GenerateToken(userID, role string) (string, error)
	ParseToken(tokenStr string) (*auth.Claims, error)
}

func JWTMiddleware(jwtManager jwtManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				adapters.WriteError(w, server.CodeUnauthorized, server.InvalidAuthorisationHeaderMsg, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")

			claims, err := jwtManager.ParseToken(tokenStr)
			if err != nil {
				adapters.WriteError(w, server.CodeUnauthorized, server.InvalidTokenMsg, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), model.CtxUserID, claims.UserID)
			ctx = context.WithValue(ctx, model.CtxRole, model.Role(claims.Role))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(model.CtxRole).(model.Role)
		if role != model.RoleAdmin {
			adapters.WriteError(w, server.CodeForbidden, server.AdminRoleRequiredMsg, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(model.CtxRole).(model.Role)
		if role != model.RoleUser {
			adapters.WriteError(w, server.CodeForbidden, server.UserRoleRequiredMsg, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserInfoFromCtx возвращает роль и айди пользователя из токена в контексте запроса
func UserInfoFromCtx(ctx context.Context) (model.Role, string) {
	id, _ := ctx.Value(model.CtxUserID).(string)
	role, _ := ctx.Value(model.CtxRole).(model.Role)
	return role, id
}
