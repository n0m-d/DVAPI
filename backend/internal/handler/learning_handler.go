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
	"github.com/n0m-d/DVAPI/internal/service"
)

type LearningHandler struct {
	learning  service.LearningService
	uploadDir string
}

func NewLearningHandler(learning service.LearningService) *LearningHandler {
	return &LearningHandler{learning: learning, uploadDir: "uploads/assignments"}
}

// SetLessonProgress godoc
// @Summary      Update lesson progress
// @Description  Mark a lesson as completed or incomplete for the authenticated student.
// @Tags         Learning
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                       true  "Lesson ID"
// @Param        body  body  domain.LessonProgressRequest  true  "Progress update"
// @Success      200
// @Failure      400  "Invalid lesson id or validation error"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not enrolled in this course"
// @Failure      404  "Lesson not found"
// @Failure      500  "Internal server error"
// @Router       /lessons/{id}/progress [put]
func (h *LearningHandler) SetLessonProgress(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson id", "status": "error"})
		return
	}
	var input domain.LessonProgressRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}
	if err := h.learning.SetLessonProgress(c.Request.Context(), lessonID, claims.UserID, input.Completed); err != nil {
		respondLearningError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Lesson progress updated"})
}

// GetCourseProgress godoc
// @Summary      Get course progress
// @Description  Get lesson completion progress for a course the student is enrolled in.
// @Tags         Learning
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path  string  true  "Course ID"
// @Success      200
// @Failure      400  "Invalid course id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not enrolled in this course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /courses/{courseId}/progress [get]
func (h *LearningHandler) GetCourseProgress(c *gin.Context) {
	claims, courseID, ok := claimsAndCourseID(c)
	if !ok {
		return
	}
	progress, err := h.learning.GetCourseProgress(c.Request.Context(), courseID, claims.UserID)
	if err != nil {
		respondLearningError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": progress})
}

// GetNextLesson godoc
// @Summary      Continue course
// @Description  Get the next lesson to continue for a course the student is enrolled in.
// @Tags         Learning
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path  string  true  "Course ID"
// @Success      200
// @Failure      400  "Invalid course id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not enrolled in this course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /courses/{courseId}/continue [get]
func (h *LearningHandler) GetNextLesson(c *gin.Context) {
	claims, courseID, ok := claimsAndCourseID(c)
	if !ok {
		return
	}
	lesson, err := h.learning.GetNextLesson(c.Request.Context(), courseID, claims.UserID)
	if err != nil {
		respondLearningError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": lesson})
}

// GetGrades godoc
// @Summary      Get grades
// @Description  List graded assignment submissions for the authenticated student.
// @Tags         Learning
// @Produce      json
// @Security     BearerAuth
// @Success      200
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal server error"
// @Router       /grades [get]
func (h *LearningHandler) GetGrades(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	grades, err := h.learning.GetGrades(c.Request.Context(), claims.UserID)
	if err != nil {
		respondLearningError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": grades})
}

// Resubmit godoc
// @Summary      Resubmit assignment
// @Description  Submit a new version of an existing assignment submission with text and file.
// @Tags         Assignments
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        id               path      string  true  "Submission ID"
// @Param        submission_text  formData  string  true  "Submission text"
// @Param        file             formData  file    true  "Uploaded file"
// @Success      200
// @Failure      400  "Invalid submission id or validation error"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Submission not found"
// @Failure      500  "Internal server error"
// @Router       /submissions/{id} [put]
func (h *LearningHandler) Resubmit(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	submissionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission id", "status": "error"})
		return
	}
	var input domain.AssignmentSubmissionRequest
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}
	fileName, filePath, err := h.saveResubmissionFile(c, submissionID, input.File.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file", "status": "error"})
		return
	}
	submission, err := h.learning.Resubmit(c.Request.Context(), service.ResubmitInput{
		SubmissionID: submissionID, StudentID: claims.UserID,
		SubmissionText: input.SubmissionText, FilePath: filePath, FileName: fileName,
	})
	if err != nil {
		_ = os.Remove(filePath)
		respondLearningError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": submission})
}

// ListSubmissionVersions godoc
// @Summary      List submission versions
// @Description  List all versions of an assignment submission.
// @Tags         Assignments
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Submission ID"
// @Success      200
// @Failure      400  "Invalid submission id"
// @Failure      401  "Unauthorized"
// @Failure      404  "Submission not found"
// @Failure      500  "Internal server error"
// @Router       /submissions/{id}/versions [get]
func (h *LearningHandler) ListSubmissionVersions(c *gin.Context) {
	submissionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission id", "status": "error"})
		return
	}
	versions, err := h.learning.ListSubmissionVersions(c.Request.Context(), submissionID)
	if err != nil {
		respondLearningError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": versions})
}

// CreateAnnouncement godoc
// @Summary      Create announcement
// @Description  Create a course announcement.
// @Tags         Announcements
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path  string                           true  "Course ID"
// @Param        body      body  domain.CreateAnnouncementRequest  true  "Announcement payload"
// @Success      201
// @Failure      400  "Invalid course id or validation error"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /courses/{courseId}/announcements [post]
func (h *LearningHandler) CreateAnnouncement(c *gin.Context) {
	claims, courseID, ok := claimsAndCourseID(c)
	if !ok {
		return
	}
	var input domain.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}
	announcement, err := h.learning.CreateAnnouncement(c.Request.Context(), courseID, claims.UserID, input)
	if err != nil {
		respondLearningError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": announcement})
}

// ListStudentAnnouncements godoc
// @Summary      List course announcements
// @Description  List published announcements for a course the student is enrolled in.
// @Tags         Announcements
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path  string  true  "Course ID"
// @Success      200
// @Failure      400  "Invalid course id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not enrolled in this course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /courses/{courseId}/announcements [get]
func (h *LearningHandler) ListStudentAnnouncements(c *gin.Context) {
	h.listAnnouncements(c, false)
}

// ListInstructorAnnouncements godoc
// @Summary      List course announcements
// @Description  List all announcements for a course you own.
// @Tags         Announcements
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path  string  true  "Course ID"
// @Success      200
// @Failure      400  "Invalid course id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /my-courses/{courseId}/announcements [get]
func (h *LearningHandler) ListInstructorAnnouncements(c *gin.Context) {
	h.listAnnouncements(c, true)
}

func (h *LearningHandler) listAnnouncements(c *gin.Context, instructor bool) {
	claims, courseID, ok := claimsAndCourseID(c)
	if !ok {
		return
	}
	announcements, err := h.learning.ListAnnouncements(c.Request.Context(), courseID, claims.UserID, instructor)
	if err != nil {
		respondLearningError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": announcements})
}

// UpdateAnnouncement godoc
// @Summary      Update announcement
// @Description  Update an announcement in a course you own.
// @Tags         Announcements
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                            true  "Announcement ID"
// @Param        body  body  domain.UpdateAnnouncementRequest  true  "Announcement fields to update"
// @Success      200
// @Failure      400  "Invalid announcement id or validation error"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Announcement not found"
// @Failure      500  "Internal server error"
// @Router       /announcements/{id} [patch]
func (h *LearningHandler) UpdateAnnouncement(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid announcement id", "status": "error"})
		return
	}
	var input domain.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}
	announcement, err := h.learning.UpdateAnnouncement(c.Request.Context(), id, claims.UserID, input)
	if err != nil {
		respondLearningError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": announcement})
}

// DeleteAnnouncement godoc
// @Summary      Delete announcement
// @Description  Delete a course announcement.
// @Tags         Announcements
// @Security     BearerAuth
// @Param        id  path  string  true  "Announcement ID"
// @Success      204
// @Failure      400  "Invalid announcement id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Announcement not found"
// @Failure      500  "Internal server error"
// @Router       /announcements/{id} [delete]
func (h *LearningHandler) DeleteAnnouncement(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid announcement id", "status": "error"})
		return
	}
	if err := h.learning.DeleteAnnouncement(c.Request.Context(), id, claims.UserID); err != nil {
		respondLearningError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetCourseAnalytics godoc
// @Summary      Course analytics
// @Description  Get enrollment, assignment, and progress analytics for a course you own.
// @Tags         Learning
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path  string  true  "Course ID"
// @Success      200
// @Failure      400  "Invalid course id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /courses/{courseId}/analytics [get]
func (h *LearningHandler) GetCourseAnalytics(c *gin.Context) {
	claims, courseID, ok := claimsAndCourseID(c)
	if !ok {
		return
	}
	analytics, err := h.learning.GetCourseAnalytics(c.Request.Context(), courseID, claims.UserID)
	if err != nil {
		respondLearningError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": analytics})
}

// GetStats godoc
// @Summary      Dashboard stats
// @Description  Aggregate stats for the authenticated user. Instructors get course/teaching stats; students get enrollment/progress stats.
// @Tags         Learning
// @Produce      json
// @Security     BearerAuth
// @Success      200
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      500  "Internal server error"
// @Router       /stats [get]
func (h *LearningHandler) GetStats(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}

	switch claims.Role {
	case "instructor":
		stats, err := h.learning.GetInstructorStats(c.Request.Context(), claims.UserID)
		if err != nil {
			respondLearningError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "role": "instructor", "data": stats})
	case "student":
		stats, err := h.learning.GetStudentStats(c.Request.Context(), claims.UserID)
		if err != nil {
			respondLearningError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "role": "student", "data": stats})
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Stats are only available for students and instructors", "status": "error"})
	}
}

func (h *LearningHandler) saveResubmissionFile(c *gin.Context, submissionID uuid.UUID, originalName string) (string, string, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return "", "", err
	}
	dir := filepath.Join(h.uploadDir, "resubmissions", submissionID.String())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	fileName := uuid.NewString() + strings.ToLower(filepath.Ext(filepath.Base(originalName)))
	path := filepath.Join(dir, fileName)
	if err := c.SaveUploadedFile(file, path); err != nil {
		return "", "", err
	}
	return fileName, path, nil
}

func claimsAndCourseID(c *gin.Context) (*domain.Claims, uuid.UUID, bool) {
	claims, ok := currentClaims(c)
	if !ok {
		return nil, uuid.Nil, false
	}
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course id", "status": "error"})
		return nil, uuid.Nil, false
	}
	return claims, courseID, true
}

func respondLearningError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found", "status": "error"})
	case errors.Is(err, service.ErrForbidden), errors.Is(err, service.ErrNotEnrolled):
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden", "status": "error"})
	case errors.Is(err, service.ErrAssignmentClosed), errors.Is(err, service.ErrPastDue):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
	case errors.Is(err, service.ErrAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Already exists", "status": "error"})
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "status": "error"})
	}
}
