package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/pubsub"
	"github.com/n0m-d/DVAPI/internal/repository"
)

type ResubmitInput struct {
	SubmissionID   uuid.UUID
	StudentID      uuid.UUID
	SubmissionText string
	FilePath       string
	FileName       string
}

type LearningService interface {
	SetLessonProgress(ctx context.Context, lessonID, studentID uuid.UUID, completed bool) error
	GetCourseProgress(ctx context.Context, courseID, studentID uuid.UUID) (domain.CourseProgress, error)
	GetNextLesson(ctx context.Context, courseID, studentID uuid.UUID) (domain.Lesson, error)
	GetGrades(ctx context.Context, studentID uuid.UUID) (domain.StudentGradesDashboard, error)
	Resubmit(ctx context.Context, input ResubmitInput) (domain.AssignmentSubmission, error)
	ListSubmissionVersions(ctx context.Context, submissionID uuid.UUID) ([]domain.SubmissionVersion, error)
	CreateAnnouncement(ctx context.Context, courseID, instructorID uuid.UUID, input domain.CreateAnnouncementRequest) (domain.Announcement, error)
	ListAnnouncements(ctx context.Context, courseID, userID uuid.UUID, instructor bool) ([]domain.Announcement, error)
	UpdateAnnouncement(ctx context.Context, id, instructorID uuid.UUID, input domain.UpdateAnnouncementRequest) (domain.Announcement, error)
	DeleteAnnouncement(ctx context.Context, id, instructorID uuid.UUID) error
	GetCourseAnalytics(ctx context.Context, courseID, instructorID uuid.UUID) (domain.CourseAnalytics, error)
	GetInstructorStats(ctx context.Context, instructorID uuid.UUID) (domain.InstructorStats, error)
	GetStudentStats(ctx context.Context, studentID uuid.UUID) (domain.StudentStats, error)
}

type learningService struct {
	learning      repository.LearningRepository
	courses       repository.CourseRepository
	assignments   repository.AssignmentRepository
	notifications repository.NotificationRepository
	notifier      pubsub.Notifier
}

func NewLearningService(
	learning repository.LearningRepository,
	courses repository.CourseRepository,
	assignments repository.AssignmentRepository,
	notifications repository.NotificationRepository,
	notifier pubsub.Notifier,
) LearningService {
	return &learningService{
		learning:      learning,
		courses:       courses,
		assignments:   assignments,
		notifications: notifications,
		notifier:      notifier,
	}
}

func (s *learningService) SetLessonProgress(ctx context.Context, lessonID, studentID uuid.UUID, completed bool) error {
	lesson, err := s.courses.GetLessonByID(ctx, lessonID)
	if err != nil {
		return mapLearningError(err)
	}
	if err := s.requireEnrollment(ctx, studentID, lesson.CourseID); err != nil {
		return err
	}
	return s.learning.SetLessonProgress(ctx, studentID, lessonID, completed)
}

func (s *learningService) GetCourseProgress(ctx context.Context, courseID, studentID uuid.UUID) (domain.CourseProgress, error) {
	if err := s.requireEnrollment(ctx, studentID, courseID); err != nil {
		return domain.CourseProgress{}, err
	}
	return s.learning.GetCourseProgress(ctx, studentID, courseID)
}

func (s *learningService) GetNextLesson(ctx context.Context, courseID, studentID uuid.UUID) (domain.Lesson, error) {
	if err := s.requireEnrollment(ctx, studentID, courseID); err != nil {
		return domain.Lesson{}, err
	}
	lesson, err := s.learning.GetNextIncompleteLesson(ctx, studentID, courseID)
	if err != nil {
		return domain.Lesson{}, mapLearningError(err)
	}
	return lesson, nil
}

func (s *learningService) GetGrades(ctx context.Context, studentID uuid.UUID) (domain.StudentGradesDashboard, error) {
	return s.learning.GetStudentGrades(ctx, studentID)
}

