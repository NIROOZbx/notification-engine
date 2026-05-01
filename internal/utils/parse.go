package utils

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
)

func ParseIDParam(c fiber.Ctx, paramName string) (pgtype.UUID, bool) {
	val := c.Params(paramName)

	id, err := StringToUUID(val)
	if err != nil {
		return pgtype.UUID{}, false
	}
	return id, true
}

func QueryInt32(c fiber.Ctx, key string, defaultValue int32) int32 {
	valStr := c.Query(key)
	if valStr == "" {
		return defaultValue
	}

	val, err := strconv.ParseInt(valStr, 10, 32)
	if err != nil {
		return defaultValue
	}

	return int32(val)
}
