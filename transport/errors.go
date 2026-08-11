package transport

import (
	"encoding/json"
	"errors"
	"fmt"
)

// BirdeyeError is a structured error from a Birdeye REST response. Message
// is decoded opportunistically from a top-level "message" field if the
// response body has one — Birdeye's docs do not document a stable error
// envelope schema, so Raw always preserves the exact response body
// verbatim rather than risking silently dropped error detail.
type BirdeyeError struct {
	HTTPStatus int
	// Code is Birdeye's application error code when the response includes
	// one. It is deliberately a string because the API has used more than
	// one representation over time.
	Code string
	// RequestID is copied from a request-id response header when present.
	RequestID string
	// Success mirrors the response envelope's top-level "success" field,
	// if the body was valid JSON with one; false otherwise.
	Success bool
	Message string
	Raw     []byte
}

// decodeErrorDetails reads optional fields without assuming a stable error
// schema. The original, size-limited response remains available in Raw.
func decodeErrorDetails(raw []byte) (code, message string) {
	var payload struct {
		Code    json.RawMessage `json:"code"`
		Message string          `json:"message"`
	}
	if json.Unmarshal(raw, &payload) != nil {
		return "", ""
	}
	if len(payload.Code) > 0 && string(payload.Code) != "null" {
		if json.Unmarshal(payload.Code, &code) != nil {
			var number json.Number
			if json.Unmarshal(payload.Code, &number) == nil {
				code = number.String()
			}
		}
	}
	return code, payload.Message
}

func (e *BirdeyeError) Error() string {
	return fmt.Sprintf("birdeye: api error: http=%d code=%s request_id=%s success=%t message=%s", e.HTTPStatus, e.Code, e.RequestID, e.Success, e.Message)
}

// Sentinel errors, mapped from HTTP status. Match with errors.Is; inspect
// the wrapped *BirdeyeError for the exact message/raw body.
var (
	ErrBadRequest       = errors.New("birdeye: bad request")
	ErrUnauthorized     = errors.New("birdeye: unauthorized: invalid or missing X-API-KEY")
	ErrForbidden        = errors.New("birdeye: forbidden: endpoint not available on your Birdeye plan")
	ErrNotFound         = errors.New("birdeye: resource not found")
	ErrRateLimited      = errors.New("birdeye: rate limit exceeded")
	ErrServerError      = errors.New("birdeye: internal server error")
	ErrResponseTooLarge = errors.New("birdeye: response body exceeds size limit")
)

// MapHTTPStatus maps an HTTP status code to a sentinel error. Returns nil
// for 2xx. Birdeye does not publish a full HTTP-status catalog as of this
// SDK's last documentation pass, so unlisted codes fall back to the
// nearest 4xx/5xx sentinel rather than being silently ignored.
func MapHTTPStatus(status int) error {
	switch {
	case status >= 200 && status < 300:
		return nil
	case status == 400:
		return ErrBadRequest
	case status == 401:
		return ErrUnauthorized
	case status == 403:
		return ErrForbidden
	case status == 404:
		return ErrNotFound
	case status == 429:
		return ErrRateLimited
	case status >= 500:
		return ErrServerError
	case status >= 400:
		return ErrBadRequest
	default:
		return nil
	}
}
