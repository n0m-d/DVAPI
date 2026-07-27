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
	GetCourseProgress(ctx context.Context, studentID, courseID uuid.UUID) (domain.CourseProgress, error)
	GetNextIncompleteLesson(ctx context.Context, studentID, courseID uuid.UUID) (domain.Lesson, error)
	GetStudentGrades(ctx context.Context, studentID uuid.UUID) (domain.StudentGradesDashboard, error)
	Resubmit(ctx context.Context, submissionID uuid.UUID, text, filePath, fileName string) (domain.AssignmentSubmission, error)
	ListSubmissionVersions(ctx context.Context, submissionID uuid.UUID) ([]domain.SubmissionVersion, error)
	CreateAnnouncement(ctx context.Context, courseID, createdBy uuid.UUID, title, content, status string) (domain.Announcement, error)
	GetAnnouncement(ctx context.Context, id uuid.UUID) (domain.Announcement, error)
	ListAnnouncements(ctx context.Context, courseID uuid.UUID, publishedOnly bool) ([]domain.Announcement, error)
	UpdateAnnouncement(ctx context.Context, announcement domain.Announcement) (domain.Announcement, error)
	DeleteAnnouncement(ctx context.Context, id uuid.UUID) error
	GetCourseAnalytics(ctx context.Context, courseID uuid.UUID) (domain.CourseAnalytics, error)
	GetInstructorStats(ctx context.Context, instructorID uuid.UUID) (domain.InstructorStats, error)
	GetStudentStats(ctx context.Context, studentID uuid.UUID) (domain.StudentStats, error)
}

type learningRepository struct {
	queries *db.Queries
}

func NewLearningRepository(queries *db.Queries) LearningRepository {
	return &learningRepository{queries: queries}
}

func (r *learningRepository) SetLessonProgress(ctx context.Context, studentID, lessonID uuid.UUID, completed bool) error {
	_, err := r.queries.UpsertLessonProgress(ctx, db.UpsertLessonProgressParams{
		StudentID: pgUUID(studentID),
		LessonID:  pgUUID(lessonID),
		Completed: completed,
	})
	return err
}

func (r *learningRepository) GetCourseProgress(ctx context.Context, studentID, courseID uuid.UUID) (domain.CourseProgress, error) {
	row, err := r.queries.GetCourseProgress(ctx, db.GetCourseProgressParams{
		StudentID: pgUUID(studentID),
		CourseID:  pgUUID(courseID),
	})
	if err != nil {
		return domain.CourseProgress{}, err
	}
	total, completed := int(row.TotalLessons), int(row.CompletedLessons)
	percentage := 0.0
	if total > 0 {
		percentage = float64(completed) / float64(total) * 100
	}
	return domain.CourseProgress{
		CourseID:         courseID,
		TotalLessons:     total,
		CompletedLessons: completed,
		Percentage:       percentage,
	}, nil
}

