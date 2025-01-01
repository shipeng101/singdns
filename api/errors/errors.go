package errors

// APIError represents an API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

// Common errors
var (
	ErrInternalServer = &APIError{Code: 500, Message: "Internal server error"}
	ErrUnauthorized   = &APIError{Code: 401, Message: "Unauthorized"}
	ErrForbidden      = &APIError{Code: 403, Message: "Forbidden"}
	ErrNotFound       = &APIError{Code: 404, Message: "Not found"}
	ErrBadRequest     = &APIError{Code: 400, Message: "Bad request"}
)

// NewAPIError creates a new API error
func NewAPIError(code int, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}
