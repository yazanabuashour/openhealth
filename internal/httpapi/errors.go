package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yazanabuashour/openhealth/internal/api/generated"
	"github.com/yazanabuashour/openhealth/internal/health"
)

func requestErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	writeError(w, r, http.StatusBadRequest, generated.VALIDATIONERROR, err.Error())
}

func responseErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	var (
		validationErr *health.ValidationError
		notFoundErr   *health.NotFoundError
		conflictErr   *health.ConflictError
	)

	switch {
	case errors.As(err, &validationErr):
		writeError(w, r, http.StatusBadRequest, generated.VALIDATIONERROR, validationErr.Error())
	case errors.As(err, &notFoundErr):
		writeError(w, r, http.StatusNotFound, generated.NOTFOUND, notFoundErr.Error())
	case errors.As(err, &conflictErr):
		writeError(w, r, http.StatusConflict, generated.CONFLICT, conflictErr.Error())
	default:
		writeError(w, r, http.StatusInternalServerError, generated.INTERNALERROR, "An unexpected error occurred")
	}
}

func writeError(w http.ResponseWriter, r *http.Request, statusCode int, code generated.ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := generated.ErrorEnvelope{
		Error: generated.ErrorDetails{
			Code:          code,
			CorrelationId: correlationIDFromContext(r.Context()),
			Message:       message,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode error response", http.StatusInternalServerError)
	}
}
