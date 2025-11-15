package stats_retriever

type AssignmentStats struct {
	ReviewerID string `json:"reviewer_id"`
	IsActive   bool   `json:"is_active"`
	PRCount    int    `json:"pr_count"`
}

type TeamsStats struct {
	TeamName string `json:"team_name"`
	TotalPRs int    `json:"total_prs"`
}

type UsersStats struct {
	IsActive bool `json:"is_active"`
	Total    int  `json:"total"`
}

// Stats represents system statistics.
type Stats struct {
	UserStats   []UsersStats      `json:"user_stats"`
	TeamStats   []TeamsStats      `json:"team_stats"`
	AssignStats []AssignmentStats `json:"assignment_stats"`
}
