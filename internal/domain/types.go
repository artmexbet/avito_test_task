package domain

// PRStatus represents the status of a pull request.
type PRStatus string

func (p *PRStatus) String() string {
	return string(*p)
}

// Set sets the PRStatus from a string value.
func (p *PRStatus) Set(value string) {
	*p = PRStatus(value)
}

// Possible values for PRStatus
const (
	PRStatusOpen   PRStatus = "open"
	PRStatusMerged PRStatus = "merged"
)

// ErrorCode defines the type for error codes.
type ErrorCode string

// Possible values for ErrorCode
const (
	ErrorCodeNotFound       ErrorCode = "not_found"
	ErrorCodeInvalidRequest ErrorCode = "invalid_request"
	ErrorCodeInternalError  ErrorCode = "internal_error"
)
