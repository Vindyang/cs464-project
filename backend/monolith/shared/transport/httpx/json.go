package httpx

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, message string, err error) {
	payload := map[string]string{"error": message}
	if err != nil {
		payload["details"] = err.Error()
	}
	WriteJSON(w, status, payload)
}

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
