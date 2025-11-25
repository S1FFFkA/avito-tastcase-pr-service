package handlers

import (
	"encoding/json"
	"net/http"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/service"
	"AVITOSAMPISHU/pkg/logger"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/users/setIsActive", h.SetIsActive)
	mux.HandleFunc("/users/getReview", h.GetUserReviews)
	mux.HandleFunc("/users/deactivateTeamMembers", h.DeactivateTeamMembers)
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondMethodNotAllowed(w, r.Method)
		return
	}

	var req domain.SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, domain.ErrFailedToDecodeJSON)
		return
	}

	if err := validateSetIsActiveRequest(&req); err != nil {
		respondError(w, err)
		return
	}

	user, err := h.userService.SetIsActive(r.Context(), &req)
	if err != nil {
		logger.Logger.Errorw("failed to set user active status", "user_id", req.UserID, "error", err)
		respondError(w, err)
		return
	}

	logger.Logger.Infow("user active status updated", "user_id", user.UserID, "is_active", user.IsActive)
	writeJSON(w, statusOK, domain.SetIsActiveResponse{User: user})
}

func (h *UserHandler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondMethodNotAllowed(w, r.Method)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		respondError(w, domain.ErrQueryParameterRequired)
		return
	}

	prs, err := h.userService.GetUserReviews(r.Context(), userID)
	if err != nil {
		logger.Logger.Errorw("failed to get user reviews", "user_id", userID, "error", err)
		respondError(w, err)
		return
	}

	logger.Logger.Infow("user reviews retrieved", "user_id", userID, "prs_count", len(prs))
	writeJSON(w, statusOK, domain.GetUserReviewsResponse{
		UserID:       userID,
		PullRequests: prs,
	})
}

func (h *UserHandler) DeactivateTeamMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondMethodNotAllowed(w, r.Method)
		return
	}

	var req domain.DeactivateTeamMembersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, domain.ErrFailedToDecodeJSON)
		return
	}

	// Валидация данных
	if err := validateDeactivateTeamMembersReq(&req); err != nil {
		respondError(w, err)
		return
	}

	res, err := h.userService.DeactivateTeamMembers(r.Context(), &req)
	if err != nil {
		logger.Logger.Errorw("failed to deactivate team members", "team_name", req.TeamName, "error", err)
		respondError(w, err)
		return
	}

	logger.Logger.Infow("team members deactivated", "team_name", req.TeamName, "deactivated_count", len(res.DeactivatedUserIDs))
	writeJSON(w, statusOK, res)
}
