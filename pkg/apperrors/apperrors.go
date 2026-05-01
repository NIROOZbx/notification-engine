package apperrors

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrForbidden         = errors.New("forbidden")
	ErrDependencyFailure = errors.New("dependency failure")
	ErrReqBody           = errors.New("invalid request body")
	ErrInternal          = errors.New("internal server error")
	ErrBadRequest        = errors.New("bad request")
	ErrUnauthorized      = errors.New("unauthorized: invalid or missing session")
	ErrLimitReached      = errors.New("plan limit reached: please upgrade your plan")

	// layout specific
	ErrDefaultExists = errors.New("a default layout already exists in this workspace")
	ErrDuplicateName = errors.New("name already exists in this workspace")

	// template specific
	ErrTemplateNotFound = errors.New("template not found")
	ErrTemplateDropped  = errors.New("cannot modify a dropped template")
	ErrTemplateNotLive  = errors.New("template is not live")
	ErrNoActiveChannels = errors.New("no active channels")

	// channel specific
	ErrDuplicateChannel    = errors.New("channel already exists for this template")
	ErrLastActiveChannel   = errors.New("cannot disable or delete the last active channel of a live template")
	ErrDefaultConfigExists = errors.New("a default config already exists for this channel")
	ErrDuplicateConfig     = errors.New("a config already exists for this provider and channel")
	ErrInactiveProvider    = errors.New("cannot set an inactive provider as the default")
	ErrDecryptionFailed    = errors.New("failed to decrypt credentials: check your secret key")
	ErrAlreadyCancelled    = errors.New("already been cancelled")
)

type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

func NewNotFoundError(resource string) error {
	return &NotFoundError{Resource: resource}
}

type AlreadyExistsError struct {
	Resource string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already exists", e.Resource)
}

func NewAlreadyExistsError(resource string) error {
	return &AlreadyExistsError{Resource: resource}
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func GetPQError(err error) *pgconn.PgError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr
	}
	return nil
}

func MapDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			switch pgErr.ConstraintName {
			case "default_layout":
				return ErrDefaultExists
			case "layouts_workspace_id_name_key":
				return ErrDuplicateName
			case "channel_configs_workspace_id_channel_provider_key":
				return ErrDuplicateConfig
			case "idx_unique_default_channel":
				return ErrDefaultConfigExists
			default:
				return ErrAlreadyExists
			}
		case "23503":
			return fmt.Errorf("%w: %s", ErrInvalidInput, pgErr.Detail)
		case "23502":
			return fmt.Errorf("%w: %s", ErrInvalidInput, pgErr.ColumnName)
		case "23514":
			return fmt.Errorf("%w: %s", ErrInvalidInput, pgErr.ConstraintName)
		}
	}

	return err
}
