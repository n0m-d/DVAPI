package handler

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/middleware"
	"github.com/n0m-d/DVAPI/internal/service"
	"github.com/n0m-d/DVAPI/internal/utils"
)

type AssignmentHandler struct {
	assignments service.AssignmentService
	uploadDir   string
}

func NewAssignmentHandler(assignments service.AssignmentService) *AssignmentHandler {
	return &AssignmentHandler{
		assignments: assignments,
		uploadDir:   "uploads/assignments",
	}
}

// ListByCourse godoc
// @Summary      List course assignments
// @Description  List published assignments for a course the student is enrolled in.
// @Tags         Assignments
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path  string  true  "Course ID"
// @Success      200
// @Failure      400  "Invalid course id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not enrolled in this course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /courses/{courseId}/assignments [get]
func (h *AssignmentHandler) ListByCourse(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course id", "status": "error"})
		return
	}

	assignments, err := h.assignments.ListPublishedByCourse(c.Request.Context(), courseID, claims.UserID)
	if err != nil {
		respondAssignmentError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": assignments})
}

// GetByID godoc
// @Summary      Get assignment by ID
// @Description  Get assignment details
// @Tags         Assignments
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Assignment ID"
// @Success      200
// @Failure      400  "Invalid assignment id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Assignment not found"
// @Failure      500  "Internal server error"
// @Router       /assignments/{id} [get]
func (h *AssignmentHandler) GetByID(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}

	assignmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment id", "status": "error"})
		return
	}

	assignment, err := h.assignments.GetAccessibleByID(
		c.Request.Context(),
		assignmentID,
		claims.UserID,
		claims.Role,
	)
	if err != nil {
		respondAssignmentError(c, err)
		return
	}

	response := domain.AssignmentResponse{
		ID:          assignment.ID,
		CourseID:    assignment.CourseID,
		Title:       assignment.Title,
		Description: assignment.Description,
		DueDate:     assignment.DueDate,
		Status:      assignment.Status,
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": response})
}

// CreateSubmission godoc
// @Summary      Submit assignment
// @Description  Submit an assignment with text and an uploaded file.
// @Tags         Assignments
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        id               path      string  true  "Assignment ID"
// @Param        submission_text  formData  string  true  "Submission text"
// @Param        file             formData  file    true  "Uploaded file"
// @Success      201
// @Failure      400  "Invalid assignment id or validation error"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not enrolled in this course"
// @Failure      404  "Assignment not found"
// @Failure      409  "Already submitted"
// @Failure      500  "Internal server error"
// @Router       /assignments/{id}/submissions [post]
func (h *AssignmentHandler) CreateSubmission(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}

	assignmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment id", "status": "error"})
		return
	}

	var input domain.AssignmentSubmissionRequest
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}

	if err := h.assignments.ValidateSubmission(c.Request.Context(), assignmentID, claims.UserID, input.SubmissionText); err != nil {
		respondAssignmentError(c, err)
		return
	}

	fileName, filePath, err := h.saveSubmissionFile(c, assignmentID, claims.UserID, input.File.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file", "status": "error"})
		return
	}

	_, err = h.assignments.CreateSubmission(c.Request.Context(), domain.CreateAssignmentSubmissionInput{
		AssignmentID:   assignmentID,
		StudentID:      claims.UserID,
		SubmissionText: input.SubmissionText,
		FilePath:       filePath,
		FileName:       fileName,
	})
	if err != nil {
		_ = os.Remove(filePath)
		respondAssignmentError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Assignment submitted successfully",
	})
}

// CreateAssignment godoc
// @Summary      Create assignment
// @Description  Create an assignment for a course
// @Tags         Assignments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  domain.CreateAssignmentInput  true  "Assignment payload"
// @Success      201
// @Failure      400  "Validation error"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /assignments [post]
// @Router       /assignment/create [post]
func (h *AssignmentHandler) CreateAssignment(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}

	var input domain.CreateAssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.MapError(err), "status": "error"})
		return
	}

	_, err := h.assignments.CreateAssignment(c.Request.Context(), input, claims.UserID)
	if err != nil {
		respondAssignmentError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Assignment created successfully",
	})
}

