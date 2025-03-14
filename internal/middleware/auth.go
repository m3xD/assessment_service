package middleware

import (
	"assessment_service/internal/util"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	jwtService util.Jwt
}

func NewAuthMiddleware(jwtService util.Jwt) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

func (auth *AuthMiddleware) AuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				util.ResponseError(w, util.Response{
					StatusCode: http.StatusUnauthorized,
					Message:    "missing authorization header",
					Data:       nil,
				})
				return
			}

			parts := strings.Split(token, " ")

			if len(parts) != 2 || parts[0] != "Bearer" {
				util.ResponseError(w, util.Response{
					StatusCode: http.StatusUnauthorized,
					Message:    "invalid authorization header",
					Data:       nil,
				})
				return
			}
			token = parts[1]

			claims, err := auth.jwtService.ValidateToken(token)
			if err != nil {
				util.ResponseError(w, util.Response{
					StatusCode: http.StatusUnauthorized,
					Message:    "unauthorized",
					Data:       nil,
				})
				return
			}

			ctx := context.WithValue(r.Context(), "user", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (auth *AuthMiddleware) ACLMiddleware(allowRoles ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value("user").(jwt.MapClaims)
			role := claims["role"].(string)
			for _, allowRole := range allowRoles {
				if role == allowRole {
					next.ServeHTTP(w, r)
					return
				}
			}
			util.ResponseError(w, util.Response{
				StatusCode: http.StatusForbidden,
				Message:    "forbidden",
				Data:       nil,
			})
		})
	}
}

func (auth *AuthMiddleware) OwnerMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value("user").(jwt.MapClaims)
			role := claims["role"].(string)
			user := claims["userID"].(string)
			if role == "admin" ||
				(role == "user" && (r.URL.Query().Get("id") == "" || user == r.URL.Query().Get("id"))) {
				next.ServeHTTP(w, r)
				return
			}
			util.ResponseError(w, util.Response{
				StatusCode: http.StatusForbidden,
				Message:    "forbidden",
				Data:       nil,
			})
		})
	}
}
