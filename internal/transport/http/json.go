package httptransport

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/endorsain/nictofilia-go-api/internal/errors"
)

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// DecodeJSON convierte errores de sintaxis o tipos incompatibles en un error público controlado.
func DecodeJSON(r *http.Request, destination any) error {
	if err := json.NewDecoder(r.Body).Decode(destination); err != nil {
		return apperrors.New(apperrors.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest)
	}

	return nil
}
