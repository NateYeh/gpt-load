package utils

// Body size limits for logging and storage
const (
	// MaxBodySize is the maximum size for request/response body logging (1MB)
	// This supports large prompts up to ~100k tokens (~400-500KB)
	MaxBodySize = 1048576 // 1MB

	// MaxPathLength is the maximum length for URL path logging
	MaxPathLength = 1024

	// MaxAddrLength is the maximum length for upstream address logging
	MaxAddrLength = 1024
)
