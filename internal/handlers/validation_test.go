package handlers

import (
	"AVITOSAMPISHU/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTeam(t *testing.T) {
	tests := []struct {
		name    string
		team    *domain.Team
		wantErr bool
	}{
		{
			name: "valid team",
			team: &domain.Team{
				TeamName: "team1",
				Members: []domain.TeamMember{
					{UserID: "user1", Username: "User1"},
					{UserID: "user2", Username: "User2"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty team name",
			team: &domain.Team{
				TeamName: "",
				Members: []domain.TeamMember{
					{UserID: "user1", Username: "User1"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty members list",
			team: &domain.Team{
				TeamName: "team1",
				Members:  []domain.TeamMember{},
			},
			wantErr: true,
		},
		{
			name: "member with empty user_id",
			team: &domain.Team{
				TeamName: "team1",
				Members: []domain.TeamMember{
					{UserID: "", Username: "User1"},
				},
			},
			wantErr: true,
		},
		{
			name: "member with empty username",
			team: &domain.Team{
				TeamName: "team1",
				Members: []domain.TeamMember{
					{UserID: "user1", Username: ""},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTeam(tt.team)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidRequest)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSetIsActiveRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *domain.SetIsActiveRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &domain.SetIsActiveRequest{
				UserID:   "user1",
				IsActive: true,
			},
			wantErr: false,
		},
		{
			name: "empty user_id",
			req: &domain.SetIsActiveRequest{
				UserID:   "",
				IsActive: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSetIsActiveRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidRequest)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDeactivateTeamMembersReq(t *testing.T) {
	tests := []struct {
		name    string
		req     *domain.DeactivateTeamMembersReq
		wantErr bool
	}{
		{
			name: "valid request with user_ids",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "team1",
				UserIDs:  []string{"user1", "user2"},
			},
			wantErr: false,
		},
		{
			name: "valid request without user_ids",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "team1",
				UserIDs:  []string{},
			},
			wantErr: false,
		},
		{
			name: "empty team_name",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "",
				UserIDs:  []string{"user1"},
			},
			wantErr: true,
		},
		{
			name: "empty user_id in list",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "team1",
				UserIDs:  []string{"user1", ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDeactivateTeamMembersReq(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidRequest)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCreatePullRequestReq(t *testing.T) {
	tests := []struct {
		name    string
		req     *domain.CreatePullRequestReq
		wantErr bool
	}{
		{
			name: "valid request",
			req: &domain.CreatePullRequestReq{
				PullRequestID:   "pr1",
				PullRequestName: "PR Name",
				AuthorID:        "user1",
			},
			wantErr: false,
		},
		{
			name: "empty pull_request_id",
			req: &domain.CreatePullRequestReq{
				PullRequestID:   "",
				PullRequestName: "PR Name",
				AuthorID:        "user1",
			},
			wantErr: true,
		},
		{
			name: "empty pull_request_name",
			req: &domain.CreatePullRequestReq{
				PullRequestID:   "pr1",
				PullRequestName: "",
				AuthorID:        "user1",
			},
			wantErr: true,
		},
		{
			name: "empty author_id",
			req: &domain.CreatePullRequestReq{
				PullRequestID:   "pr1",
				PullRequestName: "PR Name",
				AuthorID:        "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreatePullRequestReq(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidRequest)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMergePullRequestReq(t *testing.T) {
	tests := []struct {
		name    string
		req     *domain.MergePullRequestReq
		wantErr bool
	}{
		{
			name: "valid request",
			req: &domain.MergePullRequestReq{
				PullRequestID: "pr1",
			},
			wantErr: false,
		},
		{
			name: "empty pull_request_id",
			req: &domain.MergePullRequestReq{
				PullRequestID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMergePullRequestReq(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidRequest)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateReassignReviewerReq(t *testing.T) {
	tests := []struct {
		name    string
		req     *domain.ReassignReviewerReq
		wantErr bool
	}{
		{
			name: "valid request",
			req: &domain.ReassignReviewerReq{
				PullRequestID: "pr1",
				OldUserID:     "user1",
			},
			wantErr: false,
		},
		{
			name: "empty pull_request_id",
			req: &domain.ReassignReviewerReq{
				PullRequestID: "",
				OldUserID:     "user1",
			},
			wantErr: true,
		},
		{
			name: "empty old_user_id",
			req: &domain.ReassignReviewerReq{
				PullRequestID: "pr1",
				OldUserID:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateReassignReviewerReq(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidRequest)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