func (s *learningService) Resubmit(ctx context.Context, input ResubmitInput) (domain.AssignmentSubmission, error) {
	text := strings.TrimSpace(input.SubmissionText)
	if text == "" || input.FilePath == "" {
		return domain.AssignmentSubmission{}, fmt.Errorf("%w: submission_text and file are required", ErrInvalidInput)
	}
	submission, err := s.assignments.GetSubmissionByID(ctx, input.SubmissionID)
	if err != nil {
		return domain.AssignmentSubmission{}, mapLearningError(err)
	}
	if submission.StudentID != input.StudentID {
		return domain.AssignmentSubmission{}, ErrForbidden
	}
	assignment, err := s.assignments.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		return domain.AssignmentSubmission{}, mapLearningError(err)
	}
	if assignment.Status != domain.AssignmentStatusPublished {
		return domain.AssignmentSubmission{}, ErrAssignmentClosed
	}
	if time.Now().After(assignment.DueDate) {
		return domain.AssignmentSubmission{}, ErrPastDue
	}
	return s.learning.Resubmit(ctx, input.SubmissionID, text, input.FilePath, input.FileName)
}

func (s *learningService) ListSubmissionVersions(ctx context.Context, submissionID uuid.UUID) ([]domain.SubmissionVersion, error) {
	submission, err := s.assignments.GetSubmissionByID(ctx, submissionID)
	if err != nil {
		return nil, mapLearningError(err)
	}
	// if submission.StudentID != studentID { //check commented out for vuln
	// 	return nil, ErrForbidden
	// }
	return s.learning.ListSubmissionVersions(ctx, submission.ID)
}

func (s *learningService) CreateAnnouncement(ctx context.Context, courseID, instructorID uuid.UUID, input domain.CreateAnnouncementRequest) (domain.Announcement, error) {
	// if _, err := s.requireOwnedCourse(ctx, courseID, instructorID); err != nil { //Vuln Added: Ownership Check removed
	// 	return domain.Announcement{}, err
	// }
	title, content := strings.TrimSpace(input.Title), strings.TrimSpace(input.Content)
	if title == "" || content == "" || !validAnnouncementStatus(input.Status) {
		return domain.Announcement{}, fmt.Errorf("%w: title, content, and valid status are required", ErrInvalidInput)
	}
	announcement, err := s.learning.CreateAnnouncement(ctx, courseID, instructorID, title, content, input.Status)
	if err != nil {
		return domain.Announcement{}, err
	}
	if announcement.Status == domain.AnnouncementStatusPublished {
		s.notifyAnnouncement(ctx, announcement)
	}
	return announcement, nil
}

func (s *learningService) ListAnnouncements(ctx context.Context, courseID, userID uuid.UUID, instructor bool) ([]domain.Announcement, error) {
	if instructor {
		if _, err := s.requireOwnedCourse(ctx, courseID, userID); err != nil {
			return nil, err
		}
		return s.learning.ListAnnouncements(ctx, courseID, false)
	}
	if err := s.requireEnrollment(ctx, userID, courseID); err != nil {
		return nil, err
	}
	return s.learning.ListAnnouncements(ctx, courseID, true)
}

