package domain

import "time"

type PRStatus string

// Validate проверяет валидность статуса PR
func (s PRStatus) Validate() bool {
	switch s {
	case PRStatusOpen, PRStatusMerged:
		return true
	default:
		return false
	}
}

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id" db:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name" db:"pull_request_name"`
	AuthorID          string     `json:"author_id" db:"author_id"`
	Status            PRStatus   `json:"status" db:"status"` // Используем ENUM в качестве статуса так-как скорее всего изменять его не будут
	AssignedReviewers []string   `json:"assigned_reviewers" db:"assigned_reviewers"`
	NeedMoreReviewers *bool      `json:"need_more_reviewers,omitempty"`
	CreatedAt         *time.Time `json:"createdAt" db:"created_at"`
	MergedAt          *time.Time `json:"mergedAt" db:"merged_at"`
}

type PullRequestShort struct {
	PullRequestID   string   `json:"pull_request_id"`
	PullRequestName string   `json:"pull_request_name"`
	AuthorID        string   `json:"author_id"`
	Status          PRStatus `json:"status"`
}

type CreatePullRequestReq struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type MergePullRequestReq struct {
	PullRequestID string `json:"pull_request_id"`
}

type ReassignReviewerReq struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type ReviewerReassignment struct {
	PrID          string `json:"pr_id"`
	OldReviewerID string `json:"old_reviewer_id"`
	NewReviewerID string `json:"new_reviewer_id,omitempty"`
}
