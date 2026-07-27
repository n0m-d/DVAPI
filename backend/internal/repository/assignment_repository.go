package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/n0m-d/DVAPI/internal/db"
	"github.com/n0m-d/DVAPI/internal/domain"
)

var (
	ErrAssignmentNotFound = errors.New("assignment not found")
	ErrSubmissionNotFound = errors.New("submission not found")
	ErrSubmissionExists   = errors.New("submission already exists")
	ErrEnrollmentNotFound = errors.New("enrollment not found")
)

type AssignmentRepository interface {
	Create(ctx context.Context, input CreateAssignmentInput) (domain.Assignment, error)
	Update(ctx context.Context, assignment domain.Assignment) (domain.Assignment, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Assignment, error)
	GetByCourseAndTitle(ctx context.Context, courseID uuid.UUID, title string) (domain.Assignment, error)
	ListPublishedByCourse(ctx context.Context, courseID uuid.UUID) ([]domain.Assignment, error)
	ListByCourse(ctx context.Context, params ListAssignmentsByCourseParams) ([]domain.Assignment, error)
	CountByCourse(ctx context.Context, courseID uuid.UUID, title string) (int, error)
	CreateSubmission(ctx context.Context, input domain.CreateAssignmentSubmissionInput) (domain.AssignmentSubmission, error)
	GetSubmission(ctx context.Context, assignmentID, studentID uuid.UUID) (domain.AssignmentSubmission, error)
	GetSubmissionByID(ctx context.Context, submissionID uuid.UUID) (domain.AssignmentSubmission, error)
	GradeSubmission(ctx context.Context, submissionID uuid.UUID, grade int, feedback string) (domain.AssignmentSubmission, error)
	ListSubmissionsByAssignment(ctx context.Context, params ListSubmissionsByAssignmentParams) ([]domain.InstructorSubmission, error)
	CountSubmissionsByAssignment(ctx context.Context, assignmentID uuid.UUID, name string) (int, error)
	IsEnrolled(ctx context.Context, studentID, courseID uuid.UUID) (bool, error)
	CloseOverdue(ctx context.Context) ([]domain.Assignment, error)
}

type ListAssignmentsByCourseParams struct {
	CourseID uuid.UUID
	Title    string
	Limit    int
	Offset   int
}

type ListSubmissionsByAssignmentParams struct {
	AssignmentID uuid.UUID
	Name         string
	Limit        int
	Offset       int
}

type CreateAssignmentInput struct {
	CourseID    uuid.UUID
	Title       string
	Description string
	DueDate     time.Time
	Status      string
	CreatedBy   uuid.UUID
}

type assignmentRepository struct {
	queries *db.Queries
}

func NewAssignmentRepository(queries *db.Queries) AssignmentRepository {
	return &assignmentRepository{queries: queries}
}

func (r *assignmentRepository) Create(ctx context.Context, input CreateAssignmentInput) (domain.Assignment, error) {
	row, err := r.queries.CreateAssignment(ctx, db.CreateAssignmentParams{
		CourseID:    pgUUID(input.CourseID),
		Title:       input.Title,
		Description: pgtype.Text{String: input.Description, Valid: input.Description != ""},
		DueDate:     pgtype.Timestamptz{Time: input.DueDate, Valid: true},
		Status:      db.Status(input.Status),
		CreatedBy:   pgUUID(input.CreatedBy),
	})
	if err != nil {
		return domain.Assignment{}, mapAssignmentError(err)
	}
	return toDomainAssignment(row)
}

func (r *assignmentRepository) Update(ctx context.Context, assignment domain.Assignment) (domain.Assignment, error) {
	row, err := r.queries.UpdateAssignment(ctx, db.UpdateAssignmentParams{
		ID:          pgUUID(assignment.ID),
		Title:       assignment.Title,
		Description: pgtype.Text{String: assignment.Description, Valid: assignment.Description != ""},
		DueDate:     pgtype.Timestamptz{Time: assignment.DueDate, Valid: true},
		Status:      db.Status(assignment.Status),
	})
	if err != nil {
		return domain.Assignment{}, mapAssignmentError(err)
	}
	return toDomainAssignment(row)
}

func (r *assignmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	affected, err := r.queries.DeleteAssignment(ctx, pgUUID(id))
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrAssignmentNotFound
	}
	return nil
}

func (r *assignmentRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Assignment, error) {
	row, err := r.queries.GetAssignmentByID(ctx, pgUUID(id))
	if err != nil {
		return domain.Assignment{}, mapAssignmentError(err)
	}
	return toDomainAssignment(row)
}

func (r *assignmentRepository) GetByCourseAndTitle(ctx context.Context, courseID uuid.UUID, title string) (domain.Assignment, error) {
	row, err := r.queries.GetAssignmentByCourseAndTitle(ctx, db.GetAssignmentByCourseAndTitleParams{
		CourseID: pgUUID(courseID),
		Title:    title,
	})
	if err != nil {
		return domain.Assignment{}, mapAssignmentError(err)
	}
	return toDomainAssignment(row)
}

