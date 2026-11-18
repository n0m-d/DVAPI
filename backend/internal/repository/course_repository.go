package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/n0m-d/DVAPI/internal/db"
	"github.com/n0m-d/DVAPI/internal/domain"
)

type SaveCourseInput struct {
	ID           uuid.UUID
	InstructorID uuid.UUID
	Title        string
	Slug         string
	Description  string
	Published    bool
}

type EnrolledStudentsListParams struct {
	CourseID uuid.UUID
	Name     string
	Limit    int
	Offset   int
}

type CourseRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.Course, error)
	Create(ctx context.Context, input SaveCourseInput) (domain.Course, error)
	Update(ctx context.Context, input SaveCourseInput) (domain.Course, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetLessonByID(ctx context.Context, id uuid.UUID) (domain.Lesson, error)
	ListLessons(ctx context.Context, courseID uuid.UUID) ([]domain.Lesson, error)
	CreateLesson(ctx context.Context, courseID uuid.UUID, title, content string, sortOrder int) (domain.Lesson, error)
	UpdateLesson(ctx context.Context, lesson domain.Lesson) (domain.Lesson, error)
	DeleteLesson(ctx context.Context, id uuid.UUID) error
	Enroll(ctx context.Context, studentID, courseID uuid.UUID) (domain.EnrolledStudent, error)
	Unenroll(ctx context.Context, studentID, courseID uuid.UUID) error
	IsEnrolled(ctx context.Context, studentID, courseID uuid.UUID) (bool, error)
	ListEnrolledStudents(ctx context.Context, params EnrolledStudentsListParams) ([]domain.EnrolledStudent, error)
}

type courseRepository struct {
	queries *db.Queries
}

func NewCourseRepository(queries *db.Queries) CourseRepository {
	return &courseRepository{queries: queries}
}

func (r *courseRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Course, error) {
	row, err := r.queries.GetCourseByID(ctx, db.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Course{}, ErrCourseNotFound
		}
		return domain.Course{}, err
	}
	return toDomainCourse(row)
}

func (r *courseRepository) Create(ctx context.Context, input SaveCourseInput) (domain.Course, error) {
	row, err := r.queries.CreateCourse(ctx, db.CreateCourseParams{
		InstructorID: db.PgUUID(input.InstructorID),
		Title:        input.Title,
		Slug:         input.Slug,
		Description:  pgtype.Text{String: input.Description, Valid: input.Description != ""},
		Published:    input.Published,
	})
	if err != nil {
		return domain.Course{}, err
	}
	return toDomainCourse(db.GetCourseByIDRow(row))
}

func (r *courseRepository) Update(ctx context.Context, input SaveCourseInput) (domain.Course, error) {
	row, err := r.queries.UpdateCourse(ctx, db.UpdateCourseParams{
		ID:          db.PgUUID(input.ID),
		Title:       input.Title,
		Slug:        input.Slug,
		Description: pgtype.Text{String: input.Description, Valid: input.Description != ""},
		Published:   input.Published,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Course{}, ErrCourseNotFound
		}
		return domain.Course{}, err
	}
	return toDomainCourse(db.GetCourseByIDRow(row))
}

func (r *courseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	affected, err := r.queries.DeleteCourse(ctx, db.PgUUID(id))
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrCourseNotFound
	}
	return nil
}

func (r *courseRepository) GetLessonByID(ctx context.Context, id uuid.UUID) (domain.Lesson, error) {
	row, err := r.queries.GetLessonByID(ctx, db.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Lesson{}, ErrLessonNotFound
		}
		return domain.Lesson{}, err
	}
	return toDomainLesson(row)
}

func (r *courseRepository) ListLessons(ctx context.Context, courseID uuid.UUID) ([]domain.Lesson, error) {
	rows, err := r.queries.ListCourseLessons(ctx, db.PgUUID(courseID))
	if err != nil {
		return nil, err
	}
	lessons := make([]domain.Lesson, 0, len(rows))
	for _, row := range rows {
		lesson, err := toDomainLesson(row)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}
	return lessons, nil
}

func (r *courseRepository) CreateLesson(ctx context.Context, courseID uuid.UUID, title, content string, sortOrder int) (domain.Lesson, error) {
	row, err := r.queries.CreateLesson(ctx, db.CreateLessonParams{
		CourseID:  db.PgUUID(courseID),
		Title:     title,
		SortOrder: int32(sortOrder),
		Content:   pgtype.Text{String: content, Valid: content != ""},
	})
	if err != nil {
		return domain.Lesson{}, err
	}
	return toDomainLesson(row)
}

