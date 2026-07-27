package domain

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

const (
	AssignmentStatusDraft     = "draft"
	AssignmentStatusPublished = "published"
	AssignmentStatusClosed    = "closed"
)

type Assignment struct {
	ID          uuid.UUID `json:"id"`
	CourseID    uuid.UUID `json:"course_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Status      string    `json:"status"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type AssignmentResponse struct {
	ID          uuid.UUID `json:"id"`
	CourseID    uuid.UUID `json:"course_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Status      string    `json:"status"`
}

type AssignmentSubmissionRequest struct {
	SubmissionText string                `form:"submission_text" binding:"required"`
	File           *multipart.FileHeader `form:"file" binding:"required"`
}

type CreateAssignmentSubmissionInput struct {
	AssignmentID   uuid.UUID
	StudentID      uuid.UUID
	SubmissionText string
	FilePath       string
	FileName       string
}

type GradeSubmissionRequest struct {
	Grade    *int   `json:"grade" binding:"required"`
	Feedback string `json:"feedback"`
}

type GradeSubmissionInput struct {
	SubmissionID uuid.UUID
	InstructorID uuid.UUID
	Grade        int
	Feedback     string
}

type CreateAssignmentInput struct {
	CourseID    uuid.UUID `json:"course_id" binding:"required"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description" binding:"required"`
	DueDate     time.Time `json:"due_date" binding:"required"`
	Status      string    `json:"status" binding:"required"`
}

type UpdateAssignmentRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	Status      *string    `json:"status" binding:"omitempty,oneof=draft published closed"`
}

type AssignmentSubmission struct {
	ID             uuid.UUID `json:"id"`
	AssignmentID   uuid.UUID `json:"assignment_id"`
	StudentID      uuid.UUID `json:"student_id"`
	SubmissionText string    `json:"submission_text"`
	FilePath       string    `json:"file_path"`
	FileName       string    `json:"file_name"`
	SubmittedAt    time.Time `json:"submitted_at"`
	Grade          int       `json:"grade"`
	Feedback       string    `json:"feedback"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type InstructorSubmission struct {
	ID             uuid.UUID `json:"id"`
	AssignmentID   uuid.UUID `json:"assignment_id"`
	StudentID      uuid.UUID `json:"student_id"`
	StudentName    string    `json:"student_name"`
	StudentEmail   string    `json:"student_email"`
	SubmissionText string    `json:"submission_text"`
	FilePath       string    `json:"file_path"`
	FileName       string    `json:"file_name"`
	SubmittedAt    time.Time `json:"submitted_at"`
	Grade          int       `json:"grade"`
	Feedback       string    `json:"feedback"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AssignmentListData struct {
	Assignments []Assignment `json:"assignments"`
	Pagination  *Pagination  `json:"pagination"`
}

type AssignmentListResponse struct {
	Status string              `json:"status"`
	Data   *AssignmentListData `json:"data"`
}

type SubmissionListData struct {
	Submissions []InstructorSubmission `json:"submissions"`
	Pagination  *Pagination            `json:"pagination"`
}

type SubmissionListResponse struct {
	Status string              `json:"status"`
	Data   *SubmissionListData `json:"data"`
}
