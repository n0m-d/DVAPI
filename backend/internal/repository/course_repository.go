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

type CourseListParams struct {
	Published bool
	Title     string
	Limit     int
	Offset    int
}

type EnrolledCourseListParams struct {
	StudentID uuid.UUID
	Title     string
	Limit     int
	Offset    int
}

type EnrolledStudentsListParams struct {
	CourseID uuid.UUID
	Name     string
	Limit    int
	Offset   int
}

type InstructorCourseListParams struct {
	InstructorID uuid.UUID
	Title        string
	Limit        int
	Offset       int
}

type SaveCourseInput struct {
	ID           uuid.UUID
	InstructorID uuid.UUID
	Title        string
	Slug         string
	Description  string
	Published    bool
}

type CourseRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.Course, error)
	GetCourses(ctx context.Context, params CourseListParams) ([]domain.Course, error)
	CountCourses(ctx context.Context, published bool, title string) (int, error)
	GetEnrolledCourses(ctx context.Context, params EnrolledCourseListParams) ([]domain.Course, error)
	CountEnrolledCourses(ctx context.Context, studentID uuid.UUID, title string) (int, error)
	ListEnrolledStudents(ctx context.Context, params EnrolledStudentsListParams) ([]domain.EnrolledStudent, error)
	CountEnrolledStudents(ctx context.Context, courseID uuid.UUID, name string) (int, error)
	GetInstructorCourses(ctx context.Context, params InstructorCourseListParams) ([]domain.Course, error)
	CountInstructorCourses(ctx context.Context, instructorID uuid.UUID, title string) (int, error)
	Create(ctx context.Context, input SaveCourseInput) (domain.Course, error)
	Update(ctx context.Context, input SaveCourseInput) (domain.Course, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetLessonByID(ctx context.Context, id uuid.UUID) (domain.Lesson, error)
	ListLessons(ctx context.Context, courseID uuid.UUID) ([]domain.Lesson, error)
	CreateLesson(ctx context.Context, courseID uuid.UUID, title, content string, sortOrder int) (domain.Lesson, error)
	UpdateLesson(ctx context.Context, lesson domain.Lesson) (domain.Lesson, error)
	DeleteLesson(ctx context.Context, id uuid.UUID) error
	Enroll(ctx context.Context, studentID, courseID uuid.UUID) error
	Unenroll(ctx context.Context, studentID, courseID uuid.UUID) error
	IsEnrolled(ctx context.Context, studentID, courseID uuid.UUID) (bool, error)
}

type courseRepository struct {
	queries *db.Queries
}

func NewCourseRepository(queries *db.Queries) CourseRepository {
	return &courseRepository{queries: queries}
}

