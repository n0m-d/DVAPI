package query

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
)

var errNotFound = errors.New("not found")

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

type InstructorCoursesInput struct {
	InstructorID uuid.UUID
	Title        string
	Page         int
	PageSize     int
}

type CourseStudentsInput struct {
	CourseID     uuid.UUID
	InstructorID uuid.UUID
	Name         string
	Page         int
	PageSize     int
}

// Service is the read-model API for dashboards, analytics, and list views.
type Service interface {
	AdminStats(ctx context.Context) (domain.AdminStats, error)
	AdminList(ctx context.Context, search, role string, page, pageSize int) (domain.AdminUserListResponse, error)
	InstructorStats(ctx context.Context, instructorID uuid.UUID) (domain.InstructorStats, error)
	StudentStats(ctx context.Context, studentID uuid.UUID) (domain.StudentStats, error)
	StudentGrades(ctx context.Context, studentID uuid.UUID) (domain.StudentGradesDashboard, error)
	CourseAnalytics(ctx context.Context, courseID, instructorID uuid.UUID) (domain.CourseAnalytics, error)
	CourseProgress(ctx context.Context, courseID, studentID uuid.UUID) (domain.CourseProgress, error)
	Courses(ctx context.Context, input CourseListInput) (domain.CourseResponse, error)
	EnrolledCourses(ctx context.Context, input EnrolledCoursesInput) (domain.CourseResponse, error)
	EnrolledCoursesCount(ctx context.Context, studentID uuid.UUID) (domain.EnrolledCourseCount, error)
	InstructorCourses(ctx context.Context, input InstructorCoursesInput) (domain.CourseResponse, error)
	CourseStudents(ctx context.Context, input CourseStudentsInput) (domain.EnrolledStudentsResponse, error)
}

type queryService struct {
	store Store
}

func NewService(store Store) Service {
	return &queryService{store: store}
}

func (s *queryService) AdminStats(ctx context.Context) (domain.AdminStats, error) {
	return s.store.AdminStats(ctx)
}

func (s *queryService) AdminList(ctx context.Context, search, role string, page, pageSize int) (domain.AdminUserListResponse, error) {
	search = strings.TrimSpace(search)
	role = strings.TrimSpace(role)
	if role != "" && role != "student" && role != "instructor" && role != "admin" {
		return domain.AdminUserListResponse{}, fmt.Errorf("%w: invalid role", domain.ErrInvalidInput)
	}

	page, pageSize = normalizePagination(page, pageSize)
	users, err := s.store.ListUsers(ctx, search, role, pageSize, (page-1)*pageSize)
	if err != nil {
		return domain.AdminUserListResponse{}, fmt.Errorf("list users: %w", err)
	}
	total, err := s.store.CountUsers(ctx, search, role)
	if err != nil {
		return domain.AdminUserListResponse{}, fmt.Errorf("count users: %w", err)
	}
	pagination := domain.NewPagination(total, page, pageSize)
	return domain.AdminUserListResponse{
		Status: "success",
		Data: &domain.AdminUserListData{
			Users:      users,
			Pagination: &pagination,
		},
	}, nil
}

func (s *queryService) InstructorStats(ctx context.Context, instructorID uuid.UUID) (domain.InstructorStats, error) {
	if instructorID == uuid.Nil {
		return domain.InstructorStats{}, fmt.Errorf("%w: instructor id is required", domain.ErrInvalidInput)
	}
	return s.store.InstructorStats(ctx, instructorID)
}

func (s *queryService) StudentStats(ctx context.Context, studentID uuid.UUID) (domain.StudentStats, error) {
	if studentID == uuid.Nil {
		return domain.StudentStats{}, fmt.Errorf("%w: student id is required", domain.ErrInvalidInput)
	}
	return s.store.StudentStats(ctx, studentID)
}

func (s *queryService) StudentGrades(ctx context.Context, studentID uuid.UUID) (domain.StudentGradesDashboard, error) {
	return s.store.StudentGrades(ctx, studentID)
}

func (s *queryService) CourseAnalytics(ctx context.Context, courseID, instructorID uuid.UUID) (domain.CourseAnalytics, error) {
	if _, err := s.requireOwnedCourse(ctx, courseID, instructorID); err != nil {
		return domain.CourseAnalytics{}, err
	}
	return s.store.CourseAnalytics(ctx, courseID)
}

func (s *queryService) CourseProgress(ctx context.Context, courseID, studentID uuid.UUID) (domain.CourseProgress, error) {
	if err := s.requireEnrollment(ctx, studentID, courseID); err != nil {
		return domain.CourseProgress{}, err
	}
	return s.store.CourseProgress(ctx, studentID, courseID)
}

