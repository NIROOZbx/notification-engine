package conversion

import (
	"time"

	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5/pgtype"
)


func JSONBFromMap(m map[string]any) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return sonic.Marshal(m)
}

func MapFromJSONB(b []byte) (map[string]any, error) {
	var m map[string]any
	if len(b) == 0 {
		return m, nil
	}
	return m, sonic.Unmarshal(b, &m)
}

func TimestampFromPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func TimestampFromTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: !t.IsZero()}
}

func TimeFromTimestamp(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	tCopy := t.Time
	return &tCopy
}

func TimeFromTimestampVal(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func TextFromPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func StringFromText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	sCopy := t.String
	return &sCopy
}

func ToNullString(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func PtrTimeFromTimestamp(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}