func (r *courseRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Course, error) {
	row, err := r.queries.GetCourseByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Course{}, ErrCourseNotFound
		}
		return domain.Course{}, err
	}
	return toDomainCourseFromFields(
		row.ID,
		row.InstructorID,
		row.InstructorName,
		row.Email,
		row.Title,
		row.Slug,
		row.Description,
		row.Published,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func (r *courseRepository) GetCourses(ctx context.Context, params CourseListParams) ([]domain.Course, error) {
	courses, err := r.queries.GetCourses(ctx, db.GetCoursesParams{
		Published:   params.Published,
		Title:       params.Title,
		LimitCount:  int32(params.Limit),
		OffsetCount: int32(params.Offset),
	})
	if err != nil {
		return nil, err
	}
	return toDomainCourses(courses)
}

func (r *courseRepository) CountCourses(ctx context.Context, published bool, title string) (int, error) {
	total, err := r.queries.CountCourses(ctx, db.CountCoursesParams{
		Published: published,
		Title:     title,
	})
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

func (r *courseRepository) GetEnrolledCourses(ctx context.Context, params EnrolledCourseListParams) ([]domain.Course, error) {
	courses, err := r.queries.GetEnrolledCourses(ctx, db.GetEnrolledCoursesParams{
		StudentID:   pgUUID(params.StudentID),
		Title:       params.Title,
		LimitCount:  int32(params.Limit),
		OffsetCount: int32(params.Offset),
	})
	if err != nil {
		return nil, err
	}
	return toDomainCourses(courses)
}

func (r *courseRepository) CountEnrolledCourses(ctx context.Context, studentID uuid.UUID, title string) (int, error) {
	total, err := r.queries.CountEnrolledCourses(ctx, db.CountEnrolledCoursesParams{
		StudentID: pgUUID(studentID),
		Title:     title,
	})
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

func (r *courseRepository) ListEnrolledStudents(ctx context.Context, params EnrolledStudentsListParams) ([]domain.EnrolledStudent, error) {
	rows, err := r.queries.ListEnrolledStudentsByCourse(ctx, db.ListEnrolledStudentsByCourseParams{
		CourseID:    pgUUID(params.CourseID),
		Name:        params.Name,
		LimitCount:  int32(params.Limit),
		OffsetCount: int32(params.Offset),
	})
	if err != nil {
		return nil, err
	}

	out := make([]domain.EnrolledStudent, 0, len(rows))
	for _, row := range rows {
		id, err := fromPgUUID(row.ID)
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

func (r *courseRepository) CountEnrolledStudents(ctx context.Context, courseID uuid.UUID, name string) (int, error) {
	total, err := r.queries.CountEnrolledStudentsByCourse(ctx, db.CountEnrolledStudentsByCourseParams{
		CourseID: pgUUID(courseID),
		Name:     name,
	})
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

func (r *courseRepository) GetInstructorCourses(ctx context.Context, params InstructorCourseListParams) ([]domain.Course, error) {
	courses, err := r.queries.GetInstructorCourses(ctx, db.GetInstructorCoursesParams{
		InstructorID: pgUUID(params.InstructorID),
		Title:        params.Title,
		LimitCount:   int32(params.Limit),
		OffsetCount:  int32(params.Offset),
	})
	if err != nil {
		return nil, err
	}
	return toDomainCourses(courses)
}

func (r *courseRepository) CountInstructorCourses(ctx context.Context, instructorID uuid.UUID, title string) (int, error) {
	total, err := r.queries.CountInstructorCourses(ctx, db.CountInstructorCoursesParams{
		InstructorID: pgUUID(instructorID),
		Title:        title,
	})
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

func (r *courseRepository) Create(ctx context.Context, input SaveCourseInput) (domain.Course, error) {
	row, err := r.queries.CreateCourse(ctx, db.CreateCourseParams{
		InstructorID: pgUUID(input.InstructorID),
		Title:        input.Title,
		Slug:         input.Slug,
		Description:  pgtype.Text{String: input.Description, Valid: input.Description != ""},
		Published:    input.Published,
	})
	if err != nil {
		return domain.Course{}, err
	}
	id, err := fromPgUUID(row.ID)
	if err != nil {
		return domain.Course{}, err
	}
	return r.GetByID(ctx, id)
}

func (r *courseRepository) Update(ctx context.Context, input SaveCourseInput) (domain.Course, error) {
	row, err := r.queries.UpdateCourse(ctx, db.UpdateCourseParams{
		ID:          pgUUID(input.ID),
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
	id, err := fromPgUUID(row.ID)
	if err != nil {
		return domain.Course{}, err
	}
	return r.GetByID(ctx, id)
}

func (r *courseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	affected, err := r.queries.DeleteCourse(ctx, pgUUID(id))
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrCourseNotFound
	}
	return nil
}

func (r *courseRepository) GetLessonByID(ctx context.Context, id uuid.UUID) (domain.Lesson, error) {
	row, err := r.queries.GetLessonByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Lesson{}, ErrLessonNotFound
		}
		return domain.Lesson{}, err
	}
	return toDomainLesson(row)
}

func (r *courseRepository) ListLessons(ctx context.Context, courseID uuid.UUID) ([]domain.Lesson, error) {
	rows, err := r.queries.ListCourseLessons(ctx, pgUUID(courseID))
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
		CourseID:  pgUUID(courseID),
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
		ID:        pgUUID(lesson.ID),
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
	affected, err := r.queries.DeleteLesson(ctx, pgUUID(id))
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrLessonNotFound
	}
	return nil
}

func (r *courseRepository) Enroll(ctx context.Context, studentID, courseID uuid.UUID) error {
	_, err := r.queries.CreateEnrollment(ctx, db.CreateEnrollmentParams{
		UserID:   pgUUID(studentID),
		CourseID: pgUUID(courseID),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrEnrollmentExists
		}
	}
	return err
}

func (r *courseRepository) Unenroll(ctx context.Context, studentID, courseID uuid.UUID) error {
	affected, err := r.queries.DeleteEnrollment(ctx, db.DeleteEnrollmentParams{
		UserID:   pgUUID(studentID),
		CourseID: pgUUID(courseID),
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
		UserID:   pgUUID(studentID),
		CourseID: pgUUID(courseID),
	})
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func toDomainLesson(row db.Lesson) (domain.Lesson, error) {
	id, err := fromPgUUID(row.ID)
	if err != nil {
		return domain.Lesson{}, fmt.Errorf("parse lesson id: %w", err)
	}
	courseID, err := fromPgUUID(row.CourseID)
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

func toDomainCourseFromFields(
	idPg pgtype.UUID,
	instructorIDPg pgtype.UUID,
	instructorName string,
	email string,
	title string,
	slug string,
	description pgtype.Text,
	published bool,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
) (domain.Course, error) {
	id, err := fromPgUUID(idPg)
	if err != nil {
		return domain.Course{}, fmt.Errorf("parse course id: %w", err)
	}
	instructorID, err := fromPgUUID(instructorIDPg)
	if err != nil {
		return domain.Course{}, fmt.Errorf("parse instructor id: %w", err)
	}

	return domain.Course{
		ID: id,
		Instructor: domain.Instructor{
			ID:       instructorID,
			FullName: instructorName,
			Email:    email,
		},
		Title:       title,
		Slug:        slug,
		Description: description.String,
		Published:   published,
		CreatedAt:   createdAt.Time,
		UpdatedAt:   updatedAt.Time,
	}, nil
}

func toDomainCourses(rows any) ([]domain.Course, error) {
	switch rs := rows.(type) {
	case []db.GetCoursesRow:
		out := make([]domain.Course, 0, len(rs))
		for _, r := range rs {
			c, err := toDomainCourseFromFields(r.ID, r.InstructorID, r.InstructorName, r.Email, r.Title, r.Slug, r.Description, r.Published, r.CreatedAt, r.UpdatedAt)
			if err != nil {
				return nil, err
			}
			out = append(out, c)
		}
		return out, nil
	case []db.GetEnrolledCoursesRow:
		out := make([]domain.Course, 0, len(rs))
		for _, r := range rs {
			c, err := toDomainCourseFromFields(r.ID, r.InstructorID, r.InstructorName, r.Email, r.Title, r.Slug, r.Description, r.Published, r.CreatedAt, r.UpdatedAt)
			if err != nil {
				return nil, err
			}
			out = append(out, c)
		}
		return out, nil
	case []db.GetInstructorCoursesRow:
		out := make([]domain.Course, 0, len(rs))
		for _, r := range rs {
			c, err := toDomainCourseFromFields(r.ID, r.InstructorID, r.InstructorName, r.Email, r.Title, r.Slug, r.Description, r.Published, r.CreatedAt, r.UpdatedAt)
			if err != nil {
				return nil, err
			}
			out = append(out, c)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported rows type for toDomainCourses: %T", rows)
	}
}