func (s *queryService) Courses(ctx context.Context, input CourseListInput) (domain.CourseResponse, error) {
	page, pageSize := normalizePagination(input.Page, input.PageSize)
	title := strings.TrimSpace(input.Title)
	courses, err := s.store.Courses(ctx, input.Published, title, pageSize, (page-1)*pageSize)
	if err != nil {
		return domain.CourseResponse{}, fmt.Errorf("list courses: %w", err)
	}
	total, err := s.store.CountCourses(ctx, input.Published, title)
	if err != nil {
		return domain.CourseResponse{}, fmt.Errorf("count courses: %w", err)
	}
	return courseListResponse(courses, total, page, pageSize), nil
}

func (s *queryService) EnrolledCourses(ctx context.Context, input EnrolledCoursesInput) (domain.CourseResponse, error) {
	page, pageSize := normalizePagination(input.Page, input.PageSize)
	title := strings.TrimSpace(input.Title)
	courses, err := s.store.EnrolledCourses(ctx, input.StudentID, title, pageSize, (page-1)*pageSize)
	if err != nil {
		return domain.CourseResponse{}, fmt.Errorf("get enrolled courses: %w", err)
	}
	total, err := s.store.CountEnrolledCourses(ctx, input.StudentID, title)
	if err != nil {
		return domain.CourseResponse{}, fmt.Errorf("count enrolled courses: %w", err)
	}
	return courseListResponse(courses, total, page, pageSize), nil
}

func (s *queryService) EnrolledCoursesCount(ctx context.Context, studentID uuid.UUID) (domain.EnrolledCourseCount, error) {
	count, err := s.store.CountEnrolledCourses(ctx, studentID, "")
	if err != nil {
		return domain.EnrolledCourseCount{}, fmt.Errorf("count enrolled courses: %w", err)
	}
	return domain.EnrolledCourseCount{Count: count}, nil
}

func (s *queryService) InstructorCourses(ctx context.Context, input InstructorCoursesInput) (domain.CourseResponse, error) {
	page, pageSize := normalizePagination(input.Page, input.PageSize)
	title := strings.TrimSpace(input.Title)
	courses, err := s.store.InstructorCourses(ctx, input.InstructorID, title, pageSize, (page-1)*pageSize)
	if err != nil {
		return domain.CourseResponse{}, fmt.Errorf("list instructor courses: %w", err)
	}
	total, err := s.store.CountInstructorCourses(ctx, input.InstructorID, title)
	if err != nil {
		return domain.CourseResponse{}, fmt.Errorf("count instructor courses: %w", err)
	}
	return courseListResponse(courses, total, page, pageSize), nil
}

func (s *queryService) CourseStudents(ctx context.Context, input CourseStudentsInput) (domain.EnrolledStudentsResponse, error) {
	if _, err := s.requireOwnedCourse(ctx, input.CourseID, input.InstructorID); err != nil {
		return domain.EnrolledStudentsResponse{}, err
	}

	page, pageSize := normalizePagination(input.Page, input.PageSize)
	name := strings.TrimSpace(input.Name)
	students, err := s.store.EnrolledStudents(ctx, input.CourseID, name, pageSize, (page-1)*pageSize)
	if err != nil {
		return domain.EnrolledStudentsResponse{}, fmt.Errorf("list enrolled students: %w", err)
	}
	total, err := s.store.CountEnrolledStudents(ctx, input.CourseID, name)
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

func (s *queryService) requireEnrollment(ctx context.Context, studentID, courseID uuid.UUID) error {
	if _, err := s.store.Course(ctx, courseID); err != nil {
		return mapQueryError(err)
	}
	enrolled, err := s.store.IsEnrolled(ctx, studentID, courseID)
	if err != nil {
		return err
	}
	if !enrolled {
		return domain.ErrNotEnrolled
	}
	return nil
}

func (s *queryService) requireOwnedCourse(ctx context.Context, courseID, instructorID uuid.UUID) (domain.Course, error) {
	course, err := s.store.Course(ctx, courseID)
	if err != nil {
		return domain.Course{}, mapQueryError(err)
	}
	if course.Instructor.ID != instructorID {
		return domain.Course{}, domain.ErrForbidden
	}
	return course, nil
}

func mapQueryError(err error) error {
	if errors.Is(err, errNotFound) {
		return domain.ErrNotFound
	}
	return err
}
