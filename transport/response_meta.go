package transport

import "net/http"

// ResponseMeta captures everything about a REST response beyond the
// decoded payload: HTTP status, the envelope's success/message fields,
// the raw body, and every response header verbatim. Headers is exposed
// in full (rather than picking out named rate-limit fields) because
// Birdeye's docs do not document a stable set of rate-limit header
// names — inspect Headers yourself for whatever your plan tier sends
// rather than trusting an SDK-invented field name.
type ResponseMeta struct {
	HTTPStatus int
	Success    bool
	Message    string
	RawBody    []byte
	Headers    http.Header

	// Attempts is how many HTTP attempts were made, including retries.
	Attempts int
}
