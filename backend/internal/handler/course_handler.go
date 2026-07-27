package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/middleware"
	"github.com/n0m-d/DVAPI/internal/service"
)

type CourseHandler struct {
	courseService service.CourseService
}

func NewCourseHandler(courseService service.CourseService) *CourseHandler {
	return &CourseHandler{courseService: courseService}
}

// GetCourses godoc
// @Summary      List courses
// @Description  List published courses with optional title filter and pagination.
// @Tags         Courses
// @Produce      json
// @Security     BearerAuth
// @Param        published  query  string  false  "Include published courses (default true when omitted)"
// @Param        title      query  string  false  "Filter by title"
// @Param        page       query  int     false  "Page number"  default(1)
// @Param        page_size  query  int     false  "Page size"    default(10)
// @Success      200
// @Failure      400  "Invalid pagination"
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal server error"
// @Router       /courses [get]
func (h *CourseHandler) GetCourses(c *gin.Context) {
	published := c.Query("published") == "true" || c.Query("published") == ""
	page, pageSize, ok := parsePagination(c)
	if !ok {
		return
	}

	resp, err := h.courseService.GetCourses(c.Request.Context(), service.CourseListInput{
		Published: published,
		Title:     c.Query("title"),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "status": "error"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetByID godoc
// @Summary      Get course by ID
// @Description  Get course details by course ID.
// @Tags         Courses
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path  string  true  "Course ID"
// @Success      200
// @Failure      401  "Unauthorized"
// @Failure      404  "Course not found"
// @Router       /courses/{courseId} [get]
func (h *CourseHandler) GetByID(c *gin.Context) {
	id := c.Param("courseId")
	course, err := h.courseService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found", "status": "error"})
		return
	}
	c.JSON(http.StatusOK, course.ToDetail())
}

// GetEnrolledCourses godoc
// @Summary      List enrolled courses
// @Description  List courses the authenticated student is enrolled in.
// @Tags         Courses
// @Produce      json
// @Security     BearerAuth
// @Param        title      query  string  false  "Filter by title"
// @Param        page       query  int     false  "Page number"  default(1)
// @Param        page_size  query  int     false  "Page size"    default(10)
// @Success      200
// @Failure      400  "Invalid pagination"
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal server error"
// @Router       /enrolled-courses [get]
func (h *CourseHandler) GetEnrolledCourses(c *gin.Context) {
	raw, ok := c.Get(middleware.ClaimsKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "status": "error"})
		return
	}

	claims, ok := raw.(*domain.Claims)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "status": "error"})
		return
	}

	page, pageSize, ok := parsePagination(c)
	if !ok {
		return
	}

	resp, err := h.courseService.GetEnrolledCourses(c.Request.Context(), service.EnrolledCoursesInput{
		StudentID: claims.UserID,
		Title:     c.Query("title"),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "status": "error"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetEnrolledCoursesCount godoc
// @Summary      Enrolled courses count
// @Description  Return the number of courses the authenticated student is enrolled in.
// @Tags         Courses
// @Produce      json
// @Security     BearerAuth
// @Success      200
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal server error"
// @Router       /enrolled-courses-count [get]
func (h *CourseHandler) GetEnrolledCoursesCount(c *gin.Context) {
	raw, ok := c.Get(middleware.ClaimsKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "status": "error"})
		return
	}

	claims, ok := raw.(*domain.Claims)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "status": "error"})
		return
	}

	resp, err := h.courseService.GetEnrolledCoursesCount(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":  resp.Count,
		"status": "success",
	})
}

