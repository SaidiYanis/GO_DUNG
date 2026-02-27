package errors

import "errors"

var (
	ErrValidation      = errors.New("validation")
	ErrNotFound        = errors.New("not_found")
	ErrConflict        = errors.New("conflict")
	ErrForbidden       = errors.New("forbidden")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInsufficient    = errors.New("insufficient_funds")
	ErrWrongStepOrder  = errors.New("wrong_step_order")
	ErrNotInRange      = errors.New("not_in_range")
	ErrAlreadyHandled  = errors.New("already_handled")
	ErrInvalidArgument = errors.New("invalid_argument")
)
