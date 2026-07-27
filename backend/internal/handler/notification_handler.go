package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/pubsub"
	"github.com/n0m-d/DVAPI/internal/service"
)

type NotificationHandler struct {
	notifications service.NotificationService
	notifier      pubsub.Notifier
}

func NewNotificationHandler(notifications service.NotificationService, notifier pubsub.Notifier) *NotificationHandler {
	return &NotificationHandler{notifications: notifications, notifier: notifier}
}

// List godoc
// @Summary      List notifications
// @Description  List notifications for the authenticated user with optional unread filter.
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Param        unread     query  bool  false  "Only unread notifications"
// @Param        page       query  int   false  "Page number"  default(1)
// @Param        page_size  query  int   false  "Page size"    default(10)
// @Success      200
// @Failure      400  "Invalid pagination"
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal server error"
// @Router       /notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	page, pageSize, ok := parsePagination(c)
	if !ok {
		return
	}
	unreadOnly := c.Query("unread") == "true"
	resp, err := h.notifications.List(c.Request.Context(), claims.UserID, unreadOnly, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list notifications", "status": "error"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// MarkRead godoc
// @Summary      Mark notification read
// @Description  Mark a single notification as read
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Notification ID"
// @Success      200
// @Failure      400  "Invalid notification id"
// @Failure      401  "Unauthorized"
// @Failure      404  "Notification not found"
// @Failure      500  "Internal server error"
// @Router       /notifications/{id}/read [post]
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification id", "status": "error"})
		return
	}
	n, err := h.notifications.MarkRead(c.Request.Context(), id, claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found", "status": "error"})
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification read", "status": "error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": n})
}

// MarkAllRead godoc
// @Summary      Mark all notifications read
// @Description  Mark all notifications as read
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Success      200
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal server error"
// @Router       /notifications/read-all [post]
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	if err := h.notifications.MarkAllRead(c.Request.Context(), claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications read", "status": "error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Delete godoc
// @Summary      Delete notification
// @Description  Delete a notification
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Notification ID"
// @Success      200
// @Failure      400  "Invalid notification id"
// @Failure      401  "Unauthorized"
// @Failure      404  "Notification not found"
// @Failure      500  "Internal server error"
// @Router       /notifications/{id} [delete]
func (h *NotificationHandler) Delete(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification id", "status": "error"})
		return
	}
	if err := h.notifications.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found", "status": "error"})
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification", "status": "error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *NotificationHandler) Stream(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}

	events, unsubscribe, err := h.notifier.Subscribe(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to subscribe to notifications", "status": "error"})
		return
	}
	defer unsubscribe()

	c.Writer.Header().Set("Content-Type", "text/event-stream") // Content-Type is required for SSE
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	c.SSEvent("connected", gin.H{"ok": true})
	c.Writer.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	clientGone := c.Request.Context().Done()
	for {
		select {
		case <-clientGone:
			return
		case <-ticker.C:
			c.SSEvent("ping", gin.H{"ts": time.Now().Unix()})
			c.Writer.Flush()
		case notification, open := <-events:
			if !open {
				return
			}
			c.SSEvent("notification", notification)
			c.Writer.Flush()
		}
	}
}