func (r *learningRepository) GetNextIncompleteLesson(ctx context.Context, studentID, courseID uuid.UUID) (domain.Lesson, error) {
	row, err := r.queries.GetNextIncompleteLesson(ctx, db.GetNextIncompleteLessonParams{
		StudentID: pgUUID(studentID),
		CourseID:  pgUUID(courseID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Lesson{}, ErrLessonNotFound
		}
		return domain.Lesson{}, err
	}
	return toDomainLesson(row)
}

func (r *learningRepository) GetStudentGrades(ctx context.Context, studentID uuid.UUID) (domain.StudentGradesDashboard, error) {
	rows, err := r.queries.ListStudentGrades(ctx, pgUUID(studentID))
	if err != nil {
		return domain.StudentGradesDashboard{}, err
	}
	summary, err := r.queries.GetStudentGradeSummary(ctx, pgUUID(studentID))
	if err != nil {
		return domain.StudentGradesDashboard{}, err
	}
	grades := make([]domain.StudentGrade, 0, len(rows))
	for _, row := range rows {
		submissionID, err := fromPgUUID(row.SubmissionID)
		if err != nil {
			return domain.StudentGradesDashboard{}, err
		}
		assignmentID, err := fromPgUUID(row.AssignmentID)
		if err != nil {
			return domain.StudentGradesDashboard{}, err
		}
		courseID, err := fromPgUUID(row.CourseID)
		if err != nil {
			return domain.StudentGradesDashboard{}, err
		}
		var grade *int
		if row.Grade.Valid {
			value := int(row.Grade.Int32)
			grade = &value
		}
		grades = append(grades, domain.StudentGrade{
			SubmissionID:    submissionID,
			AssignmentID:    assignmentID,
			AssignmentTitle: row.AssignmentTitle,
			CourseID:        courseID,
			CourseTitle:     row.CourseTitle,
			Grade:           grade,
			Feedback:        row.Feedback.String,
			SubmittedAt:     row.SubmittedAt.Time,
			UpdatedAt:       row.UpdatedAt.Time,
		})
	}
	return domain.StudentGradesDashboard{
		Submitted:    int(summary.Submitted),
		AverageGrade: summary.AverageGrade,
		Grades:       grades,
	}, nil
}

func (r *learningRepository) Resubmit(ctx context.Context, submissionID uuid.UUID, text, filePath, fileName string) (domain.AssignmentSubmission, error) {
	row, err := r.queries.ResubmitAssignment(ctx, db.ResubmitAssignmentParams{
		ID:             pgUUID(submissionID),
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
	rows, err := r.queries.ListSubmissionVersions(ctx, pgUUID(submissionID))
	if err != nil {
		return nil, err
	}
	versions := make([]domain.SubmissionVersion, 0, len(rows))
	for _, row := range rows {
		id, err := fromPgUUID(row.ID)
		if err != nil {
			return nil, err
		}
		parentID, err := fromPgUUID(row.SubmissionID)
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
		CourseID:  pgUUID(courseID),
		Title:     title,
		Content:   content,
		Status:    db.AnnouncementStatus(status),
		CreatedBy: pgUUID(createdBy),
	})
	if err != nil {
		return domain.Announcement{}, err
	}
	return toDomainAnnouncement(row)
}

func (r *learningRepository) GetAnnouncement(ctx context.Context, id uuid.UUID) (domain.Announcement, error) {
	row, err := r.queries.GetAnnouncementByID(ctx, pgUUID(id))
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
		rows, err = r.queries.ListPublishedAnnouncements(ctx, pgUUID(courseID))
	} else {
		rows, err = r.queries.ListInstructorAnnouncements(ctx, pgUUID(courseID))
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
		ID:      pgUUID(announcement.ID),
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
	affected, err := r.queries.DeleteAnnouncement(ctx, pgUUID(id))
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrAnnouncementNotFound
	}
	return nil
}

func (r *learningRepository) GetCourseAnalytics(ctx context.Context, courseID uuid.UUID) (domain.CourseAnalytics, error) {
	row, err := r.queries.GetCourseAnalytics(ctx, pgUUID(courseID))
	if err != nil {
		return domain.CourseAnalytics{}, err
	}
	return domain.CourseAnalytics{
		CourseID:          courseID,
		Enrollments:       int(row.Enrollments),
		Assignments:       int(row.Assignments),
		Submissions:       int(row.Submissions),
		AverageGrade:      row.AverageGrade,
		Lessons:           int(row.Lessons),
		LessonCompletions: int(row.LessonCompletions),
	}, nil
}

func (r *learningRepository) GetInstructorStats(ctx context.Context, instructorID uuid.UUID) (domain.InstructorStats, error) {
	row, err := r.queries.GetInstructorStats(ctx, pgUUID(instructorID))
	if err != nil {
		return domain.InstructorStats{}, err
	}
	return domain.InstructorStats{
		Courses:             int(row.Courses),
		PublishedCourses:    int(row.PublishedCourses),
		Enrollments:         int(row.Enrollments),
		Students:            int(row.Students),
		Lessons:             int(row.Lessons),
		Assignments:         int(row.Assignments),
		Submissions:         int(row.Submissions),
		UngradedSubmissions: int(row.UngradedSubmissions),
		Announcements:       int(row.Announcements),
	}, nil
}

func (r *learningRepository) GetStudentStats(ctx context.Context, studentID uuid.UUID) (domain.StudentStats, error) {
	row, err := r.queries.GetStudentStats(ctx, pgUUID(studentID))
	if err != nil {
		return domain.StudentStats{}, err
	}
	return domain.StudentStats{
		EnrolledCourses:    int(row.EnrolledCourses),
		CompletedLessons:   int(row.CompletedLessons),
		Submissions:        int(row.Submissions),
		GradedSubmissions:  int(row.GradedSubmissions),
		AverageGrade:       row.AverageGrade,
		PendingAssignments: int(row.PendingAssignments),
	}, nil
}

func toDomainAnnouncement(row db.Announcement) (domain.Announcement, error) {
	id, err := fromPgUUID(row.ID)
	if err != nil {
		return domain.Announcement{}, err
	}
	courseID, err := fromPgUUID(row.CourseID)
	if err != nil {
		return domain.Announcement{}, err
	}
	createdBy, err := fromPgUUID(row.CreatedBy)
	if err != nil {
		return domain.Announcement{}, err
	}
	return domain.Announcement{
		ID: id, CourseID: courseID, Title: row.Title, Content: row.Content,
		Status: string(row.Status), CreatedBy: createdBy,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}, nil
}
