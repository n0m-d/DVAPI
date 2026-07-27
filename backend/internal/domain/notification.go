package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	NotificationTypeGrade        = "grade"
	NotificationTypeAssignment   = "assignment"
	NotificationTypeAnnouncement = "announcement"
	NotificationTypeSystem       = "system"
)

// Notification is an inbox event persisted in Postgres and optionally pushed over SSE.
type Notification struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Read      bool       `json:"read"`
	CreatedAt time.Time  `json:"created_at"`
	CourseID  *uuid.UUID `json:"course_id,omitempty"`
}

type CreateNotificationInput struct {
	UserID   uuid.UUID
	Type     string
	Title    string
	Body     string
	CourseID *uuid.UUID
}

type NotificationListData struct {
	Notifications []Notification `json:"notifications"`
	UnreadCount   int            `json:"unread_count"`
	Pagination    *Pagination    `json:"pagination"`
}

type NotificationListResponse struct {
	Status string                `json:"status"`
	Data   *NotificationListData `json:"data"`
}
