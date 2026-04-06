package helpers

import "github.com/jackc/pgx/v5/pgtype"

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