package graph

import (
	"github.com/n0m-d/DVAPI/graph/model"
	"github.com/n0m-d/DVAPI/internal/domain"
)

func toModelNote(n domain.Note) *model.Note {
	return &model.Note{
		ID:        n.ID.String(),
		UserID:    n.UserID.String(),
		Title:     n.Title,
		Body:      n.Body,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

func toModelUser(u domain.User) *model.User {
	return &model.User{
		ID:        u.ID.String(),
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
