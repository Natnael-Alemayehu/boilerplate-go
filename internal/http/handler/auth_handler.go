package handler

import (
	"net/http"

	"github.com/nate/go-boilerplate/internal/http/middleware"
	"github.com/nate/go-boilerplate/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	User  *UserResponse  `json:"user"`
	Token *TokenResponse `json:"token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	if req.Email == "" || req.Password == "" {
		middleware.RespondWithError(w, badRequest("email and password are required"))
		return
	}

	result, err := h.authService.Register(r.Context(), service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, AuthResponse{
		User:  toUserResponse(result.User),
		Token: toTokenResponse(result.Tokens),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	if req.Email == "" || req.Password == "" {
		middleware.RespondWithError(w, badRequest("email and password are required"))
		return
	}

	result, err := h.authService.Login(r.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, AuthResponse{
		User:  toUserResponse(result.User),
		Token: toTokenResponse(result.Tokens),
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := decodeJSON(r, &req); err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	if req.RefreshToken == "" {
		middleware.RespondWithError(w, badRequest("refresh_token is required"))
		return
	}

	tokens, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, toTokenResponse(tokens))
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := decodeJSON(r, &req); err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	if req.RefreshToken == "" {
		middleware.RespondWithError(w, badRequest("refresh_token is required"))
		return
	}

	if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondWithError(w, unauthorized("authentication required"))
		return
	}

	if err := h.authService.LogoutAll(r.Context(), userID); err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
