package domain

type CreatePullRequestResponse struct {
	PR *PullRequest `json:"pr"`
}

type MergePullRequestResponse struct {
	PR *PullRequest `json:"pr"`
}

type ReassignReviewerResponse struct {
	PR         *PullRequest `json:"pr"`
	ReplacedBy string       `json:"replaced_by"`
}
