package repositories

import (

	"github.com/NIROOZbx/notification-engine/db/sqlc"
)


type notificationRepository struct {
	queries *sqlc.Queries
}
func NewNotificationRepository(queries *sqlc.Queries) *notificationRepository {
    return &notificationRepository{queries: queries}
}

