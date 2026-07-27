package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	AnnouncementStatusDraft     = "draft"
	AnnouncementStatusPublished = "published"
)

type LessonProgressRequest struct {
	Completed bool `json:"completed"`
}

type CourseProgress struct {
	CourseID         uuid.UUID `json:"course_id"`
	TotalLessons     int       `json:"total_lessons"`
	CompletedLessons int       `json:"completed_lessons"`
	Percentage       float64   `json:"percentage"`
}

type StudentGrade struct {
	SubmissionID    uuid.UUID `json:"submission_id"`
	AssignmentID    uuid.UUID `json:"assignment_id"`
	AssignmentTitle string    `json:"assignment_title"`
	CourseID        uuid.UUID `json:"course_id"`
	CourseTitle     string    `json:"course_title"`
	Grade           *int      `json:"grade"`
	Feedback        string    `json:"feedback"`
	SubmittedAt     time.Time `json:"submitted_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type StudentGradesDashboard struct {
	Submitted    int            `json:"submitted"`
	AverageGrade float64        `json:"average_grade"`
	Grades       []StudentGrade `json:"grades"`
}

type SubmissionVersion struct {
	ID             uuid.UUID `json:"id"`
	SubmissionID   uuid.UUID `json:"submission_id"`
	Version        int       `json:"version"`
	SubmissionText string    `json:"submission_text"`
	FilePath       string    `json:"file_path"`
	FileName       string    `json:"file_name"`
	SubmittedAt    time.Time `json:"submitted_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type Announcement struct {
	ID        uuid.UUID `json:"id"`
	CourseID  uuid.UUID `json:"course_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedBy uuid.UUID `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateAnnouncementRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Status  string `json:"status" binding:"required,oneof=draft published"`
}

type UpdateAnnouncementRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
	Status  *string `json:"status" binding:"omitempty,oneof=draft published"`
}

type CourseAnalytics struct {
	CourseID          uuid.UUID `json:"course_id"`
	Enrollments       int       `json:"enrollments"`
	Assignments       int       `json:"assignments"`
	Submissions       int       `json:"submissions"`
	AverageGrade      float64   `json:"average_grade"`
	Lessons           int       `json:"lessons"`
	LessonCompletions int       `json:"lesson_completions"`
}

type InstructorStats struct {
	Courses             int `json:"courses"`
	PublishedCourses    int `json:"published_courses"`
	Enrollments         int `json:"enrollments"`
	Students            int `json:"students"`
	Lessons             int `json:"lessons"`
	Assignments         int `json:"assignments"`
	Submissions         int `json:"submissions"`
	UngradedSubmissions int `json:"ungraded_submissions"`
	Announcements       int `json:"announcements"`
}

type StudentStats struct {
	EnrolledCourses    int     `json:"enrolled_courses"`
	CompletedLessons   int     `json:"completed_lessons"`
	Submissions        int     `json:"submissions"`
	GradedSubmissions  int     `json:"graded_submissions"`
	AverageGrade       float64 `json:"average_grade"`
	PendingAssignments int     `json:"pending_assignments"`
}
