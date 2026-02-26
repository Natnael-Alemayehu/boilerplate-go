package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	apperrors "github.com/nate/go-boilerplate/internal/errors"
)

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
