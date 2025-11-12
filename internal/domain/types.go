package domain

type PRStatus string

func (p *PRStatus) String() string {
	return string(*p)
}

func (p *PRStatus) Set(value string) {
	*p = PRStatus(value)
}

const (
	PRStatusOpen   PRStatus = "open"
	PRStatusMerged PRStatus = "merged"
)

type ErrorCode string

const (
	ErrorCodeNotFound       ErrorCode = "not_found"
	ErrorCodeInvalidRequest ErrorCode = "invalid_request"
	ErrorCodeInternalError  ErrorCode = "internal_error"
)
