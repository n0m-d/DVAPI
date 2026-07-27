package repository

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrCourseNotFound       = errors.New("course not found")
	ErrLessonNotFound       = errors.New("lesson not found")
	ErrEnrollmentExists     = errors.New("enrollment already exists")
	ErrEnrollmentMissing    = errors.New("enrollment not found")
	ErrAnnouncementNotFound = errors.New("announcement not found")
	ErrNotificationNotFound = errors.New("notification not found")
	ErrNoteNotFound         = errors.New("note not found")
)
