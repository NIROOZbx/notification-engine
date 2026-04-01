package utils

import (

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
)

func ParseIDParam(c fiber.Ctx, paramName string) (pgtype.UUID, bool) {
	val := c.Params(paramName)

	id, err :=StringToUUID(val)
	if err != nil {
		return pgtype.UUID{}, false
	}
	return id, true
}