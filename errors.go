package igris

import "fmt"

// AuthenticationError is returned on 401/403.
type AuthenticationError struct {
	StatusCode int
	Message    string
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication failed (%d): %s", e.StatusCode, e.Message)
}

// NetworkError is returned on connection failures.
type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %s", e.Err.Error())
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// RateLimitError is returned on 429.
type RateLimitError struct {
	Message string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %s", e.Message)
}

// ValidationError is returned on 400/422.
type ValidationError struct {
	StatusCode int
	Message    string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error (%d): %s", e.StatusCode, e.Message)
}

// APIError is returned on other HTTP errors.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
}
