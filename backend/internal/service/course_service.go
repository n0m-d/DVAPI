package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/email"
	"github.com/n0m-d/DVAPI/internal/pubsub"
	"github.com/n0m-d/DVAPI/internal/repository"
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

type courseService struct {
	courseRepository repository.CourseRepository
	notifications    repository.NotificationRepository
	notifier         pubsub.Notifier
	mail             email.Queue
}

func NewCourseService(courseRepository repository.CourseRepository, notifications repository.NotificationRepository,
	notifier pubsub.Notifier, mail email.Queue) *courseService {
	return &courseService{courseRepository: courseRepository, notifications: notifications,
		notifier: notifier, mail: mail}
}

func (s *courseService) GetByID(ctx context.Context, id string) (domain.Course, error) {
	idUUID, err := uuid.Parse(id)
	if err != nil {
		return domain.Course{}, fmt.Errorf("%w: invalid course id", domain.ErrInvalidInput)
	}
	course, err := s.courseRepository.GetByID(ctx, idUUID)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			return domain.Course{}, domain.ErrNotFound
		}
		return domain.Course{}, err
	}
	return course, nil
}

func (s *courseService) Create(ctx context.Context, instructorID uuid.UUID, input domain.CreateCourseRequest) (domain.Course, error) {
	title := strings.TrimSpace(input.Title)
	if instructorID == uuid.Nil || title == "" {
		return domain.Course{}, fmt.Errorf("%w: title is required", domain.ErrInvalidInput)
	}
	return s.courseRepository.Create(ctx, repository.SaveCourseInput{
		InstructorID: instructorID,
		Title:        title,
		Slug:         slug.Make(title),
		Description:  strings.TrimSpace(input.Description),
		Published:    input.Published,
	})
}

func (s *courseService) Update(ctx context.Context, courseID, instructorID uuid.UUID, input domain.UpdateCourseRequest) (domain.Course, error) {
	if input.Title == nil && input.Description == nil && input.Published == nil {
		return domain.Course{}, fmt.Errorf("%w: at least one field is required", domain.ErrInvalidInput)
	}
	course, err := s.ownedCourse(ctx, courseID, instructorID)
	if err != nil {
		return domain.Course{}, err
	}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return domain.Course{}, fmt.Errorf("%w: title cannot be empty", domain.ErrInvalidInput)
		}
		course.Title = title
		course.Slug = slug.Make(title)
	}
	if input.Description != nil {
		course.Description = strings.TrimSpace(*input.Description)
	}
	if input.Published != nil {
		course.Published = *input.Published
	}
	return s.courseRepository.Update(ctx, repository.SaveCourseInput{
		ID:           course.ID,
		InstructorID: instructorID,
		Title:        course.Title,
		Slug:         course.Slug,
		Description:  course.Description,
		Published:    course.Published,
	})
}

func (s *courseService) Delete(ctx context.Context, courseID, instructorID uuid.UUID) error {
	if _, err := s.ownedCourse(ctx, courseID, instructorID); err != nil {
		return err
	}
	if err := s.courseRepository.Delete(ctx, courseID); err != nil {
		return mapCourseError(err)
	}
	return nil
}

func (s *courseService) ListLessons(ctx context.Context, courseID, userID uuid.UUID, requireEnrollment bool) ([]domain.Lesson, error) {
	course, err := s.courseRepository.GetByID(ctx, courseID)
	if err != nil {
		return nil, mapCourseError(err)
	}
	if requireEnrollment {
		enrolled, err := s.courseRepository.IsEnrolled(ctx, userID, courseID)
		if err != nil {
			return nil, err
		}
		// if !enrolled || !course.Published { Published Status check removal
		// 	return nil, domain.ErrNotEnrolled
		// }
		if !enrolled {
			return nil, domain.ErrNotEnrolled
		}
	} else if course.Instructor.ID != userID {
		return nil, domain.ErrForbidden
	}
	return s.courseRepository.ListLessons(ctx, courseID)
}

