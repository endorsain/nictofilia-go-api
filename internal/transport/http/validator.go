package httptransport

import (
	"errors"
	"fmt"
	"net/http"

	apperrors "github.com/endorsain/nictofilia-go-api/internal/errors"
	"github.com/go-playground/validator/v10"
)

// RequestValidator valida structs de entrada HTTP sin exponer errores de la librería.
type RequestValidator struct {
	validator *validator.Validate
}

// NewRequestValidator crea una instancia reutilizable durante la vida de la aplicación.
func NewRequestValidator() *RequestValidator {
	return &RequestValidator{validator: validator.New()}
}

func (v *RequestValidator) Validate(request any) error {
	if err := v.validator.Struct(request); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return apperrors.New(apperrors.CodeValidationError, "request validation failed", http.StatusBadRequest)
		}
		return fmt.Errorf("validate request DTO: %w", err)
	}

	return nil
}
