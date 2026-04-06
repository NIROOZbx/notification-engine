package utils

import (
	"errors"

	"github.com/NIROOZbx/notification-engine/consts"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
)

type CallerContext struct {
    WorkspaceID pgtype.UUID
    CallerID    pgtype.UUID
    CallerRole  string
}

func GetWID(c fiber.Ctx) (pgtype.UUID, error) {
    wid, ok := c.Locals(consts.WID).(pgtype.UUID)
    if !ok {
        return pgtype.UUID{}, errors.New("no workspace associated with session")
    }
    return wid, nil
}

func GetUID(c fiber.Ctx) (pgtype.UUID, error) {
    uid, ok := c.Locals(consts.UID).(pgtype.UUID)
    if !ok {
        return pgtype.UUID{}, errors.New("no user associated with session")
    }
    return uid, nil
}

func GetRole(c fiber.Ctx)(string, error) {
    role, ok := c.Locals(consts.Role).(string)
    if !ok {
        return "", errors.New("no role associated with session")
    }
    return role, nil
}

func GetEnvID(c fiber.Ctx) (pgtype.UUID, error) {
    envID, ok := c.Locals(consts.ENVID).(pgtype.UUID)
    if !ok {
        return pgtype.UUID{}, errors.New("no environment associated with session")
    }
    return envID, nil
}

func GetIsTest(c fiber.Ctx) bool {
    isTest, ok := c.Locals(consts.ISTEST).(bool)
    if !ok {
        return false // default to false if not present
    }
    return isTest
}

func GetCallerContext(c fiber.Ctx) (CallerContext, error) {
    workspaceID, err := GetWID(c)
    if err != nil {
        return CallerContext{}, err
    }
    callerID, err := GetUID(c)
    if err != nil {
        return CallerContext{}, err
    }
    role, err := GetRole(c)
    if err != nil {
        return CallerContext{}, err
    }
    return CallerContext{
        WorkspaceID: workspaceID,
        CallerID:    callerID,
        CallerRole:  role,
    }, nil
}