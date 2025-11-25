package domain

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type DeactivateTeamMembersReq struct {
	TeamName string   `json:"team_name"`
	UserIDs  []string `json:"user_ids,omitempty"` // Если пустой деактивируем всех пользователей команды
}

type DeactivateTeamMembersRes struct {
	DeactivatedUserIDs []string               `json:"deactivated_user_ids"`
	Reassignments      []ReviewerReassignment `json:"reassignments"`
}
