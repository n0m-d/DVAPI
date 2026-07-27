package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/repository"
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

type CourseListInput struct {
	Published bool
	Title     string
	Page      int
	PageSize  int
}

type EnrolledCoursesInput struct {
	StudentID uuid.UUID
	Title     string
	Page      int
	PageSize  int
}

type ListCourseStudentsInput struct {
	CourseID     uuid.UUID
	InstructorID uuid.UUID
	Name         string
	Page         int
	PageSize     int
}

type InstructorCoursesInput struct {
	InstructorID uuid.UUID
	Title        string
	Page         int
	PageSize     int
}

type CourseService interface {
	GetCourses(ctx context.Context, input CourseListInput) (domain.CourseResponse, error)
	GetEnrolledCourses(ctx context.Context, input EnrolledCoursesInput) (domain.CourseResponse, error)
	GetByID(ctx context.Context, id string) (domain.Course, error)
	GetEnrolledCoursesCount(ctx context.Context, studentID uuid.UUID) (domain.EnrolledCourseCount, error)
	ListStudents(ctx context.Context, input ListCourseStudentsInput) (domain.EnrolledStudentsResponse, error)
	ListMyCourses(ctx context.Context, input InstructorCoursesInput) (domain.CourseResponse, error)
	Create(ctx context.Context, instructorID uuid.UUID, input domain.CreateCourseRequest) (domain.Course, error)
	Update(ctx context.Context, courseID, instructorID uuid.UUID, input domain.UpdateCourseRequest) (domain.Course, error)
	Delete(ctx context.Context, courseID, instructorID uuid.UUID) error
	ListLessons(ctx context.Context, courseID, userID uuid.UUID, requireEnrollment bool) ([]domain.Lesson, error)
	CreateLesson(ctx context.Context, courseID, instructorID uuid.UUID, input domain.CreateLessonRequest) (domain.Lesson, error)
	UpdateLesson(ctx context.Context, lessonID, instructorID uuid.UUID, input domain.UpdateLessonRequest) (domain.Lesson, error)
	DeleteLesson(ctx context.Context, lessonID, instructorID uuid.UUID) error
	Enroll(ctx context.Context, courseID, studentID uuid.UUID) error
	Unenroll(ctx context.Context, courseID, studentID uuid.UUID) error
}

type courseService struct {
	courseRepository repository.CourseRepository
}

func NewCourseService(courseRepository repository.CourseRepository) CourseService {
	return &courseService{courseRepository: courseRepository}
}

func (s *courseService) GetCourses(ctx context.Context, input CourseListInput) (domain.CourseResponse, error) {
	page, pageSize := normalizePagination(input.Page, input.PageSize)
	offset := (page - 1) * pageSize
	title := strings.TrimSpace(input.Title)

	courses, err := s.courseRepository.GetCourses(ctx, repository.CourseListParams{
		Published: input.Published,
		Title:     title,
		Limit:     pageSize,
		Offset:    offset,
	})
	if err != nil {
		return domain.CourseResponse{}, fmt.Errorf("list courses: %w", err)
	}

	total, err := s.courseRepository.CountCourses(ctx, input.Published, title)
	if err != nil {
		return domain.CourseResponse{}, fmt.Errorf("count courses: %w", err)
	}

	return courseListResponse(courses, total, page, pageSize), nil
}

func (s *courseService) GetByID(ctx context.Context, id string) (domain.Course, error) {
	idUUID, err := uuid.Parse(id)
	if err != nil {
		return domain.Course{}, fmt.Errorf("%w: invalid course id", ErrInvalidInput)
	}
	course, err := s.courseRepository.GetByID(ctx, idUUID)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			return domain.Course{}, ErrNotFound
		}
		return domain.Course{}, err
	}
	return course, nil
}

func (s *courseService) GetEnrolledCourses(ctx context.Context, input EnrolledCoursesInput) (domain.CourseResponse, error) {
	page, pageSize := normalizePagination(input.Page, input.PageSize)
	offset := (page - 1) * pageSize
	title := strings.TrimSpace(input.Title)

	courses, err := s.courseRepository.GetEnrolledCourses(ctx, repository.EnrolledCourseListParams{
		StudentID: input.StudentID,
		Title:     title,
		Limit:     pageSize,
		Offset:    offset,
	})
	if err != nil {
		return domain.CourseResponse{}, fmt.Errorf("get enrolled courses: %w", err)
	}

	total, err := s.courseRepository.CountEnrolledCourses(ctx, input.StudentID, title)
	if err != nil {
		return domain.CourseResponse{}, fmt.Errorf("count enrolled courses: %w", err)
	}

	return courseListResponse(courses, total, page, pageSize), nil
}

func (s *courseService) GetEnrolledCoursesCount(ctx context.Context, studentID uuid.UUID) (domain.EnrolledCourseCount, error) {
	count, err := s.courseRepository.CountEnrolledCourses(ctx, studentID, "")
	if err != nil {
		return domain.EnrolledCourseCount{}, fmt.Errorf("count enrolled courses: %w", err)
	}

	return domain.EnrolledCourseCount{Count: count}, nil
}

