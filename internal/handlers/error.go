package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
)

const (
	statusBadRequest          = 400
	statusNotFound            = 404
	statusConflict            = 409
	statusInternalServerError = 500
	statusMethodNotAllowed    = 405
	statusCreated             = 201
	statusOK                  = 200
)

// errorMapping используется для маппинга доменной ошибки на HTTP статус и код ошибки
type errorMapping struct {
	status  int
	code    domain.ErrorCode
	message string
}

func respondError(w http.ResponseWriter, err error) {
	mapping := resolveError(err)
	logger.SafeErrorw("request error", "status", mapping.status, "code", string(mapping.code), "error", err)
	writeJSON(w, mapping.status, domain.NewErrorResponse(mapping.code, mapping.message))
}

func respondMethodNotAllowed(w http.ResponseWriter, method string) {
	writeJSON(w, statusMethodNotAllowed, domain.NewErrorResponse(domain.ErrorCodeInvalidRequest, "method "+method+" not allowed"))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		// Если не удалось замаршалить, отправляем ошибку
		errorResp := domain.NewErrorResponse(domain.ErrorCodeInternalError, "failed to encode response")
		errorData, _ := json.MarshalIndent(errorResp, "", "    ")

		w.WriteHeader(statusInternalServerError)
		_, _ = w.Write(errorData)
		return
	}

	w.WriteHeader(status)
	_, err = w.Write(data)
	if err != nil {

		return
	}
}

// resolveError маппит доменную ошибку на HTTP статус и доменный код ошибки
func resolveError(err error) errorMapping {
	switch {
	case errors.Is(err, domain.ErrInvalidRequest):
		return errorMapping{statusBadRequest, domain.ErrorCodeInvalidRequest, err.Error()}
	case errors.Is(err, domain.ErrTeamExists):
		return errorMapping{statusBadRequest, domain.ErrorCodeTeamExists, domain.ErrTeamExists.Error()}
	case errors.Is(err, domain.ErrPRExists):
		return errorMapping{statusConflict, domain.ErrorCodePRExists, domain.ErrPRExists.Error()}
	case errors.Is(err, domain.ErrPRMerged):
		return errorMapping{statusConflict, domain.ErrorCodePRMerged, domain.ErrPRMerged.Error()}
	case errors.Is(err, domain.ErrNotAssigned):
		return errorMapping{statusConflict, domain.ErrorCodeNotAssigned, domain.ErrNotAssigned.Error()}
	case errors.Is(err, domain.ErrNoCandidate):
		return errorMapping{statusConflict, domain.ErrorCodeNoCandidate, domain.ErrNoCandidate.Error()}
	case errors.Is(err, domain.ErrNotFound):
		return errorMapping{statusNotFound, domain.ErrorCodeNotFound, domain.ErrNotFound.Error()}
	case errors.Is(err, domain.ErrFailedToDecodeJSON):
		return errorMapping{statusBadRequest, domain.ErrorCodeFailedToDecodeJSON, domain.ErrFailedToDecodeJSON.Error()}
	case errors.Is(err, domain.ErrQueryParameterRequired):
		return errorMapping{statusBadRequest, domain.ErrorCodeQueryParameterRequired, domain.ErrQueryParameterRequired.Error()}
	default:
		return errorMapping{statusInternalServerError, domain.ErrorCodeInternalError, domain.ErrInternalError.Error()}
	}
}
