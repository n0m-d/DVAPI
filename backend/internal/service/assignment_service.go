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

var (
	ErrNotEnrolled      = errors.New("not enrolled in course")
	ErrAssignmentClosed = errors.New("assignment is not open for submissions")
	ErrPastDue          = errors.New("assignment due date has passed")
	ErrForbidden        = errors.New("forbidden")
)

type AssignmentService interface {
	CreateAssignment(ctx context.Context, input domain.CreateAssignmentInput, userId uuid.UUID) (domain.Assignment, error)
	UpdateAssignment(ctx context.Context, assignmentID, instructorID uuid.UUID, input domain.UpdateAssignmentRequest) (domain.Assignment, error)
	DeleteAssignment(ctx context.Context, assignmentID, instructorID uuid.UUID) error
	ListPublishedByCourse(ctx context.Context, courseID, studentID uuid.UUID) ([]domain.Assignment, error)
	ListCourseAssignments(ctx context.Context, input ListCourseAssignmentsInput) (domain.AssignmentListResponse, error)
	ListSubmissions(ctx context.Context, input ListSubmissionsInput) (domain.SubmissionListResponse, error)
	GetByID(ctx context.Context, assignmentID, studentID uuid.UUID) (domain.Assignment, error)
	GetAccessibleByID(ctx context.Context, assignmentID, userID uuid.UUID, role string) (domain.Assignment, error)
	ValidateSubmission(ctx context.Context, assignmentID, studentID uuid.UUID, submissionText string) error
	CreateSubmission(ctx context.Context, input domain.CreateAssignmentSubmissionInput) (domain.AssignmentSubmission, error)
	GetMySubmission(ctx context.Context, assignmentID, studentID uuid.UUID) (domain.AssignmentSubmission, error)
	GradeSubmission(ctx context.Context, input domain.GradeSubmissionInput) (domain.AssignmentSubmission, error)
	GetSubmissionForInstructor(ctx context.Context, submissionID, instructorID uuid.UUID) (domain.AssignmentSubmission, error)
	CloseOverdue(ctx context.Context) ([]domain.Assignment, error)
}

type ListCourseAssignmentsInput struct {
	CourseID     uuid.UUID
	InstructorID uuid.UUID
	Title        string
	Page         int
	PageSize     int
}

type ListSubmissionsInput struct {
	AssignmentID uuid.UUID
	Name         string
	Page         int
	PageSize     int
}

type assignmentService struct {
	assignments   repository.AssignmentRepository
	courses       repository.CourseRepository
	notifications repository.NotificationRepository
	notifier      pubsub.Notifier
}

func NewAssignmentService(
	assignments repository.AssignmentRepository,
	courses repository.CourseRepository,
	notifications repository.NotificationRepository,
	notifier pubsub.Notifier,
) AssignmentService {
	return &assignmentService{
		assignments:   assignments,
		courses:       courses,
		notifications: notifications,
		notifier:      notifier,
	}
}

func (s *assignmentService) CreateAssignment(ctx context.Context, input domain.CreateAssignmentInput, userId uuid.UUID) (domain.Assignment, error) {
	if input.CourseID == uuid.Nil || strings.TrimSpace(input.Title) == "" {
		return domain.Assignment{}, fmt.Errorf("%w: course_id and title are required", ErrInvalidInput)
	}
	if input.Status != domain.AssignmentStatusDraft && input.Status != domain.AssignmentStatusPublished && input.Status != domain.AssignmentStatusClosed {
		return domain.Assignment{}, fmt.Errorf("%w: invalid status value", ErrInvalidInput)
	}
	if input.DueDate.IsZero() {
		return domain.Assignment{}, fmt.Errorf("%w: due_date is required", ErrInvalidInput)
	}

	course, err := s.courses.GetByID(ctx, input.CourseID)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			return domain.Assignment{}, ErrNotFound
		}
		return domain.Assignment{}, err
	}
	if course.Instructor.ID != userId {
		return domain.Assignment{}, ErrForbidden
	}

	assignment, err := s.assignments.Create(ctx, repository.CreateAssignmentInput{
		CourseID:    input.CourseID,
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		DueDate:     input.DueDate,
		Status:      input.Status,
		CreatedBy:   userId,
	})
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			return domain.Assignment{}, ErrNotFound
		}
		return domain.Assignment{}, err
	}
	if assignment.Status == domain.AssignmentStatusPublished {
		s.notifyAssignmentPublished(ctx, assignment, course.Title)
	}
	return assignment, nil
}