func (s *courseService) ListStudents(ctx context.Context, input ListCourseStudentsInput) (domain.EnrolledStudentsResponse, error) {
	course, err := s.courseRepository.GetByID(ctx, input.CourseID)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			return domain.EnrolledStudentsResponse{}, ErrNotFound
		}
		return domain.EnrolledStudentsResponse{}, err
	}
	if course.Instructor.ID != input.InstructorID {
		return domain.EnrolledStudentsResponse{}, ErrForbidden
	}

	page, pageSize := normalizePagination(input.Page, input.PageSize)
	offset := (page - 1) * pageSize
	name := strings.TrimSpace(input.Name)

	students, err := s.courseRepository.ListEnrolledStudents(ctx, repository.EnrolledStudentsListParams{
		CourseID: input.CourseID,
		Name:     name,
		Limit:    pageSize,
		Offset:   offset,
	})
	if err != nil {
		return domain.EnrolledStudentsResponse{}, fmt.Errorf("list enrolled students: %w", err)
	}

	total, err := s.courseRepository.CountEnrolledStudents(ctx, input.CourseID, name)
	if err != nil {
		return domain.EnrolledStudentsResponse{}, fmt.Errorf("count enrolled students: %w", err)
	}

	pagination := domain.NewPagination(total, page, pageSize)
	return domain.EnrolledStudentsResponse{
		Status: "success",
		Data: &domain.EnrolledStudentsData{
			Students:   students,
			Pagination: &pagination,
		},
	}, nil
}

func (s *courseService) ListMyCourses(ctx context.Context, input InstructorCoursesInput) (domain.CourseResponse, error) {
	page, pageSize := normalizePagination(input.Page, input.PageSize)
	offset := (page - 1) * pageSize
	title := strings.TrimSpace(input.Title)

	courses, err := s.courseRepository.GetInstructorCourses(ctx, repository.InstructorCourseListParams{
		InstructorID: input.InstructorID,
		Title:        title,
		Limit:        pageSize,
		Offset:       offset,
	})
	if err != nil {
		return domain.CourseResponse{}, fmt.Errorf("list instructor courses: %w", err)
	}

	total, err := s.courseRepository.CountInstructorCourses(ctx, input.InstructorID, title)
	if err != nil {
		return domain.CourseResponse{}, fmt.Errorf("count instructor courses: %w", err)
	}

	return courseListResponse(courses, total, page, pageSize), nil
}

func (s *courseService) Create(ctx context.Context, instructorID uuid.UUID, input domain.CreateCourseRequest) (domain.Course, error) {
	title := strings.TrimSpace(input.Title)
	if instructorID == uuid.Nil || title == "" {
		return domain.Course{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
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
		return domain.Course{}, fmt.Errorf("%w: at least one field is required", ErrInvalidInput)
	}
	course, err := s.ownedCourse(ctx, courseID, instructorID)
	if err != nil {
		return domain.Course{}, err
	}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return domain.Course{}, fmt.Errorf("%w: title cannot be empty", ErrInvalidInput)
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
		// 	return nil, ErrNotEnrolled
		// }
		if !enrolled {
			return nil, ErrNotEnrolled
		}
	} else if course.Instructor.ID != userID {
		return nil, ErrForbidden
	}
	return s.courseRepository.ListLessons(ctx, courseID)
}

func (s *courseService) CreateLesson(ctx context.Context, courseID, instructorID uuid.UUID, input domain.CreateLessonRequest) (domain.Lesson, error) {
	if _, err := s.ownedCourse(ctx, courseID, instructorID); err != nil {
		return domain.Lesson{}, err
	}
	title := strings.TrimSpace(input.Title)
	if title == "" || input.SortOrder < 0 {
		return domain.Lesson{}, fmt.Errorf("%w: valid title and sort_order are required", ErrInvalidInput)
	}
	return s.courseRepository.CreateLesson(ctx, courseID, title, strings.TrimSpace(input.Content), input.SortOrder)
}

func (s *courseService) UpdateLesson(ctx context.Context, lessonID, instructorID uuid.UUID, input domain.UpdateLessonRequest) (domain.Lesson, error) {
	if input.Title == nil && input.SortOrder == nil && input.Content == nil {
		return domain.Lesson{}, fmt.Errorf("%w: at least one field is required", ErrInvalidInput)
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
			return domain.Lesson{}, fmt.Errorf("%w: title cannot be empty", ErrInvalidInput)
		}
	}
	if input.SortOrder != nil {
		if *input.SortOrder < 0 {
			return domain.Lesson{}, fmt.Errorf("%w: sort_order cannot be negative", ErrInvalidInput)
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

	if err := s.courseRepository.Enroll(ctx, studentID, course.ID); err != nil {
		if errors.Is(err, repository.ErrEnrollmentExists) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *courseService) Unenroll(ctx context.Context, courseID, studentID uuid.UUID) error {
	if err := s.courseRepository.Unenroll(ctx, studentID, courseID); err != nil {
		if errors.Is(err, repository.ErrEnrollmentMissing) {
			return ErrNotEnrolled
		}
		return err
	}
	return nil
}

func (s *courseService) ownedCourse(ctx context.Context, courseID, instructorID uuid.UUID) (domain.Course, error) {
	course, err := s.courseRepository.GetByID(ctx, courseID)
	if err != nil {
		return domain.Course{}, mapCourseError(err)
	}
	if course.Instructor.ID != instructorID {
		return domain.Course{}, ErrForbidden
	}
	return course, nil
}

func mapCourseError(err error) error {
	switch {
	case errors.Is(err, repository.ErrCourseNotFound), errors.Is(err, repository.ErrLessonNotFound):
		return ErrNotFound
	default:
		return err
	}
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
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}
