package serviceerror

import "errors"

var ErrNotFound = errors.New("not found")

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
type ValidationError struct{ Fields []FieldError }

func (e ValidationError) Error() string   { return "validation failed" }
func (e ValidationError) HasFields() bool { return len(e.Fields) > 0 }
func NewValidation(fields []FieldError) error {
	if len(fields) == 0 {
		return nil
	}
	return ValidationError{Fields: fields}
}
