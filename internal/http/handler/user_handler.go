package handler

import (
	"net/http"

	"github.com/nate/go-boilerplate/internal/http/middleware"
	"github.com/nate/go-boilerplate/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	opts := parseListOptions(r)

	result, err := h.userService.List(r.Context(), opts)
	if err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	users := make([]UserResponse, len(result.Items))
	for i, u := range result.Items {
		users[i] = *toUserResponse(&u)
	}

	middleware.WriteJSON(w, http.StatusOK, UserListResponse{
		Users:      users,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	})
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUIDParam(r, "id")
	if err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	if err := h.userService.Delete(r.Context(), userID); err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