func (r *courseRepository) UpdateLesson(ctx context.Context, lesson domain.Lesson) (domain.Lesson, error) {
	row, err := r.queries.UpdateLesson(ctx, db.UpdateLessonParams{
		ID:        db.PgUUID(lesson.ID),
		Title:     lesson.Title,
		SortOrder: int32(lesson.SortOrder),
		Content:   pgtype.Text{String: lesson.Content, Valid: lesson.Content != ""},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Lesson{}, ErrLessonNotFound
		}
		return domain.Lesson{}, err
	}
	return toDomainLesson(row)
}

func (r *courseRepository) DeleteLesson(ctx context.Context, id uuid.UUID) error {
	affected, err := r.queries.DeleteLesson(ctx, db.PgUUID(id))
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrLessonNotFound
	}
	return nil
}

func (r *courseRepository) Enroll(ctx context.Context, studentID, courseID uuid.UUID) (domain.EnrolledStudent, error) {
	row, err := r.queries.CreateEnrollment(ctx, db.CreateEnrollmentParams{
		UserID:   db.PgUUID(studentID),
		CourseID: db.PgUUID(courseID),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.EnrolledStudent{}, ErrEnrollmentExists
		}
		return domain.EnrolledStudent{}, err
	}
	id, err := db.FromPgUUID(row.UserID)
	if err != nil {
		return domain.EnrolledStudent{}, fmt.Errorf("parse student id: %w", err)
	}
	return domain.EnrolledStudent{
		ID:         id,
		Email:      row.Email,
		FullName:   row.FullName,
		EnrolledAt: row.EnrolledAt.Time,
	}, nil
}

func (r *courseRepository) Unenroll(ctx context.Context, studentID, courseID uuid.UUID) error {
	affected, err := r.queries.DeleteEnrollment(ctx, db.DeleteEnrollmentParams{
		UserID:   db.PgUUID(studentID),
		CourseID: db.PgUUID(courseID),
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrEnrollmentMissing
	}
	return nil
}

func (r *courseRepository) IsEnrolled(ctx context.Context, studentID, courseID uuid.UUID) (bool, error) {
	_, err := r.queries.GetEnrollment(ctx, db.GetEnrollmentParams{
		UserID:   db.PgUUID(studentID),
		CourseID: db.PgUUID(courseID),
	})
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func (r *courseRepository) ListEnrolledStudents(ctx context.Context, params EnrolledStudentsListParams) ([]domain.EnrolledStudent, error) {
	rows, err := r.queries.ListEnrolledStudentsByCourse(ctx, db.ListEnrolledStudentsByCourseParams{
		CourseID:    db.PgUUID(params.CourseID),
		Name:        params.Name,
		LimitCount:  int32(params.Limit),
		OffsetCount: int32(params.Offset),
	})
	if err != nil {
		return nil, err
	}

	out := make([]domain.EnrolledStudent, 0, len(rows))
	for _, row := range rows {
		id, err := db.FromPgUUID(row.ID)
		if err != nil {
			return nil, fmt.Errorf("parse student id: %w", err)
		}
		out = append(out, domain.EnrolledStudent{
			ID:         id,
			Email:      row.Email,
			FullName:   row.FullName,
			EnrolledAt: row.EnrolledAt.Time,
		})
	}
	return out, nil
}

func toDomainLesson(row db.Lesson) (domain.Lesson, error) {
	id, err := db.FromPgUUID(row.ID)
	if err != nil {
		return domain.Lesson{}, fmt.Errorf("parse lesson id: %w", err)
	}
	courseID, err := db.FromPgUUID(row.CourseID)
	if err != nil {
		return domain.Lesson{}, fmt.Errorf("parse lesson course id: %w", err)
	}
	return domain.Lesson{
		ID:        id,
		CourseID:  courseID,
		Title:     row.Title,
		SortOrder: int(row.SortOrder),
		Content:   row.Content.String,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

func toDomainCourse(row db.GetCourseByIDRow) (domain.Course, error) {
	id, err := db.FromPgUUID(row.ID)
	if err != nil {
		return domain.Course{}, fmt.Errorf("parse course id: %w", err)
	}
	instructorID, err := db.FromPgUUID(row.InstructorID)
	if err != nil {
		return domain.Course{}, fmt.Errorf("parse instructor id: %w", err)
	}

	return domain.Course{
		ID: id,
		Instructor: domain.Instructor{
			ID:       instructorID,
			FullName: row.InstructorName,
			Email:    row.Email,
		},
		Title:       row.Title,
		Slug:        row.Slug,
		Description: row.Description.String,
		Published:   row.Published,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}