func (s *assignmentService) UpdateAssignment(ctx context.Context, assignmentID, instructorID uuid.UUID, input domain.UpdateAssignmentRequest) (domain.Assignment, error) {
	if input.Title == nil && input.Description == nil && input.DueDate == nil && input.Status == nil {
		return domain.Assignment{}, fmt.Errorf("%w: at least one field is required", ErrInvalidInput)
	}
	assignment, err := s.ownedAssignment(ctx, assignmentID, instructorID)
	if err != nil {
		return domain.Assignment{}, err
	}
	previousStatus := assignment.Status
	if input.Title != nil {
		assignment.Title = strings.TrimSpace(*input.Title)
		if assignment.Title == "" {
			return domain.Assignment{}, fmt.Errorf("%w: title cannot be empty", ErrInvalidInput)
		}
	}
	if input.Description != nil {
		assignment.Description = strings.TrimSpace(*input.Description)
	}
	if input.DueDate != nil {
		if input.DueDate.IsZero() {
			return domain.Assignment{}, fmt.Errorf("%w: due_date is invalid", ErrInvalidInput)
		}
		assignment.DueDate = *input.DueDate
	}
	if input.Status != nil {
		if !validAssignmentStatus(*input.Status) {
			return domain.Assignment{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
		}
		assignment.Status = *input.Status
	}
	updated, err := s.assignments.Update(ctx, assignment)
	if err != nil {
		return domain.Assignment{}, err
	}
	if previousStatus != domain.AssignmentStatusPublished && updated.Status == domain.AssignmentStatusPublished {
		courseTitle := "your course"
		if course, courseErr := s.courses.GetByID(ctx, updated.CourseID); courseErr == nil {
			courseTitle = course.Title
		}
		s.notifyAssignmentPublished(ctx, updated, courseTitle)
	}
	return updated, nil
}

func (s *assignmentService) DeleteAssignment(ctx context.Context, assignmentID, instructorID uuid.UUID) error {
	if _, err := s.ownedAssignment(ctx, assignmentID, instructorID); err != nil {
		return err
	}
	if err := s.assignments.Delete(ctx, assignmentID); err != nil {
		if errors.Is(err, repository.ErrAssignmentNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *assignmentService) ListPublishedByCourse(ctx context.Context, courseID, studentID uuid.UUID) ([]domain.Assignment, error) {
	if courseID == uuid.Nil || studentID == uuid.Nil {
		return nil, fmt.Errorf("%w: course_id and student are required", ErrInvalidInput)
	}

	enrolled, err := s.assignments.IsEnrolled(ctx, studentID, courseID)
	if err != nil {
		return nil, err
	}
	if !enrolled {
		return nil, ErrNotEnrolled
	}

	return s.assignments.ListPublishedByCourse(ctx, courseID)
}

func (s *assignmentService) ListCourseAssignments(ctx context.Context, input ListCourseAssignmentsInput) (domain.AssignmentListResponse, error) {
	course, err := s.courses.GetByID(ctx, input.CourseID)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			return domain.AssignmentListResponse{}, ErrNotFound
		}
		return domain.AssignmentListResponse{}, err
	}
	if course.Instructor.ID != input.InstructorID {
		return domain.AssignmentListResponse{}, ErrForbidden
	}

	page, pageSize := normalizePagination(input.Page, input.PageSize)
	offset := (page - 1) * pageSize
	title := strings.TrimSpace(input.Title)

	assignments, err := s.assignments.ListByCourse(ctx, repository.ListAssignmentsByCourseParams{
		CourseID: input.CourseID,
		Title:    title,
		Limit:    pageSize,
		Offset:   offset,
	})
	if err != nil {
		return domain.AssignmentListResponse{}, fmt.Errorf("list course assignments: %w", err)
	}

	total, err := s.assignments.CountByCourse(ctx, input.CourseID, title)
	if err != nil {
		return domain.AssignmentListResponse{}, fmt.Errorf("count course assignments: %w", err)
	}

	pagination := domain.NewPagination(total, page, pageSize)
	return domain.AssignmentListResponse{
		Status: "success",
		Data: &domain.AssignmentListData{
			Assignments: assignments,
			Pagination:  &pagination,
		},
	}, nil
}

func (s *assignmentService) ListSubmissions(ctx context.Context, input ListSubmissionsInput) (domain.SubmissionListResponse, error) {
	assignment, err := s.assignments.GetByID(ctx, input.AssignmentID)
	if err != nil {
		if errors.Is(err, repository.ErrAssignmentNotFound) {
			return domain.SubmissionListResponse{}, ErrNotFound
		}
		return domain.SubmissionListResponse{}, err
	}

	page, pageSize := normalizePagination(input.Page, input.PageSize)
	offset := (page - 1) * pageSize
	name := strings.TrimSpace(input.Name)

	submissions, err := s.assignments.ListSubmissionsByAssignment(ctx, repository.ListSubmissionsByAssignmentParams{
		AssignmentID: assignment.ID,
		Name:         name,
		Limit:        pageSize,
		Offset:       offset,
	})
	if err != nil {
		return domain.SubmissionListResponse{}, fmt.Errorf("list submissions: %w", err)
	}

	total, err := s.assignments.CountSubmissionsByAssignment(ctx, assignment.ID, name)
	if err != nil {
		return domain.SubmissionListResponse{}, fmt.Errorf("count submissions: %w", err)
	}

	pagination := domain.NewPagination(total, page, pageSize)
	return domain.SubmissionListResponse{
		Status: "success",
		Data: &domain.SubmissionListData{
			Submissions: submissions,
			Pagination:  &pagination,
		},
	}, nil
}

func (s *assignmentService) GetByID(ctx context.Context, assignmentID, studentID uuid.UUID) (domain.Assignment, error) {
	assignment, err := s.assignments.GetByID(ctx, assignmentID)
	if err != nil {
		if errors.Is(err, repository.ErrAssignmentNotFound) {
			return domain.Assignment{}, ErrNotFound
		}
		return domain.Assignment{}, err
	}

	enrolled, err := s.assignments.IsEnrolled(ctx, studentID, assignment.CourseID)
	if err != nil {
		return domain.Assignment{}, err
	}
	if !enrolled {
		return domain.Assignment{}, ErrNotEnrolled
	}

	if assignment.Status != domain.AssignmentStatusPublished && assignment.Status != domain.AssignmentStatusClosed {
		return domain.Assignment{}, ErrNotFound
	}

	return assignment, nil
}

func (s *assignmentService) GetAccessibleByID(ctx context.Context, assignmentID, userID uuid.UUID, role string) (domain.Assignment, error) {
	assignment, err := s.assignments.GetByID(ctx, assignmentID)
	if err != nil {
		if errors.Is(err, repository.ErrAssignmentNotFound) {
			return domain.Assignment{}, ErrNotFound
		}
		return domain.Assignment{}, err
	}

	switch role {
	case "admin":
		return assignment, nil
	case "instructor":
		course, err := s.courses.GetByID(ctx, assignment.CourseID)
		if err != nil {
			if errors.Is(err, repository.ErrCourseNotFound) {
				return domain.Assignment{}, ErrNotFound
			}
			return domain.Assignment{}, err
		}
		if course.Instructor.ID != userID {
			return domain.Assignment{}, ErrForbidden
		}
		return assignment, nil
	case "student":
		return s.GetByID(ctx, assignmentID, userID)
	default:
		return domain.Assignment{}, ErrForbidden
	}
}

func (s *assignmentService) ValidateSubmission(ctx context.Context, assignmentID, studentID uuid.UUID, submissionText string) error {
	submissionText = strings.TrimSpace(submissionText)
	if assignmentID == uuid.Nil || studentID == uuid.Nil {
		return fmt.Errorf("%w: assignment_id and student are required", ErrInvalidInput)
	}
	if submissionText == "" {
		return fmt.Errorf("%w: submission_text is required", ErrInvalidInput)
	}

	assignment, err := s.assignments.GetByID(ctx, assignmentID)
	if err != nil {
		if errors.Is(err, repository.ErrAssignmentNotFound) {
			return ErrNotFound
		}
		return err
	}

	if assignment.Status != domain.AssignmentStatusPublished {
		return ErrAssignmentClosed
	}
	if time.Now().After(assignment.DueDate) {
		return ErrPastDue
	}

	enrolled, err := s.assignments.IsEnrolled(ctx, studentID, assignment.CourseID)
	if err != nil {
		return err
	}
	if !enrolled {
		return ErrNotEnrolled
	}

	_, err = s.assignments.GetSubmission(ctx, assignmentID, studentID)
	if err == nil {
		return ErrAlreadyExists
	}
	if !errors.Is(err, repository.ErrSubmissionNotFound) {
		return err
	}

	return nil
}

func (s *assignmentService) CreateSubmission(ctx context.Context, input domain.CreateAssignmentSubmissionInput) (domain.AssignmentSubmission, error) {
	input.SubmissionText = strings.TrimSpace(input.SubmissionText)
	if input.FilePath == "" {
		return domain.AssignmentSubmission{}, fmt.Errorf("%w: file is required", ErrInvalidInput)
	}

	if err := s.ValidateSubmission(ctx, input.AssignmentID, input.StudentID, input.SubmissionText); err != nil {
		return domain.AssignmentSubmission{}, err
	}

	submission, err := s.assignments.CreateSubmission(ctx, input)
	if err != nil {
		if errors.Is(err, repository.ErrSubmissionExists) {
			return domain.AssignmentSubmission{}, ErrAlreadyExists
		}
		return domain.AssignmentSubmission{}, err
	}
	s.notifySubmissionCreated(ctx, submission)
	return submission, nil
}

func (s *assignmentService) GetMySubmission(ctx context.Context, assignmentID, studentID uuid.UUID) (domain.AssignmentSubmission, error) {
	if _, err := s.GetByID(ctx, assignmentID, studentID); err != nil {
		return domain.AssignmentSubmission{}, err
	}

	submission, err := s.assignments.GetSubmission(ctx, assignmentID, studentID)
	if err != nil {
		if errors.Is(err, repository.ErrSubmissionNotFound) {
			return domain.AssignmentSubmission{}, ErrNotFound
		}
		return domain.AssignmentSubmission{}, err
	}
	return submission, nil
}

func (s *assignmentService) GradeSubmission(ctx context.Context, input domain.GradeSubmissionInput) (domain.AssignmentSubmission, error) {
	if input.SubmissionID == uuid.Nil || input.InstructorID == uuid.Nil {
		return domain.AssignmentSubmission{}, fmt.Errorf("%w: submission and instructor are required", ErrInvalidInput)
	}
	// if input.Grade < 0 || input.Grade > 100 { //Vuln: Grade input
	// 	return domain.AssignmentSubmission{}, fmt.Errorf("%w: grade must be between 0 and 100", ErrInvalidInput)
	// }

	submission, err := s.assignments.GetSubmissionByID(ctx, input.SubmissionID)
	if err != nil {
		if errors.Is(err, repository.ErrSubmissionNotFound) {
			return domain.AssignmentSubmission{}, ErrNotFound
		}
		return domain.AssignmentSubmission{}, err
	}

	assignment, err := s.assignments.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		if errors.Is(err, repository.ErrAssignmentNotFound) {
			return domain.AssignmentSubmission{}, ErrNotFound
		}
		return domain.AssignmentSubmission{}, err
	}

	course, err := s.courses.GetByID(ctx, assignment.CourseID)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			return domain.AssignmentSubmission{}, ErrNotFound
		}
		return domain.AssignmentSubmission{}, err
	}
	if course.Instructor.ID != input.InstructorID {
		return domain.AssignmentSubmission{}, ErrForbidden
	}

	graded, err := s.assignments.GradeSubmission(
		ctx,
		input.SubmissionID,
		input.Grade,
		strings.TrimSpace(input.Feedback),
	)
	if err != nil {
		if errors.Is(err, repository.ErrSubmissionNotFound) {
			return domain.AssignmentSubmission{}, ErrNotFound
		}
		return domain.AssignmentSubmission{}, err
	}
	publishNotification(s.notifications, s.notifier, graded.StudentID, domain.Notification{
		Type:     domain.NotificationTypeGrade,
		Title:    "Grade posted",
		Body:     fmt.Sprintf("Your grade for %s is available.", assignment.Title),
		CourseID: courseIDPtr(assignment.CourseID),
	})
	return graded, nil
}

