package handlers

import (
	"encoding/json"
	"net/http"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/service"
	"AVITOSAMPISHU/pkg/logger"
)

type PullRequestHandler struct {
	prService service.PullRequestService
}

func NewPullRequestHandler(prService service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{prService: prService}
}

func (h *PullRequestHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/pullRequest/create", h.CreatePullRequest)
	mux.HandleFunc("/pullRequest/merge", h.MergePullRequest)
	mux.HandleFunc("/pullRequest/reassign", h.ReassignReviewer)
}

func (h *PullRequestHandler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondMethodNotAllowed(w, r.Method)
		return
	}

	var req domain.CreatePullRequestReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, domain.ErrFailedToDecodeJSON)
		return
	}

	// Валидация данных
	if err := validateCreatePullRequestReq(&req); err != nil {
		respondError(w, err)
		return
	}

	pr, err := h.prService.CreatePullRequest(r.Context(), &req)
	if err != nil {
		logger.SafeErrorw("failed to create pull request", "pr_id", req.PullRequestID, "error", err)
		respondError(w, err)
		return
	}

	logger.SafeInfow("pull request created", "pr_id", pr.PullRequestID, "reviewers_count", len(pr.AssignedReviewers))
	writeJSON(w, statusCreated, domain.CreatePullRequestResponse{PR: pr})
}

func (h *PullRequestHandler) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondMethodNotAllowed(w, r.Method)
		return
	}

	var req domain.MergePullRequestReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, domain.ErrFailedToDecodeJSON)
		return
	}

	// Валидация данных
	if err := validateMergePullRequestReq(&req); err != nil {
		respondError(w, err)
		return
	}

	pr, err := h.prService.MergePullRequest(r.Context(), &req)
	if err != nil {
		logger.SafeErrorw("failed to merge pull request", "pr_id", req.PullRequestID, "error", err)
		respondError(w, err)
		return
	}

	logger.SafeInfow("pull request merged", "pr_id", pr.PullRequestID)
	writeJSON(w, statusOK, domain.MergePullRequestResponse{PR: pr})
}

func (h *PullRequestHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondMethodNotAllowed(w, r.Method)
		return
	}

	var req domain.ReassignReviewerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, domain.ErrFailedToDecodeJSON)
		return
	}

	// Валидация данных
	if err := validateReassignReviewerReq(&req); err != nil {
		respondError(w, err)
		return
	}

	pr, newReviewerID, err := h.prService.ReassignReviewer(r.Context(), &req)
	if err != nil {
		logger.SafeErrorw("failed to reassign reviewer", "pr_id", req.PullRequestID, "old_reviewer_id", req.OldUserID, "error", err)
		respondError(w, err)
		return
	}

	logger.SafeInfow("reviewer reassigned", "pr_id", pr.PullRequestID, "old_reviewer_id", req.OldUserID, "new_reviewer_id", newReviewerID)
	writeJSON(w, statusOK, domain.ReassignReviewerResponse{
		PR:         pr,
		ReplacedBy: newReviewerID,
	})
}
