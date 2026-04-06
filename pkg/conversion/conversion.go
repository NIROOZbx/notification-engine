package conversion

import (
	"time"

	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5/pgtype"
)

func JSONBFromMap(m map[string]any) ([]byte, error) {
	return sonic.Marshal(m)
}

func MapFromJSONB(b []byte) (map[string]any, error) {
	var m map[string]any
	return m, sonic.Unmarshal(b, &m)
}

func TimestampFromTime(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}