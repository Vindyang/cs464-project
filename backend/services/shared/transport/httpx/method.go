package httpx

import "net/http"

// RequireMethod validates the expected HTTP method and writes 405 on mismatch.
func RequireMethod(w http.ResponseWriter, r *http.Request, expected string) bool {
	if r.Method != expected {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return false
	}
	return true
}
