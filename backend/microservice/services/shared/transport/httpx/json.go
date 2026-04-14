package httpx

import (
	"encoding/json"
	"net/http"
)

// WriteJSON writes a JSON response with the supplied status code.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteError writes a stable JSON error shape used across services.
func WriteError(w http.ResponseWriter, status int, message string, err error) {
	payload := map[string]string{"error": message}
	if err != nil {
		payload["details"] = err.Error()
	}
	WriteJSON(w, status, payload)
}

// WriteErrorWithCode writes a stable JSON error shape with an optional machine-readable code.
// The code is omitted if empty.
func WriteErrorWithCode(w http.ResponseWriter, status int, message string, code string, err error) {
	payload := map[string]string{"error": message}
	if code != "" {
		payload["code"] = code
	}
	if err != nil {
		payload["details"] = err.Error()
	}
	WriteJSON(w, status, payload)
}