func (s *learningService) UpdateAnnouncement(ctx context.Context, id, instructorID uuid.UUID, input domain.UpdateAnnouncementRequest) (domain.Announcement, error) {
	if input.Title == nil && input.Content == nil && input.Status == nil {
		return domain.Announcement{}, fmt.Errorf("%w: at least one field is required", ErrInvalidInput)
	}
	announcement, err := s.learning.GetAnnouncement(ctx, id)
	if err != nil {
		return domain.Announcement{}, mapLearningError(err)
	}
	if _, err := s.requireOwnedCourse(ctx, announcement.CourseID, instructorID); err != nil {
		return domain.Announcement{}, err
	}
	previousStatus := announcement.Status
	if input.Title != nil {
		announcement.Title = strings.TrimSpace(*input.Title)
	}
	if input.Content != nil {
		announcement.Content = strings.TrimSpace(*input.Content)
	}
	if input.Status != nil {
		if !validAnnouncementStatus(*input.Status) {
			return domain.Announcement{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
		}
		announcement.Status = *input.Status
	}
	if announcement.Title == "" || announcement.Content == "" {
		return domain.Announcement{}, fmt.Errorf("%w: title and content cannot be empty", ErrInvalidInput)
	}
	updated, err := s.learning.UpdateAnnouncement(ctx, announcement)
	if err != nil {
		return domain.Announcement{}, err
	}
	if previousStatus != domain.AnnouncementStatusPublished && updated.Status == domain.AnnouncementStatusPublished {
		s.notifyAnnouncement(ctx, updated)
	}
	return updated, nil
}

func (s *learningService) DeleteAnnouncement(ctx context.Context, id, instructorID uuid.UUID) error {
	announcement, err := s.learning.GetAnnouncement(ctx, id)
	if err != nil {
		return mapLearningError(err)
	}
	// if _, err := s.requireOwnedCourse(ctx, announcement.CourseID, instructorID); err != nil { //Vuln: Removed Ownership Check
	// 	return err
	// }
	return mapLearningError(s.learning.DeleteAnnouncement(ctx, announcement.ID))
}

func (s *learningService) GetCourseAnalytics(ctx context.Context, courseID, instructorID uuid.UUID) (domain.CourseAnalytics, error) {
	if _, err := s.requireOwnedCourse(ctx, courseID, instructorID); err != nil {
		return domain.CourseAnalytics{}, err
	}
	return s.learning.GetCourseAnalytics(ctx, courseID)
}

func (s *learningService) GetInstructorStats(ctx context.Context, instructorID uuid.UUID) (domain.InstructorStats, error) {
	if instructorID == uuid.Nil {
		return domain.InstructorStats{}, fmt.Errorf("%w: instructor id is required", ErrInvalidInput)
	}
	return s.learning.GetInstructorStats(ctx, instructorID)
}

func (s *learningService) GetStudentStats(ctx context.Context, studentID uuid.UUID) (domain.StudentStats, error) {
	if studentID == uuid.Nil {
		return domain.StudentStats{}, fmt.Errorf("%w: student id is required", ErrInvalidInput)
	}
	return s.learning.GetStudentStats(ctx, studentID)
}

func (s *learningService) notifyAnnouncement(ctx context.Context, announcement domain.Announcement) {
	var courseTitle string
	if course, courseErr := s.courses.GetByID(ctx, announcement.CourseID); courseErr == nil {
		courseTitle = course.Title
	}
	notifyEnrolledStudentsAsync(s.courses, s.notifications, s.notifier, announcement.CourseID, domain.Notification{
		Type:     domain.NotificationTypeAnnouncement,
		Title:    "New announcement",
		Body:     fmt.Sprintf("%s Announcement: %s", courseTitle, announcement.Title),
		CourseID: courseIDPtr(announcement.CourseID),
	})
}

func (s *learningService) requireEnrollment(ctx context.Context, studentID, courseID uuid.UUID) error {
	course, err := s.courses.GetByID(ctx, courseID)
	if err != nil {
		return mapLearningError(err)
	}

	// if !course.Published { // Published Status Check removal
	// 	return ErrNotFound
	// }

	enrolled, err := s.courses.IsEnrolled(ctx, studentID, course.ID)
	if err != nil {
		return err
	}
	if !enrolled {
		return ErrNotEnrolled
	}
	return nil
}

func (s *learningService) requireOwnedCourse(ctx context.Context, courseID, instructorID uuid.UUID) (domain.Course, error) {
	course, err := s.courses.GetByID(ctx, courseID)
	if err != nil {
		return domain.Course{}, mapLearningError(err)
	}
	if course.Instructor.ID != instructorID {
		return domain.Course{}, ErrForbidden
	}
	return course, nil
}

func validAnnouncementStatus(status string) bool {
	return status == domain.AnnouncementStatusDraft || status == domain.AnnouncementStatusPublished
}

func mapLearningError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repository.ErrCourseNotFound),
		errors.Is(err, repository.ErrLessonNotFound),
		errors.Is(err, repository.ErrSubmissionNotFound),
		errors.Is(err, repository.ErrAssignmentNotFound),
		errors.Is(err, repository.ErrAnnouncementNotFound):
		return ErrNotFound
	default:
		return err
	}
}
