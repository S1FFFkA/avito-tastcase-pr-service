package domain

import "errors"

// Сгненерированные с помощью OPEN-API ошибки домена
var (
	ErrInvalidRequest         = errors.New("invalid request")
	ErrTeamExists             = errors.New("team_name already exists")
	ErrPRExists               = errors.New("PR id already exists")
	ErrPRMerged               = errors.New("cannot reassign on merged PR")
	ErrNotAssigned            = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate            = errors.New("no active replacement candidate in team")
	ErrNotFound               = errors.New("resource not found")
	ErrInternalError          = errors.New("internal server error")
	ErrFailedToDecodeJSON     = errors.New("failed to decode JSON")
	ErrQueryParameterRequired = errors.New("query parameter is required")
)

type ErrorCode string

const (
	ErrorCodeInvalidRequest         ErrorCode = "INVALID_REQUEST"
	ErrorCodeTeamExists             ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists               ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged               ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned            ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate            ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound               ErrorCode = "NOT_FOUND"
	ErrorCodeInternalError          ErrorCode = "INTERNAL_ERROR"
	ErrorCodeFailedToDecodeJSON     ErrorCode = "FAILED_TO_DECODE_JSON"
	ErrorCodeQueryParameterRequired ErrorCode = "QUERY_PARAMETER_REQUIRED"
)

type ErrorResponse struct {
	Error struct {
		Code    ErrorCode `json:"code"`
		Message string    `json:"message"`
	} `json:"error"`
}

func NewErrorResponse(code ErrorCode, message string) ErrorResponse {
	var resp ErrorResponse
	resp.Error.Code = code
	resp.Error.Message = message
	return resp
}
