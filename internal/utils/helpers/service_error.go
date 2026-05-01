package helpers

import (
	"errors"

	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleServiceError(c fiber.Ctx, err error, log zerolog.Logger) error {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.NotFound:
			return response.NotFound(c, st.Message())
		case codes.PermissionDenied, codes.Unauthenticated:
			return response.Unauthorized(c, st.Message())
		case codes.FailedPrecondition:
			if st.Message() == "already been cancelled" {
				return response.Conflict(c, st.Message())
			}
			return response.BadRequest(c, nil, st.Message())
		case codes.ResourceExhausted:
			return response.Forbidden(c, nil, st.Message())
		case codes.InvalidArgument:
			return response.BadRequest(c, nil, st.Message())
		}
	}

	switch {
	case errors.Is(err, apperrors.ErrNotFound), errors.Is(err, apperrors.ErrTemplateNotFound):
		return response.NotFound(c, err.Error())

	case errors.Is(err, apperrors.ErrAlreadyExists), errors.Is(err, apperrors.ErrDefaultExists), errors.Is(err, apperrors.ErrDuplicateName), errors.Is(err, apperrors.ErrDuplicateChannel), errors.Is(err, apperrors.ErrDefaultConfigExists), errors.Is(err, apperrors.ErrDuplicateConfig), errors.Is(err, apperrors.ErrDependencyFailure), errors.Is(err, apperrors.ErrAlreadyCancelled):
		return response.Conflict(c, err.Error())

	case errors.Is(err, apperrors.ErrInvalidInput), errors.Is(err, apperrors.ErrTemplateNotLive), errors.Is(err, apperrors.ErrNoActiveChannels), errors.Is(err, apperrors.ErrTemplateDropped), errors.Is(err, apperrors.ErrLastActiveChannel), errors.Is(err, apperrors.ErrReqBody), errors.Is(err, apperrors.ErrBadRequest), errors.Is(err, apperrors.ErrInactiveProvider):
		return response.BadRequest(c, nil, err.Error())

	case errors.Is(err, apperrors.ErrForbidden), errors.Is(err, apperrors.ErrLimitReached):
		return response.Forbidden(c, nil, err.Error())

	case errors.Is(err, apperrors.ErrUnauthorized):
		return response.Unauthorized(c, err.Error())
	default:
		log.Error().Err(err).Msg("unexpected error")
		return response.InternalServerError(c)
	}
}