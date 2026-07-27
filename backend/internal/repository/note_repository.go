package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/n0m-d/DVAPI/internal/db"
	"github.com/n0m-d/DVAPI/internal/domain"
)

type NoteRepository interface {
	Create(ctx context.Context, input domain.CreateNoteInput) (domain.Note, error)
	GetByID(ctx context.Context, id string) (domain.Note, error)
	List(ctx context.Context, userID *uuid.UUID) ([]domain.Note, error)
	Update(ctx context.Context, id uuid.UUID, input domain.UpdateNoteInput) (domain.Note, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type noteRepository struct {
	queries *db.Queries
	db      db.DBTX
}

func NewNoteRepository(queries *db.Queries, database db.DBTX) NoteRepository {
	return &noteRepository{queries: queries, db: database}
}

func (r *noteRepository) Create(ctx context.Context, input domain.CreateNoteInput) (domain.Note, error) {
	row, err := r.queries.CreateNote(ctx, db.CreateNoteParams{
		UserID: pgUUID(input.UserID),
		Title:  input.Title,
		Body:   input.Body,
	})
	if err != nil {
		return domain.Note{}, err
	}
	return toDomainNote(row)
}

func (r *noteRepository) GetByID(ctx context.Context, id string) (domain.Note, error) {
	q := fmt.Sprintf(`
		SELECT id, user_id, title, body, created_at, updated_at
		FROM notes
		WHERE id = '%s'
	`, id) //Raw Query:Vulnerable to SQL injection
	var row db.Note
	err := r.db.QueryRow(ctx, q).Scan(&row.ID, &row.UserID, &row.Title, &row.Body, &row.CreatedAt, &row.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Note{}, ErrNoteNotFound
		}
		return domain.Note{}, err
	}
	return toDomainNote(row)
}

func (r *noteRepository) List(ctx context.Context, userID *uuid.UUID) ([]domain.Note, error) {
	rows, err := r.queries.ListNotes(ctx, optionalPgUUID(userID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.Note, 0, len(rows))
	for _, row := range rows {
		n, err := toDomainNote(row)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func (r *noteRepository) Update(ctx context.Context, id uuid.UUID, input domain.UpdateNoteInput) (domain.Note, error) {
	row, err := r.queries.UpdateNote(ctx, db.UpdateNoteParams{
		ID:    pgUUID(id),
		Title: optionalPgText(input.Title),
		Body:  optionalPgText(input.Body),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Note{}, ErrNoteNotFound
		}
		return domain.Note{}, err
	}
	return toDomainNote(row)
}

func (r *noteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	affected, err := r.queries.DeleteNote(ctx, pgUUID(id))
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNoteNotFound
	}
	return nil
}

func optionalPgText(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *v, Valid: true}
}

func toDomainNote(row db.Note) (domain.Note, error) {
	id, err := fromPgUUID(row.ID)
	if err != nil {
		return domain.Note{}, fmt.Errorf("note id: %w", err)
	}
	userID, err := fromPgUUID(row.UserID)
	if err != nil {
		return domain.Note{}, fmt.Errorf("note user id: %w", err)
	}
	return domain.Note{
		ID:        id,
		UserID:    userID,
		Title:     row.Title,
		Body:      row.Body,
		CreatedAt: row.CreatedAt.Time.UTC(),
		UpdatedAt: row.UpdatedAt.Time.UTC(),
	}, nil
}
