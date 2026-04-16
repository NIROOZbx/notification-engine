package helpers

import (
	"errors"

	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func HandleServiceError(c fiber.Ctx, err error, log zerolog.Logger) error {
    switch {
    case errors.Is(err, apperrors.ErrNotFound), errors.Is(err, apperrors.ErrTemplateNotFound):
        return response.NotFound(c, err.Error())

    case errors.Is(err, apperrors.ErrAlreadyExists), errors.Is(err, apperrors.ErrDefaultExists), errors.Is(err, apperrors.ErrDuplicateName), errors.Is(err, apperrors.ErrDuplicateChannel), errors.Is(err, apperrors.ErrDefaultConfigExists), errors.Is(err, apperrors.ErrDuplicateConfig), errors.Is(err, apperrors.ErrDependencyFailure):
        return response.Conflict(c, err.Error())

    case errors.Is(err, apperrors.ErrInvalidInput), errors.Is(err, apperrors.ErrTemplateNotLive), errors.Is(err, apperrors.ErrNoActiveChannels), errors.Is(err, apperrors.ErrTemplateDropped), errors.Is(err, apperrors.ErrLastActiveChannel), errors.Is(err, apperrors.ErrReqBody), errors.Is(err, apperrors.ErrBadRequest), errors.Is(err, apperrors.ErrInactiveProvider):
        return response.BadRequest(c, nil, err.Error())

    case errors.Is(err, apperrors.ErrForbidden):
        return response.Forbidden(c, nil, err.Error())

    case errors.Is(err, apperrors.ErrUnauthorized):
        return response.Unauthorized(c, err.Error())
    default:
        log.Error().Err(err).Msg("unexpected error")
        return response.InternalServerError(c)
    }
}