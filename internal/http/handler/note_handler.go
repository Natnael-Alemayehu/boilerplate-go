package handler

import (
	"net/http"

	"github.com/nate/go-boilerplate/internal/http/middleware"
	"github.com/nate/go-boilerplate/internal/service"
	"github.com/nate/go-boilerplate/pkg/response"
	"github.com/nate/go-boilerplate/pkg/validation"
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
	Title   string `json:"title" validate:"required,min=1,max=255"`
	Content string `json:"content" validate:"max=10000"`
}

type UpdateNoteRequest struct {
	Title   string `json:"title" validate:"required,min=1,max=255"`
	Content string `json:"content" validate:"max=10000"`
}

func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, unauthorized("authentication required"))
		return
	}

	opts := parseNoteListOptions(r)

	result, err := h.noteService.ListByUserID(r.Context(), userID, opts)
	if err != nil {
		response.Error(w, err)
		return
	}

	notes := make([]NoteResponse, len(result.Items))
	for i, n := range result.Items {
		notes[i] = *toNoteResponse(&n)
	}

	response.JSON(w, http.StatusOK, NoteListResponse{
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
		response.Error(w, unauthorized("authentication required"))
		return
	}

	noteID, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}

	note, err := h.noteService.GetByIDAndUserID(r.Context(), noteID, userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, toNoteResponse(note))
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, unauthorized("authentication required"))
		return
	}

	var req CreateNoteRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, err)
		return
	}

	if errs := validation.Validate(req); len(errs) > 0 {
		response.JSON(w, http.StatusBadRequest, errs)
		return
	}

	note, err := h.noteService.Create(r.Context(), service.CreateNoteInput{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, toNoteResponse(note))
}

func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, unauthorized("authentication required"))
		return
	}

	noteID, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}

	var req UpdateNoteRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, err)
		return
	}

	if errs := validation.Validate(req); len(errs) > 0 {
		response.JSON(w, http.StatusBadRequest, errs)
		return
	}

	note, err := h.noteService.Update(r.Context(), service.UpdateNoteInput{
		ID:      noteID,
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, toNoteResponse(note))
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, unauthorized("authentication required"))
		return
	}

	noteID, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}

	if err := h.noteService.SoftDelete(r.Context(), noteID, userID); err != nil {
		response.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NoteHandler) Restore(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, unauthorized("authentication required"))
		return
	}

	noteID, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}

	if err := h.noteService.Restore(r.Context(), noteID, userID); err != nil {
		response.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
