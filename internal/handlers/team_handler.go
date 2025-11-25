package handlers

import (
	"encoding/json"
	"net/http"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/service"
	"AVITOSAMPISHU/pkg/logger"
)

type TeamHandler struct {
	teamService service.TeamService
}

func NewTeamHandler(teamService service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/team/add", h.CreateTeam)
	mux.HandleFunc("/team/get", h.GetTeam)
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondMethodNotAllowed(w, r.Method)
		return
	}

	var team domain.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		respondError(w, domain.ErrFailedToDecodeJSON)
		return
	}

	// Валидация данных
	if err := validateTeam(&team); err != nil {
		respondError(w, err)
		return
	}

	createdTeam, err := h.teamService.CreateTeam(r.Context(), &team)
	if err != nil {
		logger.Logger.Errorw("failed to create team", "team_name", team.TeamName, "error", err)
		respondError(w, err)
		return
	}

	logger.Logger.Infow("team created successfully", "team_name", createdTeam.TeamName, "members_count", len(createdTeam.Members))
	writeJSON(w, statusCreated, domain.CreateTeamResponse{Team: createdTeam})
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondMethodNotAllowed(w, r.Method)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		respondError(w, domain.ErrQueryParameterRequired)
		return
	}

	team, err := h.teamService.GetTeam(r.Context(), teamName)
	if err != nil {
		logger.Logger.Errorw("failed to get team", "team_name", teamName, "error", err)
		respondError(w, err)
		return
	}

	logger.Logger.Infow("team retrieved successfully", "team_name", teamName, "members_count", len(team.Members))
	writeJSON(w, statusOK, team)
}
