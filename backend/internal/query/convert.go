package query

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/n0m-d/DVAPI/internal/db"
	"github.com/n0m-d/DVAPI/internal/domain"
)

func toDomainUser(user db.User) (domain.User, error) {
	id, err := db.FromPgUUID(user.ID)
	if err != nil {
		return domain.User{}, fmt.Errorf("parse user id: %w", err)
	}
	return domain.User{
		ID:        id,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

// courseListRow is the shared shape of sqlc course-list queries (identical SELECT lists).
type courseListRow struct {
	ID             pgtype.UUID        `json:"id"`
	InstructorID   pgtype.UUID        `json:"instructor_id"`
	Title          string             `json:"title"`
	Slug           string             `json:"slug"`
	Description    pgtype.Text        `json:"description"`
	Published      bool               `json:"published"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	UpdatedAt      pgtype.Timestamptz `json:"updated_at"`
	InstructorName string             `json:"instructor_name"`
	Email          string             `json:"email"`
}

func toDomainCourse(r courseListRow) (domain.Course, error) {
	cid, err := db.FromPgUUID(r.ID)
	if err != nil {
		return domain.Course{}, fmt.Errorf("parse course id: %w", err)
	}
	iid, err := db.FromPgUUID(r.InstructorID)
	if err != nil {
		return domain.Course{}, fmt.Errorf("parse instructor id: %w", err)
	}
	return domain.Course{
		ID: cid,
		Instructor: domain.Instructor{
			ID:       iid,
			FullName: r.InstructorName,
			Email:    r.Email,
		},
		Title:       r.Title,
		Slug:        r.Slug,
		Description: r.Description.String,
		Published:   r.Published,
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
	}, nil
}

func toDomainCourses[T ~struct {
	ID             pgtype.UUID        `json:"id"`
	InstructorID   pgtype.UUID        `json:"instructor_id"`
	Title          string             `json:"title"`
	Slug           string             `json:"slug"`
	Description    pgtype.Text        `json:"description"`
	Published      bool               `json:"published"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	UpdatedAt      pgtype.Timestamptz `json:"updated_at"`
	InstructorName string             `json:"instructor_name"`
	Email          string             `json:"email"`
}](rows []T) ([]domain.Course, error) {
	out := make([]domain.Course, 0, len(rows))
	for i := range rows {
		c, err := toDomainCourse(courseListRow(rows[i]))
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func courseListResponse(courses []domain.Course, total, page, pageSize int) domain.CourseResponse {
	pagination := domain.NewPagination(total, page, pageSize)
	return domain.CourseResponse{
		Status: "success",
		Data: &domain.CourseData{
			Courses:    courses,
			Pagination: &pagination,
		},
	}
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
