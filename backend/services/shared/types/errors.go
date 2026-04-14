package types

// ErrorResponse is the standard JSON error shape emitted by all services.
// The Code field is a machine-readable string (e.g. "FILE_NOT_FOUND") that
// the frontend error registry uses to show contextual help.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}
