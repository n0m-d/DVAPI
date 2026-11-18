package graph

import (
	"context"

	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

// NoteService is the subset of note operations the resolvers depend on.
type NoteService interface {
	Create(ctx context.Context, input domain.CreateNoteInput) (domain.Note, error)
	GetByID(ctx context.Context, id string) (domain.Note, error)
	List(ctx context.Context, userID *uuid.UUID) ([]domain.Note, error)
	Update(ctx context.Context, id uuid.UUID, input domain.UpdateNoteInput) (domain.Note, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserLister is the subset of user operations the resolvers depend on.
type UserLister interface {
	AdminList(ctx context.Context, search, role string, page, pageSize int) (domain.AdminUserListResponse, error)
}

type Resolver struct {
	NoteService NoteService
	UserService UserLister
}