// ListStudents godoc
// @Summary      List course students
// @Description  List students enrolled in a course
// @Tags         Courses
// @Produce      json
// @Security     BearerAuth
// @Param        courseId   path   string  true   "Course ID"
// @Param        name       query  string  false  "Filter by student name"
// @Param        page       query  int     false  "Page number"  default(1)
// @Param        page_size  query  int     false  "Page size"    default(10)
// @Success      200
// @Failure      400  "Invalid course id or pagination"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /courses/{courseId}/students [get]
func (h *CourseHandler) ListStudents(c *gin.Context) {
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

	resp, err := h.courseService.ListStudents(c.Request.Context(), service.ListCourseStudentsInput{
		CourseID:     courseID,
		InstructorID: claims.UserID,
		Name:         c.Query("name"),
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Course not found", "status": "error"})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only view students for your own courses", "status": "error"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "status": "error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListMyCourses godoc
// @Summary      List my courses
// @Description  List courses owned
// @Tags         Courses
// @Produce      json
// @Security     BearerAuth
// @Param        title      query  string  false  "Filter by title"
// @Param        page       query  int     false  "Page number"  default(1)
// @Param        page_size  query  int     false  "Page size"    default(10)
// @Success      200
// @Failure      400  "Invalid pagination"
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal server error"
// @Router       /my-courses [get]
func (h *CourseHandler) ListMyCourses(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}

	page, pageSize, ok := parsePagination(c)
	if !ok {
		return
	}

	resp, err := h.courseService.ListMyCourses(c.Request.Context(), service.InstructorCoursesInput{
		InstructorID: claims.UserID,
		Title:        c.Query("title"),
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "status": "error"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Create godoc
// @Summary      Create course
// @Description  Create a new course as the authenticated instructor.
// @Tags         Courses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  domain.CreateCourseRequest  true  "Course payload"
// @Success      201
// @Failure      400  "Validation error"
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal server error"
// @Router       /courses [post]
func (h *CourseHandler) Create(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	var input domain.CreateCourseRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}
	course, err := h.courseService.Create(c.Request.Context(), claims.UserID, input)
	if err != nil {
		respondCourseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": course})
}

// Update godoc
// @Summary      Update course
// @Description  Update a course owned
// @Tags         Courses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path  string                      true  "Course ID"
// @Param        body      body  domain.UpdateCourseRequest  true  "Course fields to update"
// @Success      200
// @Failure      400  "Invalid course id or validation error"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /courses/{courseId} [patch]
func (h *CourseHandler) Update(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course id", "status": "error"})
		return
	}
	var input domain.UpdateCourseRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}
	course, err := h.courseService.Update(c.Request.Context(), courseID, claims.UserID, input)
	if err != nil {
		respondCourseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": course})
}

// Delete godoc
// @Summary      Delete course
// @Description  Delete a course owned
// @Tags         Courses
// @Security     BearerAuth
// @Param        courseId  path  string  true  "Course ID"
// @Success      204
// @Failure      400  "Invalid course id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /courses/{courseId} [delete]
func (h *CourseHandler) Delete(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course id", "status": "error"})
		return
	}
	if err := h.courseService.Delete(c.Request.Context(), courseID, claims.UserID); err != nil {
		respondCourseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListStudentLessons godoc
// @Summary      List course lessons
// @Description  List lessons for a course the authenticated student is enrolled in.
// @Tags         Lessons
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path  string  true  "Course ID"
// @Success      200
// @Failure      400  "Invalid course id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not enrolled in this course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /courses/{courseId}/lessons [get]
func (h *CourseHandler) ListStudentLessons(c *gin.Context) {
	h.listLessons(c, true)
}

// ListInstructorLessons godoc
// @Summary      List course lessons
// @Description  List all lessons for a course
// @Tags         Lessons
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path  string  true  "Course ID"
// @Success      200
// @Failure      400  "Invalid course id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /my-courses/{courseId}/lessons [get]
func (h *CourseHandler) ListInstructorLessons(c *gin.Context) {
	h.listLessons(c, false)
}

func (h *CourseHandler) listLessons(c *gin.Context, requireEnrollment bool) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course id", "status": "error"})
		return
	}
	lessons, err := h.courseService.ListLessons(c.Request.Context(), courseID, claims.UserID, requireEnrollment)
	if err != nil {
		respondCourseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": lessons})
}

// CreateLesson godoc
// @Summary      Create lesson
// @Description  Create a lesson in a course
// @Tags         Lessons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path  string                      true  "Course ID"
// @Param        body      body  domain.CreateLessonRequest  true  "Lesson payload"
// @Success      201
// @Failure      400  "Invalid course id or validation error"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /courses/{courseId}/lessons [post]
func (h *CourseHandler) CreateLesson(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course id", "status": "error"})
		return
	}
	var input domain.CreateLessonRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}
	lesson, err := h.courseService.CreateLesson(c.Request.Context(), courseID, claims.UserID, input)
	if err != nil {
		respondCourseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": lesson})
}

// UpdateLesson godoc
// @Summary      Update lesson
// @Description  Update a lesson in a course
// @Tags         Lessons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                      true  "Lesson ID"
// @Param        body  body  domain.UpdateLessonRequest  true  "Lesson fields to update"
// @Success      200
// @Failure      400  "Invalid lesson id or validation error"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Lesson not found"
// @Failure      500  "Internal server error"
// @Router       /lessons/{id} [patch]
func (h *CourseHandler) UpdateLesson(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson id", "status": "error"})
		return
	}
	var input domain.UpdateLessonRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}
	lesson, err := h.courseService.UpdateLesson(c.Request.Context(), lessonID, claims.UserID, input)
	if err != nil {
		respondCourseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": lesson})
}

// DeleteLesson godoc
// @Summary      Delete lesson
// @Description  Delete a lesson from a course
// @Tags         Lessons
// @Security     BearerAuth
// @Param        id  path  string  true  "Lesson ID"
// @Success      204
// @Failure      400  "Invalid lesson id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not your course"
// @Failure      404  "Lesson not found"
// @Failure      500  "Internal server error"
// @Router       /lessons/{id} [delete]
func (h *CourseHandler) DeleteLesson(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson id", "status": "error"})
		return
	}
	if err := h.courseService.DeleteLesson(c.Request.Context(), lessonID, claims.UserID); err != nil {
		respondCourseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Enroll godoc
// @Summary      Enroll in course
// @Description  Enroll the authenticated student in a course.
// @Tags         Courses
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path  string  true  "Course ID"
// @Success      201
// @Failure      400  "Invalid course id"
// @Failure      401  "Unauthorized"
// @Failure      404  "Course not found"
// @Failure      409  "Already enrolled"
// @Failure      500  "Internal server error"
// @Router       /courses/{courseId}/enrollments/me [post]
func (h *CourseHandler) Enroll(c *gin.Context) {
	h.changeEnrollment(c, true)
}

// Unenroll godoc
// @Summary      Unenroll from course
// @Description  Remove the authenticated student's enrollment from a course.
// @Tags         Courses
// @Security     BearerAuth
// @Param        courseId  path  string  true  "Course ID"
// @Success      204
// @Failure      400  "Invalid course id"
// @Failure      401  "Unauthorized"
// @Failure      403  "Not enrolled in this course"
// @Failure      404  "Course not found"
// @Failure      500  "Internal server error"
// @Router       /courses/{courseId}/enrollments/me [delete]
func (h *CourseHandler) Unenroll(c *gin.Context) {
	h.changeEnrollment(c, false)
}

func (h *CourseHandler) changeEnrollment(c *gin.Context, enroll bool) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course id", "status": "error"})
		return
	}
	if enroll {
		err = h.courseService.Enroll(c.Request.Context(), courseID, claims.UserID)
	} else {
		err = h.courseService.Unenroll(c.Request.Context(), courseID, claims.UserID)
	}
	if err != nil {
		respondCourseError(c, err)
		return
	}
	if enroll {
		c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "Enrolled successfully"})
		return
	}
	c.Status(http.StatusNoContent)
}

func respondCourseError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Course or lesson not found", "status": "error"})
	case errors.Is(err, service.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only manage your own courses", "status": "error"})
	case errors.Is(err, service.ErrNotEnrolled):
		c.JSON(http.StatusForbidden, gin.H{"error": "Not enrolled in this course", "status": "error"})
	case errors.Is(err, service.ErrAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Already enrolled in this course", "status": "error"})
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "status": "error"})
	}
}

func parsePagination(c *gin.Context) (page, pageSize int, ok bool) {
	page, err := parsePositiveIntQuery(c.Query("page"), 1)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page", "status": "error"})
		return 0, 0, false
	}
	pageSize, err = parsePositiveIntQuery(c.Query("page_size"), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page_size", "status": "error"})
		return 0, 0, false
	}
	return page, pageSize, true
}

func parsePositiveIntQuery(raw string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	return n, nil
}
