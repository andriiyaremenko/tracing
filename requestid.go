package tracing

import "net/http"

var (
	// RequestID reader from Header using default Header name.
	DefaultRequestIDReadHeader = RequestIDReadHeader(HeaderRequestID)
	// RequestID writer to Header using default Header name.
	DefaultRequestIDWriteHeader = RequestIDWriteHeader(HeaderRequestID)
	// RequestID options with default Header name.
	DefaultRequestIDOptions = RequestIDOptionsWithHeader(HeaderRequestID)
)

// RequestID carries additional information not used in command execution.
type RequestID string

// Creates new RequestID.
func NewRequestID(id string) RequestID {
	return RequestID(id)
}

// New RequestID for next event in execution chain.
func NextRequestID(r RequestID, _ string) RequestID {
	return r
}

// Checks if RequestID is valid.
func ValidRequestID(r RequestID) bool {
	return r != ""
}

// RequestID reader from Header using provided Header name.
// Will canonicalize provided name.
func RequestIDReadHeader(requestID string) func(header http.Header, id string) (RequestID, bool) {
	return func(header http.Header, id string) (RequestID, bool) {
		r := RequestID(header.Get(requestID))

		if ValidRequestID(r) {
			return r, true
		}

		return NewRequestID(id), false
	}
}

// RequestID writer to Header using provided Header name.
// Will canonicalize provided name.
func RequestIDWriteHeader(requestID string) func(http.Header, RequestID) {
	return func(header http.Header, r RequestID) {
		header.Set(requestID, string(r))
	}
}

// RequestID options with provided Header reader and writer.
func RequestIDOptions(
	read func(http.Header, string) (RequestID, bool),
	write func(http.Header, RequestID),
) Options[RequestID] {
	return func() (ReadHeader[RequestID], WriteHeader[RequestID], Next[RequestID]) {
		return read, write, NextRequestID
	}
}

// RequestID options with provided Header names.
func RequestIDOptionsWithHeader(requestID string) Options[RequestID] {
	return RequestIDOptions(RequestIDReadHeader(requestID), RequestIDWriteHeader(requestID))
}
