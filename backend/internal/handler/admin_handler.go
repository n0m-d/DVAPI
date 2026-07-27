package handler

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	logFile string
}

func NewAdminHandler(logFile string) *AdminHandler {
	return &AdminHandler{logFile: logFile}
}

// ViewLogs godoc
// @Summary      Tail application logs (admin only)
// @Description  Returns the last lines of a log file under the server log directory.
// @Tags         Admin
// @Produce      json
// @Param        lines  query  int     false  "Number of lines" default(100)
// @Success      200  "Log output"
// @Failure      400  "Missing file"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      500  "Failed to read logs"
// @Security     BearerAuth
// @Router       /admin/logs [get]
func (h *AdminHandler) ViewLogs(c *gin.Context) {
	lines := c.DefaultQuery("lines", "100")

	if strings.TrimSpace(h.logFile) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Log file not configured", "status": "error"})
		return
	}

	// Command Injection: lines are interpolated into a shell command.
	cmd := exec.CommandContext(c.Request.Context(), "sh", "-c",
		fmt.Sprintf("tail -n %s %s", lines, h.logFile),
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to read logs",
			"status": "error",
			"detail": string(out),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"file":      h.logFile,
		"lines":     lines,
		"fetchedAt": time.Now().UTC(),
		"data":      string(out),
	})
}
