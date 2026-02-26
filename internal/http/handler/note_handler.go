package handler

import (
	"net/http"

	"github.com/nate/go-boilerplate/internal/http/middleware"
	"github.com/nate/go-boilerplate/internal/service"
)

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
