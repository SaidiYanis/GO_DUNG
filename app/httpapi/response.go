package httpapi

import (
	apperrors "dungeons/app/errors"
	"dungeons/app/models"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func JSON(c *gin.Context, status int, payload interface{}) {
	c.JSON(status, payload)
}

func JSONError(c *gin.Context, err error) {
	status, code := MapError(err)
	c.JSON(status, models.ErrorEnvelope{
		Error: models.ErrorPayload{
			Code:    code,
			Message: err.Error(),
		},
	})
}

func MapError(err error) (int, string) {
	var validationErr validator.ValidationErrors
	var syntaxErr *json.SyntaxError
	switch {
	case errors.As(err, &validationErr), errors.As(err, &syntaxErr):
		return http.StatusBadRequest, "bad_request"
	case errors.Is(err, apperrors.ErrValidation), errors.Is(err, apperrors.ErrInvalidArgument):
		return http.StatusBadRequest, "bad_request"
	case errors.Is(err, apperrors.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, apperrors.ErrForbidden):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, apperrors.ErrNotFound):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, apperrors.ErrWrongStepOrder):
		return http.StatusConflict, "WRONG_STEP_ORDER"
	case errors.Is(err, apperrors.ErrNotInRange):
		return http.StatusConflict, "NOT_IN_RANGE"
	case errors.Is(err, apperrors.ErrAlreadyHandled):
		return http.StatusConflict, "ATTEMPT_ALREADY_HANDLED"
	case errors.Is(err, apperrors.ErrConflict):
		return http.StatusConflict, "conflict"
	case errors.Is(err, apperrors.ErrInsufficient):
		return http.StatusConflict, "insufficient_funds"
	default:
		return http.StatusInternalServerError, "internal_error"
	}
}
