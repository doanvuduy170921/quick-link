package usecase

import "github.com/jackc/pgx/v5/pgtype"

func toPgInt8(userID *int64) pgtype.Int8 {
	if userID == nil {
		return pgtype.Int8{Valid: false} // NULL
	}
	return pgtype.Int8{Int64: *userID, Valid: true}
}