func (s *courseService) CreateLesson(ctx context.Context, courseID, instructorID uuid.UUID, input domain.CreateLessonRequest) (domain.Lesson, error) {
	if _, err := s.ownedCourse(ctx, courseID, instructorID); err != nil {
		return domain.Lesson{}, err
	}
	title := strings.TrimSpace(input.Title)
	if title == "" || input.SortOrder < 0 {
		return domain.Lesson{}, fmt.Errorf("%w: valid title and sort_order are required", domain.ErrInvalidInput)
	}
	return s.courseRepository.CreateLesson(ctx, courseID, title, strings.TrimSpace(input.Content), input.SortOrder)
}

func (s *courseService) UpdateLesson(ctx context.Context, lessonID, instructorID uuid.UUID, input domain.UpdateLessonRequest) (domain.Lesson, error) {
	if input.Title == nil && input.SortOrder == nil && input.Content == nil {
		return domain.Lesson{}, fmt.Errorf("%w: at least one field is required", domain.ErrInvalidInput)
	}
	lesson, err := s.courseRepository.GetLessonByID(ctx, lessonID)
	if err != nil {
		return domain.Lesson{}, mapCourseError(err)
	}
	if _, err := s.ownedCourse(ctx, lesson.CourseID, instructorID); err != nil {
		return domain.Lesson{}, err
	}
	if input.Title != nil {
		lesson.Title = strings.TrimSpace(*input.Title)
		if lesson.Title == "" {
			return domain.Lesson{}, fmt.Errorf("%w: title cannot be empty", domain.ErrInvalidInput)
		}
	}
	if input.SortOrder != nil {
		if *input.SortOrder < 0 {
			return domain.Lesson{}, fmt.Errorf("%w: sort_order cannot be negative", domain.ErrInvalidInput)
		}
		lesson.SortOrder = *input.SortOrder
	}
	if input.Content != nil {
		lesson.Content = strings.TrimSpace(*input.Content)
	}
	return s.courseRepository.UpdateLesson(ctx, lesson)
}

func (s *courseService) DeleteLesson(ctx context.Context, lessonID, instructorID uuid.UUID) error {
	lesson, err := s.courseRepository.GetLessonByID(ctx, lessonID)
	if err != nil {
		return mapCourseError(err)
	}
	if _, err := s.ownedCourse(ctx, lesson.CourseID, instructorID); err != nil {
		return err
	}
	return mapCourseError(s.courseRepository.DeleteLesson(ctx, lessonID))
}

func (s *courseService) Enroll(ctx context.Context, courseID, studentID uuid.UUID) error {
	course, err := s.courseRepository.GetByID(ctx, courseID)
	if err != nil {
		return mapCourseError(err)
	}

	student, err := s.courseRepository.Enroll(ctx, studentID, course.ID)
	if err != nil {
		if errors.Is(err, repository.ErrEnrollmentExists) {
			return domain.ErrAlreadyExists
		}
		return err
	}
	notifyEnrollmentChange(s.notifications, s.notifier, course, studentID, true)

	if s.mail != nil && student.Email != "" {
		if err := s.mail.EnqueueEnrollment(ctx, student.Email, student.FullName, course.Title); err != nil {
			slog.Error("failed to enqueue enrollment email", "error", err)
		}
	}

	return nil
}

func (s *courseService) Unenroll(ctx context.Context, courseID, studentID uuid.UUID) error {
	course, err := s.courseRepository.GetByID(ctx, courseID)
	if err != nil {
		return mapCourseError(err)
	}
	if err := s.courseRepository.Unenroll(ctx, studentID, course.ID); err != nil {
		if errors.Is(err, repository.ErrEnrollmentMissing) {
			return domain.ErrNotEnrolled
		}
		return err
	}
	notifyEnrollmentChange(s.notifications, s.notifier, course, studentID, false)
	return nil
}

func (s *courseService) ownedCourse(ctx context.Context, courseID, instructorID uuid.UUID) (domain.Course, error) {
	course, err := s.courseRepository.GetByID(ctx, courseID)
	if err != nil {
		return domain.Course{}, mapCourseError(err)
	}
	if course.Instructor.ID != instructorID {
		return domain.Course{}, domain.ErrForbidden
	}
	return course, nil
}

func mapCourseError(err error) error {
	switch {
	case errors.Is(err, repository.ErrCourseNotFound), errors.Is(err, repository.ErrLessonNotFound):
		return domain.ErrNotFound
	default:
		return err
	}
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}
