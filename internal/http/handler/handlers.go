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

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
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

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
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

type NoteHandler struct {
	noteService *service.NoteService
}

func NewNoteHandler(noteService *service.NoteService) *NoteHandler {
	return &NoteHandler{noteService: noteService}
}

type NoteResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	DeletedAt *string `json:"deleted_at,omitempty"`
}

type NoteListResponse struct {
	Notes      []NoteResponse `json:"notes"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

type CreateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondWithError(w, unauthorized("authentication required"))
		return
	}

	opts := parseNoteListOptions(r)

	result, err := h.noteService.ListByUserID(r.Context(), userID, opts)
	if err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	notes := make([]NoteResponse, len(result.Items))
	for i, n := range result.Items {
		notes[i] = *toNoteResponse(&n)
	}

	middleware.WriteJSON(w, http.StatusOK, NoteListResponse{
		Notes:      notes,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	})
}

func (h *NoteHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondWithError(w, unauthorized("authentication required"))
		return
	}

	noteID, err := parseUUIDParam(r, "id")
	if err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	note, err := h.noteService.GetByIDAndUserID(r.Context(), noteID, userID)
	if err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, toNoteResponse(note))
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondWithError(w, unauthorized("authentication required"))
		return
	}

	var req CreateNoteRequest
	if err := decodeJSON(r, &req); err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	if req.Title == "" {
		middleware.RespondWithError(w, validationError("title is required"))
		return
	}

	note, err := h.noteService.Create(r.Context(), service.CreateNoteInput{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, toNoteResponse(note))
}

func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondWithError(w, unauthorized("authentication required"))
		return
	}

	noteID, err := parseUUIDParam(r, "id")
	if err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	var req UpdateNoteRequest
	if err := decodeJSON(r, &req); err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	if req.Title == "" {
		middleware.RespondWithError(w, validationError("title is required"))
		return
	}

	note, err := h.noteService.Update(r.Context(), service.UpdateNoteInput{
		ID:      noteID,
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, toNoteResponse(note))
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondWithError(w, unauthorized("authentication required"))
		return
	}

	noteID, err := parseUUIDParam(r, "id")
	if err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	if err := h.noteService.SoftDelete(r.Context(), noteID, userID); err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NoteHandler) Restore(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondWithError(w, unauthorized("authentication required"))
		return
	}

	noteID, err := parseUUIDParam(r, "id")
	if err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	if err := h.noteService.Restore(r.Context(), noteID, userID); err != nil {
		middleware.RespondWithError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