func (s *assignmentService) GetSubmissionForInstructor(ctx context.Context, submissionID, instructorID uuid.UUID) (domain.AssignmentSubmission, error) {
	submission, err := s.assignments.GetSubmissionByID(ctx, submissionID)
	if err != nil {
		if errors.Is(err, repository.ErrSubmissionNotFound) {
			return domain.AssignmentSubmission{}, ErrNotFound
		}
		return domain.AssignmentSubmission{}, err
	}
	if _, err := s.ownedAssignment(ctx, submission.AssignmentID, instructorID); err != nil {
		return domain.AssignmentSubmission{}, err
	}
	return submission, nil
}

func (s *assignmentService) ownedAssignment(ctx context.Context, assignmentID, instructorID uuid.UUID) (domain.Assignment, error) {
	assignment, err := s.assignments.GetByID(ctx, assignmentID)
	if err != nil {
		if errors.Is(err, repository.ErrAssignmentNotFound) {
			return domain.Assignment{}, ErrNotFound
		}
		return domain.Assignment{}, err
	}
	course, err := s.courses.GetByID(ctx, assignment.CourseID)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			return domain.Assignment{}, ErrNotFound
		}
		return domain.Assignment{}, err
	}
	if course.Instructor.ID != instructorID {
		return domain.Assignment{}, ErrForbidden
	}
	return assignment, nil
}