// UpdateAssignment godoc
// @Summary      Update assignment
// @Description  Update an assignment
// @Tags         Assignments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                          true  "Assignment ID"
// @Param        body  body  domain.UpdateAssignmentRequest  true  "Assignment fields to update"
// @Success      200
// @Failure      400  "Invalid assignment id or validation error"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Assignment not found"
// @Failure      500  "Internal server error"
// @Router       /assignments/{id} [patch]
func (h *AssignmentHandler) UpdateAssignment(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	assignmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment id", "status": "error"})
		return
	}
	var input domain.UpdateAssignmentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}
	assignment, err := h.assignments.UpdateAssignment(c.Request.Context(), assignmentID, claims.UserID, input)
	if err != nil {
		respondAssignmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": assignment})
}

// DeleteAssignment godoc
// @Summary      Delete assignment
// @Description  Delete an assignment.
// @Tags         Assignments
// @Security     BearerAuth
// @Param        id  path  string  true  "Assignment ID"
// @Success      204
// @Failure      400  "Invalid assignment id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Assignment not found"
// @Failure      500  "Internal server error"
// @Router       /assignments/{id} [delete]
func (h *AssignmentHandler) DeleteAssignment(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	assignmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment id", "status": "error"})
		return
	}
	if err := h.assignments.DeleteAssignment(c.Request.Context(), assignmentID, claims.UserID); err != nil {
		respondAssignmentError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetMySubmission godoc
// @Summary      Get my submission
// @Description  Get the authenticated student's submission for an assignment.
// @Tags         Assignments
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Assignment ID"
// @Success      200
// @Failure      400  "Invalid assignment id"
// @Failure      401  "Unauthorized"
// @Failure      404  "Submission not found"
// @Failure      500  "Internal server error"
// @Router       /assignments/{id}/submissions/me [get]
func (h *AssignmentHandler) GetMySubmission(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}

	assignmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment id", "status": "error"})
		return
	}

	submission, err := h.assignments.GetMySubmission(c.Request.Context(), assignmentID, claims.UserID)
	if err != nil {
		respondAssignmentError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": submission})
}

// ListCourseAssignments godoc
// @Summary      List course assignments
// @Description  List assignments for a course
// @Tags         Assignments
// @Produce      json
// @Security     BearerAuth
// @Param        courseId   path   string  true   "Course ID"
// @Param        title      query  string  false  "Filter by title"
// @Param        page       query  int     false  "Page number"  default(1)
// @Param        page_size  query  int     false  "Page size"    default(10)
// @Success      200
// @Failure      400  "Invalid course id or pagination"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /my-courses/{courseId}/assignments [get]
func (h *AssignmentHandler) ListCourseAssignments(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course id", "status": "error"})
		return
	}

	page, pageSize, ok := parsePagination(c)
	if !ok {
		return
	}

	resp, err := h.assignments.ListCourseAssignments(c.Request.Context(), service.ListCourseAssignmentsInput{
		CourseID:     courseID,
		InstructorID: claims.UserID,
		Title:        c.Query("title"),
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		respondAssignmentError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListSubmissions godoc
// @Summary      List assignment submissions
// @Description  List submissions for an assignment with optional name filter and pagination.
// @Tags         Assignments
// @Produce      json
// @Security     BearerAuth
// @Param        id         path   string  true   "Assignment ID"
// @Param        name       query  string  false  "Filter by student name"
// @Param        page       query  int     false  "Page number"  default(1)
// @Param        page_size  query  int     false  "Page size"    default(10)
// @Success      200
// @Failure      400  "Invalid assignment id or pagination"
// @Failure      401  "Unauthorized"
// @Failure      404  "Assignment not found"
// @Failure      500  "Internal server error"
// @Router       /assignments/{id}/submissions [get]
func (h *AssignmentHandler) ListSubmissions(c *gin.Context) {

	assignmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment id", "status": "error"})
		return
	}

	page, pageSize, ok := parsePagination(c)
	if !ok {
		return
	}

	resp, err := h.assignments.ListSubmissions(c.Request.Context(), service.ListSubmissionsInput{
		AssignmentID: assignmentID,
		Name:         c.Query("name"),
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		respondAssignmentError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GradeSubmission godoc
// @Summary      Grade submission
// @Description  Grade a student submission for an assignment.
// @Tags         Assignments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                         true  "Submission ID"
// @Param        body  body  domain.GradeSubmissionRequest  true  "Grade and feedback"
// @Success      200
// @Failure      400  "Invalid submission id or missing grade"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Submission not found"
// @Failure      500  "Internal server error"
// @Router       /submissions/{id}/grade [patch]
func (h *AssignmentHandler) GradeSubmission(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}

	submissionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission id", "status": "error"})
		return
	}

	var input domain.GradeSubmissionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "grade is required", "status": "error"})
		return
	}

	submission, err := h.assignments.GradeSubmission(c.Request.Context(), domain.GradeSubmissionInput{
		SubmissionID: submissionID,
		InstructorID: claims.UserID,
		Grade:        *input.Grade,
		Feedback:     input.Feedback,
	})
	if err != nil {
		respondAssignmentError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Submission graded successfully",
		"data":    submission,
	})
}

