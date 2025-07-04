package errors

import (
	"github.com/pkg/errors"
)

var (
	ErrNotFound           = errors.New("common.not_found")
	ErrInvalidArgument    = errors.New("common.invalid_argument")
	ErrUnauthenticated    = errors.New("common.unauthenticated")
	ErrPermissionDenied   = errors.New("common.permission_denied")
	ErrPreconditionFailed = errors.New("common.precondition_failed")
	ErrAlreadyExists      = errors.New("common.already_exists")
	ErrExternal           = errors.New("common.external_error")
	ErrUnknown            = errors.New("common.unknown")
)
