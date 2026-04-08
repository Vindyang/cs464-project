package httpx

import "strings"

// errorRule maps an error message fragment to a machine-readable code.
// Rules are checked in order; first match wins.
type errorRule struct {
	fragment string
	code     string
}

var providerErrorRules = []errorRule{
	{"timeout", "PROVIDER_TIMEOUT"},
	{"deadline exceeded", "PROVIDER_TIMEOUT"},
	{"quota", "PROVIDER_QUOTA_EXCEEDED"},
	{"storage full", "PROVIDER_QUOTA_EXCEEDED"},
	{"insufficient storage", "PROVIDER_QUOTA_EXCEEDED"},
	{"auth", "PROVIDER_AUTH_EXPIRED"},
	{"unauthorized", "PROVIDER_AUTH_EXPIRED"},
	{"token", "PROVIDER_AUTH_EXPIRED"},
}

var uploadErrorRules = []errorRule{
	{"insufficient healthy providers", "PROVIDER_UNAVAILABLE"},
	{"invalid erasure coding", "INVALID_ERASURE_PARAMS"},
	{"failed to shard", "SHARD_ENCODE_FAILED"},
	{"failed to register", "FILE_REGISTER_FAILED"},
	{"failed to record shards", "SHARD_RECORD_FAILED"},
	{"shards succeeded", "SHARD_UPLOAD_PARTIAL"},
}

func classifyError(msg string, rules []errorRule) string {
	lower := strings.ToLower(msg)
	for _, r := range rules {
		if strings.Contains(lower, r.fragment) {
			return r.code
		}
	}
	return "UNKNOWN_ERROR"
}

// ClassifyProviderError maps a provider error message to a machine-readable code.
func ClassifyProviderError(msg string) string {
	return classifyError(msg, providerErrorRules)
}

// ClassifyUploadError maps an orchestrator upload error message to a machine-readable code.
func ClassifyUploadError(msg string) string {
	return classifyError(msg, uploadErrorRules)
}
