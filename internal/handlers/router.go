package handlers

import (
	"net/http"

	"AVITOSAMPISHU/internal/service"
)

func RegisterRoutes(
	mux *http.ServeMux,
	teamService service.TeamService,
	userService service.UserService,
	prService service.PullRequestService,
) {
	NewTeamHandler(teamService).Register(mux)
	NewUserHandler(userService).Register(mux)
	NewPullRequestHandler(prService).Register(mux)
}
