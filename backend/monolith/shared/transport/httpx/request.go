package httpx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func DecodeJSON(r *http.Request, dst interface{}, maxBytes int64) error {
	if maxBytes > 0 {
		r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("request body must contain a single JSON object")
	}

	return nil
}
