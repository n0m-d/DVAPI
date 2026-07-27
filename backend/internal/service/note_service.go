package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/repository"
)

type NoteService interface {
	Create(ctx context.Context, input domain.CreateNoteInput) (domain.Note, error)
	GetByID(ctx context.Context, id string) (domain.Note, error)
	List(ctx context.Context, userID *uuid.UUID) ([]domain.Note, error) //*uuid.UUID (a pointer) instead of uuid.UUID (a value) is the idiomatic Go way to make a parameter optional
	Update(ctx context.Context, id uuid.UUID, input domain.UpdateNoteInput) (domain.Note, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type noteService struct {
	notes repository.NoteRepository
}

func NewNoteService(notes repository.NoteRepository) NoteService {
	return &noteService{notes: notes}
}

func (s *noteService) Create(ctx context.Context, input domain.CreateNoteInput) (domain.Note, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.UserID == uuid.Nil {
		return domain.Note{}, fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}
	if input.Title == "" {
		return domain.Note{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	note, err := s.notes.Create(ctx, input)
	if err != nil {
		return domain.Note{}, err
	}
	return note, nil
}

func (s *noteService) GetByID(ctx context.Context, id string) (domain.Note, error) {
	if id == "" {
		return domain.Note{}, fmt.Errorf("%w: note id is required", ErrInvalidInput)
	}
	note, err := s.notes.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNoteNotFound) {
			return domain.Note{}, ErrNotFound
		}
		return domain.Note{}, err
	}
	return note, nil
}

func (s *noteService) List(ctx context.Context, userID *uuid.UUID) ([]domain.Note, error) {
	notes, err := s.notes.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	return notes, nil
}

func (s *noteService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateNoteInput) (domain.Note, error) {
	if id == uuid.Nil {
		return domain.Note{}, fmt.Errorf("%w: note id is required", ErrInvalidInput)
	}
	if input.Title == nil && input.Body == nil {
		return domain.Note{}, fmt.Errorf("%w: at least one field is required", ErrInvalidInput)
	}
	if input.Title != nil {
		trimmed := strings.TrimSpace(*input.Title)
		if trimmed == "" {
			return domain.Note{}, fmt.Errorf("%w: title cannot be empty", ErrInvalidInput)
		}
		input.Title = &trimmed
	}
	note, err := s.notes.Update(ctx, id, input)
	if err != nil {
		if errors.Is(err, repository.ErrNoteNotFound) {
			return domain.Note{}, ErrNotFound
		}
		return domain.Note{}, err
	}
	return note, nil
}

func (s *noteService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: note id is required", ErrInvalidInput)
	}
	if err := s.notes.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNoteNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