func (r *assignmentRepository) ListPublishedByCourse(ctx context.Context, courseID uuid.UUID) ([]domain.Assignment, error) {
	rows, err := r.queries.ListPublishedAssignmentsByCourse(ctx, pgUUID(courseID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.Assignment, 0, len(rows))
	for _, row := range rows {
		a, err := toDomainAssignment(row)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

func (r *assignmentRepository) ListByCourse(ctx context.Context, params ListAssignmentsByCourseParams) ([]domain.Assignment, error) {
	rows, err := r.queries.ListAssignmentsByCourse(ctx, db.ListAssignmentsByCourseParams{
		CourseID:    pgUUID(params.CourseID),
		Title:       params.Title,
		LimitCount:  int32(params.Limit),
		OffsetCount: int32(params.Offset),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.Assignment, 0, len(rows))
	for _, row := range rows {
		a, err := toDomainAssignment(row)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

func (r *assignmentRepository) CountByCourse(ctx context.Context, courseID uuid.UUID, title string) (int, error) {
	total, err := r.queries.CountAssignmentsByCourse(ctx, db.CountAssignmentsByCourseParams{
		CourseID: pgUUID(courseID),
		Title:    title,
	})
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

func (r *assignmentRepository) CreateSubmission(ctx context.Context, input domain.CreateAssignmentSubmissionInput) (domain.AssignmentSubmission, error) {
	row, err := r.queries.CreateAssignmentSubmission(ctx, db.CreateAssignmentSubmissionParams{
		AssignmentID:   pgUUID(input.AssignmentID),
		StudentID:      pgUUID(input.StudentID),
		SubmissionText: pgtype.Text{String: input.SubmissionText, Valid: true},
		FilePath:       pgtype.Text{String: input.FilePath, Valid: input.FilePath != ""},
		FileName:       pgtype.Text{String: input.FileName, Valid: input.FileName != ""},
	})
	if err != nil {
		return domain.AssignmentSubmission{}, mapAssignmentError(err)
	}
	return toDomainSubmission(
		row.ID,
		row.AssignmentID,
		row.StudentID,
		row.SubmissionText,
		row.FilePath,
		row.FileName,
		row.SubmittedAt,
		row.Grade,
		row.Feedback,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func (r *assignmentRepository) GetSubmission(ctx context.Context, assignmentID, studentID uuid.UUID) (domain.AssignmentSubmission, error) {
	row, err := r.queries.GetSubmissionByAssignmentAndStudent(ctx, db.GetSubmissionByAssignmentAndStudentParams{
		AssignmentID: pgUUID(assignmentID),
		StudentID:    pgUUID(studentID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AssignmentSubmission{}, ErrSubmissionNotFound
		}
		return domain.AssignmentSubmission{}, err
	}
	return toDomainSubmission(
		row.ID,
		row.AssignmentID,
		row.StudentID,
		row.SubmissionText,
		row.FilePath,
		row.FileName,
		row.SubmittedAt,
		row.Grade,
		row.Feedback,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func (r *assignmentRepository) GetSubmissionByID(ctx context.Context, submissionID uuid.UUID) (domain.AssignmentSubmission, error) {
	row, err := r.queries.GetSubmissionByID(ctx, pgUUID(submissionID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AssignmentSubmission{}, ErrSubmissionNotFound
		}
		return domain.AssignmentSubmission{}, err
	}
	return toDomainSubmission(
		row.ID,
		row.AssignmentID,
		row.StudentID,
		row.SubmissionText,
		row.FilePath,
		row.FileName,
		row.SubmittedAt,
		row.Grade,
		row.Feedback,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func (r *assignmentRepository) GradeSubmission(ctx context.Context, submissionID uuid.UUID, grade int, feedback string) (domain.AssignmentSubmission, error) {
	row, err := r.queries.GradeSubmission(ctx, db.GradeSubmissionParams{
		ID:       pgUUID(submissionID),
		Grade:    pgtype.Int4{Int32: int32(grade), Valid: true},
		Feedback: pgtype.Text{String: feedback, Valid: feedback != ""},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AssignmentSubmission{}, ErrSubmissionNotFound
		}
		return domain.AssignmentSubmission{}, err
	}
	return toDomainSubmission(
		row.ID,
		row.AssignmentID,
		row.StudentID,
		row.SubmissionText,
		row.FilePath,
		row.FileName,
		row.SubmittedAt,
		row.Grade,
		row.Feedback,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func (r *assignmentRepository) ListSubmissionsByAssignment(ctx context.Context, params ListSubmissionsByAssignmentParams) ([]domain.InstructorSubmission, error) {
	rows, err := r.queries.ListSubmissionsByAssignment(ctx, db.ListSubmissionsByAssignmentParams{
		AssignmentID: pgUUID(params.AssignmentID),
		Name:         params.Name,
		LimitCount:   int32(params.Limit),
		OffsetCount:  int32(params.Offset),
	})
	if err != nil {
		return nil, err
	}

	out := make([]domain.InstructorSubmission, 0, len(rows))
	for _, row := range rows {
		s, err := toDomainInstructorSubmission(row)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *assignmentRepository) CountSubmissionsByAssignment(ctx context.Context, assignmentID uuid.UUID, name string) (int, error) {
	total, err := r.queries.CountSubmissionsByAssignment(ctx, db.CountSubmissionsByAssignmentParams{
		AssignmentID: pgUUID(assignmentID),
		Name:         name,
	})
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

func (r *assignmentRepository) IsEnrolled(ctx context.Context, studentID, courseID uuid.UUID) (bool, error) {
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

func (r *assignmentRepository) CloseOverdue(ctx context.Context) ([]domain.Assignment, error) {
	rows, err := r.queries.CloseOverdueAssignments(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Assignment, 0, len(rows))
	for _, row := range rows {
		a, err := toDomainAssignment(row)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

func toDomainAssignment(row db.Assignment) (domain.Assignment, error) {
	id, err := fromPgUUID(row.ID)
	if err != nil {
		return domain.Assignment{}, fmt.Errorf("parse assignment id: %w", err)
	}
	courseID, err := fromPgUUID(row.CourseID)
	if err != nil {
		return domain.Assignment{}, fmt.Errorf("parse course id: %w", err)
	}
	createdBy, err := fromPgUUID(row.CreatedBy)
	if err != nil {
		return domain.Assignment{}, fmt.Errorf("parse created_by: %w", err)
	}
	return domain.Assignment{
		ID:          id,
		CourseID:    courseID,
		Title:       row.Title,
		Description: row.Description.String,
		DueDate:     row.DueDate.Time,
		Status:      string(row.Status),
		CreatedBy:   createdBy,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

func toDomainSubmission(
	idPg, assignmentIDPg, studentIDPg pgtype.UUID,
	submissionText, filePath, fileName pgtype.Text,
	submittedAt pgtype.Timestamptz,
	gradePg pgtype.Int4,
	feedback pgtype.Text,
	createdAt, updatedAt pgtype.Timestamptz,
) (domain.AssignmentSubmission, error) {
	id, err := fromPgUUID(idPg)
	if err != nil {
		return domain.AssignmentSubmission{}, fmt.Errorf("parse submission id: %w", err)
	}
	assignmentID, err := fromPgUUID(assignmentIDPg)
	if err != nil {
		return domain.AssignmentSubmission{}, fmt.Errorf("parse assignment id: %w", err)
	}
	studentID, err := fromPgUUID(studentIDPg)
	if err != nil {
		return domain.AssignmentSubmission{}, fmt.Errorf("parse student id: %w", err)
	}
	grade := 0
	if gradePg.Valid {
		grade = int(gradePg.Int32)
	}
	return domain.AssignmentSubmission{
		ID:             id,
		AssignmentID:   assignmentID,
		StudentID:      studentID,
		SubmissionText: submissionText.String,
		FilePath:       filePath.String,
		FileName:       fileName.String,
		SubmittedAt:    submittedAt.Time,
		Grade:          grade,
		Feedback:       feedback.String,
		CreatedAt:      createdAt.Time,
		UpdatedAt:      updatedAt.Time,
	}, nil
}

func toDomainInstructorSubmission(row db.ListSubmissionsByAssignmentRow) (domain.InstructorSubmission, error) {
	submission, err := toDomainSubmission(
		row.ID,
		row.AssignmentID,
		row.StudentID,
		row.SubmissionText,
		row.FilePath,
		row.FileName,
		row.SubmittedAt,
		row.Grade,
		row.Feedback,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return domain.InstructorSubmission{}, err
	}
	return domain.InstructorSubmission{
		ID:             submission.ID,
		AssignmentID:   submission.AssignmentID,
		StudentID:      submission.StudentID,
		StudentName:    row.StudentName,
		StudentEmail:   row.StudentEmail,
		SubmissionText: submission.SubmissionText,
		FilePath:       submission.FilePath,
		FileName:       submission.FileName,
		SubmittedAt:    submission.SubmittedAt,
		Grade:          submission.Grade,
		Feedback:       submission.Feedback,
		CreatedAt:      submission.CreatedAt,
		UpdatedAt:      submission.UpdatedAt,
	}, nil
}

func mapAssignmentError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAssignmentNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrSubmissionExists
		case "23503":
			return ErrCourseNotFound
		}
	}
	return err
}
