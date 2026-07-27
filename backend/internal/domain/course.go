package domain

import (
	"time"

	"github.com/google/uuid"
)

type Instructor struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Email    string    `json:"email"`
}

// InstructorPublic is used in public course detail responses (no email).
type InstructorPublic struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
}

type Course struct {
	ID          uuid.UUID  `json:"id"`
	Instructor  Instructor `json:"instructor"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	Published   bool       `json:"published"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateCourseRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Published   bool   `json:"published"`
}

type UpdateCourseRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Published   *bool   `json:"published"`
}

type Lesson struct {
	ID        uuid.UUID `json:"id"`
	CourseID  uuid.UUID `json:"course_id"`
	Title     string    `json:"title"`
	SortOrder int       `json:"sort_order"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateLessonRequest struct {
	Title     string `json:"title" binding:"required"`
	SortOrder int    `json:"sort_order"`
	Content   string `json:"content"`
}

type UpdateLessonRequest struct {
	Title     *string `json:"title"`
	SortOrder *int    `json:"sort_order"`
	Content   *string `json:"content"`
}

// public DTO for GET /courses/:id.
type CourseDetail struct {
	ID          uuid.UUID        `json:"id"`
	Instructor  InstructorPublic `json:"instructor"`
	Title       string           `json:"title"`
	Slug        string           `json:"slug"`
	Description string           `json:"description"`
	Published   bool             `json:"published"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

func (c Course) ToDetail() CourseDetail {
	return CourseDetail{
		ID: c.ID,
		Instructor: InstructorPublic{
			ID:       c.Instructor.ID,
			FullName: c.Instructor.FullName,
		},
		Title:       c.Title,
		Slug:        c.Slug,
		Description: c.Description,
		Published:   c.Published,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

type CourseData struct {
	Courses    []Course    `json:"courses"`
	Pagination *Pagination `json:"pagination"`
}

type CourseResponse struct {
	Status string      `json:"status"`
	Data   *CourseData `json:"data"`
}

type EnrolledCourseCount struct {
	Count int `json:"count"`
}

type EnrolledStudent struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	FullName   string    `json:"full_name"`
	EnrolledAt time.Time `json:"enrolled_at"`
}

type EnrolledStudentsData struct {
	Students   []EnrolledStudent `json:"students"`
	Pagination *Pagination       `json:"pagination"`
}

type EnrolledStudentsResponse struct {
	Status string                `json:"status"`
	Data   *EnrolledStudentsData `json:"data"`
}

type EnrollmentRequest struct {
	UserID   uuid.UUID `json:"user_id" binding:"required"`
	CourseID uuid.UUID `json:"course_id" binding:"required"`
}