// GetSubmissionForInstructor godoc
// @Summary      Get submission
// @Description  Get a submission for an assignment in a course you own.
// @Tags         Assignments
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Submission ID"
// @Success      200
// @Failure      400  "Invalid submission id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Submission not found"
// @Failure      500  "Internal server error"
// @Router       /submissions/{id} [get]
func (h *AssignmentHandler) GetSubmissionForInstructor(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	submissionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission id", "status": "error"})
		return
	}
	submission, err := h.assignments.GetSubmissionForInstructor(c.Request.Context(), submissionID, claims.UserID)
	if err != nil {
		respondAssignmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": submission})
}

func (h *AssignmentHandler) saveSubmissionFile(c *gin.Context, assignmentID, studentID uuid.UUID, originalName string) (string, string, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return "", "", err
	}

	dir := filepath.Join(h.uploadDir, assignmentID.String(), studentID.String())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}

	fileName := uuid.New().String() + strings.ToLower(filepath.Ext(filepath.Base(originalName)))
	dst := filepath.Join(dir, fileName)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", "", err
	}
	return fileName, dst, nil
}

func currentClaims(c *gin.Context) (*domain.Claims, bool) {
	raw, ok := c.Get(middleware.ClaimsKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "status": "error"})
		return nil, false
	}
	claims, ok := raw.(*domain.Claims)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "status": "error"})
		return nil, false
	}
	return claims, true
}

func (h *AssignmentHandler) GetFile(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required", "status": "error"})
		return
	}

	// Resolve the path to the file; does not prevent directory traversal
	clean := filepath.Clean(filePath)

	/*Deliberately commented out for directory traversal
	uploadRoot, err := filepath.Abs(h.uploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "status": "error"})
		return
	}
	absPath, err := filepath.Abs(clean)
	if err != nil || !strings.HasPrefix(absPath, uploadRoot+string(os.PathSeparator)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path", "status": "error"})
		return
	}
	*/

	c.File(clean)

}

func respondAssignmentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Course/assignment/submission not found", "status": "error"})
	case errors.Is(err, service.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only access assignments for your own courses", "status": "error"})
	case errors.Is(err, service.ErrNotEnrolled):
		c.JSON(http.StatusForbidden, gin.H{"error": "Not enrolled in this course", "status": "error"})
	case errors.Is(err, service.ErrAssignmentClosed):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Assignment is not open for submissions", "status": "error"})
	case errors.Is(err, service.ErrPastDue):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Assignment due date has passed", "status": "error"})
	case errors.Is(err, service.ErrAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "You have already submitted this assignment", "status": "error"})
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "status": "error"})
	}
}
