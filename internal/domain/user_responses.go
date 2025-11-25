package domain

type SetIsActiveResponse struct {
	User *User `json:"user"`
}

type GetUserReviewsResponse struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}
