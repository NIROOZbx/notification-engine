package helpers

import (


	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func Timestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: !t.IsZero()}
}

func ToTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func Text(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func Bool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

func Int4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

func Int8(i int64) pgtype.Int8 {
	return pgtype.Int8{Int64: i, Valid: true}
}

func TextPtr(s *string) pgtype.Text {
    if s == nil {
        return pgtype.Text{Valid: false}
    }
    return pgtype.Text{String: *s, Valid: true}
}

func BoolPtr(b *bool) pgtype.Bool {
    if b == nil {
        return pgtype.Bool{Valid: false}
    }
    return pgtype.Bool{Bool: *b, Valid: true}
}