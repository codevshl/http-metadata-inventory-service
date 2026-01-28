package response

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/codevshl/http-metadata-inventory-service/internal/core/apperror"
	"github.com/gin-gonic/gin"
)

const traceIDKey = "trace_id"

type APIResponse struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

type ErrorBody struct {
	Kind string `json:"kind"`
}

func WriteSuccess(c *gin.Context, status int, code, message string, data interface{}) {
	c.JSON(status, APIResponse{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
		TraceID: TraceID(c),
	})
}

func WriteError(c *gin.Context, err error) {
	appErr := asAppError(err)

	c.JSON(apperror.Status(err), APIResponse{
		Success: false,
		Code:    toErrorCode(appErr.Kind),
		Message: apperror.Message(err),
		Error: &ErrorBody{
			Kind: appErr.Kind.Name(),
		},
		TraceID: TraceID(c),
	})
}

func TraceID(c *gin.Context) string {
	if v, exists := c.Get(traceIDKey); exists {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

func SetTraceID(c *gin.Context, traceID string) {
	c.Set(traceIDKey, traceID)
	c.Writer.Header().Set("X-Trace-ID", traceID)
}

func ReadTraceID(c *gin.Context) string {
	if v := strings.TrimSpace(c.GetHeader("X-Trace-ID")); v != "" {
		return v
	}
	if v := strings.TrimSpace(c.GetHeader("X-Request-ID")); v != "" {
		return v
	}
	return ""
}

func NewTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func asAppError(err error) *apperror.AppError {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperror.New(apperror.ErrInternal, "internal server error")
}

func toErrorCode(kind apperror.ErrorKind) string {
	switch kind {
	case apperror.ErrValidation:
		return "ERR_VALIDATION"
	case apperror.ErrUnauthorized:
		return "ERR_UNAUTHORIZED"
	case apperror.ErrForbidden:
		return "ERR_FORBIDDEN"
	case apperror.ErrNotFound:
		return "ERR_NOT_FOUND"
	case apperror.ErrConflict:
		return "ERR_CONFLICT"
	case apperror.ErrBusinessLogic:
		return "ERR_BUSINESS_LOGIC"
	case apperror.ErrUpstream:
		return "ERR_UPSTREAM"
	case apperror.ErrTimeout:
		return "ERR_TIMEOUT"
	default:
		return "ERR_INTERNAL"
	}
}
