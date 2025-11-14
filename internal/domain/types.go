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
