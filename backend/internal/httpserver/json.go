package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/MisterVVP/logarift/backend/internal/serviceerror"
)

type errorEnvelope struct {
	Error apiError `json:"error"`
}
type apiError struct {
	Code    string                    `json:"code"`
	Message string                    `json:"message"`
	Fields  []serviceerror.FieldError `json:"fields,omitempty"`
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("request body must contain a single JSON object")
	}
	return nil
}

func writeAPIError(w http.ResponseWriter, status int, code, message string, fields []serviceerror.FieldError) {
	writeJSON(w, status, errorEnvelope{Error: apiError{Code: code, Message: message, Fields: fields}})
}
func writeInvalidJSON(w http.ResponseWriter) {
	writeAPIError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.", nil)
}
func writeServiceError(w http.ResponseWriter, err error) {
	var validation serviceerror.ValidationError
	if errors.As(err, &validation) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Request validation failed.", validation.Fields)
		return
	}
	if errors.Is(err, serviceerror.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Requested resource was not found.", nil)
		return
	}
	writeAPIError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred.", nil)
}
func parseLimit(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	var n int64
	_, err := fmt.Sscan(raw, &n)
	if err != nil {
		return 0, serviceerror.ValidationError{Fields: []serviceerror.FieldError{{Field: "limit", Message: "must be an integer"}}}
	}
	return n, nil
}
