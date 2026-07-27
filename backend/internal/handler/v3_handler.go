package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/service"
	"github.com/n0m-d/DVAPI/internal/utils"
)

type V3Handler struct {
	courseService service.CourseService
}

func NewV3Handler(courseService service.CourseService) *V3Handler {
	return &V3Handler{courseService: courseService}
}

func (h *V3Handler) EnrollBeta(c *gin.Context) {
	h.changeEnrollmentBeta(c, true)
}

func (h *V3Handler) UnenrollBeta(c *gin.Context) {
	h.changeEnrollmentBeta(c, false)
}
func (h *V3Handler) changeEnrollmentBeta(c *gin.Context, enroll bool) {
	var req domain.EnrollmentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.MapError(err), "status": "error"})
		return
	}

	var err error
	if enroll {
		err = h.courseService.Enroll(c.Request.Context(), req.CourseID, req.UserID)
	} else {
		err = h.courseService.Unenroll(c.Request.Context(), req.CourseID, req.UserID)
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