func (s *assignmentService) notifyAssignmentPublished(_ context.Context, assignment domain.Assignment, courseTitle string) {
	notifyEnrolledStudentsAsync(s.courses, s.notifications, s.notifier, assignment.CourseID, domain.Notification{
		Type:     domain.NotificationTypeAssignment,
		Title:    "New assignment",
		Body:     fmt.Sprintf("%s — %s is now available.", courseTitle, assignment.Title),
		CourseID: courseIDPtr(assignment.CourseID),
	})
}

func (s *assignmentService) notifySubmissionCreated(ctx context.Context, submission domain.AssignmentSubmission) {
	assignment, err := s.assignments.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		return
	} //Getting the assignment and course for the submission
	course, err := s.courses.GetByID(ctx, assignment.CourseID)
	if err != nil {
		return
	} //Getting the instructor for the course
	publishNotification(s.notifications, s.notifier, course.Instructor.ID, domain.Notification{ //Sending notification to the instructor
		Type:     domain.NotificationTypeAssignment,
		Title:    "New submission",
		Body:     fmt.Sprintf("A student submitted work for %s.", assignment.Title),
		CourseID: courseIDPtr(assignment.CourseID),
	})
}

func (s *assignmentService) CloseOverdue(ctx context.Context) ([]domain.Assignment, error) {
	return s.assignments.CloseOverdue(ctx)
}

func validAssignmentStatus(status string) bool {
	return status == domain.AssignmentStatusDraft ||
		status == domain.AssignmentStatusPublished ||
		status == domain.AssignmentStatusClosed
}
