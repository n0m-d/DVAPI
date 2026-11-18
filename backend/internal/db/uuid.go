package db

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func PgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func FromPgUUID(id pgtype.UUID) (uuid.UUID, error) {
	if !id.Valid {
		return uuid.Nil, fmt.Errorf("invalid uuid")
	}
	return uuid.UUID(id.Bytes), nil
}
