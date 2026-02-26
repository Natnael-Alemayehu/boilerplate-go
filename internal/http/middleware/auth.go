package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	apperrors "github.com/nate/go-boilerplate/internal/errors"
	"github.com/nate/go-boilerplate/internal/repository"
	"github.com/nate/go-boilerplate/pkg/jwt"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	EmailKey  contextKey = "email"
	RoleKey   contextKey = "role"
)

func Auth(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, apperrors.Unauthorized("missing authorization header"))
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				respondWithError(w, apperrors.Unauthorized("invalid authorization header format"))
				return
			}

			claims, err := jwtManager.ValidateAccessToken(parts[1])
			if err != nil {
				respondWithError(w, apperrors.InvalidToken("invalid or expired token"))
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

func GetEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(EmailKey).(string)
	return email, ok
}

func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return role, ok
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetUserID(r.Context())
		if !ok {
			respondWithError(w, apperrors.Unauthorized("authentication required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Authorize(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetRole(r.Context())
			if !ok {
				respondWithError(w, apperrors.Unauthorized("authentication required"))
				return
			}

			for _, role := range roles {
				if userRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			respondWithError(w, apperrors.Forbidden("insufficient permissions"))
		})
	}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RespondWithError(w http.ResponseWriter, err error) {
	if appErr, ok := apperrors.IsAppError(err); ok {
		WriteJSON(w, appErr.HTTPStatus, ErrorResponse{
			Error:   http.StatusText(appErr.HTTPStatus),
			Code:    string(appErr.Code),
			Message: appErr.Message,
		})
		return
	}

	WriteJSON(w, http.StatusInternalServerError, ErrorResponse{
		Error:   http.StatusText(http.StatusInternalServerError),
		Code:    string(apperrors.CodeInternal),
		Message: "An internal error occurred",
	})
}

func respondWithError(w http.ResponseWriter, err error) {
	RespondWithError(w, err)
}

func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type RateLimiter struct {
	repo   repository.RateLimitRepository
	limit  int
	window int64
}

func NewRateLimiter(repo repository.RateLimitRepository, limit int, windowSeconds int64) *RateLimiter {
	return &RateLimiter{
		repo:   repo,
		limit:  limit,
		window: windowSeconds,
	}
}

func (rl *RateLimiter) Limit() int {
	return rl.limit
}

func (rl *RateLimiter) Check(ctx context.Context, key string) (count int64, ttl int64, err error) {
	count, err = rl.repo.Increment(ctx, key, rl.window)
	if err != nil {
		return 0, 0, err
	}

	ttl, err = rl.repo.GetTTL(ctx, key)
	if err != nil {
		ttl = rl.window
	}

	return count, ttl, nil
}

func RateLimit(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := getRateLimitKey(r)

			count, ttl, err := limiter.Check(r.Context(), key)
			if err != nil {
				respondWithError(w, apperrors.Internal("rate limit check failed", err))
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.Limit()))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, limiter.Limit()-int(count))))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(ttl, 10))

			if count > int64(limiter.Limit()) {
				respondWithError(w, apperrors.RateLimited("rate limit exceeded"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getRateLimitKey(r *http.Request) string {
	if userID, ok := GetUserID(r.Context()); ok {
		return fmt.Sprintf("user:%s", userID)
	}
	return fmt.Sprintf("ip:%s", r.RemoteAddr)
}

func Logger(next http.Handler) http.Handler {
	return middleware.Logger(next)
}

func RequestID(next http.Handler) http.Handler {
	return middleware.RequestID(next)
}

func Recoverer(next http.Handler) http.Handler {
	return middleware.Recoverer(next)
}

func RealIP(next http.Handler) http.Handler {
	return middleware.RealIP(next)
}
