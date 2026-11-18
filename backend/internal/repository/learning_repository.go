package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/n0m-d/DVAPI/internal/db"
	"github.com/n0m-d/DVAPI/internal/domain"
)

type LearningRepository interface {
	SetLessonProgress(ctx context.Context, studentID, lessonID uuid.UUID, completed bool) error
	GetNextIncompleteLesson(ctx context.Context, studentID, courseID uuid.UUID) (domain.Lesson, error)
	Resubmit(ctx context.Context, submissionID uuid.UUID, text, filePath, fileName string) (domain.AssignmentSubmission, error)
	ListSubmissionVersions(ctx context.Context, submissionID uuid.UUID) ([]domain.SubmissionVersion, error)
	CreateAnnouncement(ctx context.Context, courseID, createdBy uuid.UUID, title, content, status string) (domain.Announcement, error)
	GetAnnouncement(ctx context.Context, id uuid.UUID) (domain.Announcement, error)
	ListAnnouncements(ctx context.Context, courseID uuid.UUID, publishedOnly bool) ([]domain.Announcement, error)
	UpdateAnnouncement(ctx context.Context, announcement domain.Announcement) (domain.Announcement, error)
	DeleteAnnouncement(ctx context.Context, id uuid.UUID) error
}

type learningRepository struct {
	queries *db.Queries
}

func NewLearningRepository(queries *db.Queries) LearningRepository {
	return &learningRepository{queries: queries}
}

func (r *learningRepository) SetLessonProgress(ctx context.Context, studentID, lessonID uuid.UUID, completed bool) error {
	_, err := r.queries.UpsertLessonProgress(ctx, db.UpsertLessonProgressParams{
		StudentID: db.PgUUID(studentID),
		LessonID:  db.PgUUID(lessonID),
		Completed: completed,
	})
	return err
}

func (r *learningRepository) GetNextIncompleteLesson(ctx context.Context, studentID, courseID uuid.UUID) (domain.Lesson, error) {
	row, err := r.queries.GetNextIncompleteLesson(ctx, db.GetNextIncompleteLessonParams{
		StudentID: db.PgUUID(studentID),
		CourseID:  db.PgUUID(courseID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Lesson{}, ErrLessonNotFound
		}
		return domain.Lesson{}, err
	}
	return toDomainLesson(row)
}

func (r *learningRepository) Resubmit(ctx context.Context, submissionID uuid.UUID, text, filePath, fileName string) (domain.AssignmentSubmission, error) {
	row, err := r.queries.ResubmitAssignment(ctx, db.ResubmitAssignmentParams{
		ID:             db.PgUUID(submissionID),
		SubmissionText: pgtype.Text{String: text, Valid: text != ""},
		FilePath:       pgtype.Text{String: filePath, Valid: filePath != ""},
		FileName:       pgtype.Text{String: fileName, Valid: fileName != ""},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AssignmentSubmission{}, ErrSubmissionNotFound
		}
		return domain.AssignmentSubmission{}, err
	}
	return toDomainSubmission(
		row.ID, row.AssignmentID, row.StudentID, row.SubmissionText, row.FilePath,
		row.FileName, row.SubmittedAt, row.Grade, row.Feedback, row.CreatedAt, row.UpdatedAt,
	)
}

func (r *learningRepository) ListSubmissionVersions(ctx context.Context, submissionID uuid.UUID) ([]domain.SubmissionVersion, error) {
	rows, err := r.queries.ListSubmissionVersions(ctx, db.PgUUID(submissionID))
	if err != nil {
		return nil, err
	}
	versions := make([]domain.SubmissionVersion, 0, len(rows))
	for _, row := range rows {
		id, err := db.FromPgUUID(row.ID)
		if err != nil {
			return nil, err
		}
		parentID, err := db.FromPgUUID(row.SubmissionID)
		if err != nil {
			return nil, err
		}
		versions = append(versions, domain.SubmissionVersion{
			ID:             id,
			SubmissionID:   parentID,
			Version:        int(row.Version),
			SubmissionText: row.SubmissionText.String,
			FilePath:       row.FilePath.String,
			FileName:       row.FileName.String,
			SubmittedAt:    row.SubmittedAt.Time,
			CreatedAt:      row.CreatedAt.Time,
		})
	}
	return versions, nil
}

func (r *learningRepository) CreateAnnouncement(ctx context.Context, courseID, createdBy uuid.UUID, title, content, status string) (domain.Announcement, error) {
	row, err := r.queries.CreateAnnouncement(ctx, db.CreateAnnouncementParams{
		CourseID:  db.PgUUID(courseID),
		Title:     title,
		Content:   content,
		Status:    db.AnnouncementStatus(status),
		CreatedBy: db.PgUUID(createdBy),
	})
	if err != nil {
		return domain.Announcement{}, err
	}
	return toDomainAnnouncement(row)
}

func (r *learningRepository) GetAnnouncement(ctx context.Context, id uuid.UUID) (domain.Announcement, error) {
	row, err := r.queries.GetAnnouncementByID(ctx, db.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Announcement{}, ErrAnnouncementNotFound
		}
		return domain.Announcement{}, err
	}
	return toDomainAnnouncement(row)
}

func (r *learningRepository) ListAnnouncements(ctx context.Context, courseID uuid.UUID, publishedOnly bool) ([]domain.Announcement, error) {
	var rows []db.Announcement
	var err error
	if publishedOnly {
		rows, err = r.queries.ListPublishedAnnouncements(ctx, db.PgUUID(courseID))
	} else {
		rows, err = r.queries.ListInstructorAnnouncements(ctx, db.PgUUID(courseID))
	}
	if err != nil {
		return nil, err
	}
	out := make([]domain.Announcement, 0, len(rows))
	for _, row := range rows {
		announcement, err := toDomainAnnouncement(row)
		if err != nil {
			return nil, err
		}
		out = append(out, announcement)
	}
	return out, nil
}

func (r *learningRepository) UpdateAnnouncement(ctx context.Context, announcement domain.Announcement) (domain.Announcement, error) {
	row, err := r.queries.UpdateAnnouncement(ctx, db.UpdateAnnouncementParams{
		ID:      db.PgUUID(announcement.ID),
		Title:   announcement.Title,
		Content: announcement.Content,
		Status:  db.AnnouncementStatus(announcement.Status),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Announcement{}, ErrAnnouncementNotFound
		}
		return domain.Announcement{}, err
	}
	return toDomainAnnouncement(row)
}

func (r *learningRepository) DeleteAnnouncement(ctx context.Context, id uuid.UUID) error {
	affected, err := r.queries.DeleteAnnouncement(ctx, db.PgUUID(id))
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrAnnouncementNotFound
	}
	return nil
}

func toDomainAnnouncement(row db.Announcement) (domain.Announcement, error) {
	id, err := db.FromPgUUID(row.ID)
	if err != nil {
		return domain.Announcement{}, err
	}
	courseID, err := db.FromPgUUID(row.CourseID)
	if err != nil {
		return domain.Announcement{}, err
	}
	createdBy, err := db.FromPgUUID(row.CreatedBy)
	if err != nil {
		return domain.Announcement{}, err
	}
	return domain.Announcement{
		ID: id, CourseID: courseID, Title: row.Title, Content: row.Content,
		Status: string(row.Status), CreatedBy: createdBy,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}, nil
}